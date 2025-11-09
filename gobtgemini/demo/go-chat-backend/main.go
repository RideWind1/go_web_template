package main

import (
	"go-chat-backend/config"
	"go-chat-backend/database"
	"go-chat-backend/handlers"
	"go-chat-backend/middleware"
	"go-chat-backend/services"
	"go-chat-backend/websocket"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// 加载配置
	config.LoadConfig()

	// 初始化日志
	setupLogger()

	// 连接数据库
	db, err := database.InitDB()
	if err != nil {
		logrus.Fatal("数据库连接失败:", err)
	}

	// 初始化服务
	userService := services.NewUserService(db)
	chatService := services.NewChatService(db)
	llmService := services.NewLLMService()
	chromaService,err_chroma := services.NewChromaService()
	if err_chroma != nil {
        logrus.Fatalf("初始化Chroma服务失败: %v", err_chroma)
    }

	// 初始化处理器
	authHandler := handlers.NewAuthHandler(userService)
	chatHandler := handlers.NewChatHandler(chatService, llmService, chromaService)
	chatHandler.SetUserService(userService) // 设置用户服务

	// 初始化WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()
	wsHandler := websocket.NewHandler(hub)
	chatHandler.SetWebSocketHub(hub) // 设置WebSocket Hub

	// 设置路由
	router := setupRouter(authHandler, chatHandler, wsHandler)

	// 启动服务器
	port := config.GetString("PORT", "8080")
	logrus.Info("服务器启动在端口: ", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func setupLogger() {
	logLevel := config.GetString("LOG_LEVEL", "info")
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.JSONFormatter{})
}

func setupRouter(authHandler *handlers.AuthHandler, chatHandler *handlers.ChatHandler, wsHandler *websocket.Handler) *gin.Engine {
	// 设置Gin模式
	ginMode := config.GetString("GIN_MODE", "debug")
	gin.SetMode(ginMode)

	router := gin.Default()

	// CORS中间件
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"message": "Go Chat Backend is running",
		})
	})

	// API路由组
	api := router.Group("/api/v1")
	{
		// 认证相关路由
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", middleware.JWTAuthMiddleware(), authHandler.RefreshToken)
		}

		// 受保护的路由
		protected := api.Group("/")
		protected.Use(middleware.JWTAuthMiddleware())
		{
			// 用户相关
			protected.GET("/user/profile", authHandler.GetProfile)
			protected.PUT("/user/profile", authHandler.UpdateProfile)

			// 聊天相关
			chat := protected.Group("/chat")
			{
				chat.POST("/send", chatHandler.SendMessage)
				chat.GET("/history", chatHandler.GetChatHistory)
				chat.DELETE("/history/:id", chatHandler.DeleteMessage)
				chat.POST("/clear", chatHandler.ClearHistory)
				chat.GET("/conversations", chatHandler.GetConversations)//获取对话列表
				chat.POST("/conversations", chatHandler.CreateConversation)//创建对话列表
				chat.GET("/conversations/:id/history", chatHandler.GetOneConversationHistory)
				
			}

			// WebSocket连接
			protected.GET("/ws/chat", wsHandler.HandleWebSocket)
		}
	}

	return router
}
