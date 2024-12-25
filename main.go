package main

import (
	"fmt"
	"openeyes/config"
	"openeyes/controllers"
	"openeyes/services"
	"github.com/gin-gonic/gin"
	"openeyes/database"
    "openeyes/handlers"
	"openeyes/middleware"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	db, err := database.InitDB()
    if err != nil {
        fmt.Printf("Failed to initialize database:", err)
    }
    defer db.Close()
	fmt.Println("Successfully connected to sql")


	// Initialize services
	searchService := services.NewSearchService(cfg)

	// Initialize controller
	searchController := controllers.NewSearchController(searchService)

    // Test koneksi ke ELK
    if err := searchService.TestElkConnection(); err != nil {
        fmt.Printf("Failed to connect to Elasticsearch: %v\n", err)
        return
    }
    fmt.Println("Successfully connected to Elasticsearch")

	// Setup Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Routes
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("/search", searchController.Search)
		api.POST("/login", handlers.LoginHandler)
    	api.POST("/logout", handlers.LogoutHandler)
    	api.POST("/create-user", handlers.CreateUserHandler)
    	api.POST("/reset-password", handlers.ResetPasswordHandler)
    	api.POST("/change-password", handlers.ChangePasswordHandler)
		api.POST("/delete-user", handlers.DeleteUserHandler)
		api.GET("/getall-users", handlers.GetAllUsersHandler)
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
	}
	


	// Start server
	r.Run(":8080")
}