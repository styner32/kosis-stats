package tasks

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

// This file defines the "types" and "payloads" for our async tasks.

// Task type names
const (
	TypeTaskAnalyzeReport = "task:analyze_report"
	TypeTaskFetchReports  = "task:fetch_reports"
)

// --- FetchFinancials Task ---

// FetchFinancialsPayload is the data a job needs to run
type FetchReportsPayload struct {
	CorpCode *string `json:"corp_code"`
	Limit    *int    `json:"limit"`
}

// NewFetchReportsTask creates a new task for asynq
func NewFetchReportsTask(corpCode *string, limit *int) (*asynq.Task, error) {
	payload := FetchReportsPayload{
		CorpCode: corpCode,
		Limit:    limit,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeTaskFetchReports, payloadBytes), nil
}
