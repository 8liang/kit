package crosstime

import (
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

func Cross(c TimeCrosser) {
    last, current := touch(c)
    if last.IsSameHour(current) {
        return
    }
    c.CrossHour()
    if last.IsSameDay(current) {
        return
    }
    c.CrossDay()

    if !last.StartOfWeek().IsSameDay(current.StartOfWeek()) {
        c.CrossWeek()
    }
    if !last.StartOfMonth().IsSameDay(current.StartOfMonth()) {
        c.CrossMonth()
    }
}

func touch(c TimeCrosser) (carbon.Carbon, carbon.Carbon) {
    last := c.GetTouchedAt()
    c.SetTouchedAt(time.Now())
    return carbon.CreateFromStdTime(last).SetWeekStartsAt(carbon.Monday), carbon.Now().SetWeekStartsAt(carbon.Monday)
}
