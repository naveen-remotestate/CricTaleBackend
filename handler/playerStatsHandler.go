package handler

import (
	"CricTail_Backend/database/dbHelper"
	"CricTail_Backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func GetPlayerStats(c *gin.Context) {
	userID := c.GetString("user_id")
	playerStats, err := dbHelper.GetPlayerStatsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get player Stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"player-stats": playerStats,
	})
}

func GetPlayers(c *gin.Context) {
	search := c.Query("search")

	players, err := dbHelper.GetPlayers(search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get players"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"players": players,
		"total":   len(players),
	})
}

func UpdatePlayerProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	var req models.UpdatePlayer
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request- you may send these fields in body-full_name, batting_style, bowling_style",
		})
		return
	}

	err := dbHelper.UpdatePlayerProfile(
		userID,
		req.FullName,
		req.BattingStyle,
		req.BowlingStyle,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "update player failed"}) //"update player failed"
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "player profile updated successfully",
	})
}

func ProcessBattingCareerStats(
	tx *sqlx.Tx,
	matchID string,
) error {

	battingStats, err := dbHelper.GetMatchBattingStats(tx, matchID)
	if err != nil {
		return err
	}

	for _, stat := range battingStats {

		err = dbHelper.UpdateBattingCareerStats(
			tx,
			stat,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func ProcessBowlingCareerStats(
	tx *sqlx.Tx,
	matchID string,
) error {

	bowlingStats, err := dbHelper.GetMatchBowlingStats(tx, matchID)
	if err != nil {
		return err
	}

	for _, stat := range bowlingStats {

		err = dbHelper.UpdateBowlingCareerStats(
			tx,
			stat,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func ProcessPlayerCareerStats(tx *sqlx.Tx, matchID string, winnerTeamID *string) error {

	err := ProcessBattingCareerStats(
		tx,
		matchID,
	)
	if err != nil {
		return err
	}

	err = ProcessBowlingCareerStats(tx, matchID)
	if err != nil {
		return err
	}

	err = dbHelper.UpdateMatchesPlayed(tx, matchID)
	if err != nil {
		return err
	}

	if winnerTeamID != nil {
		err = dbHelper.UpdateMatchesWon(tx, *winnerTeamID)
		if err != nil {
			return err
		}
	}

	return nil
}
