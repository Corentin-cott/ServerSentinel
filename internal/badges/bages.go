package badges

import (
	"fmt"

	"github.com/Corentin-cott/ServerSentinel/internal/db"
)

func CheckVanillaPlayed(playerUUID string, addBadge bool) error {
	// We just need to check if player has stats for vanilla server ID
	needBadge := db.CheckMinecraftPlayerGameStatisticsExists(playerUUID, 1)
	if needBadge && addBadge {
		playerID, _ := db.GetPlayerIdByAccountId(playerUUID)

		err := db.AddBadgeToPlayer(playerID, 6)
		if err != nil {
			return fmt.Errorf("ERROR WHILE ADDING BADGE TO PLAYER : %v", err)
		} else {
			fmt.Println("Badge added to player")
		}
	}

	return nil
}
