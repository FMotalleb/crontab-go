package concurrency

import (
	"errors"
)

// ConcurrentPool implements a simple semaphore-like structure to limit
// the number of concurrent goroutines working together.
type ConcurrentPool struct {
	available  uint             // total capacity of the pool
	used       uint             // number of slots currently in use
	changeChan chan interface{} // channel for signaling changes in the pool's state
}

// NewConcurrentPool creates a new ConcurrentPool with the specified capacity.
// It panics if the capacity is 0.
func NewConcurrentPool(capacity uint) *ConcurrentPool {
	if capacity == 0 {
		panic(errors.New("capacity value of a concurrent poll cannot be 0"))
	}
	return &ConcurrentPool{
		available:  capacity,
		used:       0,
		changeChan: make(chan interface{}),
	}
}

// Lock acquires a lock from the pool, waiting if necessary until a slot becomes available.
// It increments the used count using the reserveSlot method.
func (p *ConcurrentPool) Lock() {
	defer p.reserveSlot()
	if p.available > p.used {
		return
	}
	for range p.changeChan {
		if p.available > p.used {
			break
		}
	}
}

// Unlock releases a lock, making a slot available for other goroutines.
// It decrements the used count and sends a signal on the changeChan to notify waiting goroutines.
func (p *ConcurrentPool) Unlock() {
	if p.used == 0 {
		panic(errors.New("unlock called on a totally free pool"))
	}
	p.used--
	p.changeChan <- false
}

// reserveSlot is a helper method that increments the used count.
// It is called using defer in the Lock method to ensure that the used count is incremented
// even if the Lock method panics or returns early.
func (p *ConcurrentPool) reserveSlot() {
	p.used++
}
