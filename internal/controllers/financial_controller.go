package controllers

import (
	"kosis/internal/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FinancialController struct {
	DB *gorm.DB
}

// GetCompanies returns a list of all companies
func (fc *FinancialController) GetCompanies(c *gin.Context) {
	ctx := c.Request.Context()
	limit := getLimitWithDefault(c, 100)

	companies, err := gorm.G[models.Company](fc.DB).Where("category <> ?", "E").Order("last_modified_date DESC").Limit(limit).Find(ctx)
	if err != nil {
		log.Printf("failed to get companies: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"companies": companies,
	})
}

// GetReports returns a list of reports, optionally filtered by corp_code
func (fc *FinancialController) GetReports(c *gin.Context) {
	corpCode := c.Param("corp_code")

	var company models.Company
	err := fc.DB.Model(&models.Company{}).Where("corp_code = ?", corpCode).First(&company).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
			return
		}

		log.Printf("failed to get company: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	limit := getLimitWithDefault(c, 10)

	baseQuery := fc.DB.Model(&models.Analysis{}).Joins("JOIN raw_reports ON analyses.raw_report_id = raw_reports.id").Where("raw_reports.corp_code = ?", corpCode).Order("created_at DESC").Limit(limit)

	var analyses []models.Analysis
	if err := baseQuery.Find(&analyses).Error; err != nil {
		log.Printf("failed to get company reports: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reports": analyses,
	})
}

// GetRawReport returns a raw report for the given corp code and report number
func (fc *FinancialController) GetRawReport(c *gin.Context) {
	corpCode := c.Param("corp_code")
	rawReportID := c.Param("raw_report_id")

	var rawReport models.RawReport
	err := fc.DB.Model(&models.RawReport{}).Where("corp_code = ? AND id = ?", corpCode, rawReportID).First(&rawReport).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Raw report not found"})
			return
		}

		log.Printf("failed to get raw report: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"raw_report": string(rawReport.BlobData),
	})
}

func getLimitWithDefault(c *gin.Context, defaultValue int) int {
	var err error
	limit := defaultValue
	if c.Query("limit") != "" {
		limit, err = strconv.Atoi(c.Query("limit"))
		if err != nil {
			log.Printf("failed to parse limit: %v, using default value: %d", err, defaultValue)
			return defaultValue
		}
	}
	return limit
}
