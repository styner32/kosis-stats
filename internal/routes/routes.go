package routes

import (
	"kosis/internal/config"
	"kosis/internal/controllers"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter initializes all services, controllers, and API routes
func SetupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	financialController := controllers.FinancialController{DB: db}

	// Set up Gin router
	router := gin.Default()

	// Parse allowed origins
	var allowedOrigins []string
	if cfg != nil && cfg.AllowedOrigins != "" {
		allowedOrigins = strings.Split(cfg.AllowedOrigins, ",")
	}

	// CORS middleware
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		allowed := false
		if origin != "" {
			for _, o := range allowedOrigins {
				if origin == strings.TrimSpace(o) {
					allowed = true
					break
				}
			}
		}

		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

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

		// MCP-friendly endpoints
		api.GET("/mcp/reports/by-corp-name", financialController.GetReportsByCorpName)

		// Reports by corp code endpoints
		api.GET("/reports/:corp_code", financialController.GetReportsByCorpCode)

		// Raw reports endpoints
		api.GET("/reports/:corp_code/:raw_report_id", financialController.GetRawReport)

		// Summary + raw report by receipt number
		api.GET("/reports/receipt/:receipt_number", financialController.GetReportSummaryByReceiptNumber)

		// Summary + raw report by receipt number
		api.GET("/mcp/reports/receipt/:receipt_number", financialController.GetReportSummaryByReceiptNumber)

		// Reports endpoints
		api.GET("/reports", financialController.GetAllReports)

		// Reports endpoints
		api.GET("/mcp/reports", financialController.GetAllReports)
	}

	return router
}
