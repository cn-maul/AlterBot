package monitor

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/cn-maul/AlterBot/config"
)

// ExtractResult 表示从网页中提取的单个结果项
type ExtractResult map[string]interface{}

type Extractor struct {
	containerSelector string
	itemSelector      string
	fields            []config.FieldConfig
}

func NewExtractor(selectors config.SiteSelectors) *Extractor {
	return &Extractor{
		containerSelector: selectors.Container,
		itemSelector:      selectors.Item,
		fields:            selectors.Fields,
	}
}

func (e *Extractor) Extract(html string) ([]ExtractResult, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	var results []ExtractResult

	doc.Find(e.containerSelector).Find(e.itemSelector).Each(func(_ int, s *goquery.Selection) {
		result := make(ExtractResult)
		for _, field := range e.fields {
			if value := e.extractField(s, field); value != nil {
				result[field.Name] = value
			}
		}
		if len(result) > 0 {
			results = append(results, result)
		}
	})

	return results, nil
}

func (e *Extractor) extractField(s *goquery.Selection, field config.FieldConfig) interface{} {
	sel := s.Find(field.Selector)
	if sel.Length() == 0 {
		return nil
	}

	var value string
	switch field.Type {
	case "attr":
		attr := field.Attr
		if attr == "" {
			attr = "href"
		}
		value, _ = sel.Attr(attr)
	case "text":
		value = strings.TrimSpace(sel.Text())
	default:
		return nil
	}

	if field.Transform != "" {
		value = applyTransform(value, field.Transform)
	}

	return value
}

func applyTransform(value, transform string) string {
	// 转换逻辑实现...
	return value
}
