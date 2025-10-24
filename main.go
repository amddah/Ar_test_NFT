// Package main is the entry point of the application
package main

import (
	"log"

	"quizmasterapi/config"
	_ "quizmasterapi/docs"
	"quizmasterapi/handlers"
	"quizmasterapi/middleware"
	"quizmasterapi/models"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/gin-contrib/cors"
)

// @title           QuizMaster API
// @version         1.0
// @description     RESTful API for a mobile quiz application with role-based access, quiz lifecycle, scoring, and leaderboard features.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.quizmaster.io/support
// @contact.email  support@quizmaster.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	config.LoadConfig()

	// Setup logging
	if err := config.SetupLogger(); err != nil {
		log.Fatal("Failed to setup logger:", err)
	}

	// Connect to database
	if err := config.ConnectDatabase(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize Gin router
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		// Mettez ici l'URL de votre frontend.
		// Exemples : "http://localhost:3000" pour React, "http://localhost:4200" for Angular, "http://localhost:5173" for Vite
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
	// Initialize handlers
	userHandler := handlers.NewUserHandler()
	quizHandler := handlers.NewQuizHandler()
	attemptHandler := handlers.NewAttemptHandler()
	leaderboardHandler := handlers.NewLeaderboardHandler()

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public routes
	api := router.Group("/api/v1")
	{
		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok", "message": "QuizMaster API is running"})
		})

		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}
	}

	// Protected routes
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		// User routes
		users := protected.Group("/users")
		{
			users.GET("/profile", userHandler.GetProfile)
		}

		// Quiz routes
		quizzes := protected.Group("/quizzes")
		{
			quizzes.GET("", quizHandler.GetQuizzes)
			quizzes.GET("/:id", quizHandler.GetQuizByID)
			quizzes.POST("", quizHandler.CreateQuiz)
			quizzes.DELETE("/:id", quizHandler.DeleteQuiz)

			// Professor-only routes
			quizzes.PUT("/:id/:action",
				middleware.RequireRole(models.RoleProfessor),
				quizHandler.ApproveRejectQuiz)
		}

		// Quiz attempt routes (Students only)
		attempts := protected.Group("/attempts")
		attempts.Use(middleware.RequireRole(models.RoleStudent))
		{
			attempts.POST("/start", attemptHandler.StartAttempt)
			attempts.POST("/answer", attemptHandler.SubmitAnswer)
			attempts.PUT("/:id/complete", attemptHandler.CompleteAttempt)
			attempts.GET("/:id", attemptHandler.GetAttemptByID)
			attempts.GET("", attemptHandler.GetMyAttempts)
		}

	
		// Leaderboard routes
		leaderboards := protected.Group("/leaderboards")
		{
			leaderboards.GET("/quiz/:quiz_id", leaderboardHandler.GetQuizLeaderboard)
			leaderboards.GET("/quiz/:quiz_id/my-rank", leaderboardHandler.GetMyRank)
			leaderboards.GET("/global", leaderboardHandler.GetGlobalLeaderboard)
		}
	}

	// Start server
	port := ":" + config.AppConfig.ServerPort
	log.Printf("Server starting on port %s", port)
	if err := router.Run(port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
