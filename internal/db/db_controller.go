package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Corentin-cott/ServeurSentinel/config"
	"github.com/Corentin-cott/ServeurSentinel/internal/models"
	"github.com/Corentin-cott/ServeurSentinel/internal/services"
	_ "github.com/go-sql-driver/mysql"
)

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

	fmt.Println("Here is the conexion string : ", dsn)

	// Try to connect to the database
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("ERROR OPENING DATABASE: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("ERROR WHILE PINGING DATABASE WITH CONNECTION STRING: (%v) ! ERROR: %v", dsn, err)
	}

	fmt.Println("✔ Successfully connected to the database.")
	return nil
}

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
----------------------------------------------------- */

// GetAllServers returns all the servers from the database
func GetAllServers() ([]models.Server, error) {
	query := "SELECT * FROM serveurs"
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("FAILED TO GET SERVERS: %v", err)
	}
	defer rows.Close()

	var servers []models.Server
	for rows.Next() {
		var serv models.Server
		if err := rows.Scan(&serv.ID, &serv.Nom, &serv.Jeu, &serv.Version, &serv.Modpack, &serv.ModpackURL, &serv.NomMonde, &serv.EmbedColor, &serv.PathServ, &serv.StartScript, &serv.Actif, &serv.Global); err != nil {
			return nil, fmt.Errorf("FAILED TO SCAN SERVER: %v", err)
		}
		servers = append(servers, serv)
	}

	return servers, nil
}

// GetAllMineCraftServers returns all the Minecraft servers from the database
func GetAllMinecraftServers() ([]models.Server, error) {
	query := "SELECT * FROM serveurs WHERE jeu = 'Minecraft'"
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("FAILED TO GET MINECRAFT SERVERS: %v", err)
	}
	defer rows.Close()

	var servers []models.Server
	for rows.Next() {
		var serv models.Server
		if err := rows.Scan(&serv.ID, &serv.Nom, &serv.Jeu, &serv.Version, &serv.Modpack, &serv.ModpackURL, &serv.NomMonde, &serv.EmbedColor, &serv.PathServ, &serv.StartScript, &serv.Actif, &serv.Global); err != nil {
			return nil, fmt.Errorf("FAILED TO SCAN MINECRAFT SERVER: %v", err)
		}
		servers = append(servers, serv)
	}

	return servers, nil
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

// Setter to set the primary server
func SetPrimaryServerId(serverID int) error {
	query := "UPDATE serveurs_parameters SET id_serv_primaire = ?"
	_, err := db.Exec(query, serverID)
	if err != nil {
		return fmt.Errorf("FAILED TO SET PRIMARY SERVER: %v", err)
	}

	return nil
}

// Setter to set the secondary server
func SetSecondaryServerId(serverID int) error {
	query := "UPDATE serveurs_parameters SET id_serv_secondaire = ?"
	_, err := db.Exec(query, serverID)
	if err != nil {
		return fmt.Errorf("FAILED TO SET SECONDARY SERVER: %v", err)
	}

	return nil
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

// Getter to get the server name by the server id
func GetServerNameById(serverID int) (string, error) {
	query := "SELECT nom FROM serveurs WHERE id = ?"
	var serverName string

	err := db.QueryRow(query, serverID).Scan(&serverName)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("SERVER NOT FOUND: %d", serverID)
		}
		return "", fmt.Errorf("FAILED TO GET SERVER NAME: %v", err)
	}

	return serverName, nil
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

// Getter to get the server color by the server id
func GetServerColorByName(serverName string) (string, error) {
	query := "SELECT embed_color FROM serveurs WHERE nom = ?"
	var serverColor string

	err := db.QueryRow(query, serverName).Scan(&serverColor)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("SERVER NOT FOUND: %s", serverName)
		}
		return "", fmt.Errorf("FAILED TO GET SERVER COLOR: %v", err)
	}

	return serverColor, nil
}

/* -----------------------------------------------------
Table joueurs_connections_log {
    id INT [pk, increment]
    serveur_id INT [ref: > serveurs.id, not null]
    joueurs_id INT [ref: > joueurs.id, null]
    date DATETIME
}
----------------------------------------------------- */

// SaveConnectionLog saves a connection log for a player
func SaveConnectionLog(playerID int, serverID int) error {
	query := "INSERT INTO joueurs_connections_log (serveur_id, joueur_id, date) VALUES (?, ?, ?)"
	_, err := db.Exec(query, serverID, playerID, GetGoodDatetime())
	if err != nil {
		return fmt.Errorf("FAILED TO SAVE CONNECTION LOG: %v", err)
	}

	fmt.Println("Connection log saved successfully for player ID", playerID)

	return nil
}

