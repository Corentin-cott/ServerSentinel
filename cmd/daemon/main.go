package main

import (
	"fmt"
	"log"
	"os"
	"database/sql"
	"time"

	"github.com/Corentin-cott/ServerSentinel/config"
	"github.com/Corentin-cott/ServerSentinel/internal/console"
	"github.com/Corentin-cott/ServerSentinel/internal/db"
	"github.com/Corentin-cott/ServerSentinel/internal/db_stats"
	periodic "github.com/Corentin-cott/ServerSentinel/internal/events"
	"github.com/Corentin-cott/ServerSentinel/internal/triggers"
	"github.com/Corentin-cott/ServerSentinel/internal/minecraft_stats"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "serversentinel",
		Short: "ServerSentinel manages Minecraft and Palworld servers in tmux sessions.",
	}

	// Command: serversentinel daemon
	var daemonCmd = &cobra.Command{
		Use:   "daemon",
		Short: "Runs ServerSentinel in daemon mode",
		Run:   runDaemon,
	}

	// Add commands to root
	rootCmd.AddCommand(daemonCmd)

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

	save_stats_pls()

	// Create a list of triggers and create a wait group
	// triggersList := triggers.GetTriggers([]string{"MinecraftServerStarted", "MinecraftServerStopped", "PlayerJoinedMinecraftServer"}) // Example with selected triggers
	triggersList := triggers.GetTriggers([]string{})
	fmt.Println("✔ Triggers loaded : ", len(triggersList), " triggers.")
	console.ProcessLogFiles("/opt/serversentinel/serverslog/", triggersList)

	fmt.Println("♦ Server Sentinel daemon stopped.")
}

func save_stats_pls() {
	err := db.ConnectToDatabase()

	dsn := "serveursentinel:8CdKBKiM8KD!4spf?!xEtTd$9DymczgShSM5eycn@tcp(antredesloutres.fr:32768)/adl_global?parseTime=true"
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer sqlDB.Close()

	db_stats.Init(sqlDB)

	fmt.Println("🔄 Synchronisation des stats Minecraft...")
	if err := minecraft_stats.SyncMinecraftStats(); err != nil {
		log.Fatal("Erreur synchronisation:", err)
	}
	fmt.Println("✅ Stats Minecraft synchronisées.")
}