package tmux

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Corentin-cott/ServeurSentinel/internal/db"
	"github.com/Corentin-cott/ServeurSentinel/internal/models"
)

// Check if the active servers match the servers in the database
func CheckRunningServers() (string, error) {
	var message strings.Builder
	errorMessages := ""

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
	activeSessions, err := GetTmuxSessions()
	if err != nil {
		return "", fmt.Errorf("ERROR WHILE GETTING ACTIVE TMUX SESSIONS: %v", err)
	}

	// Check if the active servers match the servers in the database
	for _, session := range activeSessions {
		isSupposedToBeRunning, err := IsServerSupposedToBeRunning(session)
		if err != nil {
			errorMessages += fmt.Sprintf("ERROR WHILE CHECKING IF %s SHOULD BE RUNNING: %v", session, err)
			fmt.Fprintf(&message, "%s", errorMessages)
			continue
		}

		// If the server is not supposed to be running, stop it
		if !isSupposedToBeRunning {
			err := StopServerTmux(session)
			if err != nil {
				errorMessages += fmt.Sprintf("ERROR WHILE STOPPING %s: %v", session, err)
				fmt.Fprintf(&message, "%s", errorMessages)
			} else {
				fmt.Fprintf(&message, "✘ Stopped server: %s (not supposed to be running) ", session)
			}
		}
	}

	// Check if the espected servers are running
	for _, server := range []models.Server{primaryServer, secondaryServer} {
		isRunning, err := IsServerRunning(server.Nom)
		if err != nil {
			errorMessages += fmt.Sprintf("ERROR WHILE CHECKING IF %s IS RUNNING: %v", server.Nom, err)
			fmt.Fprintf(&message, "%s", errorMessages)
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
				errorMessages += fmt.Sprintf("SERVER %s IS NOT PRIMARY NOR SECONDARY", server.Nom)
				fmt.Fprintf(&message, "%s", errorMessages)
				continue
			}

			err := StartServerTmux(sessionID, server)
			if err != nil {
				errorMessages += fmt.Sprintf("ERROR WHILE STARTING %s: %v", server.Nom, err)
				fmt.Fprintf(&message, "%s", errorMessages)
			} else {
				fmt.Fprintf(&message, "✔ Started server: %s (supposed to be running) ", server.Nom)
			}
		}
	}

	// Empty message means all servers are running
	if message.Len() == 0 {
		message.WriteString("✔ Nothing to do, all servers are running as expected.")
	}

	if errorMessages != "" {
		return message.String(), fmt.Errorf("%s", errorMessages)
	} else {
		return message.String(), nil
	}
}

// Check if a server is currently running in a tmux session
func IsServerRunning(serverName string) (bool, error) {
	command := fmt.Sprintf("tmux list-sessions -F '#{session_name}' | grep -w \"%s\" | wc -l", serverName)
	commandOutput, err := exec.Command("bash", "-c", command).Output()
	if err != nil {
		return false, fmt.Errorf("ERROR WHILE CHECKING THE TMUX SESSION: %v", err)
	}

	sessionCount, err := strconv.Atoi(strings.TrimSpace(string(commandOutput)))
	if err != nil {
		return false, fmt.Errorf("FAILED TO CONVERT SESSION COUNT TO INTEGER: %v", err)
	}

	return sessionCount > 0, nil
}

// Check if a server is supposed to be running
func IsServerSupposedToBeRunning(serverName string) (bool, error) {
	server, err := db.GetServerByName(serverName)
	if err != nil {
		return false, fmt.Errorf("ERROR WHILE GETTING SERVER BY NAME: %v", err)
	}

	primaryServerID := db.GetPrimaryServerId()
	secondaryServerID := db.GetSecondaryServerId()

	return server.ID == primaryServerID || server.ID == secondaryServerID, nil
}

// StartServerTmux starts a Minecraft server in a tmux session
func StartServerTmux(sessionID int, server models.Server) error {
	// Check if the server is already running
	isRunning, err := IsServerRunning(server.Nom)
	if err != nil {
		return fmt.Errorf("ERROR WHILE CHECKING THE TMUX SESSION: %v", err)
	}
	if isRunning {
		return fmt.Errorf("SERVER %s IS ALREADY RUNNING", server.Nom)
	}

	fmt.Println("Starting the tmux session for", server.Nom+"...")

	// Extract the main version of Minecraft
	versionParts := strings.Split(server.Version, ".")
	if len(versionParts) < 2 {
		return fmt.Errorf("INVALID MINECRAFT VERSION: %s", server.Version)
	}
	mcMainVersion := versionParts[0] + "." + versionParts[1]

	// Map the main version of Minecraft to the corresponding Java version
	javaVersion, err := GetJavaVersionForMinecraftVersion(mcMainVersion, server.Modpack)
	fmt.Println("Java version for Minecraft version", mcMainVersion, ":", javaVersion)

	if err != nil {
		return fmt.Errorf("ERROR WHILE GETTING JAVA VERSION FOR MINECRAFT VERSION: %v", err)
	}

	// Build the full command to start the server using its StartScript
	command := fmt.Sprintf(
		"cd %s && tmux new-session -d -s '%s' './%s | tee /opt/serversentinel/serverslog/%s.log'",
		server.PathServ, server.Nom, server.StartScript, strconv.Itoa(sessionID),
	)

	// Execute the command
	err = exec.Command("bash", "-c", command).Run()
	if err != nil {
		return fmt.Errorf("ERROR WHILE STARTING THE TMUX SESSION: %v", err)
	}

	fmt.Printf("✔ Server %s started using StartScript: %s\n", server.Nom, server.StartScript)
	return nil
}

