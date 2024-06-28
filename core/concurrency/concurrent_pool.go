// Package concurrency provides utility functions for working with goroutines.
package concurrency

import (
	"errors"
	"sync"
)

// ConcurrentPool implements a simple semaphore-like structure to limit
// the number of concurrent goroutines working together.
type ConcurrentPool struct {
	internalSync sync.Locker
	available    uint             // total capacity of the pool
	used         uint             // number of slots currently in use
	changeChan   chan interface{} // channel for signaling changes in the pool's state
}

// NewConcurrentPool creates a new ConcurrentPool with the specified capacity.
// It panics if the capacity is 0.
func NewConcurrentPool(capacity uint) (*ConcurrentPool, error) {
	if capacity == 0 {
		return nil, errors.New("capacity value of a concurrent poll cannot be 0")
	}
	return &ConcurrentPool{
		internalSync: &sync.Mutex{},
		available:    capacity,
		used:         0,
		changeChan:   make(chan interface{}),
	}, nil
}

// Lock acquires a lock from the pool, waiting if necessary until a slot becomes available.
// It increments the used count using the reserveSlot method.
func (p *ConcurrentPool) Lock() {
	defer p.increase()
	if p.available > p.get() {
		return
	}
	for range p.changeChan {
		if p.available > p.get() {
			break
		}
	}
}

// Unlock releases a lock, making a slot available for other goroutines.
// It decrements the used count and sends a signal on the changeChan to notify waiting goroutines.
func (p *ConcurrentPool) Unlock() {
	if p.get() == 0 {
		panic(errors.New("unlock called on a totally free pool"))
	}
	p.decrease()
	go func() { p.changeChan <- false }()
}

func (p *ConcurrentPool) get() uint {
	return p.access(get)
}

func (p *ConcurrentPool) increase() {
	p.access(increase)
}

func (p *ConcurrentPool) decrease() {
	p.access(decrease)
}

// access is the only way to access the internal state of the pool's `used` count.
// inorder to maintain the integrity of the pool, it is protected by the internalSync mutex.
// every operation (get,increase,decrease) is encapsulated in a function that takes the pool as argument
func (p *ConcurrentPool) access(action func(p *ConcurrentPool) uint) uint {
	p.internalSync.Lock()
	defer p.internalSync.Unlock()
	return action(p)
}

func get(p *ConcurrentPool) uint {
	return p.used
}

func decrease(p *ConcurrentPool) uint {
	p.used--
	return p.used
}

func increase(p *ConcurrentPool) uint {
	p.used++
	return p.used
}
