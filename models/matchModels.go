package models

import "time"

type TeamPlayerInput struct {
	UserID    string `json:"user_id"`
	IsCaptain bool   `json:"is_captain"`
}

type TeamInput struct {
	Name    string            `json:"name"`
	Players []TeamPlayerInput `json:"players"`
}

type CreateMatchRequest struct {
	TeamA TeamInput `json:"team_a"`
	TeamB TeamInput `json:"team_b"`

	Overs    int    `json:"overs"`
	HostedBy string `json:"hosted_by"`

	TossWinnerTeam string `json:"toss_winner_team"` // A or B
	TossDecision   string `json:"toss_decision"`    // BAT or BOWL

	//for innings table

	//for live match table
	StrikerID    string `json:"striker_id"`
	NonStrikerID string `json:"non_striker_id"`
	BowlerID     string `json:"current_bowler_id"`
}

type MatchResponse struct {
	// match fields
	MatchID string `db:"match_id" json:"match_id"`

	TossWinnerTeamID *string `db:"toss_winner_team_id" json:"toss_winner_team_id"`
	WinnerTeamID     *string `db:"winner_team_id" json:"winner_team_id"`

	TossDecision string `db:"toss_decision" json:"toss_decision"`
	HostID       string `db:"hosted_by" json:"hosted_by"`

	CurrentInningNo int `db:"current_innings_no" json:"current_innings_no"`

	Overs int `db:"overs" json:"overs"`

	StartTime *time.Time `db:"start_time" json:"start_time"`
	EndTime   *time.Time `db:"end_time" json:"end_time"`

	TeamAID   string `db:"team_a_id" json:"team_a_id"`
	TeamAName string `db:"team_a_name" json:"team_a_name"`

	TeamBID   string `db:"team_b_id" json:"team_b_id"`
	TeamBName string `db:"team_b_name" json:"team_b_name"`

	// Live Score
	CurrentTotalRuns    int `db:"current_total_runs" json:"current_total_runs"`
	CurrentTotalWickets int `db:"current_total_wickets" json:"current_total_wickets"`
	LegalBalls          int `db:"legal_balls" json:"legal_balls"`

	PreviousInningsScore      *int `db:"previous_innings_score" json:"previous_innings_score"`
	PreviousInningsLegalBalls *int `db:"previous_innings_legal_balls" json:"previous_innings_legal_balls"`

	CurrentInningID string `db:"current_inning_id" json:"current_inning_id"`

	CurrentInningCompleted bool `db:"is_completed" json:"is_completed"`

	BattingTeamID string `db:"batting_team_id" json:"batting_team_id"`
	BowlingTeamID string `db:"bowling_team_id" json:"bowling_team_id"`

	// Striker details
	StrikerID    *string `db:"striker_id" json:"striker_id"`
	StrikerName  *string `db:"striker_name" json:"striker_name"`
	StrikerRuns  int     `db:"striker_runs" json:"striker_runs"`
	StrikerBalls int     `db:"striker_balls" json:"striker_balls"`

	// nonstriker details
	NonStrikerID    *string `db:"non_striker_id" json:"non_striker_id"`
	NonStrikerName  *string `db:"non_striker_name" json:"non_striker_name"`
	NonStrikerRuns  int     `db:"non_striker_runs" json:"non_striker_runs"`
	NonStrikerBalls int     `db:"non_striker_balls" json:"non_striker_balls"`

	// current bowler details
	BowlerID         *string `db:"bowler_id" json:"bowler_id"`
	BowlerName       *string `db:"bowler_name" json:"bowler_name"`
	BowlerRunsGiven  int     `db:"bowler_runs_given" json:"bowler_runs_given"`
	BowlerLegalBalls int     `db:"bowler_legal_balls" json:"bowler_legal_balls"`
	BowlerWickets    int     `db:"bowler_wickets" json:"bowler_wickets"`
}

