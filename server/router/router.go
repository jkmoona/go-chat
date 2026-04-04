package router

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/jkmoona/go-chat/server/internal/auth"
	"github.com/jkmoona/go-chat/server/internal/config"
	"github.com/jkmoona/go-chat/server/internal/middleware"
	"github.com/jkmoona/go-chat/server/internal/room"
	"github.com/jkmoona/go-chat/server/internal/user"
	"github.com/jkmoona/go-chat/server/internal/ws"
)

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "same-origin")
		c.Next()
	}
}

func NewRouter(cfg *config.Config, userHandler *user.Handler, roomHandler *room.Handler, wsHandler *ws.Handler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger(slog.Default()))
	r.Use(securityHeaders())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.ClientURL},
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	authLimiter := middleware.NewRateLimiter(5, 10)
	roomLimiter := middleware.NewRateLimiter(2, 5)

	r.POST("/register", authLimiter.Middleware(), userHandler.CreateUser)
	r.POST("/login", authLimiter.Middleware(), userHandler.Login)
	r.GET("/logout", userHandler.Logout)
	r.POST("/refresh", userHandler.RefreshToken)

	r.GET("/ws/room/:roomId", auth.OptionalAuthMiddleware(), roomHandler.GetRoom)
	r.POST("/ws/room/:roomId/verify", roomHandler.VerifyPIN)
	r.GET("/ws/guest/joinRoom/:roomId", authLimiter.Middleware(), wsHandler.GuestJoinRoom)

	wsGroup := r.Group("/ws")
	wsGroup.Use(auth.AuthMiddleware())
	{
		wsGroup.POST("/createRoom", roomLimiter.Middleware(), roomHandler.CreateRoom)
		wsGroup.GET("/getRooms", roomHandler.ListRooms)
		wsGroup.GET("/joinRoom/:roomId", wsHandler.JoinRoom)
		wsGroup.GET("/getClients/:roomId", wsHandler.GetClients)
		wsGroup.PATCH("/room/:roomId", roomHandler.ExtendRoom)
		wsGroup.DELETE("/room/:roomId", roomHandler.DeleteRoom)
		wsGroup.POST("/room/:roomId/kick", roomHandler.KickClient)
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})

	return r
}
