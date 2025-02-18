package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// DatabaseConfig is a struct that contains the configuration for the database
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// BotConfig is a struct that contains the configuration for the bot
type BotConfig struct {
	BotToken         string `json:"botToken"`
	DiscordChannelID string `json:"discordChannelID"`
}

// Config is a struct that contains every configuration needed for ServeurSentinel
type Config struct {
	Bot               BotConfig      `json:"bot"`
	DB                DatabaseConfig `json:"db"`
	LogPath           string         `json:"logPath"`
	PeriodicEventsMin int            `json:"periodicEventsMin"`
}

var AppConfig Config

// LoadConfig loads the configuration from a JSON file
func LoadConfig(configPath string) error {
	file, err := os.Open(configPath)
	if err != nil {
		return fmt.Errorf("error opening configuration file: %v", err)
	}
	defer file.Close()

	// Decode the JSON file
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&AppConfig); err != nil {
		return fmt.Errorf("error decoding configuration: %v", err)
	}

	fmt.Printf("✔ Configuration loaded successfully\n")
	return nil
}