// StopServerTmux stops a Minecraft server in a tmux session
func StopServerTmux(serverName string) error {
	// Check if the server is running
	isRunning, err := IsServerRunning(serverName)
	if err != nil {
		return fmt.Errorf("ERROR WHILE CHECKING THE TMUX SESSION: %v", err)
	}
	if !isRunning {
		return fmt.Errorf("SERVER %s IS NOT RUNNING", serverName)
	}

	fmt.Println("Stopping the tmux session for", serverName+"...")

	// Send the stop command to the tmux session
	command := fmt.Sprintf("tmux send-keys -t '%s' 'stop' C-m", serverName)
	err = exec.Command("bash", "-c", command).Run()
	if err != nil {
		return fmt.Errorf("ERROR WHILE STOPPING THE SERVER: %v", err)
	}
	time.Sleep(1 * time.Second) // Wait for the server to stop

	// Just to be sure, send the exit command to the tmux session if it's still running
	isRunning, err = IsServerRunning(serverName)
	if err != nil {
		return fmt.Errorf("ERROR WHILE CHECKING THE TMUX SESSION: %v", err)
	}

	if isRunning {
		// Send the exit command to the tmux session
		command = fmt.Sprintf("tmux send-keys -t '%s' 'exit' C-m", serverName)
		err = exec.Command("bash", "-c", command).Run()
		if err != nil {
			return fmt.Errorf("ERROR WHILE STOPPING THE TMUX SESSION: %v", err)
		}
		time.Sleep(1 * time.Second) // Wait for the server to stop again
	}

	fmt.Printf("✔ Server %s stopped\n", serverName)
	return nil
}

// Returns opened tmux sessions
func GetTmuxSessions() ([]string, error) {
	commandOutput, err := exec.Command("tmux", "list-sessions", "-F", "#{session_name}").Output()
	if err != nil {
		// Error can just be that there are no tmux sessions, so we return an empty array
		fmt.Println("✘ No tmux sessions found, or error while getting tmux sessions:", err)
		return []string{}, nil
	}

	// Remove the newline character and split the output into an array
	sessions := strings.Split(strings.TrimSpace(string(commandOutput)), "\n")

	return sessions, nil
}

// Returns the appropriate Java version for a given Minecraft version
func GetJavaVersionForMinecraftVersion(mcVersion string, mcModpack string) (string, error) {
	// Extract the main version of Minecraft
	versionParts := strings.Split(mcVersion, ".")
	if len(versionParts) < 2 {
		return "", fmt.Errorf("INVALID MINECRAFT VERSION: %s", mcVersion)
	}
	mcMainVersion := versionParts[0] + "." + versionParts[1]

	// Map the main version of Minecraft to the corresponding Java version
	javaVersionMap := map[string]string{
		"1.7":  "11",
		"1.8":  "11",
		"1.9":  "11",
		"1.10": "11",
		"1.11": "11",
		"1.12": "11",
		"1.16": "11",
		"1.17": "16",
		"1.18": "17",
		"1.19": "17",
		"1.20": "21",
		"1.21": "21",
		"1.22": "21",
	}
	if mcModpack != "Minecraft Vanilla" && mcModpack != "Vanilla" {
		javaVersionMap = map[string]string{
			"1.7":  "8",
			"1.8":  "8",
			"1.9":  "8",
			"1.10": "8",
			"1.11": "8",
			"1.12": "8",
			"1.16": "8",
			"1.17": "16",
			"1.18": "17",
			"1.19": "17",
			"1.20": "21",
			"1.21": "21",
			"1.22": "21",
		}
	}

	// Check if the main version of Minecraft is supported
	javaVersion, exists := javaVersionMap[mcMainVersion]
	if !exists {
		return "", fmt.Errorf("UNSUPPORTED MINECRAFT VERSION: %s", mcMainVersion)
	}

	return javaVersion, nil
}
