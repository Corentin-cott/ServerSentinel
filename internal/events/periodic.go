package periodic

import (
	"fmt"
	"time"

	"github.com/Corentin-cott/ServeurSentinel/internal/db"
	"github.com/Corentin-cott/ServeurSentinel/internal/services"
)

// Task to run periodically
func Task() {
	fmt.Println("♟ Periodic task executed at", time.Now().Format("02/01/2006 15:04:05"))
}

// Check if the active servers match the servers in the database
func CheckActiveServers() string {
	primaryServer, err := db.GetServerById(db.GetPrimaryServerId())
	if err != nil {
		return fmt.Sprintf("Error while getting the primary server from the database: %v", err)
	}

	secondaryServer, err := db.GetServerById(db.GetSecondaryServerId())
	if err != nil {
		return fmt.Sprintf("Error while getting the secondary server from the database: %v", err)
	}

	// Check if the primary server is running
	isPrimaryServerRunning, err := services.CheckServerTmux(primaryServer.Nom)
	if err != nil {
		return fmt.Sprintf("Error while checking the primary server tmux session: %v", err)
	} else {
		fmt.Println("Is the correct primary server running:", isPrimaryServerRunning, "("+primaryServer.Nom+")")
		if !isPrimaryServerRunning {
			err = services.StartServerTmux(1, primaryServer)
			if err != nil {
				return fmt.Sprintf("Error while starting the primary server tmux session: %v", err)
			}
		}
	}

	// Check if the secondary server is running
	isSecondaryServerRunning, err := services.CheckServerTmux(secondaryServer.Nom)
	if err != nil {
		return fmt.Sprintf("Error while checking the secondary server tmux session: %v", err)
	} else {
		fmt.Println("Is the correct secondary server running:", isSecondaryServerRunning, "("+secondaryServer.Nom+")")
		if !isSecondaryServerRunning {
			err = services.StartServerTmux(2, secondaryServer)
			if err != nil {
				return fmt.Sprintf("Error while starting the secondary server tmux session: %v", err)
			}
		}
	}

	return "Active servers checked successfully"
}

// Start the periodic task
func StartPeriodicTask(PeriodicEventsMin int) error {
	if PeriodicEventsMin <= 0 {
		return fmt.Errorf("ERROR: PERIODIC EVENTS MINUTES MUST BE GREATER THAN 0, CURRENTLY %d", PeriodicEventsMin)
	}

	interval := time.Duration(PeriodicEventsMin) * time.Minute
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		Task()
		_ = CheckActiveServers()
	}

	return nil
}
