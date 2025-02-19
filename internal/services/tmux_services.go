package services

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/Corentin-cott/ServeurSentinel/internal/models"
)

// Check if a server is currently running in a tmux session
func CheckServerTmux(serverName string) (bool, error) {
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

// StartServerTmux starts a Minecraft server in a tmux session
func StartServerTmux(sessionID int, server models.Server) error {
	// Check if the server is already running
	isRunning, err := CheckServerTmux(server.Nom)
	if err != nil {
		return fmt.Errorf("ERROR WHILE CHECKING THE TMUX SESSION: %v", err)
	}
	if isRunning {
		return fmt.Errorf("SERVER %s IS ALREADY RUNNING", server.Nom)
	}

	fmt.Println("Starting the tmux session for", server.Nom)

	// Extract the main version of Minecraft
	versionParts := strings.Split(server.Version, ".")
	if len(versionParts) < 2 {
		return fmt.Errorf("INVALID MINECRAFT VERSION: %s", server.Version)
	}
	mcMainVersion := versionParts[0] + "." + versionParts[1]

	// Map the main version of Minecraft to the corresponding Java version
	javaVersion, err := GetJavaVersionForMinecraftVersion(mcMainVersion, server.Modpack)
	if err != nil {
		return fmt.Errorf("ERROR WHILE GETTING JAVA VERSION FOR MINECRAFT VERSION: %v", err)
	}

	// Start the tmux session with the good Java version
	javaPath := fmt.Sprintf("/usr/lib/jvm/java-%s-openjdk-amd64/bin/java", javaVersion)
	command := fmt.Sprintf(
		"cd %s && tmux new-session -d -s '%s' '%s -Xmx1024M -Xms1024M -jar server.jar nogui | tee /opt/serversentinel/serverslog/%s.log'",
		server.PathServ, server.Nom, javaPath, strconv.Itoa(sessionID),
	)

	// Execute the command
	err = exec.Command("bash", "-c", command).Run()
	if err != nil {
		return fmt.Errorf("ERROR WHILE STARTING THE TMUX SESSION: %v", err)
	}

	fmt.Printf("✔ Server %s started using Java %s\n", server.Nom, javaVersion)
	return nil
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
