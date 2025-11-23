package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
)

type CorrectionReportJSON struct {
	DocID   string `json:"doc_id"`
	DocType string `json:"doc_type"`
	Issuer  struct {
		Name    string `json:"name"`
		AregCIK string `json:"areg_cik"`
	} `json:"issuer"`
	Dates struct {
		FirstFiled          string `json:"first_filed"`
		CorrectionAnnounced string `json:"correction_announced"`
	} `json:"dates"`
	Tranches []struct {
		Name      string `json:"name"`
		Seniority string `json:"seniority"`
		AmountKRW int64  `json:"amount_krw"`
	} `json:"tranches"`
	Totals struct {
		AmountKRW              int64   `json:"amount_krw"`
		WACBeforePCT           float64 `json:"wac_before_pct"`
		WACAfterPCT            float64 `json:"wac_after_pct"`
		WACDeltaBp             int64   `json:"wac_delta_bp"`
		AnnualInterestDeltaKRW int64   `json:"annual_interest_delta_krw"`
	} `json:"totals"`
	ReasonOfCorrection string  `json:"reason_of_correction"`
	SpreadAfterBp      float64 `json:"spread_after_bp"`
	ImpactScore        struct {
		EquityImpact0to5    float64 `json:"equity_impact_0to5"`
		CreditImpact0to5    float64 `json:"credit_impact_0to5"`
		LiquidityImpact0to5 float64 `json:"liquidity_impact_0to5"`
	} `json:"impact_score"`
	Notes string `json:"notes"`
}

type Score struct {
	Direction      string   `json:"direction"`  // down, up
	Magnitude      float64  `json:"magnitude"`  // 0-100
	Confidence     float64  `json:"confidence"` // 0.0-1.0
	Horizons       []string `json:"horizons"`   // "ST"
	RationaleShort string   `json:"rationale_short"`
}

type SupplyExtract struct {
	DocID     string `json:"doc_id"`
	CorpName  string `json:"corp_name"`
	Title     string `json:"report_title"`
	EventCode string `json:"event_code"` // "SUPPLY"
	Amendment struct {
		Reason         string  `json:"reason"`
		PrevAmountKRW  int64   `json:"prev_amount_krw"`
		NewAmountKRW   int64   `json:"new_amount_krw"`
		PrevRatioSales float64 `json:"prev_ratio_to_sales"`
		NewRatioSales  float64 `json:"new_ratio_to_sales"`
	} `json:"amendment"`
	Contract struct {
		Name                  string  `json:"name"`
		Counterparty          string  `json:"counterparty"`
		AmountKRW             int64   `json:"amount_krw"`
		CompanyRecentSalesKRW int64   `json:"company_recent_sales_krw"`
		CounterpartySalesKRW  int64   `json:"counterparty_recent_sales_krw"`
		Country               string  `json:"country"`
		TermFrom              string  `json:"term_from"` // YYYY-MM-DD
		TermTo                string  `json:"term_to"`   // YYYY-MM-DD
		ProgressPct           float64 `json:"progress_pct"`
	} `json:"contract"`
	Score Score `json:"score"`
}

type IssuanceTermsExtract struct {
	DocID  string `json:"doc_id"`
	Issuer struct {
		Name    string `json:"name"`
		AregCik string `json:"areg_cik"`
	} `json:"issuer"`
	EventCode string `json:"event_code"` // "SECURITY_ISSUANCE_TERMS"
	Tranches  []struct {
		Name            string  `json:"name"`
		Seniority       string  `json:"seniority"` // "senior"|"subordinated"|null
		AmountKRW       int64   `json:"amount_krw"`
		CouponBeforePct float64 `json:"coupon_before_pct"`
		CouponAfterPct  float64 `json:"coupon_after_pct"`
		CouponDeltaBp   float64 `json:"coupon_delta_bp"`
	} `json:"tranches"`
	Totals struct {
		AmountKRW    int64   `json:"amount_krw"`
		WacBeforePct float64 `json:"wac_before_pct"`
		WacAfterPct  float64 `json:"wac_after_pct"`
		WacDeltaBp   float64 `json:"wac_delta_bp"`
	} `json:"totals"`
	ReasonOfCorrection string   `json:"reason_of_correction"`
	SpreadAfterBp      float64  `json:"spread_after_bp"`
	Notes              []string `json:"notes"`
	Score              Score    `json:"score"`
}

const (
	defaultModel     = shared.ResponsesModel("gpt-5.1")
	previewByteLimit = 128 * 1024 // cap what we send to the model
)

var (
	// ErrMissingAPIKey is returned when OPENAI_API_KEY was not configured.
	ErrMissingAPIKey = errors.New("OPENAI_API_KEY is not set")
)

// FileAnalyzer is a thin wrapper around the OpenAI responses client that can
// analyze a local file using the latest SDK.
type FileAnalyzer struct {
	client *openai.Client
	model  shared.ResponsesModel
}

// NewFileAnalyzerFromEnv builds a FileAnalyzer using the OPENAI_API_KEY env var.
func NewFileAnalyzerFromEnv() (*FileAnalyzer, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &FileAnalyzer{client: &client, model: defaultModel}, nil
}