/* -----------------------------------------------------
Table joueurs {
    id INT [pk, increment]
    utilisateur_id INT [ref: > utilisateurs_discord.id, null]
    jeu VARCHAR(255) [not null]
    compte_id VARCHAR(255) [pk]
    premiere_co DATETIME
    derniere_co DATETIME
}
----------------------------------------------------- */

// GetAllPlayers returns all the players from the database
func GetAllPlayers() ([]models.Player, error) {
	return nil, nil // TODO
}

// GetAllMinecraftPlayers returns all the Minecraft players from the database
func GetAllMinecraftPlayers() ([]models.Player, error) {
	query := "SELECT * FROM joueurs WHERE jeu = 'Minecraft'"
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("FAILED TO GET MINECRAFT PLAYERS: %v", err)
	}
	defer rows.Close()

	var players []models.Player
	for rows.Next() {
		var player models.Player
		var utilisateurID sql.NullInt64 // Use sql.NullInt64 to handle NULL values

		if err := rows.Scan(&player.ID, &utilisateurID, &player.Jeu, &player.CompteID, &player.PremiereCo, &player.DerniereCo); err != nil {
			return nil, fmt.Errorf("FAILED TO SCAN MINECRAFT PLAYER: %v", err)
		}

		// If the utilisateurID is NULL, set it to 0
		if utilisateurID.Valid {
			player.UtilisateurID = int(utilisateurID.Int64)
		} else {
			player.UtilisateurID = 0
		}

		players = append(players, player)
	}

	return players, nil
}

// CheckAndInsertPlayer checks if a player exists in the database and inserts it if it doesn't
func CheckAndInsertPlayerWithPlayerName(playerName string, serverID int, timeConf string) (int, error) {
	getPlayerUUID, err := GetPlayerAccountIdByPlayerName(playerName, "Minecraft")
	if err != nil {
		return -1, fmt.Errorf("FAILED TO GET PLAYER UUID BY PLAYER NAME: %v", err)
	}

	return CheckAndInsertPlayerWithPlayerUUID(getPlayerUUID, serverID, timeConf)
}

// InsertPlayer inserts a player in the database. if utilisateurID is -1, then null is inserted
func InsertPlayer(utilisateurID int, jeu string, compteID string, premiereCo time.Time, derniereCo time.Time) (int, error) {
	var insertQuery string
	var err error
	if utilisateurID == -1 {
		insertQuery = "INSERT INTO joueurs (utilisateur_id, jeu, compte_id, premiere_co, derniere_co) VALUES (null, ?, ?, ?, ?)"
		_, err = db.Exec(insertQuery, jeu, compteID, premiereCo, derniereCo)
	} else {
		insertQuery = "INSERT INTO joueurs (utilisateur_id, jeu, compte_id, premiere_co, derniere_co) VALUES (?, ?, ?, ?, ?)"
		_, err = db.Exec(insertQuery, utilisateurID, jeu, compteID, premiereCo, derniereCo)
	}
	if err != nil {
		return -1, fmt.Errorf("FAILED TO INSERT PLAYER: %v", err)
	}

	playerID, err := GetPlayerIdByAccountId(compteID)
	if err != nil {
		return -1, fmt.Errorf("FAILED TO GET PLAYER ID: %v", err)
	}

	return playerID, nil
}

// CheckAndInsertPlayerWithPlayerUUID checks if a player exists in the database and inserts it if it doesn't
func CheckAndInsertPlayerWithPlayerUUID(playerUUID string, serverID int, timeConf string) (int, error) {
	var datetime time.Time
	if timeConf == "now" {
		datetime = GetGoodDatetime()
	} else {
		datetime, _ = time.Parse("2006-01-02 15:04:05", "2000-01-01 00:00:00") // Default value if we don't know the time
	}

	// Check that playerUUID is not empty
	if playerUUID == "" || playerUUID == "null" {
		return -1, fmt.Errorf("PLAYER UUID IS EMPTY")
	}

	// Get server game
	jeu, err := GetServerGameById(serverID)
	if err != nil {
		return -1, fmt.Errorf("FAILED TO GET SERVER GAME: %v", err)
	}

	// Check if the player already exists
	playerID, _ := GetPlayerIdByAccountId(playerUUID)
	if playerID != -1 {
		fmt.Printf("Player already exists with ID (this is not a problem) %d\n", playerID)
		return playerID, nil // Player already exists, return its ID
	}

	// If the player does not exist, insert it
	fmt.Println("Player does not exist. Inserting new player:", playerUUID)
	playerID, err = InsertPlayer(-1, jeu, playerUUID, datetime, datetime)
	if err != nil {
		return -1, fmt.Errorf("ERROR: %v", err)
	}

	return playerID, nil
}

