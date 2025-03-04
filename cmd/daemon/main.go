package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Corentin-cott/ServeurSentinel/config"
	"github.com/Corentin-cott/ServeurSentinel/internal/console"
	"github.com/Corentin-cott/ServeurSentinel/internal/db"
	"github.com/Corentin-cott/ServeurSentinel/internal/discord"
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

	// Test the discord bot
	discord.SendDiscordMessage(config.AppConfig.Bots["multiloutreBot"], config.AppConfig.DiscordChannels.MinecraftChatChannelID, "Mineotter parlait à moi <@383676607434457088>")

	// Create a list of triggers and create a wait group
	// triggersList := triggers.GetTriggers([]string{"MinecraftServerStarted", "MinecraftServerStopped", "PlayerJoinedMinecraftServer"}) // Example with selected triggers
	triggersList := triggers.GetTriggers([]string{})
	fmt.Println("✔ Triggers loaded : ", len(triggersList), " triggers.")
	console.ProcessLogFiles("/opt/serversentinel/serverslog/", triggersList)

	fmt.Println("♦ Server Sentinel daemon stopped.")
}
