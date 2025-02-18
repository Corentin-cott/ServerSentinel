package console

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Trigger is a struct that represents a trigger
type Trigger struct {
	Name      string            // Trigger name
	Condition func(string) bool // Condition of the trigger
	Action    func(string)      // Function to execute when the condition is met
}

// StartFileLogListener starts listening to a log file in real time
func StartFileLogListener(logFilePath string, triggers []Trigger) error {
	file, err := os.Open(logFilePath)
	if err != nil {
		return fmt.Errorf("ERROR WHILE OPENING LOG FILE NAMED %s : %v", logFilePath, err)
	}
	defer file.Close()

	// Position the cursor at the end of the file
	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("ERROR WHILE SEEKING TO THE END OF THE FILE NAMED %s : %v", logFilePath, err)
	}

	fmt.Printf("✔ Started listening to log file %s\n", logFilePath)

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

		// Remove leading and trailing whitespaces
		line = strings.TrimSpace(line)
		if line != "" {
			for _, trigger := range triggers {
				if trigger.Condition(line) {
					trigger.Action(line)
				}
			}
		}
	}
}
