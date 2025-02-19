package services

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
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

// Start a server in a new tmux session
func StartServerTmux(emplacementID int, serverName string, serverVersion string, serverPath string, serverScript string, logFile string) error {
	fmt.Println("Starting the tmux session for", serverName)

	// Full command: Start tmux session and redirect logs
	command := fmt.Sprintf("cd %s && tmux new-session -d -s '%s' 'java -Xmx1024M -Xms1024M -jar server.jar nogui | tee /opt/serversentinel/serverslog/%s.log'",
		serverPath, serverName, strconv.Itoa(emplacementID))

	err := exec.Command("bash", "-c", command).Run()
	if err != nil {
		return fmt.Errorf("ERROR WHILE STARTING THE TMUX SESSION: %v", err)
	}

	return nil
}
