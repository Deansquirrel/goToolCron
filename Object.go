package goToolCron

import (
	"github.com/robfig/cron"
	"time"
)

type taskState struct {
	Key     string
	Cron    *cron.Cron
	Running bool
	Prev    time.Time

	Func        func()
	PanicHandle func(interface{})
}
