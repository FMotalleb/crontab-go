package ctxutils

type ContextKey string

var (
	scope      ContextKey = ContextKey("scope")
	logger     ContextKey = ContextKey("logger")
	RetryCount ContextKey = ContextKey("retry-count")
)
