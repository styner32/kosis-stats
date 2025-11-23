package routes

import (
	"kosis/internal/config"
	"kosis/internal/controllers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter initializes all services, controllers, and API routes
func SetupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	financialController := controllers.FinancialController{DB: db}

	// Set up Gin router
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Simple health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP"})
	})

	// Group API routes under /api/v1
	api := router.Group("/api/v1")
	{
		// Companies endpoints
		api.GET("/companies", financialController.GetCompanies)

		// Reports endpoints
		api.GET("/reports/:corp_code", financialController.GetReports)
	}

	return router
}
