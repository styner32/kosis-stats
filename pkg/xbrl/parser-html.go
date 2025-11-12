package xbrl

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func ParseHTML(raw []byte) ([]byte, error) {
	rawString := string(raw)
	rawString = strings.ReplaceAll(rawString, "<TU", "<TD")
	rawString = strings.ReplaceAll(rawString, "</TU>", "</TD>")
	rawString = strings.ReplaceAll(rawString, "<TE", "<TD")
	rawString = strings.ReplaceAll(rawString, "</TE>", "</TD>")

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawString))
	if err != nil {
		return nil, err
	}

	report := &UsefulReport{}
	report.ReportTitle = doc.Find("DOCUMENT-NAME").First().Text()
	companyNameSelection := doc.Find("COMPANY-NAME").First()
	report.CompanyName = companyNameSelection.Text()
	companyCIK, exists := companyNameSelection.Attr("aregcik")
	if exists {
		report.CompanyCIK = companyCIK
	} else {
		report.CompanyCIK = ""
	}

	var tables [][][]string
	doc.Find("TABLE").Each(func(i int, table *goquery.Selection) {
		rows := [][]string{}
		table.Find("TR").Each(func(i int, row *goquery.Selection) {
			cells := []string{}
			log.Println(row.Text())

			row.Children().Each(func(i int, s *goquery.Selection) {
				tag := goquery.NodeName(s)
				if strings.EqualFold(tag, "td") || strings.EqualFold(tag, "th") || strings.EqualFold(tag, "tu") || strings.EqualFold(tag, "te") {
					childText := ""
					if s.Text() != "" {
						childText += s.Text()
					}

					for _, node := range s.Nodes {
						if node.Type == html.TextNode {
							childText += node.Data
						}
					}
					cells = append(cells, childText)
				}
			})

			rows = append(rows, cells)
		})
		tables = append(tables, rows)
	})

	report.Tables = tables

	doc.Find("P").Each(func(i int, p *goquery.Selection) {
		if trimmed := strings.TrimSpace(p.Text()); trimmed != "" {
			report.KeyParagraphs = append(report.KeyParagraphs, trimmed)
		}
	})

	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("json marshal error: %w", err)
	}
	return b, nil
}
