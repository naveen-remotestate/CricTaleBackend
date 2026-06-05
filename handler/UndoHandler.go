package handler

import (
	"CricTail_Backend/database"
	"CricTail_Backend/database/dbHelper"
	"CricTail_Backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func UndoLastBall(c *gin.Context) {

	matchID := c.Param("matchID")

	match, err := dbHelper.GetMatchByID(matchID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	lastBall, err :=
		dbHelper.GetLastBallEvent(match.CurrentInningID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ball not found",
		})
		return
	}

	var dismissedBatsmanID string

	//updating innigs table after deletion
	inningsUpdate := models.InningsUpdate{
		TotalRunsIncrement: -lastBall.TotalRuns, //sending negative value as update to reduce total runs
	}

	if lastBall.IsWicket {
		if lastBall.DismissedPlayerID != nil {
			dismissedBatsmanID = *lastBall.DismissedPlayerID
		}
		if lastBall.WicketType != nil {
			switch *lastBall.WicketType {
			case "BOWLED",
				"CAUGHT",
				"LBW",
				"RUN_OUT",
				"HIT_WICKET",
				"STUMPED":
				inningsUpdate.WicketIncrement = -1
			}
		}
	}

	if lastBall.IsLegalDelivery {
		inningsUpdate.LegalBallIncrement = -1
	}

	if lastBall.ExtraRuns > 0 || lastBall.ExtraType != nil {

		inningsUpdate.ExtrasIncrement = -lastBall.ExtraRuns
		if lastBall.ExtraType != nil {
			switch *lastBall.ExtraType {

			case "WIDE":
				inningsUpdate.WidesIncrement = -1
				inningsUpdate.ExtrasIncrement -= 1

			case "NO_BALL":
				inningsUpdate.NoBallsIncrement = -1
				inningsUpdate.ExtrasIncrement -= 1

			case "BYE":
				inningsUpdate.ByesIncrement = -lastBall.ExtraRuns

			case "LEG_BYE":
				inningsUpdate.LegByesIncrement = -lastBall.ExtraRuns
			}
		}
	}

	liveMatchUpdate := models.LiveMatchUpdate{}
	liveMatchUpdate.TotalRunsIncrement = -lastBall.TotalRuns
	if lastBall.IsWicket &&
		lastBall.WicketType != nil {
		switch *lastBall.WicketType {
		case "BOWLED",
			"CAUGHT",
			"LBW",
			"RUN_OUT",
			"HIT_WICKET",
			"STUMPED":
			liveMatchUpdate.TotalWicketsIncrement = -1
		}
	}
	if lastBall.IsLegalDelivery {
		liveMatchUpdate.LegalBallsIncrement = -1
	}
	liveMatchUpdate.StrikerID = lastBall.StrikerID
	liveMatchUpdate.NonStrikerID = lastBall.NonStrikerID
	liveMatchUpdate.BowlerID = lastBall.BowlerID

	battingUpdate := models.BattingScorecardUpdate{}
	battingUpdate.RunsIncrement = -lastBall.RunsOffBat

	if lastBall.IsLegalDelivery {
		battingUpdate.BallsIncrement = -1
	}
	if lastBall.IsBoundaryFour {
		battingUpdate.FoursIncrement = -1
	}
	if lastBall.IsBoundarySix {
		battingUpdate.SixesIncrement = -1
	}

	//bowling scorecard table
	bowlingUpdate := models.BowlingScorecardUpdate{}

	bowlingUpdate.RunsConcededIncrement = -lastBall.TotalRuns

	if lastBall.ExtraType != nil {
		switch *lastBall.ExtraType {
		case "BYE", "LEG_BYE": //bye legbye runs do not adds to bowler account
			bowlingUpdate.RunsConcededIncrement = 0
		}
	}
	if lastBall.IsLegalDelivery {
		bowlingUpdate.LegalBallsIncrement = -1
	}
	if lastBall.IsWicket && lastBall.WicketType != nil {
		switch *lastBall.WicketType {
		case "BOWLED",
			"CAUGHT",
			"LBW",
			"STUMPED",
			"HIT_WICKET":
			bowlingUpdate.WicketsIncrement = -1
		}
	}
	if lastBall.ExtraType != nil {
		switch *lastBall.ExtraType {
		case "WIDE":
			bowlingUpdate.WidesIncrement = -1
		case "NO_BALL":
			bowlingUpdate.NoBallsIncrement = -1
		}
	}

	txErr := database.Tx(func(tx *sqlx.Tx) error {
		// first updating all the tables then at end we will delete the ball event
		// innings Table
		err := dbHelper.UpdateInningsAfterBall(tx, lastBall.InningsID, inningsUpdate)
		if err != nil {
			return err
		}
		//live match table
		err = dbHelper.UpdateLiveMatchAfterBall(tx, match.MatchID, liveMatchUpdate)
		if err != nil {
			return err
		}
		//batting scorecard table
		err = dbHelper.UpdateBattingScorecardAfterBall(tx, lastBall.InningsID, lastBall.StrikerID, battingUpdate)
		if err != nil {
			return err
		}

		if lastBall.IsWicket && lastBall.DismissedPlayerID != nil {
			falseVal := false
			restoreDismissal := models.BattingScorecardUpdate{
				IsOut:               &falseVal,
				DismissalType:       nil,
				DismissedByBowlerID: nil,
				FielderID:           nil,
				ClearDismissal:      true,
			}

			err = dbHelper.UpdateBattingScorecardAfterBall(tx, lastBall.InningsID, dismissedBatsmanID, restoreDismissal)
			if err != nil {
				return err
			}
		}

		err = dbHelper.UpdateBowlingScorecardAfterBall(tx, lastBall.InningsID, lastBall.BowlerID, bowlingUpdate)
		if err != nil {
			return err
		}

		err = dbHelper.ReopenInnings(tx, lastBall.InningsID) // called always because is_complete is being set to false which is already false
		if err != nil {
			return err
		}

		err = dbHelper.ReopenMatch(tx, match.MatchID) //same as reopen innings
		if err != nil {
			return err
		}

		err = dbHelper.DeleteBallEvent(tx, lastBall.ID)
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
		"message": "ball deleted",
	})
}
