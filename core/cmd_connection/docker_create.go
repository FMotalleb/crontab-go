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
	"github.com/FMotalleb/crontab-go/ctxutils"
	"github.com/FMotalleb/crontab-go/helpers"
)

// DockerCreateConnection is a struct that manages the creation and execution of Docker containers.
type DockerCreateConnection struct {
	conn            *config.TaskConnection
	log             *logrus.Entry
	cli             *client.Client
	containerConfig *container.Config
	hostConfig      *container.HostConfig
	networkConfig   *network.NetworkingConfig
	ctx             context.Context
}

// NewDockerCreateConnection initializes a new DockerCreateConnection instance.
// Parameters:
// - log: A logrus.Entry instance for logging.
// - conn: A TaskConnection instance containing the connection configuration.
// Returns:
// - A new instance of DockerCreateConnection.
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

// Prepare sets up the Docker container configuration based on the provided task.
// Parameters:
// - ctx: A context.Context instance for managing the lifecycle of the container.
// - task: A Task instance containing the task configuration.
// Returns:
// - An error if the preparation fails, otherwise nil.
func (d *DockerCreateConnection) Prepare(ctx context.Context, task *config.Task) error {
	shell, shellArgs, env := reshapeEnviron(task.Env, d.log)
	d.ctx = ctx
	if d.conn.DockerConnection == "" {
		d.log.Debug("No explicit docker connection specified, using default: `unix:///var/run/docker.sock`")
		d.conn.DockerConnection = "unix:///var/run/docker.sock"
	}

	params := ctx.Value(ctxutils.EventData).([]string)
	cmd := append(
		[]string{shell},
		append(
			shellArgs,
			append(
				[]string{task.Command},
				params...,
			)...,
		)...,
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

// Connect establishes a connection to the Docker daemon.
// Returns:
// - An error if the connection fails, otherwise nil.
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

// Execute creates, starts, and logs the output of the Docker container.
// Returns:
// - A byte slice containing the command output.
// - An error if the execution fails, otherwise nil.
func (d *DockerCreateConnection) Execute() ([]byte, error) {
	ctx := d.ctx
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
	defer helpers.WarnOnErrIgnored(
		d.log,
		func() error {
			return d.cli.ContainerRemove(ctx, exec.ID,
				container.RemoveOptions{
					Force: true,
				},
			)
		},
		"cannot remove the container: %s",
	)

	for {
		err = d.cli.ContainerStart(
			ctx,
			exec.ID,
			container.StartOptions{},
		)

		if err == nil {
			break
		}
	}

	d.log.Tracef("container started: %v", exec)

	for {
		_, err := d.cli.ContainerStats(
			ctx,
			exec.ID,
			false,
		)

		if err == nil {
			break
		}
	}

	d.log.Debugf("container ready to attach: %v", exec)
	// Attach to the exec instance
	resp, err := d.cli.ContainerLogs(
		ctx,
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
	defer helpers.WarnOnErrIgnored(
		d.log,
		func() error {
			return resp.Close()
		},
		"cannot close the container's logs: %s",
	)

	writer := bytes.NewBuffer([]byte{})
	// Print the command output
	_, err = io.Copy(writer, resp)
	if err != nil {
		return writer.Bytes(), err
	}
	return writer.Bytes(), nil
}

// Disconnect closes the connection to the Docker daemon.
// Returns:
// - An error if the disconnection fails, otherwise nil.
func (d *DockerCreateConnection) Disconnect() error {
	return d.cli.Close()
}
