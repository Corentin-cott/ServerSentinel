package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Corentin-cott/ServeurSentinel/internal/models"
)

// SendDiscordMessage() sends a message to a Discord channel
func SendDiscordMessage(bot models.BotConfig, channelID string, message string) error {
	if !bot.Activated {
		// fmt.Println("Bot " + bot.BotToken + " is not activated")
		return nil // If the bot is not activated, we don't send the message
	}

	// Checks if one of the parameters is missing
	botToken := bot.BotToken
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

func SendDiscordEmbed(bot models.BotConfig, channelID string, title string, description string, color string) error {
	if !bot.Activated {
		// fmt.Println("Bot " + bot.BotToken + " is not activated")
		return nil // If the bot is not activated, we don't send the message
	}

	// Check required parameters
	botToken := bot.BotToken
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

func SendDiscordEmbedWithModel(bot models.BotConfig, channelID string, embed models.EmbedConfig) error {
	if !bot.Activated {
		// fmt.Println("Bot " + bot.BotToken + " is not activated")
		return nil // If the bot is not activated, we don't send the message
	}

	// Check required parameters
	botToken := bot.BotToken
	switch {
	case botToken == "" && channelID == "":
		return fmt.Errorf("ERROR: BOT TOKEN AND CHANNEL ID NOT SET")
	case botToken == "":
		return fmt.Errorf("ERROR: BOT TOKEN NOT SET")
	case channelID == "":
		return fmt.Errorf("ERROR: CHANNEL ID NOT SET")
	}

	// Convert hex color to integer
	colorInt, err := strconv.ParseInt(strings.TrimPrefix(embed.Color, "#"), 16, 32)
	if err != nil {
		return fmt.Errorf("ERROR: INVALID COLOR FORMAT: %v", err)
	}

	// Timestamp
	var timestamp string
	if embed.Timestamp {
		timestamp = time.Now().UTC().Format(time.RFC3339)
	} else {
		timestamp = ""
	}

	// Create the correct payload format
	payload := map[string]interface{}{
		"content": "", // Required but can be empty
		"embeds": []map[string]interface{}{
			{
				"title":       embed.Title,
				"url":         embed.TitleURL,
				"description": embed.Description,
				"color":       colorInt,
				"thumbnail": map[string]string{
					"url": embed.Thumbnail,
				},
				"image": map[string]string{
					"url": embed.MainImage,
				},
				"footer": map[string]string{
					"text": embed.Footer,
				},
				"author": map[string]string{
					"name":     embed.Author,
					"icon_url": embed.AuthorIcon,
				},
				"timestamp": timestamp,
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
		return fmt.Errorf("ERROR WHILE SENDING EMBED TO DISCORD : %v", err)
	}
	defer resp.Body.Close()

	// Check response
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("ERROR WHILE SENDING EMBED TO DISCORD, RESPONSE STATUS: %v, RESPONSE BODY: %s", resp.Status, string(body))
	}

	return nil
}
