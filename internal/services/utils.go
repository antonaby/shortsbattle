package services

import "time"

type Countdown struct {
	timer    *time.Timer
	started  time.Time
	duration time.Duration
}

func NewCountdown(d time.Duration) Countdown {
	return Countdown{
		timer:    time.NewTimer(d),
		started:  time.Now(),
		duration: d,
	}
}

func (c *Countdown) Remaining() time.Duration {
	if c.timer == nil {
		return 0
	}

	r := c.duration - time.Since(c.started)
	if r < 0 {
		return 0
	}

	return r
}

func (c *Countdown) Reset(d time.Duration) {
	if !c.timer.Stop() {
		select {
		case <-c.timer.C:
		default:
		}
	}
	c.timer.Reset(d)
	c.started = time.Now()
	c.duration = d
}

func (c *Countdown) Stop() {
	if c.timer != nil {
		c.timer.Stop()
	}
	c.timer = nil
}
