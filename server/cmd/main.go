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

	hub := ws.NewHub()
	wsHandler := ws.NewHandler(hub, cfg.ClientURL)
	go hub.Run()

	r := router.NewRouter(cfg, userHandler, wsHandler)

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
