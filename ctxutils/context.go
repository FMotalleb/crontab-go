package ctxutils

import "context"

type Context struct {
	ctx context.Context
}

func NewContext(section string) Context {
	return Context{
		context.WithValue(
			context.Background(),
			ScopeKey,
			section,
		),
	}
}

func (ctx *Context) getLogger() {
}
