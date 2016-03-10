package game

import (
	"testing"
	"fmt"
	"math/rand"
	"time"
)

func Test_Scheduler(t *testing.T) {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	s := NewHeapScheduler(100)
	for i := 0; i < 100; i++{
		t :=  rand.Int63n(2000)
		s.ScheduleAfterDelay(func(){fmt.Println("task ", t, " execute")}, t)
	}
	for s.Len() > 0{
		s.TrySchedule()
	}
}