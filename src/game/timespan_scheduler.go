package game

import (
	"container/heap"
	"../utility/len_fixed_que"
	"../scheduler"
)
/*
type Scheduler interface {
	ScheduleAfterDelay(action func(), delayMs int64) (int, error)
	RemoveTask(taskId int) bool
	TrySchedule() int
}
*/

type taskObj struct{
	taskId int
	action func()
}

type unAllocTask struct {
	execTime int64
	taskobj  taskObj
}

type unAllocTaskHeap []unAllocTask

func (this *unAllocTaskHeap) Len() int {
	return len(*this)
}

func (this *unAllocTaskHeap) Less(i, j int) bool{
	return (*this)[i].execTime < (*this)[j].execTime
}

func (this *unAllocTaskHeap) Swap(i, j int){
	(*this)[i], (*this)[j] = (*this)[j], (*this)[i]
}

func (this *unAllocTaskHeap) Push(elem interface{}){
	*this = append(*this, elem.(unAllocTask))
}

func (this *unAllocTaskHeap) Pop() interface{}{
	len := len(*this)
	ret := (*this)[len - 1]
	*this = (*this)[:len - 1]
	return ret
}

type timeSlot struct {
	execTime int64
	tasks    []taskObj
}


func (this *timeSlot) Reset(execTime int64){
	this.tasks = this.tasks[:0]
	this.execTime = execTime
}

type TimeSpanScheduler struct{
	slotQue *len_fixed_que.LenFixedQue
	unAlloced *unAllocTaskHeap
	nextTaskId int
	spanTimeScale int64
}

func NewTimeSpanScheduler(spanTimeScale int64, spanCount int) *TimeSpanScheduler{
	now := GetTimeStampMs()
	unAllocHeap := make(unAllocTaskHeap, 0, 128)
	ret := &TimeSpanScheduler{
		slotQue:len_fixed_que.New(spanCount),
		nextTaskId:0,
		spanTimeScale:spanTimeScale,
		unAlloced:&unAllocHeap,
	}
	for i := 0; i < spanCount; i++{
		ret.slotQue.Enqueue(timeSlot{execTime: now + int64(i) * spanTimeScale, tasks : make([]taskObj, 0, 8)})
	}
	return ret
}

func (this *TimeSpanScheduler) ScheduleAfterDelay(action func(), delayMs int64) (int, error){
	if delayMs < 0{
		return -1, scheduler.TimeError("delay time must not be less than 0 ms")
	}
	tId := this.nextTaskId
	this.nextTaskId ++
	headSlotTime := this.slotQue.GetHeadElem().(timeSlot).execTime
	now := GetTimeStampMs()
	execTime := now + delayMs
	timeDiff := execTime - headSlotTime
	slotIndex := int(timeDiff / this.spanTimeScale)
	if slotIndex < this.slotQue.Count(){
		slot:= this.slotQue.Get(slotIndex).(timeSlot)
		slot.tasks = append(slot.tasks, taskObj{taskId:tId, action:action})
		this.slotQue.Set(slotIndex, slot)
	}else{
		heap.Push(this.unAlloced, unAllocTask{taskobj:taskObj{taskId:tId, action:action}, execTime:execTime})
	}
	return tId, nil
}

func (this *TimeSpanScheduler) RemoveTask(taskId int) bool{
	for i, imax := 0, this.slotQue.Count(); i < imax; i++{
		slot := this.slotQue.Get(i).(timeSlot)
		for j, task := range slot.tasks  {
			if task.taskId == taskId{
				task.action = nil
				slot.tasks[j] = task
				return true
			}
		}
	}
	return false
}

func (this *TimeSpanScheduler) 	TrySchedule() int {
	now := GetTimeStampMs()
	ret := 0
	for {
		headSlot := this.slotQue.GetHeadElem().(timeSlot)
		execTime := headSlot.execTime
		if execTime < now {
			for _, task := range headSlot.tasks {
				if task.action != nil{
					task.action()
					ret ++
				}
			}
			this.slotQue.Dequeue()
			tailTime := this.slotQue.GetTailElem().(timeSlot).execTime
			headSlot.Reset(tailTime + this.spanTimeScale)
			this.slotQue.Enqueue(headSlot)
		}else {
			headSlot2 := this.slotQue.GetHeadElem().(timeSlot)
			for this.unAlloced.Len() > 0{
				top := heap.Pop(this.unAlloced).(unAllocTask)
				timeDiff := int(top.execTime - headSlot2.execTime)
				if timeDiff < 0{
					timeDiff = 0
				}
				slotIndex := timeDiff / int(this.spanTimeScale)
				if slotIndex < this.slotQue.Count(){
					slot := this.slotQue.Get(slotIndex).(timeSlot)
					slot.tasks = append(slot.tasks, top.taskobj)
					this.slotQue.Set(slotIndex, slot)
				}else {
					this.unAlloced.Push(top)
					break
				}
			}
			break
		}
	}
	return  ret
}
