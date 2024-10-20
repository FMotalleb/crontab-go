package global

import (
	"context"

	"github.com/robfig/cron/v3"
)

func (c *GlobalContext) register(key Key, value any) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.Context = context.WithValue(c.Context, key, value)
}

func (c *GlobalContext) retrieve(key Key) (value any, ok bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	value = c.Value(key)
	ok = value != nil
	return value, ok
}

// func (c *GlobalContext) modify(key Key, modifier func(any)) bool {
// 	c.lock.Lock()
// 	defer c.lock.Unlock()
// 	value := c.Value(key)
// 	modifier(value)
// 	c.Context = context.WithValue(c.Context, key, value)
// 	return true
// }

func (c *GlobalContext) GetCron() *cron.Cron {
	result, ok := c.retrieve(CronKey)
	if !ok {
		result = cron.New()
		c.register(CronKey, result)
		return result.(*cron.Cron)
	}
	return result.(*cron.Cron)
}
