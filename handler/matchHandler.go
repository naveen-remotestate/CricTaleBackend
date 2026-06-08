package handler

import (
	"CricTail_Backend/database"
	"CricTail_Backend/database/dbHelper"
	"CricTail_Backend/models"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

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

		//create first innings table
		now := time.Now()
		//fmt.Printf("now = %#v\n", now)
		InningID, err := dbHelper.CreateInning(
			tx,
			matchID,
			"1",
			battingFirstTeamID,
			bowlingFirstTeamID,
			&now,
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
	if match.CurrentInningsNo != 1 {
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
		now := time.Now()
		secondInningsID, err = dbHelper.CreateInning(tx, match.MatchID, "2", battingTeamID, bowlingTeamID, &now)
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
