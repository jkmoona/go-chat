package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jkmoona/go-chat/server/internal/auth"
	"github.com/jkmoona/go-chat/server/internal/user"
	"github.com/jkmoona/go-chat/server/internal/ws"

	"time"

	"github.com/gin-contrib/cors"
)

var r *gin.Engine

func InitRouter(userHandler *user.Handler, wsHandler *ws.Handler) {
	r = gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://tempchatgo.netlify.app"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.POST("/register", userHandler.CreateUser)
	r.POST("/login", userHandler.Login)
	r.GET("/logout", userHandler.Logout)
	r.POST("/refresh", userHandler.RefreshToken)

	wsGroup := r.Group("/ws")
	wsGroup.Use(auth.AuthMiddleware())
	{
		wsGroup.POST("/createRoom", wsHandler.CreateRoom)
		wsGroup.GET("/joinRoom/:roomId", wsHandler.JoinRoom)
		wsGroup.GET("/getRooms", wsHandler.GetRooms)
		wsGroup.GET("/getClients/:roomId", wsHandler.GetClients)
	}
}

func Start(addr string) error {
	return r.Run(addr)
}
