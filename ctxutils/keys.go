package ctxutils

type ContextKey string

var (
	ScopeKey      ContextKey = ContextKey("scope")
	LoggerKey     ContextKey = ContextKey("logger")
	RetryCountKey ContextKey = ContextKey("retry-count")
	JobKey        ContextKey = ContextKey("job")
)