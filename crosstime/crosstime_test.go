package crosstime

import (
    "github.com/stretchr/testify/assert"
    "testing"
    "time"
)

type stubTimeCrosser struct {
    touchedAt time.Time
    Hour      bool
    Day       bool
    Week      bool
    Month     bool
}

func (s *stubTimeCrosser) GetTouchedAt() time.Time {
    return s.touchedAt
}

func (s *stubTimeCrosser) SetTouchedAt(t time.Time) {
    s.touchedAt = t
}

func (s *stubTimeCrosser) CrossHour() {
    s.Hour = true
}

func (s *stubTimeCrosser) CrossDay() {
    s.Day = true
}

func (s *stubTimeCrosser) CrossWeek() {
    s.Week = true
}

func (s *stubTimeCrosser) CrossMonth() {
    s.Month = true
}

func TestCross(t *testing.T) {
    tm := time.Now()

    stub := &stubTimeCrosser{touchedAt: tm.Add(-time.Hour)}
    CrossWithWeekStartsAt(stub, "Monday")
    assert.True(t, stub.Hour, "hour should be crossed")
    assert.False(t, stub.Day, "day should not be crossed")
    assert.False(t, stub.Week, "week should not be crossed")
    assert.False(t, stub.Month, "month should not be crossed")
    assert.Equal(t, time.Now().Unix(), stub.GetTouchedAt().Unix(), "touchedAt should be updated")

    stub = &stubTimeCrosser{touchedAt: tm.Add(-time.Hour * 24)}
    CrossWithWeekStartsAt(stub, "Monday")
    assert.True(t, stub.Hour, "hour should be crossed")
    assert.True(t, stub.Day, "day should be crossed")
    assert.False(t, stub.Week, "week should not be crossed")
    assert.False(t, stub.Month, "month should not be crossed")

    stub = &stubTimeCrosser{touchedAt: tm.Add(-time.Hour * 24 * 6)}
    CrossWithWeekStartsAt(stub, "Monday")
    assert.True(t, stub.Hour, "hour should be crossed")
    assert.True(t, stub.Week, "week should be crossed")
    assert.True(t, stub.Day, "day should be crossed")
}
