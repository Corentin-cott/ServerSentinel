package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Corentin-cott/ServeurSentinel/internal/models"
)

// Config is a struct that contains every configuration needed for ServeurSentinel
type Config struct {
	Bots              map[string]models.BotConfig `json:"bots"`
	DiscordChannels   models.DiscordChannels      `json:"discordChannels"`
	DB                models.DatabaseConfig       `json:"db"`
	PeriodicEvents    models.PeriodicEventsConfig `json:"periodicEvents"`
	LogPath           string                      `json:"logPath"`
	PeriodicEventsMin int                         `json:"periodicEventsMin"`
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
