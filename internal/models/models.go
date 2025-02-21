package models

// DatabaseConfig is a struct that contains the configuration for the database
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// BotConfig is a struct that contains the configuration for a bot
type BotConfig struct {
	Activated bool   `json:"activated"`
	BotToken  string `json:"botToken"`
}

// DiscordChannels is a struct that contains the configuration for the Discord channels
type DiscordChannels struct {
	BotAdminChannelID      string `json:"botAdminChannelID"`
	ServerStatusChannelID  string `json:"serverStatusChannelID"`
	MinecraftChatChannelID string `json:"minecraftChatChannelID"`
	PalworldChatChannelID  string `json:"palworldChatChannelID"`
}

// Type Server is a struct that represents a server in the database
type Server struct {
	ID          int
	Nom         string
	Jeu         string
	Version     string
	Modpack     string
	ModpackURL  string
	NomMonde    string
	EmbedColor  string
	PathServ    string
	StartScript string
	Actif       bool
	Global      bool
}
