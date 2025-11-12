package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kosis/pkg/openai"
	"kosis/pkg/xbrl"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

/*
parameters: https://kosis.kr/openapi/Param/statisticsParameterData.do?method=getList
table meta: https://kosis.kr/openapi/statisticsData.do?method=getMeta&type=TBL|ITM
search: https://kosis.kr/openapi/statisticsSearch.do?method=getList
*/

var client = &http.Client{
	Timeout: 10 * time.Second,
}

/*
	res: {
		"ORG_ID": "101",
		"ORG_NM": "통계청",
		"TBL_ID": "DT_1BPA001",
		"TBL_NM": "성 및 연령별 추계인구(1세별, 5세별) / 전국",
		"STAT_ID": "1994044",
		"STAT_NM": "장래인구추계",
		"VW_CD": "MT_ZTITLE",
		"MT_ATITLE": "인구 > 장래인구추계 > 전국(2022년 기준)",
		"FULL_PATH_ID": "A > A_6 > A41_10",
		"CONTENTS": "가정별 연령별 성별 추계인구 저위 추계(최소인구 추계: 출산율-저위 / 기대수명-저위 / 국제순이동-저위) 중위 추계(기본 추계: 출산율-중위 / 기대수명-중위 / 국제순이동-중위) 고위 추계(최대인구 추계: 출산율-고위 / 기대수명-고위 / 국제순이동-고위) 47세 48세 51세 55세 61세 62세 63세 71세 72세 75세 79세 80세 90 - 94세 8세 23세 29세 40세 45 - 49세 50 - 54세 60 - 64세 64세 65세 70 - 74세 78세 80 ...",
		"STRT_PRD_DE": "1960",
		"END_PRD_DE": "2072",
		"ITEM03": "주1) 2023년 12월에 공표한 장래인구추계 자료임. 주2) 매년 7월 1일 시점 자료임. 주3) 작성대상 인구는 국적과 상관없이 대한민국에 상주하는 인구임.(외국인 포함) 주4) 1960~2022년까지는 확정인구이며, 2023년 이후는 다음 인구추계시 변경될 수 있음. 주5) 중위추계(기본추계)는 인구변동요인별(출생, 사망, 국제이동) 중위가정을 조합한 결과, 고위추계(최대인구 추계)는 인구변동요인별(출생, 사망, 국제이동) 고위가정을 조합한 결과, 저위추계(최소인구 추계)는 인구변동요인별(출생, 사망, 국제이동) 저위가정을 조합한 결과임. 주6) 단위 : 명",
		"REC_TBL_SE": "N",
		"TBL_VIEW_URL": "https://kosis.kr/statisticsList/statisticsListIndex.do?menuId=M_01_01&vwcd=MT_ZTITLE&parmTabId=M_01_01&parentId=A.1;A_6.2;A41_10.3;",
		"LINK_URL": "http://kosis.kr/statHtml/statHtml.do?orgId=101&tblId=DT_1BPA001",
		"STAT_DB_CNT": "102897",
		"QUERY": "인구"
	}

	err: {"err":"20","errMsg":"필수요청변수값이 누락되었습니다."}
*/

type KosisSearchResponse struct {
	OrgID      string `json:"ORG_ID"`
	OrgNm      string `json:"ORG_NM"`
	TblID      string `json:"TBL_ID"`
	TblNm      string `json:"TBL_NM"`
	StatID     string `json:"STAT_ID"`
	StatNm     string `json:"STAT_NM"`
	VwCd       string `json:"VW_CD"`
	MtAtitle   string `json:"MT_ATITLE"`
	FullPathID string `json:"FULL_PATH_ID"`
	Contents   string `json:"CONTENTS"`
	StrtPrdDe  string `json:"STRT_PRD_DE"`
	EndPrdDe   string `json:"END_PRD_DE"`
	Item03     string `json:"ITEM03"`
	RecTblSe   string `json:"REC_TBL_SE"`
	TblViewURL string `json:"TBL_VIEW_URL"`
	LinkURL    string `json:"LINK_URL"`
	StatDbCnt  string `json:"STAT_DB_CNT"`
	Query      string `json:"QUERY"`
}

