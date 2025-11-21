package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"kosis/internal/config"
	"kosis/internal/models"
	"kosis/internal/pkg/dart"
	"kosis/internal/pkg/openai"
	"kosis/internal/pkg/xbrl"
	"log"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// TaskProcessor holds dependencies for our task handlers
type TaskProcessor struct {
	DB           *gorm.DB
	config       *config.Config
	dartClient   *dart.DartClient
	fileAnalyzer *openai.FileAnalyzer
}

// NewTaskProcessor creates a new TaskProcessor
func NewTaskProcessor(db *gorm.DB, config *config.Config) *TaskProcessor {
	return &TaskProcessor{
		DB:           db,
		config:       config,
		dartClient:   dart.New(config.DartAPIKey),
		fileAnalyzer: openai.NewFileAnalyzer(config.OpenAIAPIKey),
	}
}

func (p *TaskProcessor) HandleFetchReportsTask(ctx context.Context, t *asynq.Task) error {
	log.Println("Fetching reports")

	var payload FetchReportsPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	log.Printf("Fetching reports for %+v", payload)

	rawReports, err := p.dartClient.GetRawReports()
	if err != nil {
		log.Printf("failed to fetch reports: %v", err)
		return nil
	}

	for _, rawReport := range rawReports {
		count, err := gorm.G[models.RawReport](p.DB).Where("receipt_number = ?", rawReport.RceptNo).Count(ctx, "id")
		if err != nil {
			return err
		}

		if count > 0 {
			log.Printf("raw report already exists: %s", rawReport.RceptNo)
			continue
		}

		rawDocument, err := p.dartClient.GetDocument(rawReport.RceptNo)
		if err == dart.ErrDocumentNotFound {
			log.Printf("document not found: %s %s %s %s", rawReport.RceptDt, rawReport.RceptNo, rawReport.CorpName, rawReport.ReportNm)
			continue
		}

		if err != nil {
			log.Printf("failed to get document: %v", err)
			return err
		}

		doc, err := xbrl.ParseXBRL([]byte(rawDocument))
		if err != nil {
			log.Printf("failed to parse XBRL document: %v", err)
			return err
		}

		j, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			log.Printf("failed to marshal JSON: %v", err)
			return err
		}

		rawReport := models.RawReport{
			ReceiptNumber: rawReport.RceptNo,
			CorpCode:      rawReport.CorpCode,
			BlobData:      []byte(rawDocument),
			BlobSize:      len(rawDocument),
			JSONData:      j,
		}

		result := gorm.WithResult()
		err = gorm.G[models.RawReport](p.DB, result).Create(ctx, &rawReport)
		if err != nil {
			return err
		}

		reportType := ""
		if strings.Contains(doc.ReportTitle, "분기보고서") {
			reportType = "report"
		} else {
			log.Printf("unknown report type: %s", doc.ReportTitle)
		}

		analysis, err := p.fileAnalyzer.AnalyzeReport(ctx, string(j), reportType)
		if err != nil {
			log.Printf("failed to analyze report: %v", err)
			continue
		}

		if v, ok := analysis.(*openai.DefaultReport); ok {
			if v.CompanyName == "" {
				company, err := gorm.G[models.Company](p.DB).Where("corp_code = ?", rawReport.CorpCode).First(ctx)
				if err != nil {
					log.Printf("failed to get company: %v", err)
					continue
				}
				v.CompanyName = company.CorpName
			}
		}

		analysisJSON, err := json.MarshalIndent(analysis, "", "  ")
		if err != nil {
			log.Printf("failed to marshal analysis: %v", err)
			continue
		}

		result = gorm.WithResult()
		err = gorm.G[models.Analysis](p.DB, result).Create(ctx, &models.Analysis{
			RawReportID: rawReport.ID,
			Analysis:    analysisJSON,
		})
		if err != nil {
			return err
		}

		log.Printf("stored raw report: %s, %s, %d", rawReport.ReceiptNumber, rawReport.CorpCode, rawReport.BlobSize)
	}

	log.Println("Reports fetched successfully")
	return nil
}

func (p *TaskProcessor) HandleFetchCompaniesTask(ctx context.Context, t *asynq.Task) error {
	log.Println("Fetching companies")

	var payload FetchCompaniesPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	log.Printf("Fetching companies")

	companies, err := p.dartClient.GetCompanies()
	if err != nil {
		log.Printf("failed to fetch companies: %v", err)
		return err
	}

	log.Printf("fetched %d companies", len(companies))

	for _, company := range companies {
		existingCompany, err := gorm.G[models.Company](p.DB).Where("corp_code = ?", company.CorpCode).First(ctx)
		if err == gorm.ErrRecordNotFound {
			lastModifiedDate, err := time.Parse("20060102", company.ModifyDate)
			if err != nil {
				log.Printf("failed to parse last modified date: %v", err)
				return err
			}
			company := models.Company{
				CorpCode:         company.CorpCode,
				CorpName:         company.CorpName,
				CorpEngName:      company.CorpEngName,
				LastModifiedDate: lastModifiedDate,
			}

			result := gorm.WithResult()
			err = gorm.G[models.Company](p.DB, result).Create(ctx, &company)
			if err != nil {
				return err
			}
			continue
		}

		if err != nil {
			return err
		}

		log.Printf("existingCompany: %+v", existingCompany)

		existingCompany.CorpName = company.CorpName
		existingCompany.CorpEngName = company.CorpEngName
		existingCompany.LastModifiedDate, err = time.Parse("20060102", company.ModifyDate)
		if err != nil {
			return err
		}

		result := gorm.WithResult()
		_, err = gorm.G[models.Company](p.DB, result).Updates(ctx, existingCompany)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *TaskProcessor) GetDartClient() *dart.DartClient {
	return p.dartClient
}
