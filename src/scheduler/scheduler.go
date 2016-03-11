package scheduler

type Scheduler interface {
	ScheduleAfterDelay(action func(), delayMs int64) (int, error)
	RemoveTask(taskId int) bool
	TrySchedule() int
}

type TimeError string

func (this TimeError) Error() string {
	return string(this)
}