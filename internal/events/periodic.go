package periodic

import (
	"fmt"
	"strings"
	"time"

	"github.com/Corentin-cott/ServeurSentinel/internal/db"
	"github.com/Corentin-cott/ServeurSentinel/internal/models"
	"github.com/Corentin-cott/ServeurSentinel/internal/tmux"
)

// Task to run periodically
func Task() {
	fmt.Println("♟ Periodic task executed at", time.Now().Format("02/01/2006 15:04:05"))
}

// Check if the active servers match the servers in the database
func CheckRunningServers() (string, error) {
	var message strings.Builder

	// Get the primary and secondary servers from the database
	primaryServer, err := db.GetServerById(db.GetPrimaryServerId())
	if err != nil {
		return "", fmt.Errorf("ERROR WHILE GETTING PRIMARY SERVER: %v", err)
	}

	secondaryServer, err := db.GetServerById(db.GetSecondaryServerId())
	if err != nil {
		return "", fmt.Errorf("ERROR WHILE GETTING SECONDARY SERVER: %v", err)
	}

	// Get the active tmux sessions
	activeSessions, err := tmux.GetTmuxSessions()
	if err != nil {
		return "", fmt.Errorf("ERROR WHILE GETTING ACTIVE TMUX SESSIONS: %v", err)
	}

	// Check if the active servers match the servers in the database
	for _, session := range activeSessions {
		isSupposedToBeRunning, err := tmux.IsServerSupposedToBeRunning(session)
		if err != nil {
			fmt.Fprintf(&message, "ERROR WHILE CHECKING IF %s SHOULD BE RUNNING: %v", session, err)
			continue
		}

		// If the server is not supposed to be running, stop it
		if !isSupposedToBeRunning {
			err := tmux.StopServerTmux(session)
			if err != nil {
				fmt.Fprintf(&message, "ERROR WHILE STOPPING %s: %v", session, err)
			} else {
				fmt.Fprintf(&message, "✘ Stopped server: %s (not supposed to be running) ", session)
			}
		}
	}

	// Check if the espected servers are running
	for _, server := range []models.Server{primaryServer, secondaryServer} {
		isRunning, err := tmux.IsServerRunning(server.Nom)
		if err != nil {
			fmt.Fprintf(&message, "ERROR WHILE CHECKING IF %s IS RUNNING: %v", server.Nom, err)
			continue
		}

		// If the server is not running, start it
		if !isRunning {
			sessionID := -1
			if server.ID == primaryServer.ID {
				sessionID = 1
			} else if server.ID == secondaryServer.ID {
				sessionID = 2
			} else {
				fmt.Fprintf(&message, "SERVER %s IS NOT PRIMARY NOR SECONDARY", server.Nom)
				continue
			}

			err := tmux.StartServerTmux(sessionID, server)
			if err != nil {
				fmt.Fprintf(&message, "ERROR WHILE STARTING %s: %v", server.Nom, err)
			} else {
				fmt.Fprintf(&message, "✔ Started server: %s (supposed to be running) ", server.Nom)
			}
		}
	}

	// Empty message means all servers are running
	if message.Len() == 0 {
		message.WriteString("✔ Nothing to do, all servers are running as expected.")
	}

	return message.String(), nil
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
		message, err := CheckRunningServers()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("♟ Actions : " + message)
		}
	}

	return nil
}
