package crosstime

import (
    "testing"
    "time"
)

type stubTimeCrosser struct {
    touchedAt time.Time
}

func (s *stubTimeCrosser) GetTouchedAt() time.Time {
    return s.touchedAt
}

func (s *stubTimeCrosser) SetTouchedAt(t time.Time) {
    s.touchedAt = t
}

func (s *stubTimeCrosser) CrossHour() {
    //TODO implement me
    //panic("implement me")
}

func (s *stubTimeCrosser) CrossDay() {
    //TODO implement me
    //panic("implement me")
}

func (s *stubTimeCrosser) CrossWeek() {
    panic("cross")
}

func (s *stubTimeCrosser) CrossMonth() {
    //TODO implement me
    //panic("implement me")
}

func TestCross(t *testing.T) {
    last, _ := time.Parse(time.RFC3339, "2024-03-03T16:55:17.596Z")
    stub := &stubTimeCrosser{touchedAt: last}

    CrossWithWeekStartsAt(stub, "Monday")
}