func NewFileAnalyzer(apiKey string) *FileAnalyzer {
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &FileAnalyzer{client: &client, model: defaultModel}
}

// AnalyzeFile sends the file contents to the OpenAI Responses API and returns the
// assistant's answer for the provided question.
func (a *FileAnalyzer) AnalyzeFile(ctx context.Context, filePath string, docType string) (interface{}, error) {
	if a == nil || a.client == nil {
		return nil, errors.New("FileAnalyzer is not initialized")
	}

	contents, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", filePath, err)
	}

	mainPrompt := systemPrompt
	prompt := ""
	var report interface{}
	if docType == "securities_issuance_terms" {
		mainPrompt += securitiesIssuanceTermsSchema
		prompt = buildPrompt(string(contents), additionalSecuritiesIssuanceTermsSchema)
		report = &CorrectionReportJSON{}
	} else if docType == "supply" { // "SUPPLY"
		mainPrompt += supplySchema
		prompt = buildPrompt(string(contents), "")
		report = &SupplyExtract{}
	} else if docType == "report" {
		mainPrompt += reportSchema
		prompt = buildPrompt(string(contents), additionalReportSchema)
		report = &Report{}
	} else {
		mainPrompt += defaultSchema
		prompt = buildPrompt(string(contents), defaultAdditionalSchema)
		report = &DefaultReport{}
	}

	resp, err := a.client.Responses.New(ctx, responses.ResponseNewParams{
		Model: a.model,
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: responses.ResponseInputParam{
				responses.ResponseInputItemParamOfMessage(mainPrompt, responses.EasyInputMessageRoleSystem),
				responses.ResponseInputItemParamOfMessage(prompt, responses.EasyInputMessageRoleUser),
			},
		},
		//		Reasoning: shared.ReasoningParam{
		//			Effort:  shared.ReasoningEffortHigh,
		//			Summary: shared.ReasoningSummaryDetailed,
		//		},
	})

	if err != nil {
		return nil, fmt.Errorf("call OpenAI: %w", err)
	}

	output := strings.TrimSpace(resp.OutputText())
	if output == "" {
		return nil, errors.New("model returned an empty response")
	}

	if err := json.Unmarshal([]byte(output), report); err != nil {
		return nil, fmt.Errorf("unmarshal JSON: %w", err)
	}

	if docType == "report" {
		metrics := analyzeTrends(report.(*Report))
		log.Printf("Trend Metrics: %+v\n", metrics)
	}

	return report, nil
}

func (a *FileAnalyzer) AnalyzeReport(ctx context.Context, contents string, docType string) (interface{}, error) {
	mainPrompt := systemPrompt
	prompt := ""
	var report interface{}
	if docType == "securities_issuance_terms" {
		mainPrompt += securitiesIssuanceTermsSchema
		prompt = buildPrompt(string(contents), additionalSecuritiesIssuanceTermsSchema)
		report = &CorrectionReportJSON{}
	} else if docType == "supply" { // "SUPPLY"
		mainPrompt += supplySchema
		prompt = buildPrompt(string(contents), "")
		report = &SupplyExtract{}
	} else if docType == "report" {
		mainPrompt += reportSchema
		prompt = buildPrompt(string(contents), additionalReportSchema)
		report = &Report{}
	} else {
		mainPrompt += defaultSchema
		prompt = buildPrompt(string(contents), defaultAdditionalSchema)
		report = &DefaultReport{}
	}

	resp, err := a.client.Responses.New(ctx, responses.ResponseNewParams{
		Model: a.model,
		Input: responses.ResponseNewParamsInputUnion{
			OfInputItemList: responses.ResponseInputParam{
				responses.ResponseInputItemParamOfMessage(mainPrompt, responses.EasyInputMessageRoleSystem),
				responses.ResponseInputItemParamOfMessage(prompt, responses.EasyInputMessageRoleUser),
			},
		},
		//		Reasoning: shared.ReasoningParam{
		//			Effort:  shared.ReasoningEffortHigh,
		//			Summary: shared.ReasoningSummaryDetailed,
		//		},
	})

	if err != nil {
		return nil, fmt.Errorf("call OpenAI: %w", err)
	}

	output := strings.TrimSpace(resp.OutputText())
	if output == "" {
		return nil, errors.New("model returned an empty response")
	}

	if err := json.Unmarshal([]byte(output), report); err != nil {
		return nil, fmt.Errorf("unmarshal JSON: %w", err)
	}

	// if docType == "report" {
	// 	metrics := analyzeTrends(report.(*Report))
	// 	log.Printf("Trend Metrics: %+v\n", metrics)
	// }

	return report, nil
}

func buildPrompt(rawContents string, additionalPrompt string) string {
	if len(rawContents) > previewByteLimit {
		rawContents = rawContents[:previewByteLimit] + "\n\n[...truncated for brevity...]"
	}

	builder := strings.Builder{}
	builder.WriteString("다음은 DART 공시 원문이다. 필요한 정보를 추출해 스키마에 맞춰 JSON만 출력하라.")
	builder.WriteString("공시 원문:\n")
	builder.WriteString(rawContents)
	builder.WriteString("\n\n")
	builder.WriteString(`추가 지시: `)
	builder.WriteString(additionalPrompt)

	return builder.String()
}
