package timetunnel

import (
    "github.com/davecgh/go-spew/spew"
    "github.com/golang-module/carbon/v2"
    "github.com/stretchr/testify/assert"
    "testing"
    "time"
)

type stubTimeTunnel struct {
    touchedAt time.Time
    Hour      bool
    Day       bool
    Week      bool
    Month     bool
}

func (s *stubTimeTunnel) GetTouchedAt() time.Time {
    return s.touchedAt
}

func (s *stubTimeTunnel) SetTouchedAt(t time.Time) {
    s.touchedAt = t
}

func (s *stubTimeTunnel) OnHourPassed() {
    s.Hour = true
}

func (s *stubTimeTunnel) OnDayPassed() {
    s.Day = true
}

func (s *stubTimeTunnel) OnWeekPassed() {
    s.Week = true
}

func (s *stubTimeTunnel) OnMonthPassed() {
    s.Month = true
}

func TestTunnel(t *testing.T) {
    last, _ := time.Parse(time.RFC3339, "2024-03-03T15:02:03.332Z")
    current, _ := time.Parse(time.RFC3339, "2024-03-03T16:02:03.332Z")

    //last = last.Local()
    //current = current.Local()
    tt := &stubTimeTunnel{touchedAt: last}

    Pass(tt, WithWeekStartsAt(carbon.Monday), WithCurrentTime(current))
    assert.True(t, tt.Week, "week should be crossed")

    spew.Dump(last)
}
