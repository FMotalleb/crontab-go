package event

type Init struct {
	notifyChan chan any
}

// BuildTickChannel implements abstraction.Scheduler.
func (c *Init) BuildTickChannel() <-chan any {
	c.notifyChan = make(chan any)

	go func() {
		c.notifyChan <- false
		close(c.notifyChan)
	}()

	return c.notifyChan
}
