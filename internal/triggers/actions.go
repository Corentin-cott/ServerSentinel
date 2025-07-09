package triggers

// This file contains the ACTIONS functions for the triggers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/Corentin-cott/ServeurSentinel/config"
	"github.com/Corentin-cott/ServeurSentinel/internal/db"
	"github.com/Corentin-cott/ServeurSentinel/internal/discord"
	"github.com/Corentin-cott/ServeurSentinel/internal/models"
	"github.com/Corentin-cott/ServeurSentinel/internal/services"
)

// WriteToLogFile writes a line to a log file
func WriteToLogFile(logPath string, line string) error {
	// Open the log file
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("ERROR WHILE OPENING LOG FILE: %v", err)
	}
	defer file.Close()

	// Write the line to the log file
	_, err = file.WriteString(line + "\n")
	if err != nil {
		return fmt.Errorf("ERROR WHILE WRITING TO LOG FILE: %v", err)
	}
	return nil
}

func SendToDiscordWebhook(serverType string, message string) error {
	// fmt.Println("Sending message to Discord webhook for server type:", serverType)
	webhookURL := config.AppConfig.DiscordWebhooks[serverType].URL
	if webhookURL == "" {
		return fmt.Errorf("ERROR: WEBHOOK URL FOR SERVER TYPE %s NOT FOUND", serverType)
	}

	payload := map[string]string{
		"content": message,
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("ERROR MARSHALING DISCORD PAYLOAD: %v", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("ERROR WHILE SENDING DISCORD WEBHOOK: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("DISCORD WEBHOOK RETURNED NON-2XX STATUS: %d", resp.StatusCode)
	}

	return nil
}

// Define the functions for each game, here is Minecraft
func handleMinecraftPlayerMessage(line string) (string, string, string, string, error) {
	playerChatRegex := regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2})\] \[Server thread/INFO](?: \[.+?/MinecraftServer])?: <(.+?)> (.+)`)
	matches := playerChatRegex.FindStringSubmatch(line)
	if len(matches) < 4 {
		return "", "", "", "", fmt.Errorf("ERROR WHILE EXTRACTING CHAT PLAYER NAME FOR MINECRAFT")
	}
	playerName := matches[2]
	message := matches[3]

	// Get player UUID and head URL for Minecraft
	playerUUID, err := services.GetMinecraftPlayerUUID(playerName)
	if err != nil {
		return "", "", "", "", fmt.Errorf("ERROR WHILE GETTING PLAYER UUID: %v", err)
	}

	playerHeadURL, err := services.GetMinecraftPlayerHeadURL(playerUUID)
	if err != nil {
		return "", "", "", "", fmt.Errorf("ERROR WHILE GETTING PLAYER HEAD URL: %v", err)
	}

	titleURL := "https://fr.namemc.com/profile/" + playerUUID
	return playerName, message, playerHeadURL, titleURL, nil
}

// Define the functions for each game, here is Palworld
func handlePalworldPlayerMessage(line string) (string, string, string, string, error) {
	playerChatRegex := regexp.MustCompile(`\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] \[CHAT\] <(.+?)> (.+)`)
	matches := playerChatRegex.FindStringSubmatch(line)
	if len(matches) < 3 {
		return "", "", "", "", fmt.Errorf("ERROR WHILE EXTRACTING CHAT PLAYER NAME FOR PALWORLD")
	}
	playerName := matches[1]
	message := matches[2]
	return playerName, message, "", "", nil
}

// Create a map for game-specific actions
var gameActionsMap = map[string]func(string) (string, string, string, string, error){
	"Minecraft": handleMinecraftPlayerMessage,
	"Palworld":  handlePalworldPlayerMessage,
}

// Action when a player message is detected
func PlayerMessageAction(line string, serverID int) error {
	/* Let's start by sending the message to Discord */
	// Server infos
	server, err := db.GetServerById(serverID)
	if err != nil {
		return fmt.Errorf("ERROR WHILE GETTING SERVER BY ID FOR PLAYER MESSAGE: %v", err)
	}

	// Get the appropriate game action function from the map
	actionFunc, exists := gameActionsMap[server.Jeu]
	if !exists {
		return fmt.Errorf("ERROR: SERVER GAME %v IS NOT SUPPORTED", server.Jeu)
	}

	// Call the specific action function for the game
	playerName, message, playerHeadURL, titleURL, err := actionFunc(line)
	if err != nil {
		return err
	}

	// Bot config
	botName := "mineotterBot" // Can be extended similarly using the mappage if needed

	// Send the Discord embed message
	embed := models.EmbedConfig{
		Title:       playerName,
		TitleURL:    titleURL,
		Description: message,
		Color:       server.EmbedColor,
		Thumbnail:   playerHeadURL,
		Footer:      "Message venant de " + server.Nom,
	}

	err = discord.SendDiscordEmbedWithModel(config.AppConfig.Bots[botName], config.AppConfig.DiscordChannels.MinecraftChatChannelID, embed)
	if err != nil {
		return fmt.Errorf("ERROR WHILE SENDING DISCORD EMBED: %v", err)
	}

	/* Let's now send the message to the secondary Server */
	// We first need to know if the server is primary or secondary
	serverToSend, serverToSendHost, serverToSendRconPort, serverToSendRconPassword, err := db.GetRconParameters(server.Type)

	// Now we can send the message to the server
	if serverToSend.Jeu == "Minecraft" {
		command := `tellraw @a ["",{"text":"<` + playerName + `>","color":"` + server.EmbedColor + `"},{"text":" `+ message + `"}]`
		fmt.Println("RCON parameters:", serverToSendHost, serverToSendRconPort, serverToSendRconPassword)
		resp, err := services.SendRconToMinecraftServer(serverToSendHost, serverToSendRconPort, serverToSendRconPassword, command)
		if err != nil {
			return fmt.Errorf("ERROR WHILE SENDING RCON COMMAND TO MINECRAFT SERVER: %v", err)
		}
		fmt.Println("RCON response from Minecraft server:", resp)
	} else if serverToSend.Jeu == "Palworld" {
		// Not implemented yet
	}

	return nil
}

// Define the functions for each game, here is Minecraft
func handleMinecraftPlayerJoined(line string) (string, error) {
	playerJoinedRegex := regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2})\] \[Server thread/INFO](?: \[.+?/MinecraftServer])?: (.+) joined the game`)
	matches := playerJoinedRegex.FindStringSubmatch(line)
	if len(matches) < 3 {
		return "", fmt.Errorf("ERROR WHILE EXTRACTING JOINED PLAYER NAME FOR MINECRAFT SERVER")
	}
	return matches[2], nil
}

