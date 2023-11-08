package client

import (
	"time"
)

type ChangeRecordSettings interface {
	SetSelectLimit(int)
	SetCountGoroutines(int)
	SetSleepTime(time.Duration)
}

type RecordSettings struct {
	selectLimit     int
	countGoroutines int
	sleepTime       time.Duration
}

func (r *RecordSettings) SetSelectLimit(limit int) {
	r.selectLimit = limit
}

func (r *RecordSettings) SetCountGoroutines(countGoroutines int) {
	r.countGoroutines = countGoroutines
}
func (r *RecordSettings) SetSleepTime(duration time.Duration) {
	r.sleepTime = duration
}
