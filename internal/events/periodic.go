package periodic

import (
	"fmt"
	"strings"
	"time"

	"github.com/Corentin-cott/ServeurSentinel/config"
	"github.com/Corentin-cott/ServeurSentinel/internal/db"
	"github.com/Corentin-cott/ServeurSentinel/internal/discord"
	"github.com/Corentin-cott/ServeurSentinel/internal/services"
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
		discord.SendDiscordEmbed(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.ServerStatusChannelID, "♟ "+time.Now().Format("02/01/2006 15:04:05"), "Periodic task executed.", "#9adfba")

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
		tmuxSessions, err := tmux.GetTmuxSessions()
		if err != nil {
			fmt.Println(err)
		} else {
			// For each opened tmux session, we add \n- before the name
			message += "\n\n♟ Curently opened serveurs:"
			for _, session := range tmuxSessions {
				message += "\n- " + session
			}
			message += fmt.Sprintf("\n%d opened sessions.", len(tmuxSessions))
		}

		err = discord.SendDiscordEmbed(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.ServerStatusChannelID, "♟ Serveur periodic check", message, color)
		if err != nil {
			fmt.Println(err)
		}

		// Get the minecraft player game statistics
		fmt.Println("Saving Minecraft players game statistics...")
		serverListe, err := db.GetAllMinecraftServers()
		if err != nil {
			fmt.Println("❌ Error while getting the Minecraft servers list" + err.Error())
		}

		for _, server := range serverListe {
			fmt.Println("Server : " + server.Nom)
			playerID := 1
			playerUUID := "4a575f21f53346718ef3aaff4146ff8c"

			_, _, playerStats, error := services.GetMinecraftPlayerGameStatistics(playerID, playerUUID, server)
			if error != nil {
				fmt.Println(error)
			}

			if db.CheckMinecraftPlayerGameStatisticsExists(playerUUID, server.ID) {
				fmt.Println("Player " + playerUUID + " statistics already exists, updating...")
				err := db.UpdateMinecraftPlayerGameStatistics(server.ID, playerUUID, playerStats)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println("✔ Player " + playerUUID + " statistics updated successfully.")
			} else {
				err := db.SaveMinecraftPlayerGameStatistics(server.ID, playerUUID, playerStats)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println("✔ Player " + playerUUID + " statistics saved successfully.")
			}

			err = discord.SendDiscordEmbed(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.ServerStatusChannelID, "♟ Minecraft statistics update", "TODO", "#9adfba")
			if err != nil {
				fmt.Println(err)
			}
		}
		fmt.Println("✔ Minecraft players game statistics saved successfully.")
	}

	return nil
}
