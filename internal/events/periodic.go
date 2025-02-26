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
			message += "\n\nCurently opened serveurs:"
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
		serverList, err := db.GetAllMinecraftServers()
		if err != nil {
			fmt.Println("❌ Error while getting the Minecraft servers list " + err.Error())
		}

		nbPlayerSaves := 0
		nbPlayerSavesFailed := 0
		serverThatFailedSavesList := make([]string, 0)
		for _, server := range serverList {
			fmt.Println("*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-* Server " + server.Nom + " *-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*")
			playerUUIDList, err := services.GetMinecraftPlayerServerUUIDSaves(server)
			if err != nil {
				fmt.Println("❌ Error while getting the Minecraft players list " + err.Error())
				serverThatFailedSavesList = append(serverThatFailedSavesList, server.Nom)
			}

			for _, playerUUID := range playerUUIDList {
				nbPlayerSaves++

				isValid, err := services.IsValidMinecraftUUID(playerUUID)
				if err != nil {
					fmt.Println("❌ Error while validating Minecraft UUID:", err)
					continue
				}
				if !isValid {
					fmt.Println("❌ Invalid Minecraft UUID : " + playerUUID + " for server " + server.Nom)
					nbPlayerSavesFailed++

					// Since the UUID is invalid, we skip the player and continue to the next one
					continue
				}

				fmt.Println("------------------------ Player " + playerUUID + " ------------------------")
				_, err = db.CheckAndInsertPlayerWithPlayerUUID(playerUUID, 1, "nil") // 1 is the ID of the server "La Vanilla", wich will put Minecraft as the game. Not a good practice, but it's a quick fix cause i'm tired.
				if err != nil {
					fmt.Println("❌ Error while checking or inserting the player " + playerUUID + " " + err.Error())
					nbPlayerSavesFailed++
					return fmt.Errorf("FAILED TO CHECK OR INSERT PLAYER: %v", err)
				}

				player, err := db.GetPlayerByUUID(playerUUID)
				if player.UtilisateurID == -1 {
					fmt.Println("Player " + playerUUID + " doesn't have a user account linked. This is not a problem.")
				}
				if err != nil {
					fmt.Println(err)
				}

				playerID := player.ID
				playerUUID := player.CompteID

				_, _, playerStats, error := services.GetMinecraftPlayerGameStatistics(playerID, playerUUID, server)
				if error != nil {
					fmt.Println(error)
				}

				if db.CheckMinecraftPlayerGameStatisticsExists(playerUUID, server.ID) {
					fmt.Println("Player " + playerUUID + " statistics already exists, they will be updated.")
					err := db.UpdateMinecraftPlayerGameStatistics(server.ID, playerUUID, playerStats)
					if err != nil {
						nbPlayerSavesFailed++
						fmt.Println(err)
					}
				} else {
					fmt.Println("Player " + playerUUID + " statistics doesn't exist, they will be created.")
					err := db.SaveMinecraftPlayerGameStatistics(server.ID, playerUUID, playerStats)
					if err != nil {
						nbPlayerSavesFailed++
						fmt.Println(err)
					}
				}

				if err != nil {
					fmt.Println(err)
				}
			}
		}
		fmt.Println("*-*-*-*-*-*-*-*-* ✔ Minecraft players stats are saved *-*-*-*-*-*-*-*-*")

		var serverThatFailedSavesListString string
		if len(serverThatFailedSavesList) > 0 {
			serverThatFailedSavesListString = "Servers that failed to give player saves list : " + strings.Join(serverThatFailedSavesList, ", ")
			color = "#ff0000"
		} else {
			serverThatFailedSavesListString = "No server failed to give player saves list."
			color = "#9adfba"
		}

		if nbPlayerSavesFailed > 0 {
			color = "#ff8c00"
		}

		err = discord.SendDiscordEmbed(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.ServerStatusChannelID,
			"♟ Minecraft statistics update", "Minecraft players stats are saved.\n\n"+
				fmt.Sprint(nbPlayerSaves)+" players stats updated from "+fmt.Sprint(len(serverList))+" servers.\n"+
				"Failed to save stats for "+fmt.Sprint(nbPlayerSavesFailed)+" players.\n\n"+
				serverThatFailedSavesListString, color)

		if err != nil {
			fmt.Println("❌ Error while sending the Discord message " + err.Error())
		}
	}

	return nil
}
