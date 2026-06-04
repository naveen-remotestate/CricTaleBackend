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

func ReopenInnings(tx *sqlx.Tx, inningsID string) error {

	query := `
		UPDATE innings
		SET
			is_completed = FALSE,
			end_time = NULL,
			updated_at = NOW()
		WHERE id = $1
	`

	_, err := tx.Exec(query, inningsID)
	return err
}

func ReopenMatch(tx *sqlx.Tx, matchID string) error {

	query := `
		UPDATE matches
		SET
			winner_team_id = NULL,
			end_time = NULL,
			updated_at = NOW()
		WHERE id = $1
	`

	_, err := tx.Exec(query,matchID,
	)

	return err
}