package db

import (
	"database/sql"
	"fmt"

	"github.com/Corentin-cott/ServeurSentinel/config"
	"github.com/Corentin-cott/ServeurSentinel/internal/models"
	"github.com/Corentin-cott/ServeurSentinel/internal/services"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

/* -----------------------------------------------------
Table serveurs {
    id INT [pk, increment]
    nom VARCHAR(255) [not null]
    jeu VARCHAR(255) [not null]
    version VARCHAR(20) [not null]
    modpack VARCHAR(255) [default: 'Vanilla']
    modpack_url VARCHAR(255) [null]
    nom_monde VARCHAR(255) [default: 'world']
    embed_color VARCHAR(7) [default: '#000000']
    path_serv TEXT [not null]
    start_script VARCHAR(255) [not null]
    actif BOOLEAN [default: false, not null]
    global BOOLEAN [default: true, not null]
}

Table serveurs_parameters {
    id_serv_primaire INT [ref: > serveurs.id, not null]
    id_serv_secondaire INT [ref: > serveurs.id, not null]
    host_primaire VARCHAR(255) [not null]
    host_secondaire VARCHAR(255) [not null]
    rcon_password VARCHAR(255) [not null]
}
----------------------------------------------------- */

// ConnectToDatabase initialises the connection to the MySQL database
func ConnectToDatabase() error {
	// Load the database configuration
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.AppConfig.DB.User,
		config.AppConfig.DB.Password,
		config.AppConfig.DB.Host,
		config.AppConfig.DB.Port,
		config.AppConfig.DB.Name,
	)

	// Try to connect to the database
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("ERROR OPENING DATABASE: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ERROR WHILE PINGING DATABASE: %v", err)
	}

	fmt.Println("✔ Successfully connected to the database.")
	return nil
}

// SaveConnectionLog saves a connection log for a player
func SaveConnectionLog(playerName string, serverID int) error {
	// Check if the player exists and insert it if it doesn't
	_, err := CheckAndInsertPlayer(playerName, serverID)
	if err != nil {
		return fmt.Errorf("FAILED TO CHECK OR INSERT PLAYER: %v", err)
	}

	// Get player account ID with the player name
	playerAcountID, err := GetPlayerAccountIdByPlayerName(playerName, "Minecraft")
	if err != nil {
		return fmt.Errorf("FAILED TO GET PLAYER ACCOUNT ID: %v", err)
	}

	// Get player ID with the player account ID
	playerID, err := GetPlayerIdByAccountId(playerAcountID)
	if err != nil {
		return fmt.Errorf("FAILED TO GET PLAYER ID: %v", err)
	} else if playerID == -1 {
		return fmt.Errorf("PLAYER ID NOT FOUND")
	}

	// Update the last connection date of the player
	err = UpdatePlayerLastConnection(playerID)
	if err != nil {
		return fmt.Errorf("FAILED TO UPDATE LAST CONNECTION: %v", err)
	}

	// Insert log in the database
	insertQuery := `INSERT INTO joueurs_connections_log (serveur_id, joueur_id, date) VALUES (?, ?, NOW())`
	fmt.Println("Inserting connection log for player", playerID)
	_, err = db.Exec(insertQuery, serverID, playerID)
	if err != nil {
		return fmt.Errorf("FAILED TO INSERT CONNECTION LOG: %v", err)
	}

	fmt.Println("Connection log successfully saved.")
	return nil
}

// CheckAndInsertPlayer checks if a player exists in the database and inserts it if it doesn't
func CheckAndInsertPlayer(playerName string, serverID int) (int, error) {
	// Get server game
	jeu, err := GetServerGameById(serverID)
	if err != nil {
		return -1, fmt.Errorf("FAILED TO GET SERVER GAME: %v", err)
	}

	// Get player account ID
	playerAcountID, err := GetPlayerAccountIdByPlayerName(playerName, jeu)
	if err != nil {
		return -1, fmt.Errorf("FAILED TO GET PLAYER ACCOUNT ID: %v", err)
	}

	// Check if the player already exists
	fmt.Println("Checking if player exists...")
	playerID, _ := GetPlayerIdByAccountId(playerAcountID)
	if playerID != -1 {
		fmt.Printf("Player already exists with ID (this is not a problem) %d\n", playerID)
		return playerID, nil // Player already exists, return its ID
	}

	// If the player does not exist, insert it
	fmt.Println("Player does not exist. Inserting new player...")
	insertQuery := "INSERT INTO joueurs (utilisateur_id, jeu, compte_id, premiere_co, derniere_co) VALUES (NULL, ?, ?, NOW(), NOW())"
	_, err = db.Exec(insertQuery, jeu, playerAcountID)
	if err != nil {
		return -1, fmt.Errorf("FAILED TO INSERT PLAYER: %v", err)
	}
	fmt.Println("Player successfully inserted !")

	// Return the player ID of the newly inserted player
	playerID, err = GetPlayerIdByAccountId(playerAcountID)
	if err != nil {
		return -1, fmt.Errorf("FAILED TO GET PLAYER ID: %v", err)
	} else if playerID == -1 {
		return -1, fmt.Errorf("PLAYER ID NOT FOUND")
	}

	return playerID, nil
}

