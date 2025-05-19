// Package generator holds logic behind the generic generators
package generator

import (
	"reflect"

	"github.com/sirupsen/logrus"
)

type Func[I any, O any] = func(*logrus.Entry, I) O

type Core[I any, O any] struct {
	generators []Func[I, O]
}

func New[I any, O any]() *Core[I, O] {
	generators := make([]Func[I, O], 0)
	return &Core[I, O]{generators: generators}
}

func (g *Core[I, O]) Register(generator Func[I, O]) {
	g.generators = append(g.generators, generator)
}

func (g *Core[I, O]) Get(log *logrus.Entry, input I) O {
	for _, generator := range g.generators {
		item := generator(log, input)
		if !isNilInterface(item) {
			return item
		}
	}
	var zero O
	return zero
}

func isNilInterface(val any) bool {
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}
