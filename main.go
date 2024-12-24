package main

import (
	"openeyes/config"
	"openeyes/controllers"
	"openeyes/services"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize services
	searchService := services.NewSearchService(cfg)

	// Initialize controller
	searchController := controllers.NewSearchController(searchService)
// Test koneksi ke ELK
    if err := searchService.testElkConnection(); err != nil {
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
	{
		api.GET("/search", searchController.Search)
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
	}

	// Start server
	r.Run(":8080")
}
