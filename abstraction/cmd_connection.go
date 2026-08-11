package abstraction

import (
	"context"
	"io"

	"github.com/fmotalleb/crontab-go/config"
)

type CmdConnection interface {
	Prepare(context.Context, *config.Task) error
	Connect() error
	Execute(stdout, stderr io.Writer) error
	Disconnect() error
}
