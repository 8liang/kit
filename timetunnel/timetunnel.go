package timetunnel

import (
    "github.com/golang-module/carbon/v2"
    "time"
)

type Tunnel interface {
    GetTouchedAt() time.Time
    SetTouchedAt(time.Time)
    OnHourPassed()
    OnDayPassed()
    OnWeekPassed()
    OnMonthPassed()
}
type Options func(t *tunnel)

type tunnel struct {
    Tunnel
    weekStartsAt string
}

func WithWeekStartsAt(weekStartsAt string) Options {
    return func(t *tunnel) {
        t.weekStartsAt = weekStartsAt
    }
}

func Pass(t Tunnel, opts ...Options) {
    tt := &tunnel{Tunnel: t, weekStartsAt: carbon.Monday}
    for _, opt := range opts {
        opt(tt)
    }
    last, current := touch(tt)
    passThrough(tt, last, current)
}

func passThrough(tt *tunnel, last, current carbon.Carbon) {
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
    last := c.GetTouchedAt().Local()
    c.SetTouchedAt(time.Now())
    return carbon.CreateFromStdTime(last).SetWeekStartsAt(c.weekStartsAt), carbon.Now().SetWeekStartsAt(c.weekStartsAt)
}
