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
				minecraftServerStartedRegex := regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2})\] \[Server thread/INFO](?: \[.+?/MinecraftServer])?: Done \((\d+\.\d+)s\)! For help, type "help"`)
				return minecraftServerStartedRegex.MatchString(line)
			},
			Action: func(line string, serverID int) {
				server, err := db.GetServerById(serverID)
				if err != nil {
					fmt.Println("ERROR WHILE GETTING SERVER BY ID FOR MINECRAFT SERVER STARTED: " + err.Error())
					return
				}
				discord.SendDiscordEmbed(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.MinecraftChatChannelID, server.Nom+" viens d'ouvrir !", "Connectez-vous !\nLe serveur "+server.Jeu+" est en ligne !", server.EmbedColor)
			},
		},
		{
			// This trigger is used to detect when a minecraft server is stopped
			Name: "MinecraftServerStopped",
			Condition: func(line string) bool {
				return strings.Contains(line, "Stopping the server")
			},
			Action: func(line string, serverID int) {
				server, err := db.GetServerById(serverID)
				if err != nil {
					fmt.Println("ERROR WHILE GETTING SERVER BY ID FOR MINECRAFT SERVER STOPPED: " + err.Error())
					return
				}
				discord.SendDiscordEmbed(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.MinecraftChatChannelID, server.Nom+" viens de fermer !", "Merci d'avoir joué !\nLe serveur "+server.Jeu+" est désormais hors ligne.", server.EmbedColor)
			},
		},
		{
			// This trigger is used to detect when a player joins a Minecraft server
			Name: "PlayerJoinedMinecraftServer",
			Condition: func(line string) bool {
				return strings.Contains(line, "joined the game")
			},
			Action: func(line string, serverID int) {
				// Player name
				playerJoinedRegex := regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2})\] \[Server thread/INFO](?: \[.+?/MinecraftServer])?: (.+) joined the game`)
				matches := playerJoinedRegex.FindStringSubmatch(line)
				if len(matches) < 3 {
					fmt.Println("ERROR WHILE EXTRACTING JOINED PLAYER NAME")
					return
				}
				// Server infos
				server, err := db.GetServerById(serverID)
				if err != nil {
					fmt.Println("ERROR WHILE GETTING SERVER BY ID FOR MINECRAFT SERVER STOPPED: " + err.Error())
					return
				}
				// Action
				discord.SendDiscordEmbed(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.MinecraftChatChannelID, matches[2]+" à rejoint "+server.Nom, "", server.EmbedColor)
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
				// Player name
				playerDisconnectedRegex := regexp.MustCompile(`\[\d{2}:\d{2}:\d{2}\] \[Server thread/INFO\].*?: ([^\s]+) (?:left the game|disconnected|lost connection)`)
				matches := playerDisconnectedRegex.FindStringSubmatch(line)
				if len(matches[1]) < 3 { // Minecraft player names are at least 3 characters long, so this filter prevents false positives
					fmt.Println("ERROR WHILE EXTRACTING DISCONNECTED PLAYER NAME")
					return
				}
				// Server infos
				server, err := db.GetServerById(serverID)
				if err != nil {
					fmt.Println("ERROR WHILE GETTING SERVER BY ID FOR MINECRAFT SERVER STOPPED: " + err.Error())
					return
				}
				// Action
				discord.SendDiscordEmbed(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.MinecraftChatChannelID, matches[1]+" à quitté "+server.Nom, "", server.EmbedColor)
				WriteToLogFile("/var/log/serversentinel/playerdisconnected.log", matches[1])
			},
		},
		{
			// This trigger is used to detect when a palworld server is started
			Name: "PalworldServerStarted",
			Condition: func(line string) bool {
				palworldServerStartedRegex := regexp.MustCompile(`Running Palworld dedicated server on :\d+`)
				return palworldServerStartedRegex.MatchString(strings.TrimSpace(line))
			},
			Action: func(line string, serverID int) {
				// Server infos
				server, err := db.GetServerById(serverID)
				if err != nil {
					fmt.Println("ERROR WHILE GETTING SERVER BY ID FOR MINECRAFT SERVER STOPPED: " + err.Error())
					return
				}
				discord.SendDiscordEmbed(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.PalworldChatChannelID, server.Nom+" viens d'ouvrir !", "Connectez-vous !\nLe serveur "+server.Jeu+" est en ligne !", server.EmbedColor)
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
