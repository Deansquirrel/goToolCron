package goToolCron

import "github.com/Deansquirrel/goToolCommon"

var taskList goToolCommon.IObjectManager
var taskTicket map[string]chan struct{}
