// Package context implements basic functionality of context used in the application
package context

import "context"

type Context struct {
	ctx context.Context
}

func NewContext(section string) Context {
	return Context{
		context.WithValue(
			context.Background(),
			scope,
			section,
		),
	}
}

func (ctx *Context) getLogger() {
}
