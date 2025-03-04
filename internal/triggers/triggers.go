package triggers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Corentin-cott/ServeurSentinel/config"
	"github.com/Corentin-cott/ServeurSentinel/internal/db"
	"github.com/Corentin-cott/ServeurSentinel/internal/discord"
	"github.com/Corentin-cott/ServeurSentinel/internal/models"
)

// GetTriggers returns the list of triggers filtered by names
func GetTriggers(selectedTriggers []string) []models.Trigger {
	// All available triggers
	allTriggers := []models.Trigger{
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
				if isPlayerMessage(line) {
					return false
				}
				match, _ := regexp.MatchString(`.*Done\s*\(.*?\)!.*`, line)
				return match
			},
			Action: func(line string, serverID int) {
				// Server infos
				server, err := db.GetServerById(serverID)
				if err != nil {
					fmt.Println("ERROR WHILE GETTING SERVER BY ID FOR MINECRAFT SERVER STARTED: " + err.Error())
					return
				}
				discord.SendDiscordEmbed(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.MinecraftChatChannelID, server.Nom+" viens d'ouvrir !", "Connectez-vous !\nLe serveur "+server.Jeu+" est en ligne !", server.EmbedColor)
			},
		},
		{
			// This trigger is used to detect when a player sends a message in the Minecraft chat
			Name: "PlayerChatMinecraftServer",
			Condition: func(line string) bool {
				return isPlayerMessage(line)
			},
			Action: func(line string, serverID int) {
				err := PlayerMessageAction(line, serverID)
				if err != nil {
					fmt.Println("ERROR WHILE PROCESSING PLAYER MESSAGE: " + err.Error())
				}
			},
		},
		{
			// This trigger is used to detect when a player joins a Minecraft server
			Name: "PlayerJoinedMinecraftServer",
			Condition: func(line string) bool {
				if isPlayerMessage(line) {
					return false
				}
				return strings.Contains(line, "joined the game")
			},
			Action: func(line string, serverID int) {
				err := PlayerJoinedAction(line, serverID)
				if err != nil {
					fmt.Println("ERROR WHILE PROCESSING PLAYER JOINED: " + err.Error())
				}
			},
		},
		{
			// This trigger is used to detect when a player disconnects from a Minecraft server
			Name: "PlayerDisconnectedMinecraftServer",
			Condition: func(line string) bool {
				if isPlayerMessage(line) {
					return false
				}
				return strings.Contains(line, "lost connection: Disconnected")
			},
			Action: func(line string, serverID int) {
				err := PlayerLeftAction(line, serverID)
				if err != nil {
					fmt.Println("ERROR WHILE PROCESSING PLAYER DISCONNECTED: " + err.Error())
				}
			},
		},
		{
			// This trigger is used to detect when a palworld server is started
			Name: "PalworldServerStarted",
			Condition: func(line string) bool {
				if isPlayerMessage(line) {
					return false
				}
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
				discord.SendDiscordEmbed(config.AppConfig.Bots["multiloutreBot"], config.AppConfig.DiscordChannels.PalworldChatChannelID, server.Nom+" viens d'ouvrir !", "Connectez-vous !\nLe serveur "+server.Jeu+" est en ligne !", server.EmbedColor)
			},
		},
	}

	// If no specific triggers are requested, return all triggers
	if len(selectedTriggers) == 0 {
		return allTriggers
	}

	// Filter triggers based on the selected names
	var filteredTriggers []models.Trigger
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

// Détecte si c'est un message de joueur
func isPlayerMessage(line string) bool {
	match, _ := regexp.MatchString(`.*<.*?>.*`, line)
	return match
}
