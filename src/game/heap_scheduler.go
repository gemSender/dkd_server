package game

import (
	"container/heap"
)


type ScheduleObject struct {
	TaskId int
	ExecTime int64
	Action func()
	Canceled bool
}

type ScheduleHeap struct{
	slice []ScheduleObject
	nextTaskId int
}

func (this *ScheduleHeap)  Len() int{
	return len(this.slice)
}

func (this *ScheduleHeap) Less(i, j int) bool{
	slice := this.slice
	return slice[i].ExecTime < slice[j].ExecTime
}

func (this *ScheduleHeap) Swap(i, j int){
	slice := this.slice
	slice[i], slice[j] = slice[j], slice[i]
}

func (this *ScheduleHeap) Push(x interface{}){
	this.slice = append(this.slice, x.(ScheduleObject))
}

func (this *ScheduleHeap) Pop()  interface{}{
	n := len(this.slice)
	ret := this.slice[n-1]
	this.slice = this.slice[0 : n-1]
	return ret
}

type TimeError string

func (this TimeError) Error() string {
	return string(this)
}

func (this *ScheduleHeap) ScheduleAfterDelay(action func(), delayMs int64) (int, error){
	if delayMs < 0{
		return -1, TimeError("delay time must not be less than 0 ms")
	}
	taskId := this.nextTaskId
	this.nextTaskId ++
	heap.Push(this, ScheduleObject{TaskId:taskId, ExecTime:GetTimeStampMs() + delayMs, Action:action, Canceled:false})
	return taskId, nil
}

func (this *ScheduleHeap) RemoveTask(taskId int) bool{
	for i, item := range this.slice{
		item.Canceled = true
		this.slice[i] = item
		return true
	}
	return  false
}

func (this *ScheduleHeap) TrySchedule() int{
	now := GetTimeStampMs()
	ret := 0
	for this.Len() > 0 {
		obj := heap.Pop(this).(ScheduleObject)
		if now < obj.ExecTime{
			this.Push(obj)
			return ret;
		}
		if !obj.Canceled {
			ret++
			obj.Action()
		}
	}
	return ret
}

func NewHeapScheduler(cap int) *ScheduleHeap{
	ret := &ScheduleHeap{slice:make([]ScheduleObject, 0, cap), nextTaskId:0}
	heap.Init(ret)
	return ret
}
