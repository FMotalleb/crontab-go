package connection

import (
	"context"

	"github.com/FMotalleb/crontab-go/config"
)

type DockerConnection struct{
	address string
	containerMatcher  string
}

func (d *DockerConnection) Prepare(context.Context, *config.Task) error{
	return nil
}
func (d *DockerConnection) Connect() error {
	return nil
}
func (d *DockerConnection) Execute() ([]byte, error) {
	
	return []byte{},nil
}
func (d *DockerConnection) Disconnect() error {
	return nil
}
