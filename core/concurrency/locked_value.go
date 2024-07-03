package concurrency

import (
	"sync"
)

type (
	Operator[T any]    func(T) T
	LockedValue[T any] struct {
		value T
		lock  sync.Locker
	}
)

func NewLockedValue[T any](initial T) *LockedValue[T] {
	return &LockedValue[T]{
		value: initial,
		lock:  &sync.Mutex{},
	}
}

func (lv *LockedValue[T]) Get() T {
	lv.lock.Lock()
	defer lv.lock.Unlock()
	return lv.value
}

func (lv *LockedValue[T]) Set(newValue T) {
	lv.lock.Lock()
	defer lv.lock.Unlock()
	lv.value = newValue
}

func (lv *LockedValue[T]) Operate(operator Operator[T]) {
	lv.lock.Lock()
	defer lv.lock.Unlock()
	lv.value = operator(lv.value)
}
