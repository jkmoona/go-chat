package router

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/jkmoona/go-chat/server/internal/auth"
	"github.com/jkmoona/go-chat/server/internal/config"
	"github.com/jkmoona/go-chat/server/internal/httputil"
	"github.com/jkmoona/go-chat/server/internal/middleware"
	"github.com/jkmoona/go-chat/server/internal/room"
	"github.com/jkmoona/go-chat/server/internal/user"
	"github.com/jkmoona/go-chat/server/internal/ws"
)

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; connect-src 'self' ws: wss:; style-src 'self' 'unsafe-inline'; img-src 'self' data:")
		next.ServeHTTP(w, r)
	})
}

func NewRouter(cfg *config.Config, userHandler *user.Handler, roomHandler *room.Handler, wsHandler *ws.Handler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RequestLogger(slog.Default()))
	r.Use(securityHeaders)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.ClientURL},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           int((12 * time.Hour).Seconds()),
	}))

	authLimiter := middleware.NewRateLimiter(5, 10)
	roomLimiter := middleware.NewRateLimiter(2, 5)

	r.With(authLimiter.Middleware).Post("/register", userHandler.CreateUser)
	r.With(authLimiter.Middleware).Post("/login", userHandler.Login)
	r.Get("/logout", userHandler.Logout)
	r.Post("/refresh", userHandler.RefreshToken)

	r.With(auth.OptionalAuthMiddleware).Get("/ws/room/{roomId}", roomHandler.GetRoom)
	r.With(authLimiter.Middleware).Post("/ws/room/{roomId}/verify", roomHandler.VerifyPIN)
	r.With(authLimiter.Middleware).Get("/ws/guest/joinRoom/{roomId}", wsHandler.GuestJoinRoom)

	r.Route("/ws", func(r chi.Router) {
		r.Use(auth.AuthMiddleware)
		r.With(roomLimiter.Middleware).Post("/createRoom", roomHandler.CreateRoom)
		r.Get("/getRooms", roomHandler.ListRooms)
		r.Get("/joinRoom/{roomId}", wsHandler.JoinRoom)
		r.Get("/getClients/{roomId}", wsHandler.GetClients)
		r.Patch("/room/{roomId}", roomHandler.ExtendRoom)
		r.Delete("/room/{roomId}", roomHandler.DeleteRoom)
		r.Post("/room/{roomId}/kick", roomHandler.KickClient)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		httputil.ErrorJSON(w, http.StatusNotFound, "not found")
	})

	return r
}
