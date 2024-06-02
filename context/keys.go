package context

type ContextKey string

var (
	scope  ContextKey = ContextKey("scope")
	logger ContextKey = ContextKey("logger")
)