// UpdatePlayerLastConnection updates the last connection date of a player
func UpdatePlayerLastConnection(playerID int) error {
	if playerID == -1 {
		return fmt.Errorf("PLAYER ID IS -1, CANNOT UPDATE LAST CONNECTION")
	}

	fmt.Println("Updating last connection for player ID", playerID)
	updateQuery := "UPDATE joueurs SET derniere_co = NOW() WHERE id = ?"
	_, err := db.Exec(updateQuery, playerID)
	if err != nil {
		return fmt.Errorf("FAILED TO UPDATE LAST CONNECTION: %v", err)
	}

	return nil
}

// GetPlayerById returns a player from the database by its ID
func GetPlayerById(playerID int) (models.Player, error) {
	query := "SELECT * FROM joueurs WHERE id = ?"
	var player models.Player

	err := db.QueryRow(query, playerID).Scan(&player.ID, &player.UtilisateurID, &player.Jeu, &player.CompteID, &player.PremiereCo, &player.DerniereCo)
	if err != nil {
		if err == sql.ErrNoRows {
			return player, fmt.Errorf("PLAYER NOT FOUND: %d", playerID)
		}
		return player, fmt.Errorf("FAILED TO GET PLAYER BY ID: %v", err)
	}

	return player, nil
}

// GetPlayerByUUID returns a player from the database by its UUID
func GetPlayerByUUID(playerUUID string) (models.Player, error) {
	query := `
        SELECT id, utilisateur_id, jeu, compte_id, premiere_co, derniere_co 
        FROM joueurs 
        WHERE compte_id = ?`

	var player models.Player
	var utilisateurID sql.NullInt64

	err := db.QueryRow(query, playerUUID).Scan(
		&player.ID,
		&utilisateurID,
		&player.Jeu,
		&player.CompteID,
		&player.PremiereCo,
		&player.DerniereCo,
	)

	if utilisateurID.Valid {
		player.UtilisateurID = int(utilisateurID.Int64)
	} else {
		player.UtilisateurID = -1 // Set to -1 if NULL
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return player, fmt.Errorf("player not found: %s", playerUUID)
		}
		return player, fmt.Errorf("failed to get player by UUID: %w", err)
	}

	return player, nil
}

