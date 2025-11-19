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

	// Simple health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "UP"})
	})

	// Group API routes under /api/v1
	api := router.Group("/api/v1")
	{
		// Financials group
		financials := api.Group("/financials")
		{
			// GET /api/v1/financials/:corp_code
			// Retrieves stored financial data for a company
			financials.GET("/", financialController.GetFinancialsByCorpCode)
		}
	}

	return router
}