type AddBallEventRequest struct {
	MatchID string `json:"match_id"`

	RunsOffBat int `json:"runs_off_bat"`

	ExtraRuns int     `json:"extra_runs"`
	ExtraType *string `json:"extra_type"`

	IsWicket bool `json:"is_wicket"`

	WicketType *string `json:"wicket_type"`

	DismissedPlayerID    *string `json:"dismissed_player_id"`
	DismissedByFielderID *string `json:"dismissed_by_fielder_id"`

	NextBatsmanID string `json:"next_batsman_id"`

	NextBowlerID string `json:"next_bowler_id"`
}

type MatchState struct {
	MatchID string `db:"match_id"`

	CurrentInningsID string `db:"current_innings_id"`

	BattingTeamID string `db:"batting_team_id"`
	BowlingTeamID string `db:"bowling_team_id"`

	StrikerID    string `db:"striker_id"`
	NonStrikerID string `db:"non_striker_id"`

	BowlerID string `db:"bowler_id"`

	CurrentRuns    int `db:"current_runs"`
	CurrentWickets int `db:"current_wickets"`

	LegalBalls int `db:"legal_balls"`

	Overs int `db:"overs"`
}

type BallEventInsert struct {
	InningsID string `json:"innings_id"`

	BallSequence int `json:"ball_sequence"`
	OverNo       int `json:"over_no"`
	BallInOver   int `json:"ball_in_over"`

	StrikerID    string `json:"striker_id"`
	NonStrikerID string `json:"non_striker_id"`

	BowlerID string `json:"bowler_id"`

	RunsOffBat int `json:"runs_off_bat"`
	ExtraRuns  int `json:"extra_runs"`
	TotalRuns  int `json:"total_runs"`

	ExtraType *string `json:"extra_type"`

	IsLegalDelivery bool `json:"is_legal_delivery"`

	IsBoundaryFour bool `json:"is_boundary_four"`
	IsBoundarySix  bool `json:"is_boundary_six"`
	IsDotBall      bool `json:"is_dot_ball"`

	IsWicket bool `json:"is_wicket"`

	WicketType *string `json:"wicket_type"`

	DismissedPlayerID *string `json:"dismissed_player_id"`

	DismissedByFielderID *string `json:"dismissed_by_fielder_id"`
}

type InningsUpdate struct {
	TotalRunsIncrement int `json:"total_runs_increment"`

	WicketIncrement int `json:"wicket_increment"`

	LegalBallIncrement int `json:"legal_ball_increment"`

	ExtrasIncrement int `json:"extras_increment"`

	WidesIncrement int `json:"wides_increment"`

	NoBallsIncrement int `json:"no_balls_increment"`

	ByesIncrement int `json:"byes_increment"`

	LegByesIncrement int `json:"leg_byes_increment"`
}

type BattingScorecardUpdate struct {
	RunsIncrement       int     `json:"runs_increment"`
	BallsIncrement      int     `json:"balls_increment"`
	FoursIncrement      int     `json:"fours_increment"`
	SixesIncrement      int     `json:"sixes_increment"`
	IsOut               bool    `json:"is_out"`
	DismissalType       *string `json:"dismissal_type"`
	DismissedByBowlerID *string `json:"dismissed_by_bowler_id"`
	FielderID           *string `json:"fielder_id"`
}

type BowlingScorecardUpdate struct {
	LegalBallsIncrement   int `json:"legal_balls_increment"`
	RunsConcededIncrement int `json:"runs_conceded_increment"`
	WicketsIncrement      int `json:"wickets_increment"`
	WidesIncrement        int `json:"wides_increment"`
	NoBallsIncrement      int `json:"no_balls_increment"`
}

type LiveMatchUpdate struct {
	TotalRunsIncrement    int    `json:"total_runs_increment"`
	TotalWicketsIncrement int    `json:"total_wickets_increment"`
	LegalBallsIncrement   int    `json:"legal_balls_increment"`
	StrikerID             string `json:"striker_id"`
	NonStrikerID          string `json:"non_striker_id"`
	BowlerID              string `json:"bowler_id"`
}

type StartSecondInningsRequest struct {
	MatchID      string `json:"match_id"`
	StrikerID    string `json:"striker_id"`
	NonStrikerID string `json:"non_striker_id"`
	BowlerID     string `json:"bowler_id"`
}

