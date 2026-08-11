package task

import (
	"go.opentelemetry.io/otel"
)

var taskTracer = otel.Tracer("crontab-go/task")