// Define the functions for each game, here is Palworld
func handlePalworldPlayerJoined(line string) (string, error) {
	playerJoinedRegex := regexp.MustCompile(`\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] \[LOG\] (.+?) \d{1,3}(?:\.\d{1,3}){3} connected the server`)
	matches := playerJoinedRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return "", fmt.Errorf("ERROR WHILE EXTRACTING JOINED PLAYER NAME FOR PALWORLD SERVER")
	}
	return matches[1], nil
}

// Create a map for game-specific actions for joined and left actions
var gameJoinActionsMap = map[string]func(string) (string, error){
	"Minecraft": handleMinecraftPlayerJoined,
	"Palworld":  handlePalworldPlayerJoined,
}

// Action when a player joined the server
func PlayerJoinedAction(line string, serverID int) error {
	// Server infos
	server, err := db.GetServerById(serverID)
	if err != nil {
		return fmt.Errorf("ERROR WHILE GETTING SERVER BY ID FOR PLAYER JOINED: %v", err)
	}

	// Get the appropriate game action function from the map for "joined" action
	actionFunc, exists := gameJoinActionsMap[server.Jeu]
	if !exists {
		return fmt.Errorf("ERROR: SERVER GAME %v IS NOT SUPPORTED", server.Jeu)
	}

	// Call the specific action function for the game
	playerName, err := actionFunc(line)
	if err != nil {
		return err
	}

	// Bot config
	var botName string
	if server.Jeu == "Minecraft" {
		botName = "mineotterBot"
	} else {
		botName = "multiloutreBot"
	}

	// Send the Discord embed message
	discord.SendDiscordEmbed(config.AppConfig.Bots[botName], config.AppConfig.DiscordChannels.MinecraftChatChannelID, playerName+" a rejoint "+server.Nom, "", server.EmbedColor)

	// Handle player connection log in DB
	playerID, err := db.CheckAndInsertPlayerWithPlayerName(playerName, 1, "now")
	if err != nil {
		return fmt.Errorf("ERROR WHILE CHECKING OR INSERTING PLAYER: %v", err)
	}

	err = db.SaveConnectionLog(playerID, serverID)
	if err != nil {
		return fmt.Errorf("ERROR WHILE SAVING CONNECTION LOG: FOR PLAYER %v IN DATABASE: %v", playerName, err)
	}

	err = db.UpdatePlayerLastConnection(playerID)
	if err != nil {
		return fmt.Errorf("ERROR WHILE UPDATING LAST CONNECTION FOR PLAYER %v IN DATABASE: %v", playerName, err)
	}

	// Log to file
	WriteToLogFile("/var/log/serversentinel/playerjoined.log", playerName)

		/* Let's now send the message to the secondary Server */
	// We first need to know if the server is primary or secondary
	serverToSend, serverToSendHost, serverToSendRconPort, serverToSendRconPassword, err := db.GetRconParameters(server.Type)

	// Now we can send the message to the server
	if serverToSend.Jeu == "Minecraft" {
		command := `tellraw @a {"text":"` + playerName + ` a rejoint le serveur ` + server.Nom + `","color":"yellow"}`
		fmt.Println("RCON parameters:", serverToSendHost, serverToSendRconPort, serverToSendRconPassword)
		resp, err := services.SendRconToMinecraftServer(serverToSendHost, serverToSendRconPort, serverToSendRconPassword, command)
		if err != nil {
			return fmt.Errorf("ERROR WHILE SENDING RCON COMMAND TO MINECRAFT SERVER: %v", err)
		}
		fmt.Println("RCON response from Minecraft server:", resp)
	} else if serverToSend.Jeu == "Palworld" {
		// Not implemented yet
	}

	return nil
}

