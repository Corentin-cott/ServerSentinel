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

// DiscordWebhookConfig is a struct that contains the configuration for a Discord webhook
type DiscordWebhookConfig struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
}

// EmbedConfig is a struct that contains the configuration for discord embeds
type EmbedConfig struct {
	Title       string `json:"title"`
	TitleURL    string `json:"titleURL"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Thumbnail   string `json:"thumbnail"`
	MainImage   string `json:"mainImage"`
	Footer      string `json:"footer"`
	Author      string `json:"author"`
	AuthorIcon  string `json:"authorIcon"`
	Timestamp   bool   `json:"timestamp"`
}

// DiscordChannels is a struct that contains the configuration for the Discord channels
type DiscordChannels struct {
	BotAdminChannelID      string `json:"botAdminChannelID"`
	ServerStatusChannelID  string `json:"serverStatusChannelID"`
	MinecraftChatChannelID string `json:"minecraftChatChannelID"`
	PalworldChatChannelID  string `json:"palworldChatChannelID"`
}

// PeriodicEventsConfig is a struct that contains the configuration for the periodic events
type PeriodicEventsConfig struct {
	ServersCheckEnabled   bool `json:"serversCheckEnabled"`
	MinecraftStatsEnabled bool `json:"minecraftStatsEnabled"`
}

// Type Player is a struct that represents a player in the database
type Player struct {
	ID            int
	UtilisateurID int
	Jeu           string
	CompteID      string
	PremiereCo    string
	DerniereCo    string
}

// Type Server is a struct that represents a server in the database
type Server struct {
	ID          int    `json:"id"`
	Nom         string `json:"nom"`
	Jeu         string `json:"jeu"`
	Version     string `json:"version"`
	Modpack     string `json:"modpack"`
	ModpackURL  string `json:"modpack_url"`
	NomMonde    string `json:"nom_monde"`
	EmbedColor  string `json:"embed_color"`
	Contenaire  string `json:"contenaire"`
	Description string `json:"description"`
	Actif       bool   `json:"actif"`
	Global      bool   `json:"global"`
	Type        string `json:"type"`
	Image       string `json:"image"`
}

// Type MinecraftPlayer is a struct that represents a player in the database (very specific, i know)
type MinecraftPlayerGameStatistics struct {
	ID               int
	ServerID         int
	PlayerID         int
	TimePlayed       int
	Deaths           int
	Kills            int
	PlayerKills      int
	MobsKilled       map[string]int
	BlocksDestroyed  int
	BlocksPlaced     int
	TotalDistance    int
	DistanceByFoot   int
	DistanceByElytra int
	DistanceByFlight int
	ItemsCrafted     map[string]int
	ItemsBroken      map[string]int
	Achievements     map[string]bool
	LastRecordedTime string
}

// Trigger is a struct that represents a trigger
type Trigger struct {
	Name      string            // Trigger name
	Condition func(string) bool // Condition of the trigger
	Action    func(string, int) // Function to execute when the condition is met
	ServerID  int               // ID of the server
}
