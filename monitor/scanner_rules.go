package monitor

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/cn-maul/Gentry/database"
)

type scanSiteRule struct {
	name       string
	urlPattern string
	build      func(doc *goquery.Document, settings *ScanSettings) []scanStrategyResult
}

type externalScanRuleFile struct {
	Version int                `json:"version"`
	Rules   []externalScanRule `json:"rules"`
}

type externalScanRule struct {
	Name              string   `json:"name"`
	URLContains       string   `json:"url_contains"`
	ContainerSelector string   `json:"container_selector"`
	ItemSelector      string   `json:"item_selector"`
	Priority          int      `json:"priority"`
	Diagnostics       []string `json:"diagnostics"`
}

var runtimeScanRules = builtInScanRules()

func matchCount(items []ExtractResult, keywords []string) int {
	count := 0
	for _, item := range items {
		if matchKeywords(item, keywords) {
			count++
		}
	}
	return count
}

func buildRuleStrategyResult(name string, container *goquery.Selection, items []ExtractResult, keywords []string, diagnostics ...string) []scanStrategyResult {
	if container == nil || container.Length() == 0 {
		return nil
	}
	// 即使关键词无命中，也保留策略结果，让扫描规则模板和结构化策略能继续参与候选生成
	hits := matchCount(items, keywords)
	return []scanStrategyResult{{
		name:        name,
		container:   container,
		hits:        max(1, hits),
		diagnostics: diagnostics,
		priority:    60,
	}}
}

func buildSelectorRuleStrategyResult(sourceLabel, name, urlPattern, containerSelector, itemSelector string, priority int, diagnostics []string, fields []database.ScanRuleField) scanSiteRule {
	if priority <= 0 {
		priority = 60
	}
	return scanSiteRule{
		name:       name,
		urlPattern: urlPattern,
		build: func(doc *goquery.Document, settings *ScanSettings) []scanStrategyResult {
			container := doc.Find(containerSelector).First()
			if container.Length() == 0 {
				return nil
			}
			var items []ExtractResult
			container.Find(itemSelector).Each(func(_ int, item *goquery.Selection) {
				text := strings.TrimSpace(item.Text())
				if text == "" {
					return
				}
				entry := ExtractResult{"title": text}
				firstLink := item.Find("a[href]").First()
				if firstLink.Length() > 0 {
					if href, exists := firstLink.Attr("href"); exists {
						entry["url"] = href
					}
				}
				items = append(items, entry)
			})
			results := buildRuleStrategyResult(name, container, items, settings.Keywords, append([]string{sourceLabel}, diagnostics...)...)
			if len(results) > 0 {
				results[0].priority = priority
				results[0].containerSelector = containerSelector
				results[0].itemSelector = itemSelector
				// 即使模板没有定义额外字段，也记录 selector，让后续逻辑能应用模板的 selector
				results[0].fields = fields
			}
			return results
		},
	}
}

func thePaperExpressRule(doc *goquery.Document, settings *ScanSettings) []scanStrategyResult {
	container := doc.Find("ul.ant-timeline").First()
	if container.Length() == 0 {
		return nil
	}
	var items []ExtractResult
	container.ChildrenFiltered("li.ant-timeline-item").Each(func(_ int, li *goquery.Selection) {
		text := strings.TrimSpace(li.Text())
		if text == "" {
			return
		}
		items = append(items, ExtractResult{"title": text})
	})
	return buildRuleStrategyResult("rule_thepaper_express", container, items, settings.Keywords, "命中内置站点规则", fmt.Sprintf("规则选择器 %s", "ul.ant-timeline > li.ant-timeline-item"))
}

func builtInScanRules() []scanSiteRule {
	return []scanSiteRule{{
		name:       "thepaper_express",
		urlPattern: "thepaper.cn/expressNews",
		build:      thePaperExpressRule,
	}}
}

func buildUserTemplateRules() []scanSiteRule {
	db := database.GetDB()
	if db == nil {
		return nil
	}
	var templates []database.ScanRuleTemplate
	if err := db.Preload("Fields").Where("enabled = ?", true).Order("priority desc, id asc").Find(&templates).Error; err != nil {
		log.Printf("[ScannerRules] 加载数据库扫描规则模板失败: %v", err)
		return nil
	}
	result := make([]scanSiteRule, 0, len(templates))
	for _, tpl := range templates {
		diagnostics := []string{"命中用户扫描规则模板"}
		if tpl.Description != "" {
			diagnostics = append(diagnostics, tpl.Description)
		}
		result = append(result, buildSelectorRuleStrategyResult("命中用户扫描规则模板", "template_"+tpl.Name, tpl.URLContains, tpl.Container, tpl.Item, max(70, tpl.Priority), diagnostics, tpl.Fields))
	}
	return result
}

func loadExternalScanRules(path string) []scanSiteRule {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("[ScannerRules] 读取外部规则失败，回退内置规则: %v", err)
		return nil
	}
	var file externalScanRuleFile
	if err := json.Unmarshal(data, &file); err != nil {
		log.Printf("[ScannerRules] 解析外部规则失败，回退内置规则: %v", err)
		return nil
	}
	var rules []scanSiteRule
	for _, rule := range file.Rules {
		if rule.Name == "" || rule.URLContains == "" || rule.ContainerSelector == "" || rule.ItemSelector == "" {
			log.Printf("[ScannerRules] 跳过无效规则: %+v", rule)
			continue
		}
		rules = append(rules, buildSelectorRuleStrategyResult("命中外部规则", rule.Name, rule.URLContains, rule.ContainerSelector, rule.ItemSelector, rule.Priority, rule.Diagnostics, nil))
	}
	log.Printf("[ScannerRules] 已从 %s 加载 %d 条外部规则", path, len(rules))
	return rules
}

func mergeScanRules(defaults, externals []scanSiteRule) []scanSiteRule {
	merged := make(map[string]scanSiteRule)
	order := []string{}
	for _, rule := range defaults {
		merged[rule.name] = rule
		order = append(order, rule.name)
	}
	for _, rule := range externals {
		if _, exists := merged[rule.name]; !exists {
			order = append(order, rule.name)
		}
		merged[rule.name] = rule
	}
	result := make([]scanSiteRule, 0, len(order))
	seen := map[string]bool{}
	for _, name := range order {
		if seen[name] {
			continue
		}
		result = append(result, merged[name])
		seen[name] = true
	}
	return result
}

func InitScanRules(externalPath string) {
	defaults := builtInScanRules()
	externals := loadExternalScanRules(externalPath)
	runtimeScanRules = mergeScanRules(defaults, externals)
	log.Printf("[ScannerRules] 已加载 %d 条内置/外部扫描规则", len(runtimeScanRules))
}

func CurrentScanRules() []scanSiteRule {
	return runtimeScanRules
}
