package connection

import (
	"bytes"
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/config"
)

type DockerConnection struct {
	conn        *config.TaskConnection
	log         *logrus.Entry
	cli         *client.Client
	execCFG     *types.ExecConfig
	containerID string
}

func NewDockerConnection(log *logrus.Entry, conn *config.TaskConnection) *DockerConnection {
	return &DockerConnection{
		conn: conn,
		log: log.WithField(
			"connection", "docker",
		),
	}
}

func (d *DockerConnection) Prepare(ctx context.Context, task *config.Task) error {
	shell, shellArgs, env := reshapeEnviron(task, d.log)
	// Specify the container ID or name
	d.containerID = d.conn.ContainerName

	if d.conn.DockerConnection == "" {
		d.log.Debug("No explicit docker connection specified, using default: `unix:///var/run/docker.sock`")
		d.conn.DockerConnection = "unix:///var/run/docker.sock"
	}
	cmd := append(
		[]string{shell},
		append(shellArgs, task.Command)...,
	)
	// Create an exec configuration
	d.execCFG = &types.ExecConfig{
		AttachStdout: true,
		AttachStderr: true,
		Privileged:   true,
		Env:          env,
		WorkingDir:   task.WorkingDirectory,
		User:         task.UserName,
		Cmd:          cmd,
	}

	return nil
}

func (d *DockerConnection) Connect() error {
	cli, err := client.NewClientWithOpts(
		client.WithHost(d.conn.DockerConnection),
	)
	if err != nil {
		return err
	}
	d.cli = cli
	return nil
}

func (d *DockerConnection) Execute() ([]byte, error) {
	// Create the exec instance
	exec, err := d.cli.ContainerExecCreate(context.Background(), d.containerID, *d.execCFG)
	if err != nil {
		return nil, err
	}

	// Attach to the exec instance
	resp, err := d.cli.ContainerExecAttach(context.Background(), exec.ID, types.ExecStartCheck{})
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	writer := bytes.NewBuffer([]byte{})
	// Print the command output
	_, err = io.Copy(writer, resp.Reader)
	if err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

func (d *DockerConnection) Disconnect() error {
	return d.cli.Close()
}
