package goToolCron

import (
	"github.com/Deansquirrel/goToolCommon"
	"time"
)

func init() {
	taskList = goToolCommon.NewObjectManager()
	taskTicket = make(map[string]chan struct{})
}

func getInitTime() time.Time {
	return time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local)
}
