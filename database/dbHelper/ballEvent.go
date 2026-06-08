package dbHelper

import (
	"CricTail_Backend/database"
	"CricTail_Backend/models"

	"github.com/jmoiron/sqlx"
)

func InsertBallEvent(tx *sqlx.Tx, event models.BallEventInsert) error {

	query := `
		INSERT INTO ball_events (

			innings_id,

			ball_sequence,
			over_no,
			ball_in_over,

			striker_id,
			non_striker_id,
			bowler_id,

			runs_off_bat,
			extra_runs,
			total_runs,

			extra_type,

			is_legal_delivery,

			is_boundary_four,
			is_boundary_six,
			is_dot_ball,

			is_wicket,

			wicket_type,

			dismissed_player_id,
			dismissed_by_fielder_id

		)
		VALUES (

			$1,

			$2,$3,$4,

			$5,$6,$7,

			$8,$9,$10,

			$11,

			$12,

			$13,$14,$15,

			$16,

			$17,

			$18,$19
		)
	`

	_, err := tx.Exec(
		query,

		event.InningsID,

		event.BallSequence,
		event.OverNo,
		event.BallInOver,

		event.StrikerID,
		event.NonStrikerID,
		event.BowlerID,

		event.RunsOffBat,
		event.ExtraRuns,
		event.TotalRuns,

		event.ExtraType,

		event.IsLegalDelivery,

		event.IsBoundaryFour,
		event.IsBoundarySix,
		event.IsDotBall,

		event.IsWicket,

		event.WicketType,

		event.DismissedPlayerID,
		event.DismissedByFielderID,
	)

	return err
}

func UpdateInningsAfterBall(tx *sqlx.Tx, inningsID string, update models.InningsUpdate) error {

	query := `
		UPDATE innings
		SET
			total_runs = total_runs + $1,
			total_wickets = total_wickets + $2,
			legal_balls = legal_balls + $3,
			extras = extras + $4,
			wides = wides + $5,
			no_balls = no_balls + $6,
			byes = byes + $7,
			leg_byes = leg_byes + $8,
			updated_at = NOW()
		WHERE id = $9
	`

	_, err := tx.Exec(
		query,

		update.TotalRunsIncrement,
		update.WicketIncrement,
		update.LegalBallIncrement,
		update.ExtrasIncrement,
		update.WidesIncrement,
		update.NoBallsIncrement,
		update.ByesIncrement,
		update.LegByesIncrement,

		inningsID,
	)

	return err
}

func GetLastBallSequence(inningsID string) (int, error) {

	query := `
		SELECT COALESCE(
			MAX(ball_sequence),
			0
		)
		FROM ball_events
		WHERE innings_id = $1
	`

	var lastSequence int

	err := database.DB.Get(
		&lastSequence,
		query,
		inningsID,
	)
	if err != nil {
		return 0, err
	}

	return lastSequence, nil
}

func UpdateBattingScorecardAfterBall(tx *sqlx.Tx, inningsID string, batsmanID string, update models.BattingScorecardUpdate) error {

	query := `
		UPDATE batting_scorecards
		SET

			runs = runs + $1,
			balls_faced = balls_faced + $2,
			fours = fours + $3,
			sixes = sixes + $4,
			is_out = COALESCE($5, is_out),
			dismissal_type = CASE
			    WHEN $9 THEN NULL
			    ELSE COALESCE($6, dismissal_type)
			    END,
		    dismissed_by_bowler_id = CASE
		        WHEN $9 THEN NULL
		        ELSE COALESCE($7, dismissed_by_bowler_id)
		        END,
		    fielder_id = CASE
		        WHEN $9 THEN NULL
		        ELSE COALESCE($8, fielder_id)
		        END,
			updated_at = NOW()
		WHERE innings_id = $10
			AND user_id = $11
	`

	_, err := tx.Exec(
		query,

		update.RunsIncrement,
		update.BallsIncrement,
		update.FoursIncrement,
		update.SixesIncrement,
		update.IsOut,
		update.DismissalType,
		update.DismissedByBowlerID,
		update.FielderID,
		//for undo a ball event
		update.ClearDismissal,

		inningsID,
		batsmanID,
	)

	return err
}

func UpdateBowlingScorecardAfterBall(tx *sqlx.Tx, inningsID string, bowlerID string, update models.BowlingScorecardUpdate) error {

	query := `
		UPDATE bowling_scorecards
		SET
			legal_balls = legal_balls + $1,
			runs_conceded = runs_conceded + $2,
			wickets = wickets + $3,
			wides = wides + $4,
			no_balls = no_balls + $5,
			updated_at = NOW()
		WHERE innings_id = $6
			AND user_id = $7
	`

	_, err := tx.Exec(
		query,

		update.LegalBallsIncrement,
		update.RunsConcededIncrement,
		update.WicketsIncrement,
		update.WidesIncrement,
		update.NoBallsIncrement,

		inningsID,
		bowlerID,
	)

	return err
}

func UpdateLiveMatchAfterBall(tx *sqlx.Tx, matchID string, update models.LiveMatchUpdate) error {

	query := `
		UPDATE live_match
		SET
			total_runs = total_runs + $1,
			total_wickets =	total_wickets + $2,
			legal_balls = legal_balls + $3,
			striker_id = $4,
			non_striker_id = $5,
			current_bowler_id=$6,
			updated_at = NOW()
		WHERE match_id = $7
	`

	_, err := tx.Exec(
		query,

		update.TotalRunsIncrement,
		update.TotalWicketsIncrement,
		update.LegalBallsIncrement,
		update.StrikerID,
		update.NonStrikerID,
		update.BowlerID,

		matchID,
	)

	return err
}
