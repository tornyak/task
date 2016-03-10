package counter
import "sync/atomic"

type Counter struct {
	value uint64
}

func (c *Counter) AddAndGet(delta uint64) uint64 {
	return atomic.AddUint64(&c.value, delta)
}

func (c *Counter) Inc() {
	atomic.AddUint64(&c.value, 1)
}

func (c *Counter) IncAndGet() uint64 {
	return atomic.AddUint64(&c.value, 1)
}

func (c *Counter) GetAndInc() uint64 {
	return atomic.AddUint64(&c.value, 1) - 1
}

func (c *Counter) Get() uint64 {
	return atomic.LoadUint64(&c.value)
}
