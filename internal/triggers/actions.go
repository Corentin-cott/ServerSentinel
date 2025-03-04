package triggers

// This file contains the ACTIONS functions for the triggers

import (
	"fmt"
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

// Action when a player message is detected
func PlayerMessageAction(line string, serverID int) error {
	// Player name
	playerChatRegex := regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2})\] \[Server thread/INFO](?: \[.+?/MinecraftServer])?: <(.+?)> (.+)`)
	matches := playerChatRegex.FindStringSubmatch(line)
	if len(matches) < 4 {
		return fmt.Errorf("ERROR WHILE EXTRACTING CHAT PLAYER NAME... MAYBE IT'S NOT A PLAYER MESSAGE ? LINE: %v", line)
	}
	// Server infos
	server, err := db.GetServerById(serverID)
	if err != nil {
		return fmt.Errorf("ERROR WHILE GETTING SERVER BY ID FOR PLAYER MESSAGE: %v", err)
	}
	// Action
	playerName := matches[2]
	message := matches[3]

	playerUUID, err := services.GetMinecraftPlayerUUID(playerName) // Get player UUID
	if err != nil {
		return fmt.Errorf("ERROR WHILE GETTING PLAYER UUID: %v", err)
	}

	playerHeadURL, err := services.GetMinecraftPlayerHeadURL(playerUUID) // Get player head URL
	if err != nil {
		return fmt.Errorf("ERROR WHILE GETTING PLAYER HEAD URL: %v", err)
	}

	embed := models.EmbedConfig{
		Title:       playerName,
		TitleURL:    "https://fr.namemc.com/profile/" + playerUUID,
		Description: message,
		Color:       server.EmbedColor,
		Thumbnail:   playerHeadURL,
		MainImage:   "",
		Footer:      "Message venant de " + server.Nom,
		Author:      "",
		AuthorIcon:  "",
		Timestamp:   false,
	}
	err = discord.SendDiscordEmbedWithModel(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.MinecraftChatChannelID, embed)
	if err != nil {
		return fmt.Errorf("ERROR WHILE SENDING DISCORD EMBED: %v", err)
	}
	return nil
}

// Action when a player joined the server
func PlayerJoinedAction(line string, serverID int) error {
	// Player name
	playerJoinedRegex := regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2})\] \[Server thread/INFO](?: \[.+?/MinecraftServer])?: (.+) joined the game`)
	matches := playerJoinedRegex.FindStringSubmatch(line)
	if len(matches) < 3 {
		return fmt.Errorf("ERROR WHILE EXTRACTING JOINED PLAYER NAME")
	}
	// Server infos
	server, err := db.GetServerById(serverID)
	if err != nil {
		return fmt.Errorf("ERROR WHILE GETTING SERVER BY ID FOR MINECRAFT SERVER STOPPED: %v", err)
	}
	// Action
	discord.SendDiscordEmbed(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.MinecraftChatChannelID, matches[2]+" à rejoint "+server.Nom, "", server.EmbedColor)
	playerID, err := db.CheckAndInsertPlayerWithPlayerName(matches[2], 1, "now")
	if err != nil {
		return fmt.Errorf("ERROR WHILE CHECKING OR INSERTING PLAYER: %v", err)
	}
	err = db.SaveConnectionLog(playerID, serverID)
	if err != nil {
		return fmt.Errorf("ERROR WHILE SAVING CONNECTION LOG: FOR PLAYER %v IN DATABASE: %v", matches[2], err)
	}
	err = db.UpdatePlayerLastConnection(playerID)
	if err != nil {
		return fmt.Errorf("ERROR WHILE UPDATING LAST CONNECTION FOR PLAYER %v IN DATABASE: %v", matches[2], err)
	}
	WriteToLogFile("/var/log/serversentinel/playerjoined.log", matches[2])
	return nil
}

// Action when a player left the server
func PlayerLeftAction(line string, serverID int) error {
	// Player name
	playerDisconnectedRegex := regexp.MustCompile(`\[\d{2}:\d{2}:\d{2}\] \[Server thread/INFO\].*?: ([^\s]+) (?:left the game|disconnected|lost connection)`)
	matches := playerDisconnectedRegex.FindStringSubmatch(line)
	if len(matches) < 2 || len(matches[1]) < 3 { // Minecraft player names are at least 3 characters long, so this filter prevents false positives
		return fmt.Errorf("ERROR WHILE EXTRACTING DISCONNECTED PLAYER NAME")
	}
	// Server infos
	server, err := db.GetServerById(serverID)
	if err != nil {
		return fmt.Errorf("ERROR WHILE GETTING SERVER BY ID FOR MINECRAFT SERVER STOPPED: %v", err)
	}
	// Action
	discord.SendDiscordEmbed(config.AppConfig.Bots["mineotterBot"], config.AppConfig.DiscordChannels.MinecraftChatChannelID, matches[1]+" à quitté "+server.Nom, "", server.EmbedColor)
	WriteToLogFile("/var/log/serversentinel/playerdisconnected.log", matches[1])
	return nil
}
