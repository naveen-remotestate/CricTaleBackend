package handler

import (
	"CricTail_Backend/database"
	"CricTail_Backend/database/dbHelper"
	"CricTail_Backend/models"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func CreateMatch(c *gin.Context) {

	var req models.CreateMatchRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	// validate

	req.TeamA.Name = strings.TrimSpace(req.TeamA.Name)
	req.TeamB.Name = strings.TrimSpace(req.TeamB.Name)

	if req.TeamA.Name == "" || req.TeamB.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "team names required",
		})
		return
	}

	if req.TeamA.Name == req.TeamB.Name {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "both teams cannot have same name",
		})
		return
	}

	if req.Overs <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "overs must be greater than 0",
		})
		return
	}

	if req.TossWinnerTeam != "A" && req.TossWinnerTeam != "B" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid toss_winner_team",
		})
		return
	}

	if req.TossDecision != "BAT" && req.TossDecision != "BOWL" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid toss_decision",
		})
		return
	}

	// player validations for captain and one same player in both teamss
	allPlayers := make(map[string]bool)
	samePlayerCount := 0

	validateTeam := func(team models.TeamInput) error {

		if len(team.Players) == 0 {
			return fmt.Errorf("team players required")
		}

		captainCount := 0

		for _, player := range team.Players {

			if player.UserID == "" {
				return fmt.Errorf("user_id required")
			}

			if allPlayers[player.UserID] {

				samePlayerCount++

				if samePlayerCount > 1 {
					return fmt.Errorf("only one same player allowed in both teams")
				}
			}

			allPlayers[player.UserID] = true

			if player.IsCaptain {
				captainCount++
			}
		}

		if captainCount != 1 {
			return fmt.Errorf("exactly one captain required")
		}

		return nil
	}
	if err := validateTeam(req.TeamA); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := validateTeam(req.TeamB); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var matchID string

	txErr := database.Tx(func(tx *sqlx.Tx) error {

		// create team A

		teamAID, err := dbHelper.CreateTeam(
			tx,
			req.TeamA.Name,
			req.HostedBy,
		)
		if err != nil {
			return err
		}

		for _, player := range req.TeamA.Players {

			err = dbHelper.AddPlayerToTeam(
				tx,
				teamAID,
				player,
			)
			if err != nil {
				return err
			}
		}

		// create team b

		teamBID, err := dbHelper.CreateTeam(
			tx,
			req.TeamB.Name,
			req.HostedBy,
		)
		if err != nil {
			return err
		}

		for _, player := range req.TeamB.Players {

			err = dbHelper.AddPlayerToTeam(
				tx,
				teamBID,
				player,
			)
			if err != nil {
				return err
			}
		}

		// who bat first

		var battingFirstTeamID string
		var bowlingFirstTeamID string

		if req.TossWinnerTeam == "A" {

			if req.TossDecision == "BAT" {
				battingFirstTeamID = teamAID
				bowlingFirstTeamID = teamBID
			} else {
				battingFirstTeamID = teamBID
				bowlingFirstTeamID = teamAID
			}

		} else {

			if req.TossDecision == "BAT" {
				battingFirstTeamID = teamBID
				bowlingFirstTeamID = teamAID
			} else {
				battingFirstTeamID = teamAID
				bowlingFirstTeamID = teamBID
			}
		}

		tossWinnerID := teamAID
		if req.TossWinnerTeam == "B" {
			tossWinnerID = teamBID
		}

		matchID, err = dbHelper.CreateMatch(
			tx,
			teamAID,
			teamBID,
			tossWinnerID,
			req.TossDecision,
			battingFirstTeamID,
			req.Overs,
			req.HostedBy,
		)
		if err != nil {
			return err
		}

		//create innings table
		InningID, err := dbHelper.CreateInning(
			tx,
			matchID,
			"1",
			battingFirstTeamID,
			bowlingFirstTeamID,
		)
		if err != nil {
			return err
		}

		//create Live Match table
		err = dbHelper.CreateLiveMatch(
			tx,
			matchID,
			InningID,
			req.StrikerID,
			req.NonStrikerID,
			req.BowlerID,
		)
		if err != nil {
			return err
		}
		if battingFirstTeamID == teamAID {
			//for batting scorecard of team batting first
			for _, player := range req.TeamA.Players {

				err = dbHelper.CreateBattingScorecard(
					tx, InningID, player.UserID,
				)
				if err != nil {
					return err
				}
			}
			//for bowling scorecard of team bowling first
			for _, player := range req.TeamB.Players {

				err = dbHelper.CreateBowlingScorecard(
					tx, InningID, player.UserID,
				)
				if err != nil {
					return err
				}
			}

		} else {
			for _, player := range req.TeamB.Players {

				err = dbHelper.CreateBattingScorecard(
					tx, InningID, player.UserID,
				)
				if err != nil {
					return err
				}
			}

			//for bowling scorecard of team bowling first
			for _, player := range req.TeamA.Players {

				err = dbHelper.CreateBowlingScorecard(
					tx, InningID, player.UserID,
				)
				if err != nil {
					return err
				}
			}

		}

		return nil
	})

	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": txErr.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "match created successfully",
		"match_id": matchID,
	})
}