// UpdatePlayerLastConnection updates the last connection date of a player
func UpdatePlayerLastConnection(playerID int) error {
	fmt.Println("Updating last connection for player ID", playerID)
	updateQuery := "UPDATE joueurs SET derniere_co = NOW() WHERE id = ?"
	_, err := db.Exec(updateQuery, playerID)
	if err != nil {
		return fmt.Errorf("FAILED TO UPDATE LAST CONNECTION: %v", err)
	}

	return nil
}

// Getter to get the primary server
func GetPrimaryServerId() int {
	query := "SELECT id_serv_primaire FROM serveurs_parameters"
	var serverID int

	err := db.QueryRow(query).Scan(&serverID)
	if err != nil {
		fmt.Println("FAILED TO GET PRIMARY SERVER:", err)
		return -1
	}

	return serverID
}

// Getter to get the secondary server
func GetSecondaryServerId() int {
	query := "SELECT id_serv_secondaire FROM serveurs_parameters"
	var serverID int

	err := db.QueryRow(query).Scan(&serverID)
	if err != nil {
		fmt.Println("FAILED TO GET SECONDARY SERVER:", err)
		return -1
	}

	return serverID
}

// Getter to get all the server informations
func GetServerById(serverID int) (models.Server, error) {
	query := "SELECT * FROM serveurs WHERE id = ?"
	var serv models.Server

	err := db.QueryRow(query, serverID).Scan(&serv.ID, &serv.Nom, &serv.Jeu, &serv.Version, &serv.Modpack, &serv.ModpackURL, &serv.NomMonde, &serv.EmbedColor, &serv.PathServ, &serv.StartScript, &serv.Actif, &serv.Global)
	if err != nil {
		if err == sql.ErrNoRows {
			return serv, fmt.Errorf("SERVER NOT FOUND: %d", serverID)
		}
		return serv, fmt.Errorf("FAILED TO GET SERVER: %v", err)
	}

	return serv, nil
}

// Getter to get the server by the server name
func GetServerByName(serverName string) (models.Server, error) {
	query := "SELECT * FROM serveurs WHERE nom = ?"
	var serv models.Server

	err := db.QueryRow(query, serverName).Scan(&serv.ID, &serv.Nom, &serv.Jeu, &serv.Version, &serv.Modpack, &serv.ModpackURL, &serv.NomMonde, &serv.EmbedColor, &serv.PathServ, &serv.StartScript, &serv.Actif, &serv.Global)
	if err != nil {
		if err == sql.ErrNoRows {
			return serv, fmt.Errorf("SERVER NOT FOUND: %s", serverName)
		}
		return serv, fmt.Errorf("FAILED TO GET SERVER: %v", err)
	}

	return serv, nil
}

// Getter to get the server game by the server ID
func GetServerGameById(serverID int) (string, error) {
	query := "SELECT jeu FROM serveurs WHERE id = ?"
	var jeu string

	err := db.QueryRow(query, serverID).Scan(&jeu)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("GAME NOT FOUND FOR SERVER ID: %d", serverID)
		}
		return "", fmt.Errorf("FAILED TO GET SERVER GAME: %v", err)
	}

	return jeu, nil
}

// Getter to get the player ID by the account ID
func GetPlayerIdByAccountId(accountId any) (int, error) {
	query := "SELECT id FROM joueurs WHERE compte_id = ?"
	var playerID int

	err := db.QueryRow(query, accountId).Scan(&playerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return -1, nil
		}
		fmt.Println("FAILED TO GET PLAYER ID:", err)
		return -1, fmt.Errorf("FAILED TO GET PLAYER ID: %v", err)
	}

	strPlayerID := fmt.Sprintf("%d", playerID)
	fmt.Println("Player ID retrieved successfully : "+strPlayerID+" for account ID : ", accountId)
	return playerID, nil
}

// Getter to get the player account ID by the player name
func GetPlayerAccountIdByPlayerName(playerName string, jeu string) (string, error) {
	if jeu == "" {
		return "", fmt.Errorf("GAME NOT FOUND")
	}

	switch jeu {
	case "Minecraft":
		return services.GetMinecraftPlayerUUID(playerName)
	default:
		return "", fmt.Errorf("UNKNOWN GAME: %s", jeu)
	}
}
