package main

import (
	"log"
	"os"

	"github.com/jkmoona/go-chat/server/db"
	"github.com/jkmoona/go-chat/server/internal/user"
	"github.com/jkmoona/go-chat/server/internal/ws"
	"github.com/jkmoona/go-chat/server/router"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	dbConn, err := db.NewDatabase()

	if err != nil {
		log.Fatalf("Could not initialize database connection: %s", err)
	}

	userRep := user.NewRepository(dbConn.GetDB())
	userSvc := user.NewService(userRep)
	userHandler := user.NewHandler(userSvc)

	hub := ws.NewHub()
	wsHandler := ws.NewHandler(hub)
	go hub.Run()

	router.InitRouter(userHandler, wsHandler)

	// Use Railway port when available
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on port %s...", port)
	router.Start(":" + port)
}