// Getter to get the player ID by the account ID
func GetPlayerIdByAccountId(accountId any) (int, error) {
	query := "SELECT id FROM joueurs WHERE compte_id = ?"
	var playerID int

	err := db.QueryRow(query, accountId).Scan(&playerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return -1, fmt.Errorf("PLAYER ISN'T IN THE DATABASE")
		}
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

/* -----------------------------------------------------
Table joueurs_stats {
  id INT [pk, increment]
  serveur_id INT [ref: > serveurs.id, not null]
  compte_id INT [ref: > joueurs.compte_id, not null]
  tmps_jeux BIGINT [default: 0]
  nb_mort INT [default: 0]
  nb_kills INT [default: 0]
  nb_playerkill INT [default: 0]
  mob_killed JSON
  nb_blocs_detr INT [default: 0]
  nb_blocs_pose INT [default: 0]
  dist_total INT [default: 0]
  dist_pieds INT [default: 0]
  dist_elytres INT [default: 0]
  dist_vol INT [default: 0]
  item_crafted JSON
  item_broken JSON
  achievement JSON
  dern_enregistrment DATETIME [not null]
}
----------------------------------------------------- */

// CheckMinecraftPlayerGameStatisticsExists checks if the game statistics of a Minecraft player already exists
func CheckMinecraftPlayerGameStatisticsExists(playerUUID string, serverID int) bool {
	query := "SELECT COUNT(*) FROM joueurs_stats WHERE compte_id = ? AND serveur_id = ?"
	var count int

	err := db.QueryRow(query, playerUUID, serverID).Scan(&count)
	if err != nil {
		fmt.Println("FAILED TO CHECK IF PLAYER STATISTICS EXISTS:", err)
		return false
	}

	return count > 0
}

// SaveMinecraftPlayerGameStatistics saves the game statistics of a Minecraft player
func SaveMinecraftPlayerGameStatistics(serverID int, playerUUID string, playerStats models.MinecraftPlayerGameStatistics) error {
	// Prepare the SQL query
	query := `
		INSERT INTO joueurs_stats (
			serveur_id, compte_id, tmps_jeux, nb_mort, nb_kills, nb_playerkill,
			mob_killed, nb_blocs_detr, nb_blocs_pose, dist_total, dist_pieds,
			dist_elytres, dist_vol, item_crafted, item_broken, achievement, dern_enregistrment
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			tmps_jeux = VALUES(tmps_jeux),
			nb_mort = VALUES(nb_mort),
			nb_kills = VALUES(nb_kills),
			nb_playerkill = VALUES(nb_playerkill),
			mob_killed = VALUES(mob_killed),
			nb_blocs_detr = VALUES(nb_blocs_detr),
			nb_blocs_pose = VALUES(nb_blocs_pose),
			dist_total = VALUES(dist_total),
			dist_pieds = VALUES(dist_pieds),
			dist_elytres = VALUES(dist_elytres),
			dist_vol = VALUES(dist_vol),
			item_crafted = VALUES(item_crafted),
			item_broken = VALUES(item_broken),
			achievement = VALUES(achievement),
			dern_enregistrment = VALUES(dern_enregistrment)
	`

	// Convert JSON fields
	mobKilledJSON, err := json.Marshal(playerStats.MobsKilled)
	if err != nil {
		return fmt.Errorf("failed to marshal mob_killed: %v", err)
	}

	itemsCraftedJSON, err := json.Marshal(playerStats.ItemsCrafted)
	if err != nil {
		return fmt.Errorf("failed to marshal item_crafted: %v", err)
	}

	itemsBrokenJSON, err := json.Marshal(playerStats.ItemsBroken)
	if err != nil {
		return fmt.Errorf("failed to marshal item_broken: %v", err)
	}

	achievementsJSON, err := json.Marshal(playerStats.Achievements)
	if err != nil {
		return fmt.Errorf("failed to marshal achievement: %v", err)
	}

	// Execute the query with all the necessary values
	_, err = db.Exec(query,
		serverID, playerUUID, playerStats.TimePlayed,
		playerStats.Deaths, playerStats.Kills, playerStats.PlayerKills,
		mobKilledJSON, playerStats.BlocksDestroyed, playerStats.BlocksPlaced,
		playerStats.TotalDistance, playerStats.DistanceByFoot, playerStats.DistanceByElytra,
		playerStats.DistanceByFlight, itemsCraftedJSON, itemsBrokenJSON,
		achievementsJSON, GetGoodDatetime(),
	)

	if err != nil {
		return fmt.Errorf("failed to insert/update player statistics: %v", err)
	}

	return nil
}

// UpdateMinecraftPlayerGameStatistics updates the game statistics of a Minecraft player
func UpdateMinecraftPlayerGameStatistics(serverID int, playerUUID string, playerStats models.MinecraftPlayerGameStatistics) error {
	// Prepare the SQL query
	query := `
		UPDATE joueurs_stats SET
			tmps_jeux = ?,
			nb_mort = ?,
			nb_kills = ?,
			nb_playerkill = ?,
			mob_killed = ?,
			nb_blocs_detr = ?,
			nb_blocs_pose = ?,
			dist_total = ?,
			dist_pieds = ?,
			dist_elytres = ?,
			dist_vol = ?,
			item_crafted = ?,
			item_broken = ?,
			achievement = ?,
			dern_enregistrment = ?
		WHERE compte_id = ? AND serveur_id = ?
	`

	// Convert JSON fields
	mobKilledJSON, err := json.Marshal(playerStats.MobsKilled)
	if err != nil {
		return fmt.Errorf("failed to marshal mob_killed: %v", err)
	}

	itemsCraftedJSON, err := json.Marshal(playerStats.ItemsCrafted)
	if err != nil {
		return fmt.Errorf("failed to marshal item_crafted: %v", err)
	}

	itemsBrokenJSON, err := json.Marshal(playerStats.ItemsBroken)
	if err != nil {
		return fmt.Errorf("failed to marshal item_broken: %v", err)
	}

	achievementsJSON, err := json.Marshal(playerStats.Achievements)
	if err != nil {
		return fmt.Errorf("failed to marshal achievement: %v", err)
	}

	// Execute the query with all the necessary values
	_, err = db.Exec(query,
		playerStats.TimePlayed, playerStats.Deaths, playerStats.Kills, playerStats.PlayerKills,
		mobKilledJSON, playerStats.BlocksDestroyed, playerStats.BlocksPlaced,
		playerStats.TotalDistance, playerStats.DistanceByFoot, playerStats.DistanceByElytra,
		playerStats.DistanceByFlight, itemsCraftedJSON, itemsBrokenJSON,
		achievementsJSON, GetGoodDatetime(),
		playerUUID, serverID,
	)

	if err != nil {
		return fmt.Errorf("failed to update player statistics: %v", err)
	}

	return nil
}

func GetGoodDatetime() time.Time {
	// Time of now + 1 hour. Not good practice, but it's a quick fix. Again. Yeah.
	return time.Now().Add(time.Hour)
}
