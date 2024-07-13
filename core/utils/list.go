package utils

type List[T comparable] struct {
	items []T
}

func NewList[T comparable](items ...T) *List[T] {
	return &List[T]{items: items}
}

func (l *List[T]) Len() int {
	return len(l.items)
}

func (l *List[T]) IsEmpty() bool {
	return l.Len() == 0
}

func (l *List[T]) IsNotEmpty() bool {
	return !l.IsEmpty()
}

func (l *List[T]) Add(item ...T) {
	l.items = append(l.items, item...)
}

func (l *List[T]) Remove(item T) {
	clone := make([]T, len(l.items))
	for _, it := range l.items {
		if it != item {
			clone = append(clone, item)
		}
	}
	l.items = clone
}

func (l *List[T]) Contains(item T) bool {
	for _, it := range l.items {
		if it == item {
			return true
		}
	}
	return false
}

func (l *List[T]) Any(test func(item T) bool) bool {
	for _, it := range l.items {
		if test(it) {
			return true
		}
	}
	return false
}

func (l *List[T]) All(test func(item T) bool) bool {
	for _, it := range l.items {
		if !test(it) {
			return false
		}
	}
	return true
}

func Fold[T comparable, U any](l *List[T], initial U, fold func(lastValue U, current T) U) U {
	result := initial
	for _, item := range l.items {
		result = fold(result, item)
	}
	return result
}

func Map[T comparable, U comparable](l List[T], mapper func(T) U) *List[U] {
	result := NewList[U]()
	for _, item := range l.items {
		result.Add(mapper(item))
	}
	return result
}
