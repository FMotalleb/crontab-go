// Package ctxutils provides utility functions for working with context.Context.
package ctxutils

type ContextKey string

var (
	ScopeKey       ContextKey = ContextKey("scope")
	LoggerKey      ContextKey = ContextKey("logger")
	RetryCountKey  ContextKey = ContextKey("retry-count")
	JobKey         ContextKey = ContextKey("job")
	TaskKey        ContextKey = ContextKey("task")
	FailedRemotes  ContextKey = ContextKey("failed-remotes")
	EventListeners ContextKey = ContextKey("event-listeners")
	EventData      ContextKey = ContextKey("event-data")
)
