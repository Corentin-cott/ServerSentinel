package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"

	"github.com/Corentin-cott/ServeurSentinel/internal/models"
)

// GetMinecraftPlayerUUID gets the UUID of a Minecraft player by their username
func GetMinecraftPlayerUUID(playerName string) (string, error) {
	// Send a request to the Mojang API to get the player UUID by their username
	APIUrl := "https://api.mojang.com/users/profiles/minecraft/" + playerName
	fmt.Println("Getting Minecraft player UUID for player " + playerName + " with API URL : " + APIUrl + " ...")
	resp, err := http.Get(APIUrl)
	if err != nil {
		return "", fmt.Errorf("FAILED TO SEND REQUEST TO MOJANG API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK { // API returns an error
		return "", fmt.Errorf("FAILED TO GET PLAYER UUID, STATUS CODE: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil { // Failed to read response body
		return "", fmt.Errorf("FAILED TO READ RESPONSE BODY: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil { // Failed to parse JSON response
		return "", fmt.Errorf("FAILED TO READ JSON RESPONSE: %v", err)
	}

	playerUUID, ok := result["id"].(string)
	if !ok || playerUUID == "" { // Failed to find player UUID in response
		return "", fmt.Errorf("FAILED TO GET PLAYER UUID: %v", result)
	}

	fmt.Println("Player UUID retrieved successfully : " + playerUUID + " for player name : " + playerName)
	return playerUUID, nil
}

// GetMinecraftPlayerServerSave gets a list of the Minecraft player UUIDs inside a server directory
func GetMinecraftPlayerServerUUIDSaves(server models.Server) ([]string, error) {
	fmt.Println("Getting Minecraft player saves for server " + server.Nom + " inside directory " + server.PathServ + server.NomMonde + "/stats ...")

	if server.Jeu != "Minecraft" {
		return nil, fmt.Errorf("%s IS NOT A MINECRAFT SERVER", server.Nom)
	}

	// Get the file path
	serverSavesPath := server.PathServ + server.NomMonde + "/stats"

	// Check if directory exists
	if _, err := os.Stat(serverSavesPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("PLAYER DATA DIRECTORY NOT FOUND")
	}

	// Read directory
	files, err := os.ReadDir(serverSavesPath)
	if err != nil {
		return nil, fmt.Errorf("FAILED TO READ PLAYER DATA DIRECTORY: %v", err)
	}

	// Extract player UUIDs from file names that are .json files
	playerUUIDs := make([]string, 0)
	for _, file := range files {
		if file.IsDir() || file.Name()[len(file.Name())-5:] != ".json" {
			continue
		}
		playerUUIDs = append(playerUUIDs, file.Name()[:len(file.Name())-5])
	}

	nbPlayerFound := len(playerUUIDs)

	fmt.Println("Player saves retrieved successfully for server " + server.Nom + " : " + fmt.Sprint(nbPlayerFound) + " players found.")
	return playerUUIDs, nil
}

// GetMinecraftPlayerGameStatistics gets the game statistics of a Minecraft player with his server save
func GetMinecraftPlayerGameStatistics(playerID int, playerUUID string, server models.Server) (int, string, models.MinecraftPlayerGameStatistics, error) {
	fmt.Println("Getting Minecraft statistics for player " + playerUUID + " in server " + server.Nom + "...")

	if server.Jeu != "Minecraft" {
		return 0, "", models.MinecraftPlayerGameStatistics{}, fmt.Errorf("%s IS NOT A MINECRAFT SERVER", server.Nom)
	}

	// Get the file path
	playerStatsFile := server.PathServ + server.NomMonde + "/stats/" + FormatMinecraftUUID(playerUUID) + ".json"

	// Check if file exists
	if _, err := os.Stat(playerStatsFile); os.IsNotExist(err) {
		return 0, "", models.MinecraftPlayerGameStatistics{}, fmt.Errorf("PLAYER STATISTICS FILE NOT FOUND")
	}

	// Read file
	playerStatsJSON, err := os.ReadFile(playerStatsFile)
	if err != nil {
		return 0, "", models.MinecraftPlayerGameStatistics{}, fmt.Errorf("FAILED TO READ PLAYER STATISTICS FILE: %v", err)
	}

	// Define a temporary struct to match Minecraft's JSON structure
	var rawStats struct {
		Stats struct {
			Custom  map[string]int `json:"minecraft:custom"`
			Mined   map[string]int `json:"minecraft:mined"`
			Killed  map[string]int `json:"minecraft:killed"`
			Crafted map[string]int `json:"minecraft:crafted"`
			Used    map[string]int `json:"minecraft:used"`
			Broken  map[string]int `json:"minecraft:broken"`
		} `json:"stats"`
	}

	// Unmarshal JSON into temporary struct
	if err := json.Unmarshal(playerStatsJSON, &rawStats); err != nil {
		return 0, "", models.MinecraftPlayerGameStatistics{}, fmt.Errorf("FAILED TO UNMARSHAL PLAYER STATISTICS JSON: %v", err)
	}

	// Depending on the server version, the stats can be in different formats/keys
	var play_time int
	if rawStats.Stats.Custom["minecraft:play_time"] == 0 {
		play_time = rawStats.Stats.Custom["minecraft:play_one_minute"]
	} else {
		play_time = rawStats.Stats.Custom["minecraft:play_time"]
	}

	// Map extracted data to your model
	playerStats := models.MinecraftPlayerGameStatistics{
		TimePlayed:       play_time,
		Deaths:           rawStats.Stats.Custom["minecraft:deaths"],
		Kills:            rawStats.Stats.Killed["minecraft:player"],
		PlayerKills:      rawStats.Stats.Custom["minecraft:player_kills"],
		BlocksDestroyed:  sumValues(rawStats.Stats.Mined),
		BlocksPlaced:     sumValues(rawStats.Stats.Used),
		TotalDistance:    rawStats.Stats.Custom["minecraft:walk_one_cm"],
		DistanceByFoot:   rawStats.Stats.Custom["minecraft:walk_one_cm"],
		DistanceByElytra: rawStats.Stats.Custom["minecraft:aviate_one_cm"],
		DistanceByFlight: rawStats.Stats.Custom["minecraft:fly_one_cm"],
		ItemsCrafted:     rawStats.Stats.Crafted,
		ItemsBroken:      rawStats.Stats.Broken,
		MobsKilled:       rawStats.Stats.Killed,
		Achievements:     make(map[string]bool),
	}

	fmt.Println("Stats registered for player " + playerUUID + " in server " + server.Nom)
	return playerID, playerUUID, playerStats, nil
}

func FormatMinecraftUUID(uuid string) string {
	if len(uuid) != 32 {
		return uuid // Return as is if not in expected format
	}
	return uuid[:8] + "-" + uuid[8:12] + "-" + uuid[12:16] + "-" + uuid[16:20] + "-" + uuid[20:]
}

func IsValidMinecraftUUID(uuid string) (bool, error) {
	if len(uuid) != 36 {
		return false, fmt.Errorf("INVALID UUID LENGTH, EXPECTED 36 CHARACTERS, FOUND %d", len(uuid))
	}

	regex := `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`
	uuidBool := regexp.MustCompile(regex).MatchString(uuid)

	if !uuidBool {
		return false, fmt.Errorf("INVALID UUID FORMAT")
	}
	return uuidBool, nil
}

func sumValues(m map[string]int) int {
	total := 0
	for _, value := range m {
		total += value
	}
	return total
}
