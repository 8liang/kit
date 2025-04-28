package timetunnel

import (
	"time"

	"github.com/golang-module/carbon/v2"
)

type Tunnel interface {
	GetTouchedAt() time.Time
	SetTouchedAt(time.Time)
	OnMinutePassed()
	OnHourPassed()
	OnDayPassed()
	OnWeekPassed()
	OnMonthPassed()
}

type EmptyTunnel struct{}

func (e *EmptyTunnel) GetTouchedAt() time.Time {
	return time.Now()
}

func (e *EmptyTunnel) SetTouchedAt(t time.Time) {
}

func (e *EmptyTunnel) OnMinutePassed() {
}

func (e *EmptyTunnel) OnHourPassed() {
}

func (e *EmptyTunnel) OnDayPassed() {
}

func (e *EmptyTunnel) OnWeekPassed() {
}

func (e *EmptyTunnel) OnMonthPassed() {
}

type Options func(t *tunnel)

type tunnel struct {
	Tunnel
	weekStartsAt string
	current      time.Time
}

// WithWeekStartsAt 设置周起始日选项
// WithWeekStartsAt sets the week start day option
func WithWeekStartsAt(weekStartsAt string) Options {
	return func(t *tunnel) {
		t.weekStartsAt = weekStartsAt
	}
}

// WithCurrentTime 设置当前时间选项
// WithCurrentTime sets the current time option
func WithCurrentTime(currentTime time.Time) Options {
	return func(t *tunnel) {
		t.current = currentTime
	}
}

// Pass 执行时间隧道检查
// Pass performs time tunnel check
func Pass(t Tunnel, opts ...Options) {
	tt := &tunnel{Tunnel: t, weekStartsAt: carbon.Monday}
	for _, opt := range opts {
		opt(tt)
	}
	last, current := touch(tt)
	passThrough(tt, last, current)
}

func passThrough(tt Tunnel, last, current carbon.Carbon) {
	if last.IsSameMinute(current) {
		return
	}
	tt.OnMinutePassed()
	if last.IsSameHour(current) {
		return
	}
	tt.OnHourPassed()

	if last.IsSameDay(current) {
		return
	}
	tt.OnDayPassed()

	if !last.StartOfWeek().IsSameDay(current.StartOfWeek()) {
		tt.OnWeekPassed()
	}

	if !last.StartOfMonth().IsSameDay(current.StartOfMonth()) {
		tt.OnMonthPassed()
	}
}

func touch(c *tunnel) (carbon.Carbon, carbon.Carbon) {
	if c.current.IsZero() {
		c.current = time.Now()
	}
	last := c.GetTouchedAt()
	c.SetTouchedAt(c.current)

	return carbon.CreateFromStdTime(last.Local()).SetWeekStartsAt(c.weekStartsAt),
		carbon.CreateFromStdTime(c.current.Local()).SetWeekStartsAt(c.weekStartsAt)
}
