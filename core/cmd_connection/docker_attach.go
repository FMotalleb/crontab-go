package connection

import (
	"bytes"
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"

	"github.com/FMotalleb/crontab-go/abstraction"
	"github.com/FMotalleb/crontab-go/config"
)

type DockerAttachConnection struct {
	conn        *config.TaskConnection
	log         *logrus.Entry
	cli         *client.Client
	execCFG     *types.ExecConfig
	containerID string
	ctx         context.Context
}

// NewDockerAttachConnection creates a new DockerAttachConnection instance.
// It initializes the connection configuration and logging fields.
// Parameters:
// - log: A logrus.Entry instance for logging purposes.
// - conn: A TaskConnection instance containing the connection configuration.
// Returns:
// - A new instance of DockerAttachConnection implementing the CmdConnection interface.
func NewDockerAttachConnection(log *logrus.Entry, conn *config.TaskConnection) abstraction.CmdConnection {
	return &DockerAttachConnection{
		conn: conn,
		log: log.WithFields(
			logrus.Fields{
				"connection":  "docker",
				"docker-mode": "attach",
			},
		),
	}
}

// Prepare sets up the DockerAttachConnection for executing a task.
// It reshapes the environment variables, sets the context, and creates an exec configuration.
// Parameters:
// - ctx: A context.Context instance for managing the request lifetime.
// - task: A Task instance containing the task configuration.
// Returns:
// - An error if the preparation fails, otherwise nil.
func (d *DockerAttachConnection) Prepare(ctx context.Context, task *config.Task) error {
	shell, shellArgs, env := reshapeEnviron(task, d.log)
	d.ctx = ctx
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

// Connect establishes a connection to the Docker daemon.
// It initializes the Docker client with the specified connection settings.
// Returns:
// - An error if the connection fails, otherwise nil.
func (d *DockerAttachConnection) Connect() error {
	cli, err := client.NewClientWithOpts(
		client.WithHost(d.conn.DockerConnection),
	)
	if err != nil {
		return err
	}
	d.cli = cli
	return nil
}

// Execute runs the command in the Docker container and captures the output.
// It creates an exec instance, attaches to it, and reads the command output.
// Returns:
// - A byte slice containing the command output.
// - An error if the execution fails, otherwise nil.
func (d *DockerAttachConnection) Execute() ([]byte, error) {
	// Create the exec instance
	exec, err := d.cli.ContainerExecCreate(d.ctx, d.containerID, *d.execCFG)
	if err != nil {
		return nil, err
	}

	// Attach to the exec instance
	resp, err := d.cli.ContainerExecAttach(d.ctx, exec.ID, types.ExecStartCheck{})
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

// Disconnect closes the connection to the Docker daemon.
// Returns:
// - An error if the disconnection fails, otherwise nil.
func (d *DockerAttachConnection) Disconnect() error {
	return d.cli.Close()
}
