package periodic

import (
	"fmt"
	"strings"
	"time"

	"github.com/Corentin-cott/ServeurSentinel/internal/discord"
	"github.com/Corentin-cott/ServeurSentinel/internal/tmux"
)

// Task to run periodically
func Task() {
	fmt.Println("♟ Periodic task executed at", time.Now().Format("02/01/2006 15:04:05"))
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
		discord.SendDiscordEmbed("♟ "+time.Now().Format("02/01/2006 15:04:05"), "Periodic task executed.", "#9adfba")

		// Check if the right tmux servers are running
		color := "#9adfba"
		message, err := tmux.CheckRunningServers()
		if err != nil {
			// If an error occurs, we change the color to red
			color = "#ff0000"
			fmt.Println(err)
		} else {
			// If the message contains "✘", we change the color to orange (cause it means a server wasn't supposed to be running)
			if message[:3] == "✘" {
				color = "#ff8c00"
			}
			fmt.Println("♟ Actions : " + message)
		}

		// Replace the "✔" and "✘" emojis with "\n✔" and "\n✘" for a better display in the Discord embed
		message = strings.ReplaceAll(message, "✔", "\n✔")
		message = strings.ReplaceAll(message, "✘", "\n✘")
		err = discord.SendDiscordEmbed("♟ Serveur periodic check", message, color)
		if err != nil {
			fmt.Println(err)
		}

		// Get the minecraft player game statistics
		err = discord.SendDiscordEmbed("♟ Minecraft statistics update", "TODO", "#9adfba")
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}
