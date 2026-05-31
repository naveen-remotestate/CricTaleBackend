package dbHelper

import (
	"CricTail_Backend/database"
	"CricTail_Backend/models"

	"github.com/jmoiron/sqlx"
)

func CreateTeam(
	tx *sqlx.Tx,
	teamName string,
	hostedBy string,
) (string, error) {

	query := `
		INSERT INTO teams (
			name,
			created_by
		)
		VALUES ($1, $2)
		RETURNING id
	`

	var teamID string

	err := tx.Get(
		&teamID,
		query,
		teamName,
		hostedBy,
	)
	if err != nil {
		return "", err
	}

	return teamID, nil
}

func AddPlayerToTeam(
	tx *sqlx.Tx,
	teamID string,
	player models.TeamPlayerInput,
) error {

	query := `
		INSERT INTO team_players (
			team_id,
			user_id,
			is_captain
		)
		VALUES ($1, $2, $3)
	`

	_, err := tx.Exec(
		query,
		teamID,
		player.UserID,
		player.IsCaptain,
	)

	return err
}

func CreateMatch(
	tx *sqlx.Tx,
	teamAID string,
	teamBID string,
	tossWinnerID string,
	tossDecision string,
	battingFirstTeamID string,
	overs int,
	hostedBy string,
) (string, error) {

	query := `
		INSERT INTO matches (
			team_a_id,
			team_b_id,
			toss_winner_team_id,
			toss_decision,
			batting_first_team_id,
			overs,
			hosted_by
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,$7
		)
		RETURNING id
	`

	var matchID string

	err := tx.Get(
		&matchID,
		query,
		teamAID,
		teamBID,
		tossWinnerID,
		tossDecision,
		battingFirstTeamID,
		overs,
		hostedBy,
	)

	if err != nil {
		return "", err
	}

	return matchID, nil
}

func CreateInning(tx *sqlx.Tx, matchID, inningNumber, battingFirstTeamID, bowlingFirstTeamID string) (string, error) {
	query := `
		INSERT INTO innings (
			match_id,
			innings_no,
		    batting_team_id,
		    bowling_team_id
		)
		VALUES (
			$1,$2,$3,$4
		)
		RETURNING id
	`

	var InningID string

	err := tx.Get(
		&InningID,
		query,
		matchID,
		inningNumber,
		battingFirstTeamID,
		bowlingFirstTeamID,
	)

	if err != nil {
		return "", err
	}

	return InningID, nil
}

func CreateBattingScorecard(tx *sqlx.Tx, inningID, userID string) error {
	query := `
		INSERT INTO batting_scorecards (
			innings_id,
			user_id
		)
		VALUES (
			$1,$2
		)`

	_, err := tx.Exec(
		query,
		inningID,
		userID,
	)

	if err != nil {
		return err
	}
	return nil

}

func CreateBowlingScorecard(tx *sqlx.Tx, inningID, userID string) error {
	query := `
		INSERT INTO bowling_scorecards (
			innings_id,
			user_id
		)
		VALUES (
			$1,$2
		)`

	_, err := tx.Exec(
		query,
		inningID,
		userID,
	)

	if err != nil {
		return err
	}
	return nil

}

func CreateLiveMatch(tx *sqlx.Tx, matchID, InningID, StrikerID, NonStrikerID, BowlerID string) error {
	query := `
		INSERT INTO live_match (
			match_id,
			innings_id,
			striker_id,
		    non_striker_id,
		    current_bowler_id
		)
		VALUES (
			$1,$2,$3,$4,$5
		)`

	_, err := tx.Exec(
		query,
		matchID,
		InningID,
		StrikerID,
		NonStrikerID,
		BowlerID,
	)

	if err != nil {
		return err
	}
	return nil
}

