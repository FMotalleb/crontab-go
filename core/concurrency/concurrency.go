package concurrency

// ConcurrentPool implements a simple semaphore-like structure to limit
// the number of concurrent goroutines working together.
type ConcurrentPool struct {
	available  int
	used       int
	changeChan chan interface{}
}

// NewConcurrentPool creates a new ConcurrentPool with the specified capacity.
func NewConcurrentPool(capacity int) *ConcurrentPool {
	return &ConcurrentPool{
		available:  capacity,
		used:       0,
		changeChan: make(chan interface{}),
	}
}

// Lock acquires a lock, waiting if necessary until a slot becomes available.
func (p *ConcurrentPool) Lock() {
	defer p.reserverSlot()
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
func (p *ConcurrentPool) Unlock() {
	p.freeSlot()
}

func (p *ConcurrentPool) reserverSlot() {
	p.used++
	p.signal()
}

func (p *ConcurrentPool) freeSlot() {
	p.used--
	p.signal()
}

func (p *ConcurrentPool) signal() {
	p.changeChan <- false
}
