package crosstime

import (
	"fmt"
	"time"

	"github.com/8liang/kit/timepass"
	"github.com/golang-module/carbon/v2"
)

type TimeCrosser interface {
	LastTouchedAt() time.Time
	SetTouchedAt(time.Time)
	CrossHour()
	CrossDay()
	CrossWeek()
	CrossMonth()
}

type tunnel struct {
	TimeCrosser
}

func (t *tunnel) OnMinute() {
	fmt.Printf("crosstime tunnel OnMinutePassed\n")
}

func (t *tunnel) OnHour() {
	t.CrossHour()
}

func (t *tunnel) OnDay() {
	t.CrossDay()
}

func (t *tunnel) OnWeek() {
	t.CrossWeek()
}

func (t *tunnel) OnMonth() {
	t.CrossMonth()
}

// Cross 执行时间跨越检查，使用默认周一作为周起始日
// Cross performs time crossing check with Monday as default week start day
func Cross(c TimeCrosser) {
	CrossWithWeekStartsAt(c, carbon.Monday)
}

// CrossWithWeekStartsAt 执行时间跨越检查，可指定周起始日
// CrossWithWeekStartsAt performs time crossing check with specified week start day
func CrossWithWeekStartsAt(c TimeCrosser, weekStartsAt string) {
	tt := &tunnel{TimeCrosser: c}
	timepass.Advance(tt, timepass.WithWeekStartsAt(weekStartsAt))
}
