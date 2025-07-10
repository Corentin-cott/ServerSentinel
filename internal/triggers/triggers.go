package triggers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Corentin-cott/ServerSentinel/config"
	"github.com/Corentin-cott/ServerSentinel/internal/db"
	"github.com/Corentin-cott/ServerSentinel/internal/discord"
	"github.com/Corentin-cott/ServerSentinel/internal/models"
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
			// This trigger is used to detect when a player sends a message in any server chat
			Name: "PlayerChatInServer",
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
			// This trigger is used to detect when a minecraft server is stopped
			Name: "MinecraftServerStopped",
			Condition: func(line string) bool {
				if isPlayerMessage(line) {
					return false
				}
				match, _ := regexp.MatchString(`.*Stopping the server.*`, line)
				return match
			},
			Action: func(line string, serverID int) {
				// Server infos
				server, err := db.GetServerById(serverID)
				if err != nil {
					fmt.Println("ERROR WHILE GETTING SERVER BY ID FOR MINECRAFT SERVER STOPPED: " + err.Error())
					return
				}
				discord.SendDiscordEmbed(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.MinecraftChatChannelID, server.Nom+" viens de fermer !", "Le serveur "+server.Jeu+" est hors ligne !", server.EmbedColor)
			},
		},
		{
			// This trigger is used to detect when a minecraft server crashes
			Name: "MinecraftServerCrashed",
			Condition: func(line string) bool {
				if isPlayerMessage(line) {
					return false
				}
				match, _ := regexp.MatchString(`.*has crashed.*`, line)
				return match
			},
			Action: func(line string, serverID int) {
				// Server infos
				server, err := db.GetServerById(serverID)
				if err != nil {
					fmt.Println("ERROR WHILE GETTING SERVER BY ID FOR MINECRAFT SERVER CRASHED: " + err.Error())
					return
				}
				discord.SendDiscordEmbed(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.MinecraftChatChannelID, server.Nom+" vient de crash !", "Le serveur "+server.Jeu+" est hors ligne !", server.EmbedColor)
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
			// This trigger is used to detect when a Minecraft Player get an advancement
			Name: "PlayerGetAdvancement",
			Condition: func(line string) bool {
				if isPlayerMessage(line) {
					return false
				}
				return strings.Contains(line, "has made the advancement")
			},
			Action: func(line string, serverID int) {
				err := PlayerGetAdvancementAction(line, serverID)
				if err != nil {
					fmt.Println("ERROR WHILE PROCESSING PLAYER GET ADVANCEMENT: " + err.Error())
				}
			},
		},
		{
			// This trigger is used to detect when a Minecraft Player dies
			Name: "PlayerDeath",
			Condition: func(line string) bool {
				if isPlayerMessage(line) {
					return false
				}
				isDeath, _, _ := isPlayerDeathMessage(line)
				return isDeath
			},
			Action: func(line string, serverID int) {
				_, deathMessage, playername := isPlayerDeathMessage(line)
				fmt.Println("Player death detected for: " + playername)
				fmt.Println("Death message: " + deathMessage)
				err := PlayerDeathAction(deathMessage, playername, serverID)
				if err != nil {
					fmt.Println("ERROR WHILE PROCESSING PLAYER DEATH: " + err.Error())
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
					fmt.Println("ERROR WHILE GETTING SERVER BY ID FOR MINECRAFT SERVER STARTED: " + err.Error())
					return
				}
				discord.SendDiscordEmbed(config.AppConfig.Bots["multiloutreBot"], config.AppConfig.DiscordChannels.PalworldChatChannelID, server.Nom+" viens d'ouvrir !", "Connectez-vous !\nLe serveur "+server.Jeu+" est en ligne !", server.EmbedColor)
			},
		},
		{
			// This trigger is used to detect when a player joins a Palworld server
			Name: "PlayerJoinedPalworldServer",
			Condition: func(line string) bool {
				if isPlayerMessage(line) {
					return false
				}
				match, _ := regexp.MatchString(`\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] \[LOG\] .*? \d{1,3}(\.\d{1,3}){3} connected the server\. \(User id: .*?\)`, line)
				return match
			},
			Action: func(line string, serverID int) {
				err := PlayerJoinedAction(line, serverID)
				if err != nil {
					fmt.Println("ERROR WHILE PROCESSING PLAYER JOINED: " + err.Error())
				}
			},
		},
		{
			// This trigger is used to detect when a player disconnects from a Palworld server
			Name: "PlayerDisconnectedPalworldServer",
			Condition: func(line string) bool {
				if isPlayerMessage(line) {
					return false
				}
				return strings.Contains(line, "left the server.")
			},
			Action: func(line string, serverID int) {
				err := PlayerLeftAction(line, serverID)
				if err != nil {
					fmt.Println("ERROR WHILE PROCESSING PLAYER DISCONNECTED: " + err.Error())
				}
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

// Detect if a line is a player message
func isPlayerMessage(line string) bool {
	match, _ := regexp.MatchString(`.*<.*?>.*`, line)
	return match
}

// Regex améliorée pour matcher le format du log
var deathMessageRegex = regexp.MustCompile(`\[.*?\] \[.*?\]: (.*?) (was slain by|was run over by|was killed by|drowned|starved to death|blew up|withered away|fell from a high place|fell out of the world)(.*)`)

// Détecte si une ligne est un message de mort et extrait les infos
func isPlayerDeathMessage(line string) (bool, string, string) {
	matches := deathMessageRegex.FindStringSubmatch(line)
	if len(matches) > 3 {
		playerName := strings.TrimSpace(matches[1])                // Nom du joueur
		deathMessage := playerName + " " + matches[2] + matches[3] // Message de mort complet
		return true, deathMessage, playerName
	}
	return false, "", ""
}
