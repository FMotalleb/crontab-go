package event

type Init struct {
	notifyChan chan any
}

// BuildTickChannel implements abstraction.Scheduler.
func (c *Init) BuildTickChannel() <-chan any {
	c.notifyChan = make(chan any)

	go func() {
		c.notifyChan <- false
		c.Cancel()
	}()

	return c.notifyChan
}

// Cancel implements abstraction.Scheduler.
func (c *Init) Cancel() {
	close(c.notifyChan)
}