func GetMatches() ([]models.MatchResponse, error) {
	query := `
		SELECT
			--match
			m.id AS match_id,

			m.toss_winner_team_id,
			m.winner_team_id,

			m.toss_decision,
			m.hosted_by,

			m.current_innings_no,

			m.overs,

			m.start_time,
			m.end_time,

			-- team a
			ta.id AS team_a_id,
			ta.name AS team_a_name,

			-- TEAM B
			tb.id AS team_b_id,
			tb.name AS team_b_name,

			-- live score
			lm.total_runs AS current_total_runs,
			lm.total_wickets AS current_total_wickets,
			lm.legal_balls,

			-- current inning
			i.id AS current_inning_id,

			-- previous iinnng
			pi.total_runs AS previous_innings_score,
			pi.legal_balls AS previous_innings_legal_balls,

			-- striker
			s.user_id AS striker_id,
			s.full_name AS striker_name,

			sb.runs AS striker_runs,
			sb.balls_faced AS striker_balls,

			-- non striker
			ns.user_id AS non_striker_id,
			ns.full_name AS non_striker_name,

			nsb.runs AS non_striker_runs,
			nsb.balls_faced AS non_striker_balls,

			-- bowler
			b.user_id AS bowler_id,
			b.full_name AS bowler_name,

			bs.runs_conceded AS bowler_runs_given,
			bs.legal_balls AS bowler_legal_balls,
			bs.wickets AS bowler_wickets

		FROM matches m

		-- teams
		INNER JOIN teams ta
			ON ta.id = m.team_a_id

		INNER JOIN teams tb
			ON tb.id = m.team_b_id

		-- live match
		LEFT JOIN live_match lm
			ON lm.match_id = m.id

		-- current inning
		LEFT JOIN innings i
			ON i.match_id = m.id
			AND i.innings_no = m.current_innings_no

		-- previous inning
		LEFT JOIN innings pi
			ON pi.match_id = m.id
			AND pi.innings_no = m.current_innings_no - 1

		-- striker
		LEFT JOIN users s
			ON s.user_id = lm.striker_id

		LEFT JOIN batting_scorecards sb
			ON sb.innings_id = i.id
			AND sb.user_id = s.user_id

		-- non striker
		LEFT JOIN users ns
			ON ns.user_id = lm.non_striker_id

		LEFT JOIN batting_scorecards nsb
			ON nsb.innings_id = i.id
			AND nsb.user_id = ns.user_id

		-- bowler
		LEFT JOIN users b
			ON b.user_id = lm.current_bowler_id

		LEFT JOIN bowling_scorecards bs
			ON bs.innings_id = i.id
			AND bs.user_id = b.user_id

		ORDER BY m.created_at DESC
	`

	var matches []models.MatchResponse

	err := database.DB.Select(
		&matches,
		query,
	)
	if err != nil {
		return nil, err
	}

	return matches, nil
}

func GetMatchByID(matchID string) (*models.MatchResponse, error) {

	query := `
		SELECT

			-- match
			m.id AS match_id,

			m.toss_winner_team_id,
			m.winner_team_id,

			m.toss_decision,
			m.hosted_by,

			m.current_innings_no,

			m.overs,

			m.start_time,
			m.end_time,

			ta.id AS team_a_id,
			ta.name AS team_a_name,

			tb.id AS team_b_id,
			tb.name AS team_b_name,

			-- Live score
			COALESCE(lm.total_runs, 0) AS current_total_runs,
			COALESCE(lm.total_wickets, 0) AS current_total_wickets,
			COALESCE(lm.legal_balls, 0) AS legal_balls,

			i.id AS current_inning_id,
			i.is_completed,
			i.batting_team_id,
			i.bowling_team_id,

			-- PREVIOUS INNING
			COALESCE(pi.total_runs, 0) AS previous_innings_score,
			COALESCE(pi.legal_balls, 0) AS previous_innings_legal_balls,

			-- STRIKER
			s.user_id AS striker_id,
			s.full_name AS striker_name,

			COALESCE(sb.runs, 0) AS striker_runs,
			COALESCE(sb.balls_faced, 0) AS striker_balls,

			-- NON STRIKER
			ns.user_id AS non_striker_id,
			ns.full_name AS non_striker_name,

			COALESCE(nsb.runs, 0) AS non_striker_runs,
			COALESCE(nsb.balls_faced, 0) AS non_striker_balls,

			-- BOWLER
			b.user_id AS bowler_id,
			b.full_name AS bowler_name,

			COALESCE(bs.runs_conceded, 0) AS bowler_runs_given,
			COALESCE(bs.legal_balls, 0) AS bowler_legal_balls,
			COALESCE(bs.wickets, 0) AS bowler_wickets

		FROM matches m

		INNER JOIN teams ta
			ON ta.id = m.team_a_id

		INNER JOIN teams tb
			ON tb.id = m.team_b_id

		INNER JOIN live_match lm
			ON lm.match_id = m.id

		-- current innnings
		INNER JOIN innings i
			ON i.match_id = m.id
			AND i.innings_no = m.current_innings_no

		-- preveios innings
		LEFT JOIN innings pi
			ON pi.match_id = m.id
			AND pi.innings_no = m.current_innings_no - 1

		LEFT JOIN users s
			ON s.user_id = lm.striker_id

		LEFT JOIN batting_scorecards sb
			ON sb.innings_id = i.id
			AND sb.user_id = s.user_id

		LEFT JOIN users ns
			ON ns.user_id = lm.non_striker_id

		LEFT JOIN batting_scorecards nsb
			ON nsb.innings_id = i.id
			AND nsb.user_id = ns.user_id

		LEFT JOIN users b
			ON b.user_id = lm.current_bowler_id

		LEFT JOIN bowling_scorecards bs
			ON bs.innings_id = i.id
			AND bs.user_id = b.user_id

		WHERE m.id = $1
	`

	var match models.MatchResponse

	err := database.DB.Get(
		&match,
		query,
		matchID,
	)
	if err != nil {
		return nil, err
	}

	return &match, nil
}

