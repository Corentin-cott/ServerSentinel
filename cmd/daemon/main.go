package main

import (
	"fmt"
	"log"
	"path/filepath"
	"sync"
	"time"

	"github.com/Corentin-cott/ServeurSentinel/config"
	"github.com/Corentin-cott/ServeurSentinel/internal/console"
	"github.com/Corentin-cott/ServeurSentinel/internal/db"
	periodic "github.com/Corentin-cott/ServeurSentinel/internal/events"
	"github.com/Corentin-cott/ServeurSentinel/internal/triggers"
)

func main() {
	fmt.Println(time.Now().Location())
	fmt.Println("Starting the Server Sentinel daemon (" + time.Now().Format("02/01/2006 15:04:05") + ") ...")

	// Load the configuration file
	err := config.LoadConfig("/opt/serversentinel/config.json")
	if err != nil {
		log.Fatalf("FATAL ERROR LOADING CONFIG JSON FILE: %v", err)
	}

	// Check that the bot configuation exists
	if len(config.AppConfig.Bots) == 0 {
		log.Fatalf("FATAL ERROR: NO BOT CONFIGURATION FOUND")
		return
	} else {
		fmt.Println("✔ Bot configuration loaded :")
		for botName, botConfig := range config.AppConfig.Bots {
			fmt.Println("  -", botName, ":", botConfig)
		}
	}

	// Initialize the connection to the database
	err = db.ConnectToDatabase()
	if err != nil {
		log.Fatalf("FATAL ERROR TESTING DATABASE CONNECTION: %v", err)
	}

	// Start the periodic service
	go func() {
		err := periodic.StartPeriodicTask(config.AppConfig.PeriodicEventsMin)
		if err != nil {
			log.Fatalf("FATAL ERROR STARTING PERIODIC TASK: %v", err)
		}
	}()
	fmt.Println("✔ Periodic service started, interval is set to", config.AppConfig.PeriodicEventsMin, "minutes.")

	// Create a list of triggers and create a wait group
	// triggersList := triggers.GetTriggers([]string{"MinecraftServerStarted", "MinecraftServerStopped", "PlayerJoinedMinecraftServer"}) // Example with selected triggers
	triggersList := triggers.GetTriggers([]string{})
	fmt.Println("✔ Triggers loaded : ", len(triggersList), " triggers.")
	processLogFiles("/opt/serversentinel/serverslog/", triggersList)

	fmt.Println("♦ Server Sentinel daemon stopped.")
}

// Function to process all log files in a directory
func processLogFiles(logDirPath string, triggersList []console.Trigger) {
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
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			err := console.StartFileLogListener(file, triggersList)
			if err != nil {
				log.Printf("✘ Error with file %s: %v\n", file, err)
			}
		}(logFile)
	}

	// Wait for all goroutines to finish
	wg.Wait()
}