func GetMatches(c *gin.Context) {
	matches, err := dbHelper.GetMatches()
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(), //"failed to get matches"
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"matches": matches,
	})
}

func GetMatchByID(c *gin.Context) {

	matchID := c.Param("matchID")

	if matchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "matchID required",
		})
		return
	}

	match, err := dbHelper.GetMatchByID(matchID)
	if err != nil {

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "match not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"match": match,
	})
}

func AddBallEvent(c *gin.Context) {

	var req models.AddBallEventRequest

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	err = validateBallEventRequest(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// fetching current state in db
	match, err := dbHelper.GetMatchByID(req.MatchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	err = validateMatchState(match)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	//getting player count----innings ending logic
	playerCount, err := dbHelper.GetTeamPlayerCount(match.BattingTeamID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	maxWickets := playerCount - 1

	isLegalDelivery := true

	if req.ExtraType != nil {
		if *req.ExtraType == "WIDE" || *req.ExtraType == "NO_BALL" {
			isLegalDelivery = false
		}
	}

	totalRuns :=
		req.RunsOffBat +
			req.ExtraRuns

	if req.ExtraType != nil && (*req.ExtraType == "WIDE" || *req.ExtraType == "NO_BALL") {
		totalRuns += 1
	}

	lastBallSequence, err :=
		dbHelper.GetLastBallSequence(
			match.CurrentInningID,
		)
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get last ball sequence",
		})
		return
	}

	nextBallSequence :=
		lastBallSequence + 1
	event := models.BallEventInsert{

		InningsID: match.CurrentInningID,

		BallSequence: nextBallSequence,

		OverNo:     (match.LegalBalls / 6) + 1,
		BallInOver: (match.LegalBalls % 6) + 1,

		StrikerID:    *match.StrikerID,
		NonStrikerID: *match.NonStrikerID,

		BowlerID: *match.BowlerID,

		RunsOffBat: req.RunsOffBat,
		ExtraRuns:  req.ExtraRuns,

		TotalRuns: totalRuns,

		ExtraType: req.ExtraType,

		IsLegalDelivery: isLegalDelivery,
		IsBoundaryFour:  req.RunsOffBat == 4,
		IsBoundarySix:   req.RunsOffBat == 6,
		IsDotBall:       totalRuns == 0,

		IsWicket:   req.IsWicket,
		WicketType: req.WicketType,

		DismissedPlayerID:    req.DismissedPlayerID,
		DismissedByFielderID: req.DismissedByFielderID,
	}

	inningsUpdate := models.InningsUpdate{

		TotalRunsIncrement: event.TotalRuns,
	}

	// wicket increment
	if event.IsWicket && event.WicketType != nil && *event.WicketType != "RETIRED_HURT" {
		inningsUpdate.WicketIncrement = 1
	}

	// legal ball increment
	if event.IsLegalDelivery {
		inningsUpdate.LegalBallIncrement = 1
	}

	// extras breakdown
	if event.ExtraType != nil {

		inningsUpdate.ExtrasIncrement = event.ExtraRuns

		switch *event.ExtraType {

		case "WIDE":

			inningsUpdate.WidesIncrement = 1

			inningsUpdate.ExtrasIncrement = event.ExtraRuns + 1

		case "NO_BALL":

			inningsUpdate.NoBallsIncrement = 1

			inningsUpdate.ExtrasIncrement = 1

		case "BYE":

			inningsUpdate.ByesIncrement = event.ExtraRuns
			inningsUpdate.ExtrasIncrement = event.ExtraRuns

		case "LEG_BYE":

			inningsUpdate.LegByesIncrement = event.ExtraRuns

			inningsUpdate.ExtrasIncrement = event.ExtraRuns
		}
	}

	// Updating Batting scorecard after adding ball in ball_event
	battingUpdate := models.BattingScorecardUpdate{
		RunsIncrement: event.RunsOffBat,
	}

	if event.IsLegalDelivery {
		battingUpdate.BallsIncrement = 1
	}

	if event.RunsOffBat == 4 {
		battingUpdate.FoursIncrement = 1
	}

	if event.RunsOffBat == 6 {

		battingUpdate.SixesIncrement = 1
	}
	if event.IsWicket && event.WicketType != nil && *event.WicketType != "RETIRED_HURT" {
		battingUpdate.IsOut = true

		battingUpdate.DismissalType = event.WicketType

		battingUpdate.DismissedByBowlerID = &event.BowlerID

		if event.DismissedByFielderID != nil {
			battingUpdate.FielderID = event.DismissedByFielderID
		}
	}

	//---------------Updating Bowling Table

	bowlingUpdate := models.BowlingScorecardUpdate{}

	if event.IsLegalDelivery {

		bowlingUpdate.LegalBallsIncrement = 1
	}

	bowlingUpdate.RunsConcededIncrement =
		event.TotalRuns

	if event.ExtraType != nil {

		switch *event.ExtraType {

		case "BYE", "LEG_BYE":
			bowlingUpdate.RunsConcededIncrement = 0
		}
	}

	if event.ExtraType != nil {

		if *event.ExtraType == "WIDE" {
			bowlingUpdate.WidesIncrement = 1
		}

		if *event.ExtraType == "NO_BALL" {
			bowlingUpdate.NoBallsIncrement = 1
		}
	}

	if event.IsWicket && event.WicketType != nil {

		switch *event.WicketType {

		case "BOWLED",
			"CAUGHT",
			"LBW",
			"HIT_WICKET",
			"STUMPED":

			bowlingUpdate.WicketsIncrement = 1
		}
	}

	//------------updating live_match table
	liveMatchUpdate := models.LiveMatchUpdate{

		TotalRunsIncrement:    inningsUpdate.TotalRunsIncrement,
		TotalWicketsIncrement: inningsUpdate.WicketIncrement,
		LegalBallsIncrement:   inningsUpdate.LegalBallIncrement,
	}

	liveMatchUpdate.BowlerID = event.BowlerID

	// strikee rotation logic
	newStrikerID := event.StrikerID
	newNonStrikerID := event.NonStrikerID

	if IsStrikeRotating(event) {

		newStrikerID = event.NonStrikerID

		newNonStrikerID = event.StrikerID
	}

	newLegalBalls := match.LegalBalls

	if event.IsLegalDelivery {
		newLegalBalls++
	}

	newTotalRuns := match.CurrentTotalRuns + inningsUpdate.TotalRunsIncrement

	//innings ending logic...
	newTotalWickets := match.CurrentTotalWickets + inningsUpdate.WicketIncrement
	isAllOut := newTotalWickets >= maxWickets
	isOversCompleted := newLegalBalls >= match.Overs*6
	isSecondInnings := match.CurrentInningNo == 2
	previousInningsScore := 0
	if match.PreviousInningsScore != nil {
		previousInningsScore = *match.PreviousInningsScore
	}
	isTargetChased := isSecondInnings && newTotalRuns > previousInningsScore

	isInningsCompleted := isAllOut || isOversCompleted || isTargetChased

	isMatchCompleted := isSecondInnings && isInningsCompleted
	var winnerTeamID string

	if isMatchCompleted {
		firstInningsScore := 0
		if match.PreviousInningsScore != nil {
			firstInningsScore = *match.PreviousInningsScore
		}
		secondInningsScore := newTotalRuns
		if secondInningsScore > firstInningsScore {
			winnerTeamID =
				match.BattingTeamID
		} else if secondInningsScore < firstInningsScore {
			winnerTeamID =
				match.BowlingTeamID
		}
	}

	liveMatchUpdate.StrikerID = newStrikerID
	liveMatchUpdate.NonStrikerID = newNonStrikerID

	/// new bowler selection on over complition
	isOverCompleted := event.IsLegalDelivery && newLegalBalls%6 == 0
	if isOverCompleted &&
		!isInningsCompleted {

		if req.NextBowlerID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "next_bowler_id required",
			})
			return
		}

		if req.NextBowlerID == event.BowlerID {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "bowler cannot bowl consecutive overs",
			})
			return
		}

		belongsToBowlingTeam, err := dbHelper.IsPlayerInTeam(match.BowlingTeamID, req.NextBowlerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		if !belongsToBowlingTeam {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "player not in bowling team",
			})
			return
		}

		if req.NextBowlerID == liveMatchUpdate.StrikerID {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "bowler cannot be striker",
			})
			return
		}

		if req.NextBowlerID == liveMatchUpdate.NonStrikerID {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "bowler cannot be non striker",
			})
			return
		}

		liveMatchUpdate.BowlerID = req.NextBowlerID
	}

	//end over strike rotation logic
	if newLegalBalls > 0 && newLegalBalls%6 == 0 {
		newStrikerID, newNonStrikerID = newNonStrikerID, newStrikerID
	}

	//-----checks that player is not already out
	if event.IsWicket && !isInningsCompleted {
		if req.NextBatsmanID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "next_batsman_id required",
			})
			return
		}

		if req.IsWicket {

			if req.WicketType == nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "wicket_type  required",
				})
				return
			}

			if req.DismissedPlayerID == nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "dismissed_player_id required",
				})
				return
			}
		}

		// next striker cannot be current striker
		if req.NextBatsmanID == newStrikerID {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "next batsman already batting",
			})
			return
		}
		if req.NextBatsmanID == newNonStrikerID {

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "next batsman already batting",
			})
			return
		}

		if req.NextBatsmanID == event.BowlerID {

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "bowler cannot bat",
			})
			return
		}

		if req.DismissedPlayerID != nil && req.NextBatsmanID == *req.DismissedPlayerID {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "dismissed player cannot continue batting",
			})
			return
		}

		belongsToBattingTeam, err := dbHelper.IsPlayerInTeam(match.BattingTeamID, req.NextBatsmanID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		if !belongsToBattingTeam {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "player not in batting team",
			})
			return
		}

		isAlreadyOut, err := dbHelper.IsPlayerAlreadyOut(event.InningsID, req.NextBatsmanID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		if isAlreadyOut {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "player already out",
			})
			return
		}

		// replace dismissed batsman
		if req.DismissedPlayerID != nil && *req.DismissedPlayerID == newStrikerID {
			newStrikerID = req.NextBatsmanID
		} else if req.DismissedPlayerID != nil && *req.DismissedPlayerID == newNonStrikerID {
			newNonStrikerID = req.NextBatsmanID
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "dismissed player not currently batting",
			})
			return
		}

		liveMatchUpdate.StrikerID = newStrikerID

		liveMatchUpdate.NonStrikerID = newNonStrikerID
	}
	//fmt.Printf("ExtraType: %#v\n", event.ExtraType)

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		//fmt.Println("1")
		err = dbHelper.InsertBallEvent(tx, event)
		if err != nil {
			return err
		}

		//fmt.Println("2")
		err = dbHelper.UpdateInningsAfterBall(tx, event.InningsID, inningsUpdate)
		if err != nil {
			return err
		}

		//fmt.Println("3")
		// to deal with the issue of getting non-striker out and upding striker as out
		if event.IsWicket && event.WicketType != nil && *event.WicketType == "RUN_OUT" &&
			event.DismissedPlayerID != nil && *event.DismissedPlayerID != event.StrikerID {

			strikerUpdate := models.BattingScorecardUpdate{
				RunsIncrement:  battingUpdate.RunsIncrement,
				BallsIncrement: battingUpdate.BallsIncrement,
				FoursIncrement: battingUpdate.FoursIncrement,
				SixesIncrement: battingUpdate.SixesIncrement,
			}

			err = dbHelper.UpdateBattingScorecardAfterBall(tx, event.InningsID, event.StrikerID, strikerUpdate)
			if err != nil {
				return err
			}

			dismissedUpdate := models.BattingScorecardUpdate{
				IsOut:               true,
				DismissalType:       battingUpdate.DismissalType,
				DismissedByBowlerID: battingUpdate.DismissedByBowlerID,
				FielderID:           battingUpdate.FielderID,
			}

			err = dbHelper.UpdateBattingScorecardAfterBall(tx, event.InningsID,
				*event.DismissedPlayerID, dismissedUpdate)
			if err != nil {
				return err
			}

		} else {

			err = dbHelper.UpdateBattingScorecardAfterBall(
				tx,
				event.InningsID,
				event.StrikerID,
				battingUpdate,
			)
			if err != nil {
				return err
			}
		}

		//fmt.Println("4")

		err = dbHelper.UpdateBowlingScorecardAfterBall(tx, event.InningsID, event.BowlerID, bowlingUpdate)
		if err != nil {
			return err
		}

		//fmt.Println("5")
		err = dbHelper.UpdateLiveMatchAfterBall(tx, match.MatchID, liveMatchUpdate)
		if err != nil {
			return err
		}

		//fmt.Println("6")
		if isInningsCompleted {
			err = dbHelper.CompleteInnings(tx, event.InningsID)
			if err != nil {
				return err
			}
		}

		//fmt.Println("7")
		if isMatchCompleted {
			err = dbHelper.CompleteMatch(tx, match.MatchID, winnerTeamID)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": txErr.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ball event added successfully",
	})
}

