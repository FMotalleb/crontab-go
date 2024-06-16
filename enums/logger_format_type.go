package enums

import "fmt"

type LoggerFormatType string

const (
	JSONLogger  = LoggerFormatType("json")
	AnsiLogger  = LoggerFormatType("ansi")
	PlainLogger = LoggerFormatType("plain")
)

func (lf LoggerFormatType) Validate() error {
	switch lf {
	case JSONLogger, AnsiLogger, PlainLogger:
		return nil
	default:
		return fmt.Errorf("given logger format: `%s`", lf)
	}
}
