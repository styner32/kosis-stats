package kosis

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const baseURL = "https://kosis.kr/openapi"

type Client struct {
	key    string
	client *http.Client
}

func New(apiKey string) *Client {
	return &Client{
		key: apiKey,
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (c *Client) get(path string, q url.Values, v any) error {
	if c.key == "" {
		return errors.New("missing KOSIS_API_KEY")
	}
	q.Set("apiKey", c.key)
	q.Set("format", "json")
	q.Set("jsonVD", "Y")

	u := fmt.Sprintf("%s/%s?%s", baseURL, path, q.Encode())
	log.Printf("u: %s", u)
	req, _ := http.NewRequest("GET", u, nil)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("kosis http %d: %s", resp.StatusCode, string(b))
	}
	//dec := json.NewDecoder(resp.Body)
	//return dec.Decode(v)
	body, _ := io.ReadAll(resp.Body)

	log.Printf("body: %s", string(body))
	return json.Unmarshal(body, v)
}

type TableRef struct {
	OrgID string
	TblID string
}

type MetaITM struct {
	ObjID    string `json:"OBJ_ID"`
	TblID    string `json:"TBL_ID"`
	OrgID    string `json:"ORG_ID"`
	ObjNM    string `json:"OBJ_NM"`
	ItmNM    string `json:"ITM_NM"`
	ItmID    string `json:"ITM_ID"`
	ObjIDSn  string `json:"OBJ_ID_SN"` // order index as string
	UnitNM   string `json:"UNIT_NM"`
	ObjNMEng string `json:"OBJ_NM_ENG"`
}

func (c *Client) GetMetaITM(ref TableRef) ([]MetaITM, error) {
	q := url.Values{}
	q.Set("method", "getMeta")
	q.Set("type", "ITM")
	q.Set("orgId", ref.OrgID)
	q.Set("tblId", ref.TblID)
	var data []MetaITM
	if err := c.get("statisticsData.do", q, &data); err != nil {
		return nil, err
	}
	return data, nil
}

type ParamRow struct {
	OrgID  string `json:"ORG_ID"`
	TblID  string `json:"TBL_ID"`
	TblNM  string `json:"TBL_NM"`
	ITMID  string `json:"ITM_ID"`
	ITMNM  string `json:"ITM_NM"`
	UnitNM string `json:"UNIT_NM"`
	PRDSE  string `json:"PRD_SE"`
	PRDDE  string `json:"PRD_DE"`
	DT     string `json:"DT"`

	C1   string `json:"C1"`
	C1NM string `json:"C1_NM"`
	C2   string `json:"C2"`
	C2NM string `json:"C2_NM"`
	// (C3..C8 필요시 확장)
}

func (c *Client) ParamData(ref TableRef, prdSe, start, end string, itmId string, obj map[int]string) ([]ParamRow, error) {
	q := url.Values{}
	q.Set("method", "getList")
	q.Set("orgId", ref.OrgID)
	q.Set("tblId", ref.TblID)
	q.Set("prdSe", prdSe)
	if start != "" {
		q.Set("startPrdDe", start)
	}
	if end != "" {
		q.Set("endPrdDe", end)
	}
	q.Set("itmId", itmId)
	for i := 1; i <= 8; i++ {
		if v, ok := obj[i]; ok && v != "" {
			q.Set(fmt.Sprintf("objL%d", i), v)
		}
	}
	var data []ParamRow
	if err := c.get("Param/statisticsParameterData.do", q, &data); err != nil {
		return nil, err
	}
	return data, nil
}

// --------- Meta helpers (공용) ---------

type ClassGroup struct {
	ObjID  string
	Name   string
	Order  int
	Values map[string]string // code -> label
}

type DigestedMeta struct {
	Items   map[string]string
	Classes []ClassGroup
}

func DigestITM(m []MetaITM) DigestedMeta {
	items := map[string]string{}
	tmp := map[string]*ClassGroup{}
	for _, r := range m {
		if r.ObjID == "ITEM" {
			items[strings.TrimSpace(r.ItmID)] = strings.TrimSpace(r.ItmNM)
			continue
		}
		if _, ok := tmp[r.ObjID]; !ok {
			order := 99
			if n, err := strconv.Atoi(strings.TrimSpace(r.ObjIDSn)); err == nil {
				order = n
			}
			tmp[r.ObjID] = &ClassGroup{ObjID: r.ObjID, Name: strings.TrimSpace(r.ObjNM), Order: order, Values: map[string]string{}}
		}
		tmp[r.ObjID].Values[strings.TrimSpace(r.ItmID)] = strings.TrimSpace(r.ItmNM)
	}
	out := make([]ClassGroup, 0, len(tmp))
	for _, g := range tmp {
		out = append(out, *g)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Order < out[j].Order })
	return DigestedMeta{Items: items, Classes: out}
}

func FindItemIDByContains(items map[string]string, needle string) (string, bool) {
	for id, nm := range items {
		if strings.Contains(nm, needle) {
			return id, true
		}
	}
	return "", false
}

func FindClassIndexByName(classes []ClassGroup, keywords ...string) (int, bool) {
	for idx, g := range classes {
		for _, kw := range keywords {
			if strings.Contains(g.Name, kw) {
				return idx, true
			}
		}
	}
	return -1, false
}

// (선택) 숫자 파서가 필요하면 노출
func ParseNumber(s string) (float64, bool) {
	if s == "" || s == "-" {
		return math.NaN(), false
	}
	s = strings.ReplaceAll(s, ",", "")
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return math.NaN(), false
	}
	return f, true
}
