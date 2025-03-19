package tmux

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/Corentin-cott/ServeurSentinel/config"
	"github.com/Corentin-cott/ServeurSentinel/internal/db"
	"github.com/Corentin-cott/ServeurSentinel/internal/discord"
	"github.com/Corentin-cott/ServeurSentinel/internal/models"
)

// Check if the active servers match the servers in the database
func CheckRunningServers() (string, error) {
	var message strings.Builder
	errorMessages := ""

	// Get the primary, secondary, and partenariat servers from the database
	primaryServer, err := db.GetServerById(db.GetPrimaryServerId())
	if err != nil {
		return "", fmt.Errorf("ERROR WHILE GETTING PRIMARY SERVER: %v", err)
	}

	secondaryServer, err := db.GetServerById(db.GetSecondaryServerId())
	if err != nil {
		return "", fmt.Errorf("ERROR WHILE GETTING SECONDARY SERVER: %v", err)
	}

	partenariatServer, err := db.GetServerById(db.GetPartenariatServerId())
	if err != nil {
		return "", fmt.Errorf("ERROR WHILE GETTING PARTENARIAT SERVER: %v", err)
	}

	fmt.Println("Supposed to be running servers:")
	fmt.Println("- Primary:", primaryServer.Nom)
	fmt.Println("- Secondary:", secondaryServer.Nom)
	fmt.Println("- Partenariat:", partenariatServer.Nom)

	// Get the active tmux sessions
	activeSessions, err := GetTmuxSessions()
	if err != nil {
		return "", fmt.Errorf("ERROR WHILE GETTING ACTIVE TMUX SESSIONS: %v", err)
	}

	// Validate active sessions
	for _, session := range activeSessions {
		isSupposedToBeRunning, err := IsServerSupposedToBeRunning(session)
		if err != nil {
			errorMessages += fmt.Sprintf("ERROR WHILE CHECKING IF %s SHOULD BE RUNNING: %v\n", session, err)
			continue
		}

		// If a server is running but shouldn't be, stop it
		if !isSupposedToBeRunning {
			if err := StopServerTmux(session); err != nil {
				errorMessages += fmt.Sprintf("ERROR WHILE STOPPING %s: %v\n", session, err)
			} else {
				fmt.Fprintf(&message, "✘ Stopped server: %s (not supposed to be running)\n", session)
			}
		}
	}

	// Ensure expected servers are running
	for _, server := range []models.Server{primaryServer, secondaryServer, partenariatServer} {
		isRunning, err := IsServerRunning(server.Nom)
		if err != nil {
			errorMessages += fmt.Sprintf("ERROR WHILE CHECKING IF %s IS RUNNING: %v\n", server.Nom, err)
			continue
		}

		// If not running, start it
		if !isRunning {
			var sessionID int
			switch server.ID {
			case primaryServer.ID:
				sessionID = 1
			case secondaryServer.ID:
				sessionID = 2
			case partenariatServer.ID:
				sessionID = 3
			default:
				errorMessages += fmt.Sprintf("SERVER %s IS NOT PRIMARY, SECONDARY, OR PARTENARIAT\n", server.Nom)
				continue
			}

			if err := StartServerTmux(sessionID, server); err != nil {
				errorMessages += fmt.Sprintf("ERROR WHILE STARTING %s: %v\n", server.Nom, err)
			} else {
				fmt.Fprintf(&message, "✔ Started server: %s (supposed to be running)\n", server.Nom)
			}
		}
	}

	// If no messages were added, everything is fine
	if message.Len() == 0 {
		message.WriteString("✔ Nothing to do, all servers are running as expected.")
	}

	if errorMessages != "" {
		return message.String(), fmt.Errorf("%s", errorMessages)
	}
	return message.String(), nil
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
	partenariatServerID := db.GetPartenariatServerId()

	return server.ID == primaryServerID || server.ID == secondaryServerID || server.ID == partenariatServerID, nil
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

	if server.Jeu == "Minecraft" {
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
	}

	// Build the full command to start the server using its StartScript
	command := fmt.Sprintf(
		"cd %s && tmux new-session -d -s '%s' './%s | tee -a /opt/serversentinel/serverslog/%s.log'",
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

	time.Sleep(2 * time.Second)

	// Just to be sure, send the exit command to the tmux session if it's still running
	isRunning, err = IsServerRunning(serverName)
	if err != nil {
		return fmt.Errorf("ERROR WHILE CHECKING THE TMUX SESSION: %v", err)
	}

	if isRunning {
		// Kill the tmux session
		command := fmt.Sprintf("tmux kill-session -t '%s'", serverName)
		err = exec.Command("bash", "-c", command).Run()
		if err != nil {
			return fmt.Errorf("ERROR WHILE STOPPING THE TMUX SESSION: %v", err)
		}
	}

	// We send a discord message to the minecraft chat channel
	server, err := db.GetServerByName(serverName)
	if err != nil {
		return fmt.Errorf("ERROR WHILE GETTING SERVER BY NAME: %v", err)
	}

	if server.Jeu == "Minecraft" {
		discord.SendDiscordEmbed(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.MinecraftChatChannelID, serverName+" se ferme.", "Merci d'avoir joué !", server.EmbedColor)
	} else {
		discord.SendDiscordEmbed(config.AppConfig.Bots["multiloutreBot"], config.AppConfig.DiscordChannels.PalworldChatChannelID, serverName+" se ferme.", "Merci d'avoir joué !", server.EmbedColor)
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

// Return supposed sessionID for a server
func GetSessionIDForServer(serverID int) (int, error) {
	primaryServerID := db.GetPrimaryServerId()
	secondaryServerID := db.GetSecondaryServerId()
	partenariatServerID := db.GetPartenariatServerId()

	if serverID == primaryServerID {
		return 1, nil
	} else if serverID == secondaryServerID {
		return 2, nil
	} else if serverID == partenariatServerID {
		return 3, nil
	} else {
		return -1, fmt.Errorf("SERVER %d IS NOT PRIMARY NOR SECONDARY", serverID)
	}
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