func validateBallEventRequest(req models.AddBallEventRequest) error {

	if req.MatchID == "" {
		return fmt.Errorf("match_id required")
	}

	if req.RunsOffBat < 0 {
		return fmt.Errorf("invalid runs_off_bat")
	}

	if req.ExtraRuns < 0 {
		return fmt.Errorf("invalid extra_runs")
	}
	//for the times when is_wicket is false but someone sends wicket type in request
	if !req.IsWicket {

		if req.WicketType != nil {
			return fmt.Errorf("wicket_type should be empty")
		}

		if req.DismissedPlayerID != nil {
			return fmt.Errorf("dismissed_player_id should be empty")
		}
	}
	return nil
}

func validateMatchState(match *models.MatchResponse) error {

	if match.WinnerTeamID != nil {
		return fmt.Errorf("match already completed")
	}

	if match.CurrentInningCompleted {
		return fmt.Errorf("innings already completed")
	}

	playerCount, err := dbHelper.GetTeamPlayerCount(
		match.BattingTeamID,
	)
	if err != nil {
		return err
	}
	maxWickets := playerCount - 1

	if match.CurrentTotalWickets >= maxWickets {
		return fmt.Errorf("innings already all out")
	}

	if match.LegalBalls >= match.Overs*6 {
		return fmt.Errorf("overs already completed")
	}

	if match.StrikerID == nil {
		return fmt.Errorf("striker not selected")
	}

	if match.NonStrikerID == nil {
		return fmt.Errorf("non striker not selected")
	}

	if match.BowlerID == nil {
		return fmt.Errorf("bowler not selected")
	}

	// striker already out
	strikerOut, err := dbHelper.IsPlayerOut(
		match.CurrentInningID,
		*match.StrikerID,
	)
	if err != nil {
		return err
	}
	if strikerOut {
		return fmt.Errorf("striker already out")
	}

	//checks if non striker already out
	nonStrikerOut, err := dbHelper.IsPlayerOut(
		match.CurrentInningID,
		*match.NonStrikerID,
	)
	if err != nil {
		return err
	}
	if nonStrikerOut {
		return fmt.Errorf("non striker already out")
	}

	// bowler validation
	if *match.BowlerID == *match.StrikerID {
		return fmt.Errorf("bowler cannot be striker")
	}

	if *match.BowlerID == *match.NonStrikerID {
		return fmt.Errorf("bowler cannot be non striker")
	}

	//check if striker belongs to the correct team
	strikerInBattingTeam, err := dbHelper.IsPlayerInTeam(
		match.BattingTeamID,
		*match.StrikerID,
	)
	if err != nil {
		return err
	}

	if !strikerInBattingTeam {
		return fmt.Errorf("striker not in batting team")
	}
	//checks if nonstriker belongs to correct team
	nonStrikerInBattingTeam, err := dbHelper.IsPlayerInTeam(
		match.BattingTeamID,
		*match.NonStrikerID,
	)
	if err != nil {
		return err
	}

	if !nonStrikerInBattingTeam {
		return fmt.Errorf("non striker not in batting team")
	}

	//checks if bowler belongs to correct team
	bowlerInBowlingTeam, err := dbHelper.IsPlayerInTeam(
		match.BowlingTeamID,
		*match.BowlerID,
	)
	if err != nil {
		return err
	}

	if !bowlerInBowlingTeam {
		return fmt.Errorf("bowler not in bowling team")
	}

	return nil
}

