package periodic

import (
	"fmt"
	"time"

	"github.com/Corentin-cott/ServeurSentinel/internal/db"
	"github.com/Corentin-cott/ServeurSentinel/internal/models"
)

// This test function is called every time the periodic task is executed
func Task() {
	fmt.Println("♟ Periodic task executed at", time.Now().Format("02/01/2006 15:04:05"))
}

// This function checks if the active servers match the servers in the database
func CheckActiveServers() {
	primaryServer, err := db.GetServerById(db.GetPrimaryServerId())
	if err != nil {
		fmt.Println("Error while getting the primary server from the database:", err)
	} else if primaryServer == (models.Server{}) {
		fmt.Println("No primary server has been found in the database... This may or may be not normal.")
		return
	} else {
		fmt.Println("Primary server found in the database:", primaryServer.Nom)
	}

	secondaryServer, err := db.GetServerById(db.GetSecondaryServerId())
	if err != nil {
		fmt.Println("Error while getting the secondary server from the database:", err)
	} else if secondaryServer == (models.Server{}) {
		fmt.Println("No secondary server has been found in the database... This may or may be not normal.")
		return
	} else {
		fmt.Println("Secondary server found in the database:", secondaryServer.Nom)
	}
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
		CheckActiveServers()
	}

	return nil
}
