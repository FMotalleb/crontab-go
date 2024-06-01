package context

type ContextKey = string

var (
	scope  ContextKey = "scope"
	logger ContextKey = "logger"
)
