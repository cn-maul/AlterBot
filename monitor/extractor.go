package monitor

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ExtractResult 表示从网页中提取的单个结果项
type ExtractResult map[string]interface{}

// SiteSelectors 提取器选择器配置
type SiteSelectors struct {
	Container string
	Item      string
	Fields    []FieldConfig
}

// FieldConfig 提取字段配置
type FieldConfig struct {
	Name      string
	Selector  string
	Type      string
	Attr      string
	Transform string
}

type Extractor struct {
	containerSelector string
	itemSelector      string
	fields            []FieldConfig
}

func NewExtractor(selectors SiteSelectors) *Extractor {
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

func (e *Extractor) extractField(s *goquery.Selection, field FieldConfig) interface{} {
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

// applyTransform 应用转换规则
// 支持格式:
//
//	trim(chars)    — 去除两端指定字符
//	prefix(text)   — 添加前缀
//	suffix(text)   — 添加后缀
//	regexp(pat,repl) — 正则替换
func applyTransform(value, transform string) string {
	if value == "" || transform == "" {
		return value
	}

	// 解析 transform: funcName(args)
	idx := strings.Index(transform, "(")
	if idx < 0 || !strings.HasSuffix(transform, ")") {
		return value
	}

	name := transform[:idx]
	args := transform[idx+1 : len(transform)-1]

	switch name {
	case "trim":
		return strings.Trim(value, args)
	case "prefix":
		return args + value
	case "suffix":
		return value + args
	case "regexp":
		parts := strings.SplitN(args, ",", 2)
		if len(parts) == 2 {
			// 简单正则替换
			pattern := strings.TrimSpace(parts[0])
			replacement := strings.TrimSpace(parts[1])
			// 去掉可能的引号
			pattern = strings.Trim(pattern, `"'`)
			replacement = strings.Trim(replacement, `"'`)
			// 这里使用 strings.Replace 作为简单实现
			// 更复杂的正则需求可后续扩展
			return strings.ReplaceAll(value, pattern, replacement)
		}
		return value
	default:
		return value
	}
}