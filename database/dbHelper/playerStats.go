package dbHelper

import (
	"CricTail_Backend/models"

	"github.com/jmoiron/sqlx"
)

func GetMatchBattingStats(tx *sqlx.Tx, matchID string) ([]models.MatchBattingStats, error) {

	query := `
		SELECT
			bs.user_id,
			bs.runs,
			bs.balls_faced,
			bs.fours,
			bs.sixes,
			bs.is_out
		FROM batting_scorecards bs
		INNER JOIN innings i
			ON i.id = bs.innings_id

		WHERE i.match_id = $1
	`

	var stats []models.MatchBattingStats

	err := tx.Select(&stats, query, matchID)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func GetMatchBowlingStats(tx *sqlx.Tx, matchID string) ([]models.MatchBowlingStats, error) {

	query := `
		SELECT
			bs.user_id,
			bs.legal_balls,
			bs.runs_conceded,
			bs.wickets,
			bs.wides,
			bs.no_balls
		FROM bowling_scorecards bs

		INNER JOIN innings i
			ON i.id = bs.innings_id
		WHERE i.match_id = $1
	`

	var stats []models.MatchBowlingStats
	err := tx.Select(&stats, query, matchID)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func UpdateBattingCareerStats(tx *sqlx.Tx, stats models.MatchBattingStats) error {

	query := `
		UPDATE player_career_stats
		SET
		    innings_batted =innings_batted + CASE
		        WHEN $2 > 0 OR $5 = true
		            THEN 1
		        ELSE 0
		        END,
		    total_runs =total_runs + $1,
			total_balls_faced =	total_balls_faced + $2,
			total_fours =total_fours + $3,
			total_sixes =total_sixes + $4,
			total_outs =total_outs +
				CASE
					WHEN $5= true THEN 1
					ELSE 0
				END,
			highest_run =GREATEST(
					highest_run,
					$1
				),
		    ducks =	ducks +
		            CASE
		                WHEN $1 = 0
		                         AND $5 = true
		                    THEN 1
		                ELSE 0
		                END,
		    golden_ducks =	golden_ducks +
		                   CASE
		                       WHEN $1 = 0
		                                AND $2 = 1
		                                AND $5 = true
		                           	THEN 1
		                       ELSE 0
		                       END,
			fifties =fifties +
				CASE
					WHEN $1 >= 50
						AND $1 < 100
					THEN 1
					ELSE 0
				END,

			hundreds =hundreds +
				CASE
					WHEN $1 >= 100
					THEN 1
					ELSE 0
				END,
			updated_at = NOW()
		WHERE user_id = $6
	`

	_, err := tx.Exec(query,
		stats.Runs,
		stats.BallsFaced,
		stats.Fours,
		stats.Sixes,
		stats.IsOut,

		stats.UserID,
	)
	return err
}

func UpdateBowlingCareerStats(tx *sqlx.Tx, stats models.MatchBowlingStats) error {

	query := `
		UPDATE player_career_stats
		SET innings_bowled =innings_bowled +
	CASE
		WHEN $1 > 0
		THEN 1
		ELSE 0
	END,			total_balls_bowled =total_balls_bowled + $1,
			total_runs_conceded =total_runs_conceded + $2,
			total_wickets_taken =total_wickets_taken + $3,
			wides =wides + $4,
			no_balls =no_balls + $5,
			highest_wicket_taken =GREATEST(
					highest_wicket_taken,
					$3
				),
			updated_at = NOW()
		WHERE user_id = $6
	`

	_, err := tx.Exec(
		query,

		stats.LegalBalls,
		stats.RunsConceded,
		stats.Wickets,
		stats.Wides,
		stats.NoBalls,

		stats.UserID,
	)
	return err
}

func UpdateMatchesPlayed(tx *sqlx.Tx, matchID string) error {

	query := `
		UPDATE player_career_stats
		SET
			matches_played = matches_played + 1,
			updated_at = NOW()
		WHERE user_id IN (
			SELECT tp.user_id
			FROM matches m
			INNER JOIN team_players tp
				ON tp.team_id IN (
					m.team_a_id,
					m.team_b_id
				)
			WHERE m.id = $1
		)
	`

	_, err := tx.Exec(
		query,
		matchID,
	)

	return err
}

func UpdateMatchesWon(tx *sqlx.Tx, winnerTeamID string) error {

	query := `
		UPDATE player_career_stats
		SET
			matches_won = matches_won + 1,
			updated_at = NOW()
		WHERE user_id IN (
			SELECT user_id
			FROM team_players
			WHERE team_id = $1
		)
	`

	_, err := tx.Exec(query, winnerTeamID)

	return err
}
