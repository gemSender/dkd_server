package game

import (
	"testing"
	"fmt"
	"math/rand"
	"time"
	"../scheduler"
)
var s1 scheduler.Scheduler = NewTimeSpanScheduler(10, 100)
var s2 scheduler.Scheduler = NewHeapScheduler(100)
func Test_Scheduler(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	s := s2
	for i := 0; i < 100; i++{
		t :=  rand.Int63n(2000)
		s.ScheduleAfterDelay(func(){fmt.Println("task ", t, " execute")}, t)
	}
	sum := 0
	for sum < 100{
		sum += s.TrySchedule()
	}
}
