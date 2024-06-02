package enums

import "fmt"

type LoggerFormatType string

var (
	DefaultLogger = LoggerFormatType("")
	JsonLogger    = LoggerFormatType("json")
	AnsiLogger    = LoggerFormatType("ansi")
	PlainLogger   = LoggerFormatType("plain")
)

func (lf LoggerFormatType) Validate() error {
	switch lf {
	case JsonLogger, AnsiLogger, PlainLogger, DefaultLogger:
		return nil
	default:
		return fmt.Errorf("Given Logger type: `%s`", lf)
	}
}
