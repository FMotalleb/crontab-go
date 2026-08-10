// Package generator holds logic behind the generic generators
package generator

import (
	"sort"

	"go.uber.org/zap"
)

type Func[I any, O any] = func(*zap.Logger, I) (O, bool)

type entry[I any, O any] struct {
	fn       Func[I, O]
	priority int
}

type Core[I any, O any] struct {
	entries []entry[I, O]
	dirty   bool
}

func New[I any, O any]() *Core[I, O] {
	return &Core[I, O]{entries: []entry[I, O]{}}
}

// Register adds a generator with default priority (0).
func (g *Core[I, O]) Register(generator Func[I, O]) {
	g.entries = append(g.entries, entry[I, O]{fn: generator, priority: 0})
	g.dirty = true
}

// RegisterWithPriority adds a generator with an explicit priority.
// Lower values are tried first.
func (g *Core[I, O]) RegisterWithPriority(generator Func[I, O], priority int) {
	g.entries = append(g.entries, entry[I, O]{fn: generator, priority: priority})
	g.dirty = true
}

func (g *Core[I, O]) sorted() {
	if g.dirty {
		sort.SliceStable(g.entries, func(i, j int) bool {
			return g.entries[i].priority < g.entries[j].priority
		})
		g.dirty = false
	}
}

func (g *Core[I, O]) Get(log *zap.Logger, input I) (O, bool) {
	g.sorted()
	for _, e := range g.entries {
		item, ok := e.fn(log, input)
		if ok {
			return item, true
		}
	}
	var empty O
	return empty, false
}
