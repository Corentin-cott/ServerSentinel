package db_stats

import (
	"database/sql"
	"fmt"
	"strings"
    "encoding/json"

	"github.com/Corentin-cott/ServerSentinel/internal/models"
	"github.com/Corentin-cott/ServerSentinel/internal/db"
)

var DB *sql.DB

func Init(db *sql.DB) {
	DB = db
}

func GetAllMinecraftServers() ([]models.Server, error) {
	rows, err := DB.Query("SELECT * FROM serveurs WHERE jeu = 'Minecraft'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []models.Server
	for rows.Next() {
		var s models.Server
		err := rows.Scan(&s.ID, &s.Nom, &s.Jeu, &s.Version, &s.Modpack, &s.ModpackURL, &s.NomMonde, &s.EmbedColor, &s.Contenaire, &s.Description, &s.Actif, &s.Global, &s.Type, &s.Image)
		if err != nil {
			return nil, err
		}
		servers = append(servers, s)
	}
	return servers, nil
}

func SavePlayerStats(stat models.PlayerStats) error {
	//fmt.Printf("\nVoici la longueur des stats pour le joueur %s :\n", stat)
	playerID, err := db.CheckAndInsertPlayerWithPlayerUUID(stat.UUID, stat.ServeurID, "idk")
	if err != nil {
		return fmt.Errorf("❌ Erreur lors de la récupération du joueur avec UUID %s: %v\n", stat.UUID, err)
	}
	if playerID == 0 {
		return fmt.Errorf("⚠️ UUID %s non trouvé dans la base de données, impossible d'enregistrer les stats de %s\n", stat.UUID)
	}

	jsonType := "{}"
	distPieds := stat.Stats["minecraft:custom:minecraft:walk_one_cm"]
	distElytres := stat.Stats["minecraft:custom:minecraft:aviate_one_cm"]
	distVol := stat.Stats["minecraft:custom:minecraft:fly_one_cm"]
	distTotal := distPieds + distElytres + distVol
	
	mob_killed_json, err := extractStatsJSON("minecraft:killed:minecraft:", stat.Stats)
	if err != nil {
		return fmt.Errorf("❌ Erreur JSON mobs tués : %v", err)
	}

	item_crafted_json, err := extractStatsJSON("minecraft:crafted:minecraft:", stat.Stats)
	if err != nil {
		return fmt.Errorf("❌ Erreur JSON items craftés : %v", err)
	}

	item_broken_json, err := extractStatsJSON("minecraft:broken:minecraft:", stat.Stats)
	if err != nil {
		return fmt.Errorf("❌ Erreur JSON items cassés : %v", err)
	}

	nb_kills := sumStatsByPrefix("minecraft:killed:", stat.Stats)
	nb_blocs_detr := sumStatsByPrefix("minecraft:mined:minecraft:", stat.Stats)
	nb_blocs_pose := sumStatsByPrefix("minecraft:used", stat.Stats)

	query := `
	INSERT INTO joueurs_stats (
			serveur_id, compte_id, tmps_jeux, nb_mort, nb_kills, nb_playerkill,
			mob_killed, nb_blocs_detr, nb_blocs_pose, dist_total, dist_pieds,
			dist_elytres, dist_vol, item_crafted, item_broken, achievement, dern_enregistrment
	) VALUES (
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW()
	)
	ON DUPLICATE KEY UPDATE
		serveur_id = VALUES(serveur_id),
		compte_id = VALUES(compte_id),
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
		achievement = VALUES(achievement)
	`

	_, err = DB.Exec(query,
		stat.ServeurID,
		stat.UUID,
		stat.Stats["minecraft:custom:minecraft:play_time"],
		stat.Stats["minecraft:custom:minecraft:deaths"],
		nb_kills,
		stat.Stats["minecraft:custom:minecraft:player_kills"],
		mob_killed_json,
		nb_blocs_detr,
		nb_blocs_pose,
		distTotal,
		distPieds,
		distElytres,
		distVol,
		item_crafted_json,
		item_broken_json,
		jsonType,
	)
	if err != nil {
		return fmt.Errorf("❌ Erreur lors de l'éxcécution de la query : \n", err)
	}

	return err
}

func extractStatsJSON(prefix string, stats map[string]int) (string, error) {
	filtered := make(map[string]int)
	for key, value := range stats {
		if strings.HasPrefix(key, prefix) {
			name := strings.TrimPrefix(key, prefix)
			filtered[name] = value
		}
	}
	bytes, err := json.Marshal(filtered)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func sumStatsByPrefix(prefix string, stats map[string]int) int {
	total := 0
	for key, value := range stats {
		if strings.HasPrefix(key, prefix) {
			total += value
		}
	}
	return total
}
