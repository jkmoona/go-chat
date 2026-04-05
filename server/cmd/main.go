package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jkmoona/go-chat/server/db"
	"github.com/jkmoona/go-chat/server/internal/auth"
	"github.com/jkmoona/go-chat/server/internal/config"
	"github.com/jkmoona/go-chat/server/internal/room"
	"github.com/jkmoona/go-chat/server/internal/user"
	"github.com/jkmoona/go-chat/server/internal/ws"
	"github.com/jkmoona/go-chat/server/router"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	var handler slog.Handler
	if os.Getenv("GIN_MODE") == "release" {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)

	if err := auth.Setup(cfg.AccessTokenSecret, cfg.RefreshTokenSecret, cfg.SecureCookies); err != nil {
		slog.Error("failed to setup auth", slog.String("error", err.Error()))
		os.Exit(1)
	}

	dbConn, err := db.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		slog.Error("could not initialize database", slog.String("error", err.Error()))
		os.Exit(1)
	}

	userRep := user.NewRepository(dbConn.GetDB())
	userSvc := user.NewService(userRep)
	userHandler := user.NewHandler(userSvc)

	roomRep := room.NewRepository(dbConn.GetDB())
	roomSvc := room.NewService(roomRep)

	hub := ws.NewHub(func(roomID string) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := roomSvc.Deactivate(ctx, roomID); err != nil {
			slog.Error("failed to deactivate room", slog.String("room_id", roomID), slog.String("error", err.Error()))
		}
	})

	go hub.Run()

	loadActiveRooms(roomSvc, hub)

	roomHandler := room.NewHandler(roomSvc, hub)

	verifyPIN := func(roomID, pin string) error {
		return roomSvc.VerifyPIN(context.Background(), roomID, pin)
	}
	wsHandler := ws.NewHandler(hub, cfg.ClientURL, verifyPIN)

	r := router.NewRouter(cfg, userHandler, roomHandler, wsHandler)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		slog.Info("server starting", slog.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	slog.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server forced to shutdown", slog.String("error", err.Error()))
	}

	hub.Close()
	dbConn.Close()
	slog.Info("server stopped")
}

func loadActiveRooms(roomSvc room.Service, hub *ws.Hub) {
	rooms, err := roomSvc.ListActiveRooms(context.Background())
	if err != nil {
		slog.Error("failed to load active rooms", slog.String("error", err.Error()))
		return
	}

	for _, r := range rooms {
		hub.CreateRoom(&ws.Room{
			ID:        r.ID,
			Name:      r.Name,
			Clients:   make(map[string]*ws.Client),
			ExpiresAt: r.ExpiresAt,
			HasPIN:    r.PinHash != "",
		})
	}

	if len(rooms) > 0 {
		slog.Info("loaded active rooms from database", slog.Int("count", len(rooms)))
	}
}
