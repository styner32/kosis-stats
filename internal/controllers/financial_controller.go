package controllers

import (
	"encoding/base64"
	"encoding/json"
	"kosis/internal/models"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FinancialController struct {
	DB *gorm.DB
}

type CompanyResponse struct {
	ID               uint   `json:"id"`
	CorpCode         string `json:"corp_code"`
	CorpName         string `json:"corp_name"`
	CorpEngName      string `json:"corp_name_eng"`
	LastModifiedDate string `json:"last_modified_date"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type ReportResponse struct {
	ID          uint            `json:"id"`
	RawReportID uint            `json:"raw_report_id"`
	UsedTokens  int64           `json:"used_tokens"`
	Analysis    json.RawMessage `json:"analysis"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

// GetCompanies returns a list of all companies
func (fc *FinancialController) GetCompanies(c *gin.Context) {
	ctx := c.Request.Context()
	limit := getLimitWithDefault(c, 100)
	search := c.Query("search")

	corpCodes := []string{}
	err := fc.DB.Model(&models.RawReport{}).Distinct("corp_code").Pluck("corp_code", &corpCodes).Error
	if err != nil {
		log.Printf("failed to get raw reports: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	// show companies that has at least one raw report
	query := gorm.G[models.Company](fc.DB).Where("category <> ?", "E").Where("corp_code IN (?)", corpCodes)

	if search != "" {
		// Use ILIKE for case-insensitive search, supported by pg_trgm indexes
		searchTerm := "%" + search + "%"
		query = query.Where("corp_name ILIKE ? OR corp_eng_name ILIKE ?", searchTerm, searchTerm)
	}

	companies, err := query.Order("last_modified_date DESC").Limit(limit).Find(ctx)
	if err != nil {
		log.Printf("failed to get companies: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	companyResponses := []CompanyResponse{}
	for _, company := range companies {
		companyResponses = append(companyResponses, CompanyResponse{
			ID:               company.ID,
			CorpCode:         company.CorpCode,
			CorpName:         company.CorpName,
			CorpEngName:      company.CorpEngName,
			LastModifiedDate: company.LastModifiedDate.Format("2006-01-02 15:04:05"),
			CreatedAt:        company.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:        company.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"companies": companyResponses,
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
		"raw_report": base64.StdEncoding.EncodeToString(rawReport.BlobData),
	})
}

// GetReportsByCorpName returns a JSON list of recent reports for a partial corp_name.
// This is a non-streaming variant for MCP/Claude clients that expect a simple HTTP response.
func (fc *FinancialController) GetReportsByCorpName(c *gin.Context) {
	corpName := strings.TrimSpace(c.Query("corp_name"))
	if corpName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "corp_name is required"})
		return
	}

	limit := getLimitWithDefault(c, 10)

	var analyses []models.Analysis
	err := fc.DB.
		Model(&models.Analysis{}).
		Joins("JOIN raw_reports ON analyses.raw_report_id = raw_reports.id").
		Joins("JOIN companies ON companies.corp_code = raw_reports.corp_code").
		Where("companies.corp_name ILIKE ?", "%"+corpName+"%").
		Order("raw_reports.receipt_number DESC").
		Limit(limit).
		Find(&analyses).Error
	if err != nil {
		log.Printf("failed to fetch reports by corp_name (non-stream): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reports": analyses,
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
