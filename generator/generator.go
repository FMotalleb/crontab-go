// Package generator holds logic behind the generic generators
package generator

import (
	"github.com/sirupsen/logrus"
)

type Func[I any, O any] = func(*logrus.Entry, I) (O, bool)

type Core[I any, O any] struct {
	generators []Func[I, O]
}

func New[I any, O any]() *Core[I, O] {
	return &Core[I, O]{generators: []Func[I, O]{}}
}

func (g *Core[I, O]) Register(generator Func[I, O]) {
	g.generators = append(g.generators, generator)
}

func (g *Core[I, O]) Get(log *logrus.Entry, input I) (O, bool) {
	for _, generator := range g.generators {
		item, ok := generator(log, input)
		if ok {
			return item, true
		}
	}
	var empty O
	return empty, false
}
