package connection

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
)

type DockerCreateConnection struct {
	conn            *config.TaskConnection
	log             *logrus.Entry
	cli             *client.Client
	imageName       string
	containerConfig *container.Config
	hostConfig      *container.HostConfig
	networkConfig   *network.NetworkingConfig
}

func NewDockerCreateConnection(log *logrus.Entry, conn *config.TaskConnection) abstraction.CmdConnection {
	return &DockerCreateConnection{
		conn: conn,
		log: log.WithFields(
			logrus.Fields{
				"connection":  "docker",
				"docker-mode": "create",
			},
		),
	}
}

func (d *DockerCreateConnection) Prepare(ctx context.Context, task *config.Task) error {
	shell, shellArgs, env := reshapeEnviron(task, d.log)

	if d.conn.DockerConnection == "" {
		d.log.Debug("No explicit docker connection specified, using default: `unix:///var/run/docker.sock`")
		d.conn.DockerConnection = "unix:///var/run/docker.sock"
	}
	cmd := append(
		[]string{shell},
		append(shellArgs, task.Command)...,
	)
	volumes := make(map[string]struct{})
	for _, volume := range d.conn.Volumes {
		inContainer := strings.Split(volume, ":")[1]
		volumes[inContainer] = struct{}{}
	}
	// Create an exec configuration
	d.containerConfig = &container.Config{
		AttachStdout: true,
		AttachStderr: true,
		Env:          env,
		WorkingDir:   task.WorkingDirectory,
		User:         task.UserName,
		Cmd:          cmd,
		Image:        d.conn.ImageName,
		Volumes:      volumes,
		Entrypoint:   []string{},
		Shell:        []string{},
	}
	d.hostConfig = &container.HostConfig{
		Binds: d.conn.Volumes,
		// AutoRemove: true,
	}
	endpointsConfig := make(map[string]*network.EndpointSettings)
	for _, networkName := range d.conn.Networks {
		endpointsConfig[networkName] = &network.EndpointSettings{}
	}
	d.networkConfig = &network.NetworkingConfig{
		EndpointsConfig: endpointsConfig,
	}
	return nil
}

func (d *DockerCreateConnection) Connect() error {
	cli, err := client.NewClientWithOpts(
		client.WithHost(d.conn.DockerConnection),
	)
	if err != nil {
		return err
	}
	d.cli = cli
	return nil
}

func (d *DockerCreateConnection) Execute() ([]byte, error) {
	ctx := context.Background()
	// Create the exec instance
	exec, err := d.cli.ContainerCreate(
		ctx,
		d.containerConfig,
		d.hostConfig,
		d.networkConfig,
		nil,
		d.conn.ContainerName,
	)
	d.log.Debugf("container created: %v, warnings: %v", exec, exec.Warnings)
	if err != nil {
		return nil, err
	}
	defer d.cli.ContainerRemove(ctx, exec.ID,
		container.RemoveOptions{
			Force: true,
		},
	)
	err = d.cli.ContainerStart(d.log.Context, exec.ID,
		container.StartOptions{},
	)

	d.log.Debugf("container started: %v", exec)
	if err != nil {
		return nil, err
	}
	starting := true
	for starting {
		_, err := d.cli.ContainerStats(ctx, exec.ID, false)
		if err == nil {
			starting = false
		}
	}
	d.log.Debugf("container ready to attach: %v", exec)
	// Attach to the exec instance
	resp, err := d.cli.ContainerLogs(
		context.Background(),
		exec.ID,
		container.LogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     false,
			Details:    true,
		},
	)
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	writer := bytes.NewBuffer([]byte{})
	// Print the command output
	_, err = io.Copy(writer, resp)
	if err != nil {
		return writer.Bytes(), err
	}
	return writer.Bytes(), nil
}

func (d *DockerCreateConnection) Disconnect() error {
	return d.cli.Close()
}