type BowlingScorecardResponse struct {
	UserID       string `db:"user_id" json:"user_id"`
	PlayerName   string `db:"full_name" json:"player_name"`
	LegalBalls   int    `db:"legal_balls" json:"legal_balls"`
	RunsConceded int    `db:"runs_conceded" json:"runs_conceded"`
	Wickets      int    `db:"wickets" json:"wickets"`
	Wides        int    `db:"wides" json:"wides"`
	NoBalls      int    `db:"no_balls" json:"no_balls"`
}

type BattingScorecardResponse struct {
	UserID        string  `db:"user_id" json:"user_id"`
	PlayerName    string  `db:"full_name" json:"player_name"`
	Runs          int     `db:"runs" json:"runs"`
	BallsFaced    int     `db:"balls_faced" json:"balls_faced"`
	Fours         int     `db:"fours" json:"fours"`
	Sixes         int     `db:"sixes" json:"sixes"`
	IsOut         bool    `db:"is_out" json:"is_out"`
	DismissalType *string `db:"dismissal_type" json:"dismissal_type"`
}

type InningsScorecard struct {
	InningsID     string                     `db:"id" json:"innings_id"`
	InningsNo     int                        `db:"innings_no" json:"innings_no"`
	BattingTeamID string                     `db:"batting_team_id" json:"batting_team_id"`
	BowlingTeamID string                     `db:"bowling_team_id" json:"bowling_team_id"`
	TotalRuns     int                        `db:"total_runs" json:"total_runs"`
	TotalWickets  int                        `db:"total_wickets" json:"total_wickets"`
	LegalBalls    int                        `db:"legal_balls" json:"legal_balls"`
	Extras        int                        `db:"extras" json:"extras"`
	Batting       []BattingScorecardResponse `json:"batting,omitempty"`
	Bowling       []BowlingScorecardResponse `json:"bowling,omitempty"`
}

type MatchScorecardResponse struct {
	MatchID       string            `json:"match_id"`
	FirstInnings  *InningsScorecard `json:"first_innings,omitempty"`
	SecondInnings *InningsScorecard `json:"second_innings,omitempty"`
}

type BallEventResponse struct {
	ID                     string    `db:"id" json:"id"`
	BallSequence           int       `db:"ball_sequence" json:"ball_sequence"`
	OverNo                 int       `db:"over_no" json:"over_no"`
	BallInOver             int       `db:"ball_in_over" json:"ball_in_over"`
	StrikerID              string    `db:"striker_id" json:"striker_id"`
	StrikerName            *string   `db:"striker_name" json:"striker_name"`
	NonStrikerID           string    `db:"non_striker_id" json:"non_striker_id"`
	NonStrikerName         *string   `db:"non_striker_name" json:"non_striker_name"`
	BowlerID               string    `db:"bowler_id" json:"bowler_id"`
	BowlerName             *string   `db:"bowler_name" json:"bowler_name"`
	RunsOffBat             int       `db:"runs_off_bat" json:"runs_off_bat"`
	ExtraRuns              int       `db:"extra_runs" json:"extra_runs"`
	TotalRuns              int       `db:"total_runs" json:"total_runs"`
	ExtraType              *string   `db:"extra_type" json:"extra_type"`
	IsLegalDelivery        bool      `db:"is_legal_delivery" json:"is_legal_delivery"`
	IsBoundaryFour         bool      `db:"is_boundary_four" json:"is_boundary_four"`
	IsBoundarySix          bool      `db:"is_boundary_six" json:"is_boundary_six"`
	IsDotBall              bool      `db:"is_dot_ball" json:"is_dot_ball"`
	IsWicket               bool      `db:"is_wicket" json:"is_wicket"`
	WicketType             *string   `db:"wicket_type" json:"wicket_type"`
	DismissedPlayerID      *string   `db:"dismissed_player_id" json:"dismissed_player_id"`
	DismissedPlayerName    *string   `db:"dismissed_player_name" json:"dismissed_player_name"`
	DismissedByFielderID   *string   `db:"dismissed_by_fielder_id" json:"dismissed_by_fielder_id"`
	DismissedByFielderName *string   `db:"dismissed_by_fielder_name" json:"dismissed_by_fielder_name"`
	BowledAt               time.Time `db:"bowled_at" json:"bowled_at"`
}
