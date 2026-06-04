package models

import "time"

type PlayerStats struct {
	ID     string `db:"id" json:"id"`
	UserID string `db:"user_id" json:"user_id"`

	BattingStyle *string `db:"batting_style" json:"batting_style"`
	BowlingStyle *string `db:"bowling_style" json:"bowling_style"`

	MatchesPlayed int `db:"matches_played" json:"matches_played"`

	InningsBatted int `db:"innings_batted" json:"innings_batted"`
	InningsBowled int `db:"innings_bowled" json:"innings_bowled"`

	MatchesWon int `db:"matches_won" json:"matches_won"`

	TotalPoints int `db:"total_points" json:"total_points"`

	TotalRuns       int `db:"total_runs" json:"total_runs"`
	HighestRun      int `db:"highest_run" json:"highest_run"`
	TotalBallsFaced int `db:"total_balls_faced" json:"total_balls_faced"`
	TotalOuts       int `db:"total_outs" json:"total_outs"`

	TotalFours int `db:"total_fours" json:"total_fours"`
	TotalSixes int `db:"total_sixes" json:"total_sixes"`

	Ducks       int `db:"ducks" json:"ducks"`
	GoldenDucks int `db:"golden_ducks" json:"golden_ducks"`

	Fifties  int `db:"fifties" json:"fifties"`
	Hundreds int `db:"hundreds" json:"hundreds"`

	TotalWicketsTaken int `db:"total_wickets_taken" json:"total_wickets_taken"`
	TotalBallsBowled  int `db:"total_balls_bowled" json:"total_balls_bowled"`
	TotalRunsConceded int `db:"total_runs_conceded" json:"total_runs_conceded"`
	TotalMaidens      int `db:"total_maidens" json:"total_maidens"`

	Wides   int `db:"wides" json:"wides"`
	NoBalls int `db:"no_balls" json:"no_balls"`

	HighestWicketTaken int `db:"highest_wicket_taken"`

	Catches  int `db:"catches" json:"catches"`
	RunOuts  int `db:"run_outs" json:"run_outs"`
	Stumping int `db:"stumping" json:"stumping"`

	UpdatedAt string `db:"updated_at" json:"updated_at"`
}

type Player struct {
	UserID       string    `db:"user_id" json:"user_id"`
	FullName     string    `db:"full_name" json:"full_name"`
	MobileNumber string    `db:"mobile_number" json:"mobile_number"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type UpdatePlayer struct {
	FullName     string `db:"full_name" json:"full_name"`
	BattingStyle string `db:"batting_style" json:"batting_style"`
	BowlingStyle string `db:"bowling_style" json:"bowling_style"`
}
