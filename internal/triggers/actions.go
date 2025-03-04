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
	// Server infos
	server, err := db.GetServerById(serverID)
	if err != nil {
		return fmt.Errorf("ERROR WHILE GETTING SERVER BY ID FOR PLAYER MESSAGE: %v", err)
	}
	// Player name & message
	var playerName string
	var message string
	if server.Jeu == "Minecraft" {
		playerChatRegex := regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2})\] \[Server thread/INFO](?: \[.+?/MinecraftServer])?: <(.+?)> (.+)`)
		matches := playerChatRegex.FindStringSubmatch(line)
		if len(matches) < 4 {
			return fmt.Errorf("ERROR WHILE EXTRACTING CHAT PLAYER NAME... MAYBE IT'S NOT A PLAYER MESSAGE ? LINE: %v", line)
		}
		playerName = matches[2]
		message = matches[3]
	} else if server.Jeu == "Palworld" {
		playerChatRegex := regexp.MustCompile(`\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] \[CHAT\] <(.+?)> (.+)`)
		matches := playerChatRegex.FindStringSubmatch(line)
		if len(matches) < 3 {
			return fmt.Errorf("ERROR WHILE EXTRACTING CHAT PLAYER NAME FOR PALWORLD... MAYBE IT'S NOT A PLAYER MESSAGE? LINE: %v", line)
		}
		playerName = matches[1]
		message = matches[2]
	} else {
		return fmt.Errorf("ERROR WHILE GETTING PLAYER NAME: SERVER GAME IS NOT SUPPORTED")
	}
	// Get player UUID and player head URL for Minecraft servers
	var playerUUID string
	var playerHeadURL string
	var titleURL string
	if server.Jeu == "Minecraft" {
		playerUUID, err = services.GetMinecraftPlayerUUID(playerName) // Get player UUID
		if err != nil {
			return fmt.Errorf("ERROR WHILE GETTING PLAYER UUID: %v", err)
		}

		playerHeadURL, err = services.GetMinecraftPlayerHeadURL(playerUUID) // Get player head URL
		if err != nil {
			return fmt.Errorf("ERROR WHILE GETTING PLAYER HEAD URL: %v", err)
		}

		titleURL = "https://fr.namemc.com/profile/" + playerUUID
	} else {
		playerUUID = ""
		playerHeadURL = ""
		titleURL = ""
	}
	// Bot config
	var botName string
	if server.Jeu == "Minecraft" {
		botName = "mineotterBot"
	} else {
		botName = "multiloutreBot"
	}
	// Action
	embed := models.EmbedConfig{
		Title:       playerName,
		TitleURL:    titleURL,
		Description: message,
		Color:       server.EmbedColor,
		Thumbnail:   playerHeadURL,
		MainImage:   "",
		Footer:      "Message venant de " + server.Nom,
		Author:      "",
		AuthorIcon:  "",
		Timestamp:   false,
	}
	err = discord.SendDiscordEmbedWithModel(config.AppConfig.Bots[botName], config.AppConfig.DiscordChannels.MinecraftChatChannelID, embed)
	if err != nil {
		return fmt.Errorf("ERROR WHILE SENDING DISCORD EMBED: %v", err)
	}
	return nil
}

// Action when a player joined the server
func PlayerJoinedAction(line string, serverID int) error {
	// Server infos
	server, err := db.GetServerById(serverID)
	if err != nil {
		return fmt.Errorf("ERROR WHILE GETTING SERVER BY ID FOR MINECRAFT SERVER STOPPED: %v", err)
	}
	// Player name
	var playername string
	if server.Jeu == "Minecraft" {
		playerJoinedRegex := regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2})\] \[Server thread/INFO](?: \[.+?/MinecraftServer])?: (.+) joined the game`)
		matches := playerJoinedRegex.FindStringSubmatch(line)
		if len(matches) < 3 {
			return fmt.Errorf("ERROR WHILE EXTRACTING JOINED PLAYER NAME FOR MINECRAFT SERVER")
		}
		playername = matches[2]
	} else if server.Jeu == "Palworld" {
		playerJoinedRegex := regexp.MustCompile(`\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] \[LOG\] (.+?) \d{1,3}(?:\.\d{1,3}){3} connected the server`)
		matches := playerJoinedRegex.FindStringSubmatch(line)
		if len(matches) < 2 {
			return fmt.Errorf("ERROR WHILE EXTRACTING JOINED PLAYER NAME FOR PALWORLD SERVER")
		}
		playername = matches[1]
	} else {
		return fmt.Errorf("ERROR WHILE GETTING PLAYER NAME: SERVER GAME IS NOT SUPPORTED")
	}
	// Bot config
	var botName string
	if server.Jeu == "Minecraft" {
		botName = "mineotterBot"
	} else {
		botName = "multiloutreBot"
	}
	// Action
	discord.SendDiscordEmbed(config.AppConfig.Bots[botName], config.AppConfig.DiscordChannels.MinecraftChatChannelID, playername+" à rejoint "+server.Nom, "", server.EmbedColor)
	playerID, err := db.CheckAndInsertPlayerWithPlayerName(playername, 1, "now")
	if err != nil {
		return fmt.Errorf("ERROR WHILE CHECKING OR INSERTING PLAYER: %v", err)
	}
	err = db.SaveConnectionLog(playerID, serverID)
	if err != nil {
		return fmt.Errorf("ERROR WHILE SAVING CONNECTION LOG: FOR PLAYER %v IN DATABASE: %v", playername, err)
	}
	err = db.UpdatePlayerLastConnection(playerID)
	if err != nil {
		return fmt.Errorf("ERROR WHILE UPDATING LAST CONNECTION FOR PLAYER %v IN DATABASE: %v", playername, err)
	}
	WriteToLogFile("/var/log/serversentinel/playerjoined.log", playername)
	return nil
}

