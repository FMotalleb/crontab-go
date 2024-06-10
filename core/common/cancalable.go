package common

type Cancelable struct {
	cancel func()
}

func (c *Cancelable) SetCancel(cancel func()) {
	c.cancel = cancel
}

func (c *Cancelable) Cancel() {
	if c.cancel != nil {
		c.cancel()
	}
}
