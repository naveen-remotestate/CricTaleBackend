package dbHelper

import (
	"CricTail_Backend/database"
	"CricTail_Backend/models"

	"github.com/jmoiron/sqlx"
)

func GetLastBallEvent(inningsID string) (*models.BallEventResponse, error) {

	query := `
		SELECT
			*
		FROM ball_events
		WHERE innings_id = $1
		ORDER BY ball_sequence DESC
		LIMIT 1
	`

	var event models.BallEventResponse
	err := database.DB.Get(&event, query, inningsID)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

func DeleteBallEvent(tx *sqlx.Tx, ballEventID string) error {

	query := `
		DELETE FROM ball_events
		WHERE id = $1
	`

	_, err := tx.Exec(query, ballEventID)
	return err
}