// Action when a player left the server
func PlayerLeftAction(line string, serverID int) error {
	// Server infos
	server, err := db.GetServerById(serverID)
	if err != nil {
		return fmt.Errorf("ERROR WHILE GETTING SERVER BY ID FOR MINECRAFT SERVER STOPPED: %v", err)
	}
	// Player name
	var playername string
	if server.Jeu == "Minecraft" {
		playerDisconnectedRegex := regexp.MustCompile(`\[\d{2}:\d{2}:\d{2}\] \[Server thread/INFO\].*?: ([^\s]+) (?:left the game|disconnected|lost connection)`)
		matches := playerDisconnectedRegex.FindStringSubmatch(line)
		if len(matches) < 2 || len(matches[1]) < 3 { // Minecraft player names are at least 3 characters long, so this filter prevents false positives
			return fmt.Errorf("ERROR WHILE EXTRACTING DISCONNECTED PLAYER NAME FOR MINECRAFT SERVER")
		}
		playername = matches[1]
	} else if server.Jeu == "Palworld" {
		playerDisconnectedRegex := regexp.MustCompile(`\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] \[LOG\] (.+?) left the server`)
		matches := playerDisconnectedRegex.FindStringSubmatch(line)
		if len(matches) < 2 {
			return fmt.Errorf("ERROR WHILE EXTRACTING JOINED PLAYER NAME FOR PALWORLD SERVER")
		}
		playername = matches[1]
	} else {
		return fmt.Errorf("ERROR WHILE GETTING PLAYER NAME: SERVER GAME IS NOT SUPPORTED")
	}
	// Bot config
	var botName string
	if server.Jeu == "Minecraft" {
		botName = "mineotterBot"
	} else {
		botName = "multiloutreBot"
	}
	// Action
	discord.SendDiscordEmbed(config.AppConfig.Bots[botName], config.AppConfig.DiscordChannels.MinecraftChatChannelID, playername+" à quitté "+server.Nom, "", server.EmbedColor)
	WriteToLogFile("/var/log/serversentinel/playerdisconnected.log", playername)
	return nil
}
