package event

type Init struct{}

// BuildTickChannel implements abstraction.Scheduler.
func (c *Init) BuildTickChannel() <-chan []string {
	notifyChan := make(chan []string)

	go func() {
		notifyChan <- []string{"init"}
		close(notifyChan)
	}()

	return notifyChan
}
