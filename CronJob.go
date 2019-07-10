package goToolCron

import (
	"github.com/Deansquirrel/goToolCommon"
	"github.com/robfig/cron"
	"time"
)

func AddFunc(key string, spec string, cmd func(), panicHandle func(interface{})) error {
	if HasTask(key) {
		DelFunc(key)
	}

	c := cron.New()
	err := c.AddFunc(spec, getRunFunc(key, cmd, panicHandle))
	if err != nil {
		return err
	}
	ts := &taskState{
		Key:     key,
		Cron:    c,
		CronStr: spec,
		Running: false,
		Prev:    getInitTime(),

		Func:        cmd,
		PanicHandle: panicHandle,
	}

	taskList.Register() <- goToolCommon.NewObject(key, ts)
	ch := make(chan struct{}, 1)
	ch <- struct{}{}
	taskTicket[key] = ch

	c.Start()
	ts.Running = true
	return nil
}

func DelFunc(key string) {
	tsi := taskList.GetObject(key)
	if tsi == nil {
		return
	}
	ts := tsi.(*taskState)
	if ts.Running {
		ts.Cron.Stop()
	}

	taskList.Unregister() <- key
	ch, ok := taskTicket[key]
	if ok {
		delete(taskTicket, key)
		close(ch)
	}
}

func Start(key string) {
	tsi := taskList.GetObject(key)
	if tsi == nil {
		return
	}
	ts := tsi.(*taskState)
	if ts.Running {
		return
	}
	ts.Running = true
	ts.Cron.Start()
}

func Stop(key string) {
	tsi := taskList.GetObject(key)
	if tsi == nil {
		return
	}
	ts := tsi.(*taskState)
	if !ts.Running {
		return
	}
	ts.Running = false
	ts.Cron.Stop()
}

func HasTask(key string) bool {
	t := taskList.GetObject(key)
	return t != nil
}

//key没有对应task时，返回false
func IsRunning(key string) bool {
	tsi := taskList.GetObject(key)
	if tsi == nil {
		return false
	}
	ts := tsi.(*taskState)
	return ts.Running
}

//key没有对应ch时，返回false
func IsWorking(key string) bool {
	ch, ok := taskTicket[key]
	if !ok {
		return false
	}
	return len(ch) == 0
}

func CronStr(key string) string {
	tsi := taskList.GetObject(key)
	if tsi == nil {
		return ""
	}
	ts := tsi.(*taskState)
	return ts.CronStr
}

/*
根据key对应的任务获取下次执行时间
如果key没有对应的任务，则返回初始时间 1970-01-01 00:00:00
*/
func Next(key string) time.Time {
	tsi := taskList.GetObject(key)
	if tsi == nil {
		return getInitTime()
	}
	ts := tsi.(*taskState)
	eList := ts.Cron.Entries()
	if len(eList) < 1 {
		return getInitTime()
	}
	return eList[0].Next
}

func Prev(key string) time.Time {
	tsi := taskList.GetObject(key)
	if tsi == nil {
		return getInitTime()
	}
	ts := tsi.(*taskState)
	return ts.Prev
}

func Func(key string) func() {
	tsi := taskList.GetObject(key)
	if tsi == nil {
		return nil
	}
	ts := tsi.(*taskState)
	return ts.Func
}

//根据传入参数构造实际运行函数（增加状态控制内容）
func getRunFunc(key string, cmd func(), panicHandle func(interface{})) func() {
	return func() {
		//判断是否可运行（相同任务不重复运行）
		ch, ok := taskTicket[key]
		if !ok {
			return
		}
		select {
		case <-ch:
			//panic处理
			defer func() {
				err := recover()
				if err != nil && panicHandle != nil {
					panicHandle(err)
				}
			}()
			defer func() {
				ch, ok := taskTicket[key]
				if ok && len(ch) == 0 {
					ch <- struct{}{}
				}
			}()
			//更新上次运行时间
			{
				tsi := taskList.GetObject(key)
				if tsi == nil {
					return
				}
				ts := tsi.(*taskState)
				ts.Prev = time.Now()
			}
			//真实任务执行
			if cmd != nil {
				cmd()
			}
		default:
			//相同任务正在运行，跳过执行
			return
		}
	}
}
