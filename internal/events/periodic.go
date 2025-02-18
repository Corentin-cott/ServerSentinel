package periodic

import (
	"fmt"
	"time"
)

// This test function is called every time the periodic task is executed
func Task() {
	fmt.Println("♟ Periodic task executed at", time.Now().Format("02/01/2006 15:04:05"))
}

// StartPeriodicTask starts a periodic task that executes the Task function every PeriodicEventsMin minutes
func StartPeriodicTask(PeriodicEventsMin int) error {
	// Interval must be greater than 0
	if PeriodicEventsMin <= 0 {
		return fmt.Errorf("ERROR: PERIODIC EVENTS MINUTES MUST BE GREATER THAN 0, CURRENTLY %d", PeriodicEventsMin)
	}

	interval := time.Duration(PeriodicEventsMin) * time.Minute
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		Task()
	}

	return nil
}
