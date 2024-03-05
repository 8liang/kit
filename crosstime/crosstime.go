package crosstime

import (
    "github.com/8liang/kit/timetunnel"
    "github.com/golang-module/carbon/v2"
    "time"
)

type TimeCrosser interface {
    GetTouchedAt() time.Time
    SetTouchedAt(time.Time)
    CrossHour()
    CrossDay()
    CrossWeek()
    CrossMonth()
}

type tunnel struct {
    TimeCrosser
}

func (t *tunnel) OnHourPassed() {
    t.CrossHour()
}

func (t *tunnel) OnDayPassed() {
    t.CrossDay()
}

func (t *tunnel) OnWeekPassed() {
    t.CrossWeek()
}

func (t *tunnel) OnMonthPassed() {
    t.CrossMonth()
}

func Cross(c TimeCrosser) {
    CrossWithWeekStartsAt(c, carbon.Monday)
}
func CrossWithWeekStartsAt(c TimeCrosser, weekStartsAt string) {
    tt := &tunnel{TimeCrosser: c}
    timetunnel.Pass(tt, timetunnel.WithWeekStartsAt(weekStartsAt))
}
