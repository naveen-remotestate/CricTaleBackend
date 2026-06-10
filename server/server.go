package server

import (
	"CricTail_Backend/handler"
	"CricTail_Backend/middleware"
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func StartServer(serverPort string) (router *gin.Engine) {
	fmt.Println("Starting Server")
	router = gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://192.168.1.100:5173",
		}, AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"serverStatus": "Running",
		})
	})

	router.POST("/register", handler.RegisterUser)

	router.POST("/login", handler.LoginUser)
	router.POST("/forgot-password", handler.ForgotPassword)

	router.GET("/matches", handler.GetMatches)
	router.GET("/matches/:matchID", handler.GetMatchByID)

	router.GET("/matches/:matchID/scorecard", handler.GetScorecard)
	router.GET("/innings/:inningsID/ball-events", handler.GetBallEvents)

	auth := router.Group("/")
	auth.Use(middleware.AuthMiddleware())
	auth.POST("/register-guest", handler.RegisterGuest)
	auth.POST("/logout", handler.LogoutUser) //either POST(mostly) or DELETE
	auth.GET("/players", handler.GetPlayers)

	auth.POST("/matches/:matchID/retired-hurt", handler.RetiredHurt)

	auth.POST("/create-match", handler.CreateMatch)

	auth.POST("/ball-event", handler.AddBallEvent)
	auth.POST("/start-second-innings", handler.StartSecondInnings)
	auth.POST("/matches/:matchID/undo", handler.UndoLastBall)
	profile := auth.Group("/player")
	profile.GET("/stats", handler.GetPlayerStats)
	profile.PUT("/update", handler.UpdatePlayerProfile)

	return

}
