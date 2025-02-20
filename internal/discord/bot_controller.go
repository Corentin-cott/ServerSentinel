package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Corentin-cott/ServeurSentinel/config"
)

// SendDiscordMessage() sends a message to a Discord channel
func SendDiscordMessage(message string) error {
	// Get parameters from the configuration
	botToken := config.AppConfig.Bot.BotToken
	channelID := config.AppConfig.Bot.DiscordChannelID

	// Checks if one of the parameters is missing
	switch {
	case botToken == "" && channelID == "":
		return fmt.Errorf("ERROR: BOT TOKEN AND CHANNEL ID NOT SET")
	case botToken == "":
		return fmt.Errorf("ERROR: BOT TOKEN NOT SET")
	case channelID == "":
		return fmt.Errorf("ERROR: CHANNEL ID NOT SET")
	}

	// Prepare the request to the Discord API
	apiURL := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelID)
	type DiscordBotMessage struct {
		Content string `json:"content"`
	}

	// Serialize the message to JSON
	payload := DiscordBotMessage{Content: message}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("ERROR WHILE SERIALISING DISCORD MESSAGE: %v", err)
	}

	// Create the HTTP request to send the message
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("ERROR WHILE CREATING REQUEST TO DISCORD: %v", err)
	}

	// Set the headers for the request
	req.Header.Set("Authorization", "Bot "+botToken)
	req.Header.Set("Content-Type", "application/json")

	// Finally, send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ERROR WHILE SENDING MESSAGE TO DISCORD: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("ERROR WHILE SENDING MESSAGE TO DISCORD, RESPONSE STATUS: %v", resp.Status)
	} else {
		return nil
	}
}

func SendDiscordEmbed(title string, description string, color string) error {
	botToken := config.AppConfig.Bot.BotToken
	channelID := config.AppConfig.Bot.DiscordChannelID

	// Check required parameters
	switch {
	case botToken == "" && channelID == "":
		return fmt.Errorf("ERROR: BOT TOKEN AND CHANNEL ID NOT SET")
	case botToken == "":
		return fmt.Errorf("ERROR: BOT TOKEN NOT SET")
	case channelID == "":
		return fmt.Errorf("ERROR: CHANNEL ID NOT SET")
	}

	// Convert hex color to integer
	colorInt, err := strconv.ParseInt(strings.TrimPrefix(color, "#"), 16, 32)
	if err != nil {
		return fmt.Errorf("ERROR: INVALID COLOR FORMAT: %v", err)
	}

	// Create the correct payload format
	payload := map[string]interface{}{
		"content": "", // Required but can be empty
		"embeds": []map[string]interface{}{
			{
				"title":       title,
				"description": description,
				"color":       colorInt,
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("ERROR WHILE SERIALIZING DISCORD EMBED: %v", err)
	}

	// Create and send the request
	apiURL := fmt.Sprintf("https://discord.com/api/v10/channels/%s/messages", channelID)
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("ERROR WHILE CREATING REQUEST TO DISCORD: %v", err)
	}

	req.Header.Set("Authorization", "Bot "+botToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ERROR WHILE SENDING EMBED TO DISCORD: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("ERROR WHILE SENDING EMBED TO DISCORD, RESPONSE STATUS: %v, RESPONSE BODY: %s", resp.Status, string(body))
	}

	return nil
}
