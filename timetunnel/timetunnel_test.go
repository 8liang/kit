package timetunnel

import (
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
    tt := &stubTimeTunnel{touchedAt: last}
    Pass(tt, WithWeekStartsAt(carbon.Monday), WithCurrentTime(current))
    assert.True(t, tt.Week, "week should be crossed")
    assert.Equal(t, current.Unix(), tt.GetTouchedAt().Unix(), "touchedAt should be updated")

    tm := time.Now()
    stub := &stubTimeTunnel{touchedAt: tm.Add(-time.Hour)}
    Pass(stub, WithWeekStartsAt(carbon.Monday))
    assert.True(t, stub.Hour, "hour should be crossed")
    assert.False(t, stub.Day, "day should not be crossed")
    assert.False(t, stub.Week, "week should not be crossed")
    assert.False(t, stub.Month, "month should not be crossed")
    assert.Equal(t, time.Now().Unix(), stub.GetTouchedAt().Unix(), "touchedAt should be updated")

    stub = &stubTimeTunnel{touchedAt: tm.Add(-time.Hour * 24)}
    Pass(stub, WithWeekStartsAt(carbon.Monday))
    assert.True(t, stub.Hour, "hour should be crossed")
    assert.True(t, stub.Day, "day should be crossed")
    assert.False(t, stub.Week, "week should not be crossed")
    assert.False(t, stub.Month, "month should not be crossed")

    stub = &stubTimeTunnel{touchedAt: tm.Add(-time.Hour * 24 * 6)}
    Pass(stub, WithWeekStartsAt(carbon.Monday))
    assert.True(t, stub.Hour, "hour should be crossed")
    assert.True(t, stub.Week, "week should be crossed")
    assert.True(t, stub.Day, "day should be crossed")
}
