package handler

import (
	"CricTail_Backend/database"
	"CricTail_Backend/database/dbHelper"
	"CricTail_Backend/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

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
	trueVal := true
	if event.IsWicket && event.WicketType != nil && *event.WicketType != "RETIRED_HURT" {
		battingUpdate.IsOut = &trueVal

		battingUpdate.DismissalType = event.WicketType

		battingUpdate.DismissedByBowlerID = &event.BowlerID

		if event.DismissedByFielderID != nil {
			battingUpdate.FielderID = event.DismissedByFielderID
		}
	}

	//updating Bowling Table

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
	isSecondInnings := match.CurrentInningsNo == 2
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
	if isOverCompleted && !isInningsCompleted {

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
	if isOverCompleted {
		newStrikerID, newNonStrikerID = newNonStrikerID, newStrikerID
		liveMatchUpdate.StrikerID = newStrikerID
		liveMatchUpdate.NonStrikerID = newNonStrikerID
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

			trueVal := true
			dismissedUpdate := models.BattingScorecardUpdate{
				IsOut:               &trueVal,
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

			if match.CurrentInningsNo == 2 {
				var winnerTeamID *string
				if newTotalRuns > *match.PreviousInningsScore {
					winnerTeamID = &match.BattingTeamID
				} else if newTotalRuns < *match.PreviousInningsScore {
					winnerTeamID = &match.BowlingTeamID
				} else {
					winnerTeamID = nil
				}
				err = dbHelper.CompleteMatch(tx, match.MatchID, winnerTeamID)
				if err != nil {
					return err
				}

			}
		}

		//fmt.Println("7")
		if isMatchCompleted {
			err = dbHelper.CompleteMatch(tx, match.MatchID, &winnerTeamID)
			if err != nil {
				return err
			}

		}
		if isMatchCompleted {
			err = ProcessPlayerCareerStats(tx, match.MatchID, &winnerTeamID)
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

func RetiredHurt(c *gin.Context) {
	matchID := c.Param("matchID")
	var req models.RetiredHurtRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	match, err := dbHelper.GetMatchByID(matchID)
	if err != nil || match == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "match not found"})
		return
	}

	newStrikerID := *match.StrikerID
	newNonStrikerID := *match.NonStrikerID

	if req.RetiredPlayerID == newStrikerID {
		newStrikerID = req.NextBatsmanID
	} else if req.RetiredPlayerID == newNonStrikerID {
		newNonStrikerID = req.NextBatsmanID
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "retired player currently not batting"})
		return
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		retire := "RETIRED_HURT"
		retiredUpdate := models.BattingScorecardUpdate{
			DismissalType: &retire,
		}
		err = dbHelper.UpdateBattingScorecardAfterBall(tx, match.CurrentInningID, req.RetiredPlayerID, retiredUpdate)
		if err != nil {
			return err
		}

		nextUpdate := models.BattingScorecardUpdate{
			ClearDismissal: true,
		}
		err = dbHelper.UpdateBattingScorecardAfterBall(tx, match.CurrentInningID, req.NextBatsmanID, nextUpdate)
		if err != nil {
			return err
		}

		liveUpdate := models.LiveMatchUpdate{
			StrikerID:    newStrikerID,
			NonStrikerID: newNonStrikerID,
			BowlerID:     *match.BowlerID,
		}
		err = dbHelper.UpdateLiveMatchAfterBall(tx, matchID, liveUpdate)
		if err != nil {
			return err
		}
		return nil
	})

	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": txErr.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "player retired successfully"})
}
