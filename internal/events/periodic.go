package periodic

import (
	"fmt"
	"time"

	"github.com/Corentin-cott/ServeurSentinel/config"
	"github.com/Corentin-cott/ServeurSentinel/internal/discord"
)

// Var for the colors of the Discord embeds
var goodColor = "#9adfba"
var mehColor = "#ff8c00"
var badColor = "#ff0000"

// Task to run periodically
func Task() {
	fmt.Println("♟ Periodic task executed at", time.Now().Format("02/01/2006 15:04:05"))
}

// Task : Server check
func TaskServerCheck() {
	/* Deprecated, need to be updated */
}

// Task : Minecraft statistics update
func TaskMinecraftStatsUpdate() {
	/* Deprecated, need to be updated */
}

// Start the periodic task
func StartPeriodicTask(PeriodicEventsMin int) error {
	if PeriodicEventsMin <= 0 {
		return fmt.Errorf("ERROR: PERIODIC EVENTS MINUTES MUST BE GREATER THAN 0, CURRENTLY %d", PeriodicEventsMin)
	}

	interval := time.Duration(PeriodicEventsMin) * time.Minute
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		// Execute the periodic task : Log the time and send a message to Discord
		Task()
		discord.SendDiscordEmbed(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.ServerStatusChannelID, "♟ "+time.Now().Format("02/01/2006 15:04:05"), "Periodic task executed.", goodColor)

		// Check if the right tmux servers are running
		if config.AppConfig.PeriodicEvents.ServersCheckEnabled {
			TaskServerCheck()
		} else {
			fmt.Println("♟ Server check is disabled.")
		}

		// Get the minecraft player game statistics
		if config.AppConfig.PeriodicEvents.MinecraftStatsEnabled {
			TaskMinecraftStatsUpdate()
		} else {
			fmt.Println("♟ Minecraft statistics update is disabled.")
		}
	}

	return nil
}
