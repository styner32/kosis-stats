package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"kosis/internal/config"
	"kosis/internal/models"
	"kosis/internal/pkg/dart"
	"log"

	"github.com/hibiken/asynq"
	"gorm.io/gorm"
)

// TaskProcessor holds dependencies for our task handlers
type TaskProcessor struct {
	DB         *gorm.DB
	config     *config.Config
	dartClient *dart.DartClient
}

// NewTaskProcessor creates a new TaskProcessor
func NewTaskProcessor(db *gorm.DB, config *config.Config) *TaskProcessor {
	return &TaskProcessor{
		DB:         db,
		config:     config,
		dartClient: dart.New(config.DartAPIKey),
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
		if err != nil {
			log.Printf("failed to get document: %v", err)
			return err
		}

		rawReport := models.RawReport{
			ReceiptNumber: rawReport.RceptNo,
			CorpCode:      rawReport.CorpCode,
			BlobData:      []byte(rawDocument),
			BlobSize:      len(rawDocument),
		}

		result := gorm.WithResult()
		err = gorm.G[models.RawReport](p.DB, result).Create(ctx, &rawReport)
		if err != nil {
			return err
		}

		log.Printf("stored raw report: %s, %s, %d", rawReport.ReceiptNumber, rawReport.CorpCode, rawReport.BlobSize)
	}

	log.Println("Reports fetched successfully")

	return nil
}

func (p *TaskProcessor) GetDartClient() *dart.DartClient {
	return p.dartClient
}
