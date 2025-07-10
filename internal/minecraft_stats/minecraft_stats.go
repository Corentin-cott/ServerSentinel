package minecraft_stats

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Corentin-cott/ServerSentinel/internal/db_stats"
	"github.com/Corentin-cott/ServerSentinel/internal/docker"
	"github.com/Corentin-cott/ServerSentinel/internal/models"
)

type RawStats struct {
	Stats map[string]map[string]int `json:"stats"`
}

func SyncMinecraftStats() error {
	servers, err := db_stats.GetAllMinecraftServers()
	if err != nil {
		return fmt.Errorf("Erreur lors de la récupération des serveurs Minecraft: %v\n", err)
	}

	for _, serv := range servers {
		if serv.Contenaire == "depreciated" || serv.Contenaire == "NULL" || serv.Contenaire == "null" {
			fmt.Printf("🔀 No container for %s, skipping.\n", serv.Nom)
			continue
		}
		volumePath, err := docker.GetVolumePath(serv.Contenaire)
		fmt.Printf("🔄 Récupération des stats pour le serveur %s (%s)...\n", serv.Nom, volumePath)
		if err != nil {
			fmt.Printf("❌ Container %s inaccessible: %v\n", serv.Contenaire, err)
			continue
		}

		statsPath := filepath.Join(volumePath, serv.NomMonde, "stats")
		playerStats, err := readStatsFolder(statsPath)
		if err != nil {
			fmt.Printf("⚠️  Stats introuvables pour %s: %v\n", serv.Nom, err)
			continue
		}

		// Vérification si le dossier stats est vide
		if len(playerStats) == 0 {
			fmt.Printf("⚠️  Aucun fichier de stats trouvé pour le serveur %s dans %s\n", serv.Nom, statsPath)
			continue
		}

		fmt.Printf("📊 %d joueurs trouvés dans les stats de %s\n", len(playerStats), serv.Nom)
		for _, pStat := range playerStats {
			pStat.ServeurID = serv.ID
			//fmt.Printf("\n\nVOICI LES STATS POUR LE JOUEUR %s :\n\n", pStat)
			if err := db_stats.SavePlayerStats(pStat); err != nil {
				fmt.Printf("❌ Insertion stats %s: %v", pStat.UUID, err)
			} else {
				fmt.Printf("✅ Enregistrement des stats pour le joueur %s sur le serveur %s réussi !\n", pStat.UUID, serv.Nom)
			}
		}
	}

	return nil
}

func readStatsFolder(path string) ([]models.PlayerStats, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var result []models.PlayerStats
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(path, file.Name()))
		if err != nil {
			continue
		}

		var parsed RawStats
		if err := json.Unmarshal(raw, &parsed); err != nil {
			continue
		}

		flat := make(map[string]int)
		for category, items := range parsed.Stats {
			for key, val := range items {
				flat[fmt.Sprintf("%s:%s", category, key)] = val
			}
		}

		result = append(result, models.PlayerStats{
			UUID:  strings.TrimSuffix(file.Name(), ".json"),
			Stats: flat,
		})
	}

	return result, nil
}