type KosisSearchErrorResponse struct {
	Err    string `json:"err"`
	ErrMsg string `json:"errMsg"`
}

func main() {
	apiKey := os.Getenv("KOSIS_API_KEY")
	if apiKey == "" {
		log.Fatal("KOSIS_API_KEY is not set")
	}

	dartApiKey := os.Getenv("DART_API_KEY")
	if dartApiKey == "" {
		log.Fatal("DART_API_KEY is not set")
	}

	// getInfo(apiKey)
	// lbh.Call(apiKey)

	// dartClient := dart.New(dartApiKey)
	// if err := dartClient.GetList(); err != nil {
	// 	log.Fatal(err)
	// }

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
	analyzeCorrectionReport("data/receipts/01515323/", "20250814001590")
}

func analyzeCorrectionReport(folderName string, reportNumber string) {
	fileName := fmt.Sprintf("%s/%s.html", folderName, reportNumber)
	file, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	report, err := xbrl.ParseHTML(file)
	if err != nil {
		log.Fatalf("Failed to parse HTML: %v", err)
	}

	filename2 := fmt.Sprintf("%s-2.json", reportNumber)
	err = os.WriteFile(filename2, report, 0644)
	if err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}

	log.Printf("Report: %s", string(report))

	fa, err := openai.NewFileAnalyzerFromEnv()
	if err != nil {
		log.Fatalf("Failed to create file analyzer: %v", err)
	}

	answer, err := fa.AnalyzeFile(context.Background(), filename2, "report")
	if err != nil {
		log.Fatalf("Failed to analyze file: %v", err)
	}

	fmt.Printf("Answer Raw: %+v", answer)

	answerJSON, err := json.MarshalIndent(answer, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal answer: %v", err)
	}
	fmt.Printf("Answer: %s", string(answerJSON))
}

func getInfo(apiKey string) {
	log.Printf("Starting KOSIS Stats")

	// example: https://kosis.kr/openapi/statisticsSearch.do?method=getList&apiKey=MThjYWY3MTRlYTllNjFlOTk4N2U2MzlkYmQ4OWVmYWI=&format=json&jsonVD=Y&searchNm=%EC%9D%B8%EA%B5%AC&startCount=1&resultCount=5&sort=RANK
	searchURL := "https://kosis.kr/openapi/statisticsSearch.do"

	method := "getList"
	format := "json"
	jsonVD := "Y"     // don't know what this is. hardcode it for now
	startCount := 1   // page number
	resultCount := 20 // limit per page
	sort := "RANK"    // RANK: 정확도순, DATE: 최신순
	searchNm := "인구"

	queryParams := url.Values{}
	queryParams.Add("method", method)
	queryParams.Add("format", format)
	queryParams.Add("jsonVD", jsonVD)
	queryParams.Add("apiKey", apiKey)
	queryParams.Add("startCount", strconv.Itoa(startCount))
	queryParams.Add("resultCount", strconv.Itoa(resultCount))
	queryParams.Add("sort", sort)
	queryParams.Add("searchNm", searchNm)

	searchURL = fmt.Sprintf("%s?%s", searchURL, queryParams.Encode())

	searchRes := []*KosisSearchResponse{}

	err := makeRequest(searchURL, &searchRes)
	if err != nil {
		log.Fatalf("Failed to make request: %v", err)
	}

	for _, res := range searchRes {
		log.Printf("Response: %v", res.MtAtitle)
	}
}

func makeRequest(url string, resBody interface{}) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// TODO: convert it to stream in case of large response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		errBody := &KosisSearchErrorResponse{}
		if err := json.Unmarshal(body, errBody); err != nil {
			return err
		}

		return fmt.Errorf("%s: %s", errBody.Err, errBody.ErrMsg)
	}

	if err := json.Unmarshal(body, resBody); err != nil {
		return err
	}

	return nil
}