// Action when a Minecraft player get an advancement
func PlayerGetAdvancementAction(line string, serverID int) error {
	// Server infos
	server, err := db.GetServerById(serverID)
	if err != nil {
		return fmt.Errorf("ERROR WHILE GETTING SERVER BY ID FOR PLAYER GET ADVANCEMENT: %v", err)
	}

	// Minecraft player advancements regex
	playerAdvancementRegex := regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2})\] \[Server thread/INFO].*?: ([^\s]+) has made the advancement \[(.+?)]`)
	matches := playerAdvancementRegex.FindStringSubmatch(line)
	if len(matches) < 4 {
		return fmt.Errorf("ERROR WHILE EXTRACTING PLAYER ADVANCEMENT FOR MINECRAFT SERVER")
	}
	playerName := matches[2]
	advancement := matches[3]

	// Bot config
	botName := "mineotterBot"

	// Create embed model
	embed := models.EmbedConfig{
		Title:       advancement,
		TitleURL:    "https://fr.namemc.com/profile/" + playerName,
		Description: playerName + " a obtenu l'avancement \"" + advancement + "\" sur " + server.Nom + " !",
		Color:       server.EmbedColor,
		Thumbnail:   "https://media.forgecdn.net/avatars/thumbnails/851/712/256/256/638254029686192051.png",
		Footer:      "Message venant de " + server.Nom,
		Author:      "",
		AuthorIcon:  "",
		Timestamp:   true,
	}

	// Send the Discord embed message
	err = discord.SendDiscordEmbedWithModel(config.AppConfig.Bots[botName], config.AppConfig.DiscordChannels.MinecraftChatChannelID, embed)
	if err != nil {
		return fmt.Errorf("ERROR WHILE SENDING DISCORD EMBED: %v", err)
	}

	return nil
}

// Action when a Minecraft player dies
func PlayerDeathAction(deathMessage string, playername string, serverID int) error {
	// Server infos
	server, err := db.GetServerById(serverID)
	if err != nil {
		return fmt.Errorf("ERROR WHILE GETTING SERVER BY ID FOR PLAYER DEATH: %v", err)
	}

	// Bot config
	botName := "mineotterBot"

	// Create embed model
	embedtwo := models.EmbedConfig{
		Title:       playername + " est mort !",
		TitleURL:    "",
		Description: deathMessage,
		Color:       server.EmbedColor,
		Thumbnail:   "",
		Footer:      "Message venant de " + server.Nom,
		Author:      "",
		AuthorIcon:  "",
		Timestamp:   true,
	}
	err = discord.SendDiscordEmbedWithModel(config.AppConfig.Bots[botName], config.AppConfig.DiscordChannels.MinecraftChatChannelID, embedtwo)
	if err != nil {
		return fmt.Errorf("ERROR WHILE SENDING DISCORD EMBED: %v", err)
	}

	return nil
}

// Define the functions for each game, here is Minecraft
func handleMinecraftPlayerLeft(line string) (string, error) {
	playerDisconnectedRegex := regexp.MustCompile(`\[\d{2}:\d{2}:\d{2}\] \[Server thread/INFO\].*?: ([^\s]+) (?:left the game|disconnected|lost connection)`)
	matches := playerDisconnectedRegex.FindStringSubmatch(line)
	if len(matches) < 2 || len(matches[1]) < 3 { // Minecraft player names are at least 3 characters long, so this filter prevents false positives
		return "", fmt.Errorf("ERROR WHILE EXTRACTING DISCONNECTED PLAYER NAME FOR MINECRAFT SERVER")
	}
	return matches[1], nil
}

// Define the functions for each game, here is Palworld
func handlePalworldPlayerLeft(line string) (string, error) {
	playerDisconnectedRegex := regexp.MustCompile(`\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] \[LOG\] (.+?) left the server`)
	matches := playerDisconnectedRegex.FindStringSubmatch(line)
	if len(matches) < 2 {
		return "", fmt.Errorf("ERROR WHILE EXTRACTING LEFT PLAYER NAME FOR PALWORLD SERVER")
	}
	return matches[1], nil
}

// Create a map for game-specific actions for joined and left actions
var gameLeaveActionsMap = map[string]func(string) (string, error){
	"Minecraft": handleMinecraftPlayerLeft,
	"Palworld":  handlePalworldPlayerLeft,
}

// Action when a player left the server
func PlayerLeftAction(line string, serverID int) error {
	// Server infos
	server, err := db.GetServerById(serverID)
	if err != nil {
		return fmt.Errorf("ERROR WHILE GETTING SERVER BY ID FOR PLAYER LEFT: %v", err)
	}

	// Get the appropriate game action function from the map for "left" action
	actionFunc, exists := gameLeaveActionsMap[server.Jeu]
	if !exists {
		return fmt.Errorf("ERROR: SERVER GAME %v IS NOT SUPPORTED", server.Jeu)
	}

	// Call the specific action function for the game
	playerName, err := actionFunc(line)
	if err != nil {
		return err
	}

	// Bot config
	var botName string
	if server.Jeu == "Minecraft" {
		botName = "mineotterBot"
	} else {
		botName = "multiloutreBot"
	}

	// Send the Discord embed message
	discord.SendDiscordEmbed(config.AppConfig.Bots[botName], config.AppConfig.DiscordChannels.MinecraftChatChannelID, playerName+" a quitté "+server.Nom, "", server.EmbedColor)

	// Log to file
	WriteToLogFile("/var/log/serversentinel/playerdisconnected.log", playerName)

	/* Let's now send the message to the secondary Server */
	// We first need to know if the server is primary or secondary
	serverToSend, serverToSendHost, serverToSendRconPort, serverToSendRconPassword, err := db.GetRconParameters(server.Type)

	// Now we can send the message to the server
	if serverToSend.Jeu == "Minecraft" {
		command := `tellraw @a {"text":"` + playerName + ` a quitté le serveur ` + server.Nom + `","color":"yellow"}`
		fmt.Println("RCON parameters:", serverToSendHost, serverToSendRconPort, serverToSendRconPassword)
		resp, err := services.SendRconToMinecraftServer(serverToSendHost, serverToSendRconPort, serverToSendRconPassword, command)
		if err != nil {
			return fmt.Errorf("ERROR WHILE SENDING RCON COMMAND TO MINECRAFT SERVER: %v", err)
		}
		fmt.Println("RCON response from Minecraft server:", resp)
	} else if serverToSend.Jeu == "Palworld" {
		// Not implemented yet
	}

	return nil
}
