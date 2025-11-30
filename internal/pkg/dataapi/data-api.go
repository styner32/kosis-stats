package dataapi

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type DataAPIClient struct {
	key    string
	client *http.Client
}

type StockPriceResponse struct {
	Response struct {
		Header struct {
			ResultCode string `json:"resultCode"`
			ResultMsg  string `json:"resultMsg"`
		} `json:"header"`
		Body struct {
			NumOfRows  int `json:"numOfRows"`
			PageNo     int `json:"pageNo"`
			TotalCount int `json:"totalCount"`
			Items      struct {
				Item []struct {
					BasDt     string `json:"basDt"`
					SrtnCd    string `json:"srtnCd"`
					IsinCd    string `json:"isinCd"`
					ItmsNm    string `json:"itmsNm"`
					MrktCtg   string `json:"mrktCtg"`
					Clpr      string `json:"clpr"`
					Vs        string `json:"vs"`
					FltRt     string `json:"fltRt"`
					Mkp       string `json:"mkp"`
					Hipr      string `json:"hipr"`
					Lopr      string `json:"lopr"`
					Trqu      string `json:"trqu"`
					TrPrc     string `json:"trPrc"`
					LstgStCnt string `json:"lstgStCnt"`
					MktTotAmt string `json:"mktTotAmt"`
				} `json:"item"`
			} `json:"items"`
		} `json:"body"`
	} `json:"response"`
}

const baseURL = "https://apis.data.go.kr/"

func New(apiKey string) *DataAPIClient {
	return &DataAPIClient{
		key: apiKey,
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (c *DataAPIClient) GetStockPrice(name string) (interface{}, error) {
	u, err := url.Parse(baseURL + "/1160100/service/GetStockSecuritiesInfoService/getStockPriceInfo")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("serviceKey", c.key) // API Key
	q.Set("numOfRows", "1")
	q.Set("pageNo", "1")
	q.Set("resultType", "json")
	q.Set("itmsNm", name)

	u.RawQuery = q.Encode()

	resp, err := c.client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var stockPriceResponse StockPriceResponse
	if err := json.Unmarshal(body, &stockPriceResponse); err != nil {
		return nil, err
	}

	log.Printf("stockPriceResponse: %+v", stockPriceResponse)

	return stockPriceResponse.Response.Body.Items.Item[0], nil
}
