package timepass

import (
	"time"

	"github.com/golang-module/carbon/v2"
)

type Handler interface {
	LastTouchedAt() time.Time
	SetTouchedAt(time.Time)
	OnMinute()
	OnHour()
	OnDay()
	OnWeek()
	OnMonth()
}

type EmptyTunnel struct{}

func (e *EmptyTunnel) LastTouchedAt() time.Time {
	return time.Now()
}

func (e *EmptyTunnel) SetTouchedAt(t time.Time) {
}

func (e *EmptyTunnel) OnMinute() {
}

func (e *EmptyTunnel) OnHour() {
}

func (e *EmptyTunnel) OnDay() {
}

func (e *EmptyTunnel) OnWeek() {
}

func (e *EmptyTunnel) OnMonth() {
}

type Option func(t *handler)

type handler struct {
	Handler
	weekStartsAt string
	current      time.Time
}

// WithWeekStartsAt 设置周起始日选项
// WithWeekStartsAt sets the week start day option
func WithWeekStartsAt(weekStartsAt string) Option {
	return func(t *handler) {
		t.weekStartsAt = weekStartsAt
	}
}

// WithCurrentTime 设置当前时间选项
// WithCurrentTime sets the current time option
func WithCurrentTime(currentTime time.Time) Option {
	return func(t *handler) {
		t.current = currentTime
	}
}

// Advance 执行时间隧道检查
// Advance performs time tunnel check
func Advance(t Handler, opts ...Option) {
	tt := &handler{Handler: t, weekStartsAt: carbon.Monday}
	for _, opt := range opts {
		opt(tt)
	}
	last, current := touch(tt)
	passThrough(tt, last, current)
}

func passThrough(tt Handler, last, current carbon.Carbon) {
	if last.IsSameMinute(current) {
		return
	}
	tt.OnMinute()
	if last.IsSameHour(current) {
		return
	}
	tt.OnHour()

	if last.IsSameDay(current) {
		return
	}
	tt.OnDay()

	if !last.StartOfWeek().IsSameDay(current.StartOfWeek()) {
		tt.OnWeek()
	}

	if !last.StartOfMonth().IsSameDay(current.StartOfMonth()) {
		tt.OnMonth()
	}
}

func touch(c *handler) (carbon.Carbon, carbon.Carbon) {
	if c.current.IsZero() {
		c.current = time.Now()
	}
	last := c.LastTouchedAt()
	c.SetTouchedAt(c.current)

	return carbon.CreateFromStdTime(last.Local()).SetWeekStartsAt(c.weekStartsAt),
		carbon.CreateFromStdTime(c.current.Local()).SetWeekStartsAt(c.weekStartsAt)
}
