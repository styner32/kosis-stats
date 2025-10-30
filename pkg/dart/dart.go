package dart

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"kosis/pkg/xbrl"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const baseURL = "https://opendart.fss.or.kr/api"

type DartClient struct {
	key    string
	client *http.Client
}

type List struct {
	RceptNo  string `json:"rcept_no"`
	CorpCode string `json:"corp_code"`
	CorpName string `json:"corp_name"`
	ReportNm string `json:"report_nm"`
	RceptDt  string `json:"rcept_dt"`
	FlrNm    string `json:"flr_nm"`
	Rm       string `json:"rm"`
}

type ListResp struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	PageNo  int    `json:"page_no"`
	PageCnt int    `json:"page_count"`
	Total   int    `json:"total_count"`
	List    []List `json:"list"`
}

// defined error that document not found
var ErrDocumentNotFound = errors.New("document not found")

// 공시 목록 조회
// https://opendart.fss.or.kr/guide/detail.do?apiGrpCd=DS001&apiId=2019001
func (c *DartClient) getDisclosureList(apiKey, corpCode, bgnDe, endDe string, pageNo, pageCount int) (*ListResp, error) {
	u, _ := url.Parse(baseURL + "/list.json")
	q := u.Query()
	q.Set("crtfc_key", apiKey) // API Key

	if corpCode != "" {
		q.Set("corp_code", corpCode) // 8자리 기업코드(예: 삼성전자 00126380)
	}
	if bgnDe != "" {
		q.Set("bgn_de", bgnDe) // YYYYMMDD, begin date, 3 months max
	}
	if endDe != "" {
		q.Set("end_de", endDe) // YYYYMMDD, end date
	}

	// page number, 1 by default
	if pageNo > 0 {
		q.Set("page_no", fmt.Sprint(pageNo))
	}

	// page count, 10 by default, max value is 100
	if pageCount > 0 {
		q.Set("page_count", fmt.Sprint(pageCount))
	}
	u.RawQuery = q.Encode()

	resp, err := c.client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out ListResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}

	if out.Status != "000" { // 000: 정상
		return nil, fmt.Errorf("DART error %s: %s", out.Status, out.Message)
	}

	return &out, nil
}

// 공시서류원본파일
// https://opendart.fss.or.kr/guide/detail.do?apiGrpCd=DS001&apiId=2019003
func (c *DartClient) GetDocument(rceptNo string) (string, error) {
	u, _ := url.Parse(baseURL + "/document.xml")
	q := u.Query()
	q.Set("crtfc_key", c.key)  // API Key
	q.Set("rcept_no", rceptNo) // 접수번호

	u.RawQuery = q.Encode()

	resp, err := c.client.Get(u.String())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DART error %d: %s", resp.StatusCode, string(buf))
	}

	// check if mime type is application/zip
	mimeType := resp.Header.Get("Content-Type")
	if mimeType == "application/xml;charset=UTF-8" {
		if strings.Contains(string(buf), "<status>014</status>") {
			return "", ErrDocumentNotFound
		}
		return string(buf), nil
	}

	zr, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return "", err
	}

	outBuf := new(bytes.Buffer)
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			return "", err
		}

		_, err = io.Copy(outBuf, rc)
		defer rc.Close()
		if err != nil {
			return "", err
		}
	}

	return outBuf.String(), nil
}

func (c *DartClient) GetList() error {
	// 삼성전자(00126380) 2025-01-01 ~ 2025-01-31 공시 100건
	// LG화학(00356361) 2025-10-01 ~ 2025-10-31 공시 100건
	// 모든 회사
	code := "" // 00126380: 삼성전자, 00356361: LG화학
	res, err := c.getDisclosureList(c.key, code, "20251001", "20251031", 1, 10)
	if err != nil {
		return err
	}

	for _, it := range res.List {
		if err := c.processDoc(it); err != nil {
			if err == ErrDocumentNotFound {
				log.Printf("document not found: %s %s %s %s", it.RceptDt, it.RceptNo, it.CorpName, it.ReportNm)
				continue
			}

			return err
		}
	}

	return nil
}

// process doc
func (c *DartClient) processDoc(it List) error {
	fmt.Printf("%s %s %s %s\n", it.RceptDt, it.RceptNo, it.CorpName, it.ReportNm)

	folder := fmt.Sprintf("data/receipts/%s", it.CorpCode)
	filename := fmt.Sprintf("%s/%s.html", folder, it.RceptNo)

	// a file exists, skip and read the file
	doc := ""
	if _, err := os.Stat(filename); err == nil {
		b, err := os.ReadFile(filename)
		if err != nil {
			return err
		}

		doc = string(b)
	} else {
		// create a folder if not exists
		if _, err := os.Stat(folder); err != nil {
			if err := os.MkdirAll(folder, 0755); err != nil {
				return err
			}
		}

		doc, err = c.GetDocument(it.RceptNo)
		if err != nil {
			return err
		}

		f, err := os.Create(filename)
		if err != nil {
			return err
		}

		defer f.Close()
		_, err = f.WriteString(doc)
		if err != nil {
			return err
		}
	}

	j, err := xbrl.ConvertXMLToUsefulJSON([]byte(doc))
	if err != nil {
		return err
	}
	fmt.Printf("Contents: %s\n", string(j))

	if strings.Contains(it.ReportNm, "배당결정") {
		dividend, err := ParseDividendHTML(doc)
		if err != nil {
			return err
		}
		fmt.Printf("dividend: %+v\n", dividend)

		alotment, err := c.getAlotment(it.CorpCode, it.RceptDt[:4], HALF_YEAR)
		if err != nil {
			return err
		}

		b, err := json.MarshalIndent(alotment, "", "  ")
		if err == nil {
			fmt.Printf("alotment: %s\n", string(b))
		} else {
			fmt.Printf("alotment: %+v\n", alotment)
		}

		return nil
	}

	if strings.Contains(it.ReportNm, "타법인주식및출자증권처분결정") {
		disposal, err := ParseDisposalHTML(doc, it.RceptNo)
		if err != nil {
			return err
		}
		b, err := json.MarshalIndent(disposal, "", "  ")
		if err == nil {
			fmt.Printf("disposal: %s\n", string(b))
		} else {
			fmt.Printf("disposal: %+v\n", disposal)
		}
		return nil
	}

	if strings.Contains(it.ReportNm, "타법인주식및출자증권취득결정") {
		acquisition, err := ParseAcquisitionHTML(doc, it.RceptNo)
		if err != nil {
			return err
		}

		b, err := json.MarshalIndent(acquisition, "", "  ")
		if err == nil {
			fmt.Printf("acquisition: %s\n", string(b))
		} else {
			fmt.Printf("acquisition: %+v\n", acquisition)
		}

		return nil
	}

	return nil
}
