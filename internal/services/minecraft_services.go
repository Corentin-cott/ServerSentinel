package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"github.com/gorcon/rcon"
)

// SendRconToMinecraftServer sends a command to a Minecraft server using RCON
func SendRconToMinecraftServer(serverAddress, rconPort, rconPassword, command string) (string, error) {
	addr := fmt.Sprintf("%s:%s", serverAddress, rconPort)
	
	client, err := rcon.Dial(addr, rconPassword)
	if err != nil {
		return "", fmt.Errorf("failed to connect to RCON server: %w", err)
	}
	defer client.Close()

	resp, err := client.Execute(command)
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %w", err)
	}

	return resp, nil
}

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

	// Format the UUID to the standard format
	playerUUID = FormatMinecraftUUID(playerUUID)

	fmt.Println("Player UUID retrieved successfully : " + playerUUID + " for player name : " + playerName)
	return playerUUID, nil
}

// GetMinecraftPlayerHeadURL gets the URL of the head of a Minecraft player by their UUID
func GetMinecraftPlayerHeadURL(playerUUID string) (string, error) {
	// Send a request to the Crafatar API to get the player head URL by their UUID
	APIUrl := "https://minotar.net/helm/" + playerUUID + "/50.png"
	fmt.Println("Getting Minecraft player head URL for player " + playerUUID + " with API URL : " + APIUrl + " ...")
	resp, err := http.Get(APIUrl)
	if err != nil {
		return "", fmt.Errorf("FAILED TO SEND REQUEST TO CRAFATAR API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK { // API returns an error
		return "", fmt.Errorf("FAILED TO GET PLAYER HEAD URL, STATUS CODE: %d", resp.StatusCode)
	}

	fmt.Println("Player head URL retrieved successfully for player " + playerUUID + " : " + APIUrl)
	return APIUrl, nil
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
