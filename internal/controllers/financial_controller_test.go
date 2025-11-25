package controllers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"kosis/internal/config"
	"kosis/internal/db"
	"kosis/internal/models"
	"kosis/internal/routes"
	"kosis/internal/testhelpers"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

func createCompany(dbConn *gorm.DB, ctx context.Context, company *models.Company) {
	// set default value if missing
	if company.CorpEngName == "" {
		company.CorpEngName = company.CorpName + " Eng"
	}

	if company.LastModifiedDate.IsZero() {
		company.LastModifiedDate = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	result := gorm.WithResult()
	Expect(gorm.G[models.Company](dbConn, result).Create(ctx, company)).To(Succeed())
	Expect(result.RowsAffected).To(Equal(int64(1)))
}

func createRawReport(dbConn *gorm.DB, ctx context.Context, rawReport *models.RawReport) {
	result := gorm.WithResult()
	Expect(gorm.G[models.RawReport](dbConn, result).Create(ctx, rawReport)).To(Succeed())
	Expect(result.RowsAffected).To(Equal(int64(1)))
}

func createAnalysis(dbConn *gorm.DB, ctx context.Context, analysis *models.Analysis) {
	result := gorm.WithResult()
	Expect(gorm.G[models.Analysis](dbConn, result).Create(ctx, analysis)).To(Succeed())
	Expect(result.RowsAffected).To(Equal(int64(1)))
}

var _ = Describe("FinancialController", func() {
	var (
		dbConn *gorm.DB
		router *gin.Engine
	)

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)

		cfg, err := config.LoadConfig()
		Expect(err).NotTo(HaveOccurred())

		dbConn, err = db.InitDB(cfg.DatabaseURL)
		if err != nil {
			Skip("database not available: " + err.Error())
		}

		testhelpers.CleanupDB(dbConn)

		router = routes.SetupRouter(dbConn, cfg)
	})

	Describe("GET /api/v1/companies", func() {
		BeforeEach(func() {
			ctx := context.Background()

			company1 := models.Company{
				CorpCode:         "10000001",
				CorpName:         "Company 1",
				CorpEngName:      "Company 1 Eng",
				LastModifiedDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			}
			createCompany(dbConn, ctx, &company1)

			company2 := models.Company{
				CorpCode:         "10000002",
				CorpName:         "Company 2",
				CorpEngName:      "Company 2 Eng",
				LastModifiedDate: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
			}
			createCompany(dbConn, ctx, &company2)
		})

		It("returns companies", func() {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/companies", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusOK))
			var body map[string]interface{}
			Expect(json.Unmarshal(resp.Body.Bytes(), &body)).To(Succeed())
			Expect(body["companies"]).To(HaveLen(2))
		})
	})

	Describe("GET /api/v1/reports/:corp_code", func() {
		var (
			rawReportA models.RawReport
			rawReportB models.RawReport
		)

		BeforeEach(func() {
			ctx := context.Background()

			companyA := models.Company{
				CorpCode: "10000001",
				CorpName: "Company A",
			}
			createCompany(dbConn, ctx, &companyA)

			companyB := models.Company{
				CorpCode: "10000002",
				CorpName: "Company B",
			}
			createCompany(dbConn, ctx, &companyB)

			rawReportA = models.RawReport{
				ReceiptNumber: "20251123000001",
				CorpCode:      "10000001",
				BlobData:      []byte("doc1"),
				BlobSize:      4,
				JSONData:      json.RawMessage(`{"a":1}`),
			}
			createRawReport(dbConn, ctx, &rawReportA)

			rawReportB = models.RawReport{
				ReceiptNumber: "20251123000002",
				CorpCode:      "10000002",
				BlobData:      []byte("doc2"),
				BlobSize:      4,
				JSONData:      json.RawMessage(`{"b":2}`),
			}
			createRawReport(dbConn, ctx, &rawReportB)

			analysisA := models.Analysis{
				RawReportID: rawReportA.ID,
				Analysis:    json.RawMessage(`{"summary":"a"}`),
			}
			createAnalysis(dbConn, ctx, &analysisA)

			analysisB := models.Analysis{
				RawReportID: rawReportB.ID,
				Analysis:    json.RawMessage(`{"summary":"b"}`),
			}
			createAnalysis(dbConn, ctx, &analysisB)
		})

		It("returns reports for the given corp code", func() {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/10000001", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusOK))

			var body struct {
				Reports []models.Analysis `json:"reports"`
			}
			Expect(json.Unmarshal(resp.Body.Bytes(), &body)).To(Succeed())
			Expect(body.Reports).To(HaveLen(1))
			Expect(body.Reports[0].RawReportID).To(Equal(rawReportA.ID))
		})

		It("returns error if corp code is not found", func() {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/10000003", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			Expect(resp.Code).To(Equal(http.StatusNotFound))
			Expect(resp.Body.String()).To(MatchJSON(`{"error": "Company not found"}`))
		})
	})

	Describe("GET /api/v1/reports/:corp_code/:raw_report_id", func() {
		It("returns raw report for the given corp code and raw report id", func() {
			rawReportA := models.RawReport{
				ReceiptNumber: "20251123000001",
				CorpCode:      "10000001",
				BlobData:      []byte("doc1"),
				BlobSize:      4,
				JSONData:      json.RawMessage(`{"a":1}`),
			}

			ctx := context.Background()
			createRawReport(dbConn, ctx, &rawReportA)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/10000001/"+strconv.FormatUint(uint64(rawReportA.ID), 10), nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)
			Expect(resp.Code).To(Equal(http.StatusOK))

			var body struct {
				RawReport string `json:"raw_report"`
			}
			Expect(json.Unmarshal(resp.Body.Bytes(), &body)).To(Succeed())
			Expect(body.RawReport).To(Equal("doc1"))
		})
	})
})
