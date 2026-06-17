package main

import (
	"CricTail_Backend/database"
	"CricTail_Backend/server"
	"fmt"
	"os"
)

func main() {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "naveen"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "naveen"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "crictail-db"
	}
	sslMode := database.SSLModeDisable
	serverPort := ":8080"

	err := database.ConnectAndMigrate(
		dbHost,
		dbPort,
		dbName,
		dbUser,
		dbPassword,
		sslMode,
	)
	if err != nil {
		fmt.Printf("Failed to initialise and Migrate database: %v", err)
	}

	router := server.StartServer(serverPort)

	fmt.Println("Server is Running on Port:", serverPort)
	_ = router.Run(serverPort)
}
