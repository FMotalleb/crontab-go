package event

type Init struct{}

// BuildTickChannel implements abstraction.Scheduler.
func (c *Init) BuildTickChannel() <-chan any {
	notifyChan := make(chan any)

	go func() {
		notifyChan <- false
		close(notifyChan)
	}()

	return notifyChan
}
