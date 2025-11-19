package main

import (
	"context"
	"encoding/json"
	"fmt"
	"kosis/internal/pkg/dart"
	"kosis/internal/pkg/openai"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	apiKey := os.Getenv("KOSIS_API_KEY")
	if apiKey == "" {
		log.Fatal("KOSIS_API_KEY is not set")
	}

	dartApiKey := os.Getenv("DART_API_KEY")
	if dartApiKey == "" {
		log.Fatal("DART_API_KEY is not set")
	}

	// kosisClient := kosis.New(apiKey)
	// if err := kosisClient.Search(); err != nil {
	// 	log.Fatalf("Failed to search KOSIS: %v", err)
	// }

	dartClient := dart.New(dartApiKey)
	reports, err := dartClient.GetRawReports()
	if err != nil {
		log.Fatalf("Failed to get raw reports: %v", err)
	}

	fmt.Printf("Raw reports: %+v\n", reports)

	// data/receipts/01942952/20251015000221.html
	// 20251015000218.xml
	// data/receipts/01035942/20251015000213.html
	// data/receipts/01878037/20251015000214.html
	// data/receipts/01136001/20251015900291.html
	// data/receipts/00127875/20251031000217.html
	// data/receipts/00977377/20251031000579.html
	// data/receipts/01960949/20251030000572.html
	// data/receipts/00485177/20251031900992.html // 일진파워, 단일판매ㆍ공급계약체결
	// data/receipts/01515323/20250814001590.html // LG에너지솔루션, 반기보고서 (2025.06)
	// data/receipts/00126380/20250515001922.html // 삼성전자, 반기보고서 (2025.03)
	// analyzeCorrectionReport("data/receipts/00126380/", "20250515001922")
}

func analyzeCorrectionReport(folderName string, reportNumber string) {
	fileName := fmt.Sprintf("%s/%s.html", folderName, reportNumber)
	file, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	if err := dart.StoreFiles(file, reportNumber); err != nil {
		log.Fatalf("Failed to store files: %v", err)
	}

	compactFolder := fmt.Sprintf("data/compact/%s", reportNumber)
	compactFilename := fmt.Sprintf("%s/%s.json", compactFolder, reportNumber)

	fa, err := openai.NewFileAnalyzerFromEnv()
	if err != nil {
		log.Fatalf("Failed to create file analyzer: %v", err)
	}

	answer, err := fa.AnalyzeFile(context.Background(), compactFilename, "report")
	if err != nil {
		log.Fatalf("Failed to analyze file: %v", err)
	}

	fmt.Printf("Answer Raw: %+v\n", answer)

	answerJSON, err := json.MarshalIndent(answer, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal answer: %v", err)
	}

	fmt.Printf("Answer: %s\n", string(answerJSON))
}
