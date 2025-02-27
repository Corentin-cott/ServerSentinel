package triggers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Corentin-cott/ServeurSentinel/config"
	"github.com/Corentin-cott/ServeurSentinel/internal/console"
	"github.com/Corentin-cott/ServeurSentinel/internal/db"
	"github.com/Corentin-cott/ServeurSentinel/internal/discord"
)

// GetTriggers returns the list of triggers filtered by names
func GetTriggers(selectedTriggers []string) []console.Trigger {
	// All available triggers
	allTriggers := []console.Trigger{
		{
			// This is an example trigger, use it as a template to create new triggers
			Name: "ExampleTrigger",
			Condition: func(line string) bool {
				// Here you can define the condition that will trigger the action, you're most probably looking for a specific string in the server log
				return strings.Contains(line, "whatever line you're looking for here")
			},
			Action: func(line string, serverID int) {
				// Here you can define the action that will be executed
				fmt.Println("Example trigger action executed")
			},
		},
		{
			// This trigger is used to detect when a minecraft server is started
			Name: "MinecraftServerStarted",
			Condition: func(line string) bool {
				// For this trigger, we'll need regex to extract the time and the server name
				minecraftServerStartedRegex := regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2})\] \[Server thread/INFO\]: Done \((\d+\.\d+)s\)! For help, type "help"`)
				return minecraftServerStartedRegex.MatchString(line)
			},
			Action: func(line string, serverID int) {
				discord.SendDiscordEmbed(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.MinecraftChatChannelID, "Connectez-vous !", "Le serveur Minecraft est en ligne", "#9adfba")
			},
		},
		{
			// This trigger is used to detect when a minecraft server is stopped
			Name: "MinecraftServerStopped",
			Condition: func(line string) bool {
				return strings.Contains(line, "Stopping the server")
			},
			Action: func(line string, serverID int) {
				discord.SendDiscordEmbed(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.MinecraftChatChannelID, "Arrêt du serveur", "Le serveur Minecraft est désormais hors ligne", "#9adfba")
			},
		},
		{
			// This trigger is used to detect when a player joins a Minecraft server
			Name: "PlayerJoinedMinecraftServer",
			Condition: func(line string) bool {
				return strings.Contains(line, "joined the game")
			},
			Action: func(line string, serverID int) {
				playerJoinedRegex := regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2})\] \[Server thread/INFO\]: (.+) joined the game`)
				matches := playerJoinedRegex.FindStringSubmatch(line)
				if len(matches) < 3 {
					fmt.Println("ERROR WHILE EXTRACTING JOINED PLAYER NAME")
					return
				}
				discord.SendDiscordMessage(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.MinecraftChatChannelID, matches[2]+" à rejoint le serveur")
				playerID, err := db.CheckAndInsertPlayerWithPlayerName(matches[2], 1, "now")
				if err != nil {
					fmt.Println("ERROR WHILE CHECKING OR INSERTING PLAYER " + matches[2] + " IN DATABASE: " + err.Error())
				}
				err = db.SaveConnectionLog(playerID, serverID)
				if err != nil {
					fmt.Println("ERROR WHILE SAVING CONNECTION LOG FOR PLAYER " + matches[2] + " IN DATABASE: " + err.Error())
				}
				err = db.UpdatePlayerLastConnection(playerID)
				if err != nil {
					fmt.Println("ERROR WHILE UPDATING LAST CONNECTION FOR PLAYER " + matches[2] + " IN DATABASE: " + err.Error())
				}
				WriteToLogFile("/var/log/serversentinel/playerjoined.log", matches[2])
			},
		},
		{
			// This trigger is used to detect when a player disconnects from a Minecraft server
			Name: "PlayerDisconnectedMinecraftServer",
			Condition: func(line string) bool {
				return strings.Contains(line, "lost connection: Disconnected")
			},
			Action: func(line string, serverID int) {
				playerDisconnectedRegex := regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2})\] \[Server thread/INFO\]: (.+) lost connection: Disconnected`)
				matches := playerDisconnectedRegex.FindStringSubmatch(line)
				if len(matches) < 3 {
					fmt.Println("ERROR WHILE EXTRACTING DISCONNECTED PLAYER NAME")
					return
				}
				discord.SendDiscordMessage(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.MinecraftChatChannelID, matches[2]+" à quitté le serveur")
				WriteToLogFile("/var/log/serversentinel/playerdisconnected.log", matches[2])
			},
		},
		{
			// This trigger is used to detect when a palworld server is started
			Name: "PalworldServerStarted",
			Condition: func(line string) bool {
				palworldServerStartedRegex := regexp.MustCompile(`Running Palworld dedicated server on :\d+`)
				return palworldServerStartedRegex.MatchString(line)
			},
			Action: func(line string, serverID int) {
				discord.SendDiscordEmbed(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.PalworldChatChannelID, "Connectez-vous !", "Le serveur Palworld est en ligne", "#9adfba")
			},
		},
	}

	// If no specific triggers are requested, return all triggers
	if len(selectedTriggers) == 0 {
		return allTriggers
	}

	// Filter triggers based on the selected names
	var filteredTriggers []console.Trigger
	for _, trigger := range allTriggers {
		for _, name := range selectedTriggers {
			if trigger.Name == name {
				filteredTriggers = append(filteredTriggers, trigger)
				break
			}
		}
	}

	return filteredTriggers
}
