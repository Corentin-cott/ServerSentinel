package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Corentin-cott/ServeurSentinel/config"
	"github.com/Corentin-cott/ServeurSentinel/internal/console"
	"github.com/Corentin-cott/ServeurSentinel/internal/db"
	periodic "github.com/Corentin-cott/ServeurSentinel/internal/events"
	"github.com/Corentin-cott/ServeurSentinel/internal/tmux"
	"github.com/Corentin-cott/ServeurSentinel/internal/triggers"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "serversentinel",
		Short: "ServerSentinel manages Minecraft and Palworld servers in tmux sessions.",
	}

	// Command: serversentinel start-server [id]
	var startServerCmd = &cobra.Command{
		Use:   "start-server [id]",
		Short: "Starts a game server by its ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			serverID := args[0]
			fmt.Printf("Starting server with ID: %s\n", serverID)
			commandStartStopServerWithID(serverID, "start")
		},
	}

	// Command: serversentinel stop-server [id]
	var stopServerCmd = &cobra.Command{
		Use:   "stop-server [id]",
		Short: "Stops a game server by its ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			serverID := args[0]
			fmt.Printf("Stopping server with ID: %s\n", serverID)
			commandStartStopServerWithID(serverID, "stop")
		},
	}

	// Command: serversentinel check-server
	var checkServerCmd = &cobra.Command{
		Use:   "check-server",
		Short: "Check and start servers that are supposed to be running",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			// Because it's a CLI command, we need to load the configuration file
			err := config.LoadConfig("/opt/serversentinel/config.json")
			if err != nil {
				log.Fatalf("FATAL ERROR LOADING CONFIG JSON FILE: %v", err)
				return
			}

			// Initialize the connection to the database
			err = db.ConnectToDatabase()
			if err != nil {
				log.Fatalf("FATAL ERROR TESTING DATABASE CONNECTION: %v", err)
				return
			}

			// Check if the right tmux servers are running
			fmt.Printf("Checking and starting servers that are supposed to be running\n")
			message, err := tmux.CheckRunningServers()
			if err != nil {
				log.Fatalf("FATAL ERROR CHECKING RUNNING SERVERS: %v", err)
				return
			}
			fmt.Println(message)
		},
	}

	// Command: serversentinel daemon
	var daemonCmd = &cobra.Command{
		Use:   "daemon",
		Short: "Runs ServerSentinel in daemon mode",
		Run:   runDaemon,
	}

	// Add commands to root
	rootCmd.AddCommand(daemonCmd)
	rootCmd.AddCommand(startServerCmd)
	rootCmd.AddCommand(stopServerCmd)
	rootCmd.AddCommand(checkServerCmd)

	// Execute CLI
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runDaemon(cmd *cobra.Command, args []string) {
	fmt.Println("Starting the Server Sentinel daemon (" + time.Now().Format("02/01/2006 15:04:05") + ") ...")

	// Load the configuration file
	err := config.LoadConfig("/opt/serversentinel/config.json")
	if err != nil {
		log.Fatalf("FATAL ERROR LOADING CONFIG JSON FILE: %v", err)
		return
	}

	if !config.AppConfig.PeriodicEvents.ServersCheckEnabled {
		fmt.Println("♟ Periodic task : Servers check disabled.")
	}
	if !config.AppConfig.PeriodicEvents.MinecraftStatsEnabled {
		fmt.Println("♟ Periodic task : Minecraft statistics retrieval disabled.")
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
	console.ProcessLogFiles("/opt/serversentinel/serverslog/", triggersList)

	fmt.Println("♦ Server Sentinel daemon stopped.")
}

// Function to start a server by its ID. This function is use in the CLI command "start-server"
func commandStartStopServerWithID(serverID string, action string) {
	// Action can only be "start" or "stop"
	if action != "start" && action != "stop" {
		log.Fatalf("FATAL ERROR: INVALID ACTION: %s", action)
		return
	}

	// Check if the server ID is a number, and if so, convert it to an integer
	serverIDInt, err := strconv.Atoi(serverID)
	if err != nil {
		log.Fatalf("FATAL ERROR: SERVER ID IS NOT A NUMBER: %v", err)
		return
	}

	// Because it's a CLI command, we need to load the configuration file
	err = config.LoadConfig("/opt/serversentinel/config.json")
	if err != nil {
		log.Fatalf("FATAL ERROR LOADING CONFIG JSON FILE: %v", err)
		return
	}

	// Initialize the connection to the database
	err = db.ConnectToDatabase()
	if err != nil {
		log.Fatalf("FATAL ERROR TESTING DATABASE CONNECTION: %v", err)
	}

	// Now we check if the server exists in the database
	server, err := db.GetServerById(serverIDInt)
	if err != nil {
		log.Fatalf("FATAL ERROR GETTING SERVER BY ID: %v", err)
		return
	}

	// If the action is "stop", we stop the server
	if action == "stop" {
		// Stop the server
		err = tmux.StopServerTmux(server.Nom)
		if err != nil {
			log.Fatalf("FATAL ERROR STOPPING SERVER: %v", err)
			return
		}
		return
	}

	// If the action is "start", we first check if the server is primary or not
	if server.ID == 1 {
		err = db.SetPrimaryServerId(server.ID)
		if err != nil {
			log.Fatalf("FATAL ERROR SETTING PRIMARY SERVER ID: %v", err)
			return
		}
	} else {
		err = db.SetSecondaryServerId(server.ID)
		if err != nil {
			log.Fatalf("FATAL ERROR SETTING SECONDARY SERVER ID: %v", err)
			return
		}
	}

	// Start the server
	message, err := tmux.CheckRunningServers()
	if err != nil {
		log.Fatalf("FATAL ERROR CHECKING RUNNING SERVERS: %v", err)
		return
	}

	fmt.Println(message)
}
