package handler

import (
	"CricTail_Backend/database"
	"CricTail_Backend/database/dbHelper"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func UndoLastBall(c *gin.Context) {

	matchID := c.Param("matchID")

	match, err := dbHelper.GetMatchByID(
		matchID,
	)
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	lastBall, err :=
		dbHelper.GetLastBallEvent(
			match.CurrentInningID,
		)

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ball not found",
		})
		return
	}

	txErr := database.Tx(
		func(tx *sqlx.Tx) error {

			return dbHelper.DeleteBallEvent(
				tx,
				lastBall.ID,
			)
		},
	)
	if txErr != nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ball deleted",
	})
}