func IsStrikeRotating(event models.BallEventInsert) bool {

	if event.ExtraType != nil {
		switch *event.ExtraType {
		case "WIDE", "NO_BALL":
			return event.ExtraRuns%2 == 1
		case "BYE", "LEG_BYE":
			return event.ExtraRuns%2 == 1
		}
	}
	return event.RunsOffBat%2 == 1
}

func StartSecondInnings(c *gin.Context) {
	var req models.StartSecondInningsRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	match, err := dbHelper.GetMatchByID(req.MatchID)
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	if match == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "match not found",
		})
		return
	}
	if match.CurrentInningNo != 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "second innings already started",
		})
		return
	}
	if !match.CurrentInningCompleted {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "first innings not completed",
		})
		return
	}

	if req.StrikerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "striker_id required",
		})
		return
	}

	if req.NonStrikerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "non_striker_id required",
		})
		return
	}

	if req.BowlerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bowler_id required",
		})
		return
	}
	if req.StrikerID == req.NonStrikerID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "striker and non_striker cannot be same",
		})
		return
	}

	//swapping both the teams --- for second innnings
	battingTeamID := match.BowlingTeamID
	bowlingTeamID := match.BattingTeamID

	IsStrikerInBattingTeam, err := dbHelper.IsPlayerInTeam(battingTeamID, req.StrikerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	if !IsStrikerInBattingTeam {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "striker is not in batting team",
		})
		return
	}

	IsNonStrikerInBattingTeam, err := dbHelper.IsPlayerInTeam(battingTeamID, req.NonStrikerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	if !IsNonStrikerInBattingTeam {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "non striker is not in batting team",
		})
		return
	}

	IsBowlerInBowlingTeam, err := dbHelper.IsPlayerInTeam(bowlingTeamID, req.BowlerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !IsBowlerInBowlingTeam {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bowler is not in bowling team",
		})
		return
	}

	if req.BowlerID == req.StrikerID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bowler cannot be striker",
		})
		return
	}

	if req.BowlerID == req.NonStrikerID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "bowler cannot be non striker",
		})
		return
	}

	var secondInningsID string

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		secondInningsID, err = dbHelper.CreateInning(tx, match.MatchID, "2", battingTeamID, bowlingTeamID)
		if err != nil {
			return err
		}

		battingPlayers, err := dbHelper.GetTeamPlayers(battingTeamID)
		if err != nil {
			return err
		}

		for _, playerID := range battingPlayers {

			err = dbHelper.CreateBattingScorecard(tx, secondInningsID, playerID)
			if err != nil {
				return err
			}
		}

		bowlingPlayers, err :=
			dbHelper.GetTeamPlayers(bowlingTeamID)
		if err != nil {
			return err
		}

		for _, playerID := range bowlingPlayers {
			err = dbHelper.CreateBowlingScorecard(tx, secondInningsID, playerID)
			if err != nil {
				return err
			}
		}

		err = dbHelper.UpdateMatchCurrentInningsNo(tx, match.MatchID, 2)
		if err != nil {
			return err
		}

		err = dbHelper.ResetLiveMatchForSecondInnings(tx, match.MatchID, secondInningsID, req.StrikerID, req.NonStrikerID, req.BowlerID)
		if err != nil {
			return err
		}
		return nil
	})

	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": txErr.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "second innings started",

		"second_innings_id": secondInningsID,
	})
}

func GetScorecard(c *gin.Context) {

	matchID := c.Param("matchID")
	match, err := dbHelper.GetMatchByID(matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if match == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "match not found",
		})
		return
	}

	inningsList, err := dbHelper.GetMatchInnings(matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	response := models.MatchScorecardResponse{
		MatchID: matchID,
	}

	for i := range inningsList {
		batting, err := dbHelper.GetBattingScorecard(inningsList[i].InningsID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		bowling, err := dbHelper.GetBowlingScorecard(inningsList[i].InningsID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		inningsList[i].Batting = batting
		inningsList[i].Bowling = bowling
		if inningsList[i].InningsNo == 1 {
			response.FirstInnings = &inningsList[i]
		} else if inningsList[i].InningsNo == 2 {
			response.SecondInnings = &inningsList[i]
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"scorecard": response,
	})
}

func GetBallEvents(c *gin.Context) {
	inningsID := c.Param("inningsID")
	if inningsID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "inningsID required",
		})
		return
	}
	ballEvents, err := dbHelper.GetBallEvents(inningsID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ball_events": ballEvents,
	})
}
