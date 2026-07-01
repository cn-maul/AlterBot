package monitor

import (
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/cn-maul/AlterBot/database"
	"github.com/cn-maul/AlterBot/fetcher"
)

// 包级别类型定义（不可在函数内定义方法）
type scanContainerKey struct {
	tag      string
	selector string
}

type scanContainerEntry struct {
	key      scanContainerKey
	parent   *goquery.Selection
	items    []ExtractResult
	itemsSel []*goquery.Selection
	hits     int
}

// ScanResult 智能扫描结果
type ScanResult struct {
	URL        string          `json:"url"`
	Containers []ContainerInfo `json:"containers"`
}

// ContainerInfo 候选容器信息
type ContainerInfo struct {
	Selector     string          `json:"selector"`
	ContainerTag string          `json:"container_tag"`
	ContainerCSS string          `json:"container_css"`
	ItemTag      string          `json:"item_tag"`
	ItemCSS      string          `json:"item_css"`
	ItemCount    int             `json:"item_count"`
	KeywordHits  int             `json:"keyword_hits"`
	SampleItems  []ExtractResult `json:"sample_items"`
}

// ScanSettings 扫描配置
type ScanSettings struct {
	URL      string   `json:"url"`
	Keywords []string `json:"keywords"`
}

// SmartScan 智能扫描：根据 URL 和关键词自动检测内容容器
func SmartScan(settings *ScanSettings) (*ScanResult, error) {
	f := fetcher.New()
	html, err := f.Fetch(settings.URL)
	if err != nil {
		return nil, fmt.Errorf("抓取页面失败: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("解析 HTML 失败: %w", err)
	}

	// 1. 查找所有包含关键词的元素
	type match struct {
		sel     *goquery.Selection
		keyword string
	}
	var matches []match
	for _, kw := range settings.Keywords {
		if kw == "" {
			continue
		}
		kw = strings.TrimSpace(kw)
		doc.Find("body *").Each(func(_ int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if text == "" {
				return
			}
			tag := goquery.NodeName(s)
			if tag == "script" || tag == "style" || tag == "noscript" {
				return
			}
			if strings.Contains(text, kw) {
				matches = append(matches, match{sel: s, keyword: kw})
			}
		})
	}

	if len(matches) == 0 {
		return &ScanResult{URL: settings.URL}, nil
	}

	// 2. 对每个匹配，找到合适的容器祖先
	containerMap := make(map[scanContainerKey]*scanContainerEntry)
	var containerOrder []scanContainerKey

	for _, m := range matches {
		container := findBestContainer(m.sel, doc)
		if container == nil {
			continue
		}

		tag := goquery.NodeName(container)
		sel := buildElementSelector(container)
		if sel == "" {
			continue
		}
		key := scanContainerKey{tag: tag, selector: sel}

		if _, exists := containerMap[key]; !exists {
			containerMap[key] = &scanContainerEntry{
				key:    key,
				parent: container,
			}
			containerOrder = append(containerOrder, key)
		}
		containerMap[key].hits++
	}

	// 3. 对每个容器提取子项
	for _, key := range containerOrder {
		entry := containerMap[key]
		extractContainerItems(entry)
	}

	// 4. 按匹配数排序
	sortScanContainers(containerOrder, containerMap)

	maxContainers := 3
	if len(containerOrder) > maxContainers {
		containerOrder = containerOrder[:maxContainers]
	}

	result := &ScanResult{URL: settings.URL}
	for _, key := range containerOrder {
		entry := containerMap[key]
		css := buildShortSelector(entry.parent, entry.key.tag)
		itemTag, itemCSS := detectItemPattern(entry.parent)

		info := ContainerInfo{
			Selector:     entry.key.selector,
			ContainerTag: entry.key.tag,
			ContainerCSS: css,
			ItemTag:      itemTag,
			ItemCSS:      itemCSS,
			ItemCount:    len(entry.items),
			KeywordHits:  entry.hits,
		}

		maxItems := 10
		if len(entry.items) > maxItems {
			info.SampleItems = entry.items[:maxItems]
		} else {
			info.SampleItems = entry.items
		}

		result.Containers = append(result.Containers, info)
	}

	return result, nil
}

// findBestContainer 寻找最适合作为容器的祖先元素
func findBestContainer(sel *goquery.Selection, doc *goquery.Document) *goquery.Selection {
	current := sel
	depth := 0
	maxDepth := 10

	for current != nil && depth < maxDepth {
		tag := goquery.NodeName(current)
		if isContainerTag(tag) {
			children := current.Children()

			childCount := 0
			children.Each(func(_ int, c *goquery.Selection) {
				ctag := goquery.NodeName(c)
				if isListItemTag(ctag) {
					childCount++
				}
			})

			if childCount >= 2 {
				linkCount := 0
				children.Each(func(i int, c *goquery.Selection) {
					if c.Find("a").Length() > 0 {
						linkCount++
					}
				})

				if childCount > 0 && linkCount >= childCount/2 {
					return current
				}
			}

			if childCount >= 3 && depth <= 3 {
				return current
			}
		}

		parent := current.Parent()
		if parent.Length() == 0 || parent.Is("html") || parent.Is("body") {
			break
		}
		current = parent
		depth++
	}

	return nil
}

func isContainerTag(tag string) bool {
	containers := map[string]bool{
		"div": true, "section": true, "ul": true, "ol": true,
		"table": true, "tbody": true, "dl": true,
	}
	return containers[tag]
}

func isListItemTag(tag string) bool {
	items := map[string]bool{
		"li": true, "tr": true, "dd": true, "dt": true,
		"div": true, "p": true,
	}
	return items[tag]
}

func buildElementSelector(sel *goquery.Selection) string {
	if sel.Length() == 0 {
		return ""
	}
	if id, exists := sel.Attr("id"); exists && id != "" {
		return "#" + id
	}
	tag := goquery.NodeName(sel)
	if class, exists := sel.Attr("class"); exists && class != "" {
		classes := strings.Fields(class)
		for _, c := range classes {
			if len(c) > 1 && !strings.HasPrefix(c, "ng-") && !strings.HasPrefix(c, "_") {
				return tag + "." + c
			}
		}
		return tag + "." + classes[0]
	}
	return tag
}

func buildShortSelector(sel *goquery.Selection, tag string) string {
	base := buildElementSelector(sel)
	if base == "" {
		base = tag
	}
	children := sel.Children()
	childTag := ""
	childClass := ""
	childrenCount := 0

	children.Each(func(_ int, c *goquery.Selection) {
		childrenCount++
		ctag := goquery.NodeName(c)
		if childTag == "" {
			childTag = ctag
		} else if childTag != ctag {
			childTag = "mixed"
		}
		cls, exists := c.Attr("class")
		if exists && cls != "" && childClass == "" {
			childClass = cls
		}
	})

	if childTag != "" && childTag != "mixed" && childrenCount >= 2 {
		if childClass != "" {
			classes := strings.Fields(childClass)
			return base + " > " + childTag + "." + classes[0]
		}
		return base + " > " + childTag
	}
	return base
}

func detectItemPattern(sel *goquery.Selection) (tag, css string) {
	children := sel.Children()
	childTag := ""
	count := 0

	children.Each(func(_ int, c *goquery.Selection) {
		count++
		ctag := goquery.NodeName(c)
		if childTag == "" || childTag == ctag {
			childTag = ctag
		} else {
			childTag = "mixed"
		}
	})

	if count < 2 || childTag == "" || childTag == "mixed" {
		return "", ""
	}

	if cls, exists := children.First().Attr("class"); exists && cls != "" {
		classes := strings.Fields(cls)
		return childTag, childTag + "." + classes[0]
	}
	return childTag, childTag
}

func extractContainerItems(entry *scanContainerEntry) {
	entry.parent.Children().Each(func(_ int, c *goquery.Selection) {
		ctag := goquery.NodeName(c)
		if !isListItemTag(ctag) {
			return
		}

		item := make(ExtractResult)
		titleSel := c.Find("a").First()
		if titleSel.Length() > 0 {
			item["title"] = strings.TrimSpace(titleSel.Text())
			if href, exists := titleSel.Attr("href"); exists {
				item["url"] = href
			}
		} else {
			item["title"] = strings.TrimSpace(c.Text())
		}

		c.Find("span, time, small, .date, .time").Each(func(_ int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if isDateLike(text) {
				item["date"] = text
			}
		})

		if len(item) > 0 {
			entry.items = append(entry.items, item)
			entry.itemsSel = append(entry.itemsSel, c)
		}
	})
}

func isDateLike(text string) bool {
	if text == "" {
		return false
	}
	markers := []string{"202", "20", "年", "月", "日", "-", "/"}
	count := 0
	for _, m := range markers {
		if strings.Contains(text, m) {
			count++
		}
	}
	return count >= 2
}

func sortScanContainers(keys []scanContainerKey, m map[scanContainerKey]*scanContainerEntry) {
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if m[keys[j]].hits > m[keys[i]].hits {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
}

// MonitorFromScan 从扫描结果创建并启动监控器
func MonitorFromScan(name, url, containerCSS string) (*Monitor, error) {
	containerSel := containerCSS
	itemSel := ""

	if idx := strings.LastIndex(containerCSS, " > "); idx > 0 {
		containerSel = containerCSS[:idx]
		itemSel = containerCSS[idx+3:]
	}

	if itemSel == "" {
		itemSel = "a"
	}

	site := &database.Site{
		Name:          name,
		URL:           url,
		Container:     containerSel,
		Item:          itemSel,
		GroupName:     "默认",
		CheckInterval: 3600,
		IsActive:      true,
		Fields: []database.SiteField{
			{Name: "title", Selector: "a", Type: "text"},
			{Name: "url", Selector: "a", Type: "attr", Attr: "href"},
		},
	}

	if err := database.GetDB().Create(site).Error; err != nil {
		return nil, fmt.Errorf("保存站点失败: %w", err)
	}

	var savedSite database.Site
	database.GetDB().Preload("Fields").First(&savedSite, site.ID)

	m := NewMonitor(&savedSite)
	go Start(&savedSite)

	log.Printf("[智能创建] 监控器「%s」已创建并启动", name)
	return m, nil
}