func GetTeamPlayerCount(teamID string) (int, error) {

	query := `
		SELECT COUNT(*)
		FROM team_players
		WHERE team_id = $1
	`

	var count int

	err := database.DB.Get(
		&count,
		query,
		teamID,
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func IsPlayerOut(inningsID string, userID string) (bool, error) {

	query := `
		SELECT is_out
		FROM batting_scorecards
		WHERE innings_id = $1
			AND user_id = $2
	`

	var isOut bool

	err := database.DB.Get(
		&isOut,
		query,
		inningsID,
		userID,
	)
	if err != nil {
		return false, err
	}

	return isOut, nil
}

func IsPlayerInTeam(teamID string, userID string) (bool, error) {

	query := `
		SELECT EXISTS (
			SELECT 1
			FROM team_players
			WHERE team_id = $1
				AND user_id = $2
		)
	`

	var exists bool
	err := database.DB.Get(
		&exists,
		query,
		teamID,
		userID,
	)
	if err != nil {
		return false, err
	}

	return exists, nil
}
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
			is_out = $5,
			dismissal_type = COALESCE($6, dismissal_type),
			dismissed_by_bowler_id =COALESCE($7, dismissed_by_bowler_id),
			fielder_id =COALESCE($8, fielder_id),
			updated_at = NOW()
		WHERE innings_id = $9
			AND user_id = $10
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

func UpdateLiveMatchAfterBall(
	tx *sqlx.Tx,
	matchID string,
	update models.LiveMatchUpdate,
) error {

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

func IsPlayerAlreadyOut(inningsID string, userID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM batting_scorecards
			WHERE innings_id = $1
				AND user_id = $2
				AND is_out = TRUE
		)
	`

	var exists bool

	err := database.DB.Get(&exists, query, inningsID, userID)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func CompleteInnings(tx *sqlx.Tx, inningsID string) error {

	query := `
		UPDATE innings
		SET
			is_completed = TRUE,
			end_time = NOW(),
			updated_at = NOW()
		WHERE id = $1
	`
	_, err := tx.Exec(query, inningsID)
	return err
}

func UpdateMatchCurrentInningsNo(tx *sqlx.Tx, matchID string, inningsNo int) error {

	query := `
		UPDATE matches
		SET
			current_innings_no = $1,
			updated_at = NOW()
		WHERE id = $2
	`

	_, err := tx.Exec(query, inningsNo, matchID)

	return err
}

func ResetLiveMatchForSecondInnings(tx *sqlx.Tx, matchID string, inningsID string, strikerID string, nonStrikerID string, bowlerID string) error {

	query := `
		UPDATE live_match
		SET

			innings_id = $1,
			total_runs = 0,
			total_wickets = 0,
			legal_balls = 0,
			striker_id = $2,
			non_striker_id = $3,
			current_bowler_id = $4,
			updated_at = NOW()
		WHERE match_id = $5
	`

	_, err := tx.Exec(
		query,
		inningsID,
		strikerID,
		nonStrikerID,
		bowlerID,
		matchID,
	)

	return err
}

func GetTeamPlayers(teamID string) ([]string, error) {

	query := `
		SELECT user_id
		FROM team_players
		WHERE team_id = $1
	`

	var playerIDs []string

	err := database.DB.Select(
		&playerIDs,
		query,
		teamID,
	)
	if err != nil {
		return nil, err
	}

	return playerIDs, nil
}

func CompleteMatch(
	tx *sqlx.Tx,
	matchID string,
	winnerTeamID string,
) error {

	query := `
		UPDATE matches
		SET
			winner_team_id = NULLIF($1, ''),
			end_time = NOW(),
			updated_at = NOW()
		WHERE id = $2
	`

	_, err := tx.Exec(query, winnerTeamID, matchID)

	return err
}

func GetBattingScorecard(inningsID string) ([]models.BattingScorecardResponse, error) {

	query := `
		SELECT
			u.user_id,
			u.full_name,
			bs.runs,
			bs.balls_faced,
			bs.fours,
			bs.sixes,
			bs.is_out,
			bs.dismissal_type
		FROM batting_scorecards bs
		INNER JOIN users u
			ON u.user_id = bs.user_id
		WHERE bs.innings_id = $1
		ORDER BY bs.created_at
	`

	var batting []models.BattingScorecardResponse

	err := database.DB.Select(&batting, query, inningsID)
	if err != nil {
		return nil, err
	}
	return batting, nil
}

func GetBowlingScorecard(inningsID string) ([]models.BowlingScorecardResponse, error) {

	query := `
		SELECT
			u.user_id,
			u.full_name,
			bs.legal_balls,
			bs.runs_conceded,
			bs.wickets,
			bs.wides,
			bs.no_balls
		FROM bowling_scorecards bs
		INNER JOIN users u
			ON u.user_id = bs.user_id
		WHERE bs.innings_id = $1
		ORDER BY bs.created_at
	`

	var bowling []models.BowlingScorecardResponse
	err := database.DB.Select(&bowling, query, inningsID)
	if err != nil {
		return nil, err
	}
	return bowling, nil
}

func GetMatchInnings(matchID string) ([]models.InningsScorecard, error) {

	query := `
		SELECT
			id,
			innings_no,
			batting_team_id,
			bowling_team_id,
			total_runs,
			total_wickets,
			legal_balls,
			extras
		FROM innings
		WHERE match_id = $1
		ORDER BY innings_no
	`

	var innings []models.InningsScorecard
	err := database.DB.Select(&innings, query, matchID)
	if err != nil {
		return nil, err
	}
	return innings, nil
}

func GetBallEvents(inningsID string) ([]models.BallEventResponse, error) {

	query := `
		SELECT
			be.id,
			be.ball_sequence,
			be.over_no,
			be.ball_in_over,
			be.striker_id,
			s.full_name AS striker_name,
			be.non_striker_id,
			ns.full_name AS non_striker_name,
			be.bowler_id,
			b.full_name AS bowler_name,
			be.runs_off_bat,
			be.extra_runs,
			be.total_runs,
			be.extra_type,
			be.is_legal_delivery,
			be.is_boundary_four,
			be.is_boundary_six,
			be.is_dot_ball,
			be.is_wicket,
			be.wicket_type,
			be.dismissed_player_id,
			dp.full_name AS dismissed_player_name,
			be.dismissed_by_fielder_id,
			f.full_name AS dismissed_by_fielder_name,
			be.bowled_at
		FROM ball_events be
		LEFT JOIN users s
			ON s.user_id = be.striker_id
		LEFT JOIN users ns
			ON ns.user_id = be.non_striker_id
		LEFT JOIN users b
			ON b.user_id = be.bowler_id
		LEFT JOIN users dp
			ON dp.user_id = be.dismissed_player_id
		LEFT JOIN users f
			ON f.user_id = be.dismissed_by_fielder_id
		WHERE be.innings_id = $1
		ORDER BY be.ball_sequence ASC
	`

	var ballEvents []models.BallEventResponse
	err := database.DB.Select(&ballEvents, query, inningsID)
	if err != nil {
		return nil, err
	}

	return ballEvents, nil
}
