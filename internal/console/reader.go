package console

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Corentin-cott/ServeurSentinel/internal/db"
	"github.com/Corentin-cott/ServeurSentinel/internal/models"
	"github.com/Corentin-cott/ServeurSentinel/internal/triggers"
)

// StartFileLogListener starts listening to a log file in real time
func StartFileLogListener(logFilePath string, triggersVar []models.Trigger) error {
	file, err := os.Open(logFilePath)
	if err != nil {
		return fmt.Errorf("ERROR WHILE OPENING LOG FILE NAMED %s : %v", logFilePath, err)
	}
	defer file.Close()

	// Position the cursor at the end of the file
	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("ERROR WHILE SEEKING TO THE END OF THE FILE NAMED %s : %v", logFilePath, err)
	}

	fmt.Printf("✔ Started listening to log file %s with %d triggers.\n", logFilePath, len(triggersVar))

	// Read the file line by line
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n') // Define the delimiter as '\n' is the line break character
		if err != nil {
			if err.Error() == "EOF" { // If the end of the file is reached, wait for 100ms and continue
				time.Sleep(100 * time.Millisecond)
				continue
			}
			return fmt.Errorf("ERROR WHILE READING LOG FILE NAMED %s : %v", logFilePath, err)
		}

		// We check if logFilePath ends with 1.log or 2.log. If 1.log, it's primary server. If 2.log, it's secondary server.
		var serverType string
		if strings.HasSuffix(logFilePath, "1.log") {
			serverType = "primary"
		} else if strings.HasSuffix(logFilePath, "2.log") {
			serverType = "secondary"
		} else if strings.HasSuffix(logFilePath, "3.log") {
			serverType = "partner"
		} else {
			fmt.Println("✘ Error while determining server type, is the log file name correct?")
			return nil
		}

		// Depending on the server type, we set the serverID
		var serverID int
		if serverType == "primary" {
			serverID = db.GetPrimaryServerId()
		} else if serverType == "secondary" {
			serverID = db.GetSecondaryServerId()
		} else if serverType == "partner" {
			serverID = db.GetPartenariatServerId()
		}

		// We send the log in the appropriate channel by webhook
		err = triggers.SendToDiscordWebhook(serverType, line)
		if err != nil {
			fmt.Println("✘ Error while sending log to Discord webhook: " + err.Error())
		}

		// Remove leading and trailing whitespaces
		line = removeANSIcodes(strings.TrimSpace(line))
		if line != "" {
			for _, trigger := range triggersVar {
				if trigger.Condition(line) {
					trigger.Action(line, serverID)
				}
			}
		}
	}
}

// Function to process all log files in a directory
func ProcessLogFiles(logDirPath string, triggersList []models.Trigger) {
	logFiles, err := filepath.Glob(filepath.Join(logDirPath, "*.log"))
	if err != nil {
		log.Fatalf("✘ FATAL ERROR WHEN GETTING LOG FILES: %v", err)
	}

	if len(logFiles) == 0 {
		log.Println("✘ No log files found in the directory, did you forget to redirect the logs to the folder?")
		return
	}

	// Create a wait group
	var wg sync.WaitGroup

	// Start a goroutine for each log file
	for _, logFile := range logFiles {
		// If logFile is "3.log", ignore it
		if strings.HasSuffix(logFile, "3.log") { // BAD PRATICE: hardcoded value, but need something quick
			continue
		}
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			err := StartFileLogListener(file, triggersList)
			if err != nil {
				log.Printf("✘ Error with file %s: %v\n", file, err)
			}
		}(logFile)
	}

	// Wait for all goroutines to finish
	wg.Wait()
}

func removeANSIcodes(line string) string {
	// Regex to remove ANSI codes
	re := regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)
	line = re.ReplaceAllString(line, "")
	return line
}
