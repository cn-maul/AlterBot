package monitor

import (
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/cn-maul/Gentry/database"
	"github.com/cn-maul/Gentry/fetcher"
)

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

type ScanResult struct {
	URL        string          `json:"url"`
	Containers []ContainerInfo `json:"containers"`
}

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

type ScanSettings struct {
	URL      string   `json:"url"`
	Keywords []string `json:"keywords"`
}

func buildDeepSelector(sel *goquery.Selection) string {
	parts := []string{}
	current := sel
	depth := 0
	for current != nil && depth < 3 {
		if current.Is("html") || current.Is("body") {
			break
		}
		tag := goquery.NodeName(current)
		s := tag
		if id, exists := current.Attr("id"); exists && id != "" {
			s = "#" + id
		} else if class, exists := current.Attr("class"); exists && class != "" {
			classes := strings.Fields(class)
			for _, c := range classes {
				if len(c) > 1 && !strings.HasPrefix(c, "ng-") && !strings.HasPrefix(c, "_") {
					s = tag + "." + c
					break
				}
			}
		}
		parts = append([]string{s}, parts...)
		current = current.Parent()
		depth++
	}
	return strings.Join(parts, " > ")
}

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

	containerMap := make(map[scanContainerKey]*scanContainerEntry)
	var containerOrder []scanContainerKey

	for _, m := range matches {
		container := findBestContainer(m.sel, doc)
		if container == nil {
			continue
		}

		tag := goquery.NodeName(container)
		sel := buildDeepSelector(container)
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

	for _, key := range containerOrder {
		entry := containerMap[key]
		extractContainerItems(entry)
	}

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
		} else if len(entry.items) > 0 {
			info.SampleItems = entry.items
		} else {
			info.SampleItems = []ExtractResult{}
		}

		result.Containers = append(result.Containers, info)
	}

	return result, nil
}

func getTitle(item ExtractResult) string {
	if v, ok := item["title"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func setTitle(item ExtractResult, title string) {
	item["title"] = title
}

func findBestContainer(sel *goquery.Selection, doc *goquery.Document) *goquery.Selection {
	type scoredContainer struct {
		sel   *goquery.Selection
		score int
	}
	var best scoredContainer

	current := sel
	depth := 0
	maxDepth := 15

	for current != nil && depth < maxDepth {
		tag := goquery.NodeName(current)
		class, _ := current.Attr("class")

		if isContainerTag(tag) || hasArticleClass(class) {
			children := current.Children()
			childCount := 0
			linkCount := 0
			articleCardCount := 0
			textLen := 0

			children.Each(func(_ int, c *goquery.Selection) {
				ctag := goquery.NodeName(c)
				cclass, _ := c.Attr("class")

				if isListItemTag(ctag) || hasArticleItemClass(cclass) {
					childCount++
				}
				if c.Find("a").Length() > 0 {
					linkCount++
				}
				if hasArticleItemClass(cclass) {
					articleCardCount++
				}
				firstA := c.Find("a").First()
				if firstA.Length() > 0 {
					textLen += len(strings.TrimSpace(firstA.Text()))
				}
			})

			score := 0
			score += articleCardCount * 100
			score += childCount * 10
			score += linkCount * 5
			if childCount > 0 && textLen/childCount > 5 {
				score += 15
			}
			if childCount >= 2 && linkCount >= childCount/2 {
				score += 50
			}
			score += depth * 2

			lowerClass := strings.ToLower(class)
			if strings.Contains(lowerClass, "footer") || strings.Contains(lowerClass, "header") ||
				strings.Contains(lowerClass, "nav") || strings.Contains(lowerClass, "sidebar") ||
				strings.Contains(lowerClass, "menu") || strings.Contains(lowerClass, "banner") ||
				strings.Contains(lowerClass, "breadcrumb") || strings.Contains(lowerClass, "crumb") {
				score = 0
			}

			if score > best.score {
				best = scoredContainer{sel: current, score: score}
			}
		}

		parent := current.Parent()
		if parent.Length() == 0 || parent.Is("html") || parent.Is("body") {
			break
		}
		current = parent
		depth++
	}

	if best.score >= 20 {
		return best.sel
	}

	nearestList := sel.Closest("ul, ol")
	if nearestList.Length() > 0 {
		liCount := nearestList.Find("li").Length()
		if liCount >= 2 {
			return nearestList
		}
	}

	return nil
}

func isContainerTag(tag string) bool {
	containers := map[string]bool{
		"div": true, "section": true, "ul": true, "ol": true,
		"table": true, "tbody": true, "dl": true,
		"article": true, "main": true,
	}
	return containers[tag]
}

func hasArticleClass(class string) bool {
	if class == "" {
		return false
	}
	lower := strings.ToLower(class)
	markers := []string{
		"article-list", "articlelist", "post-list", "postlist",
		"news-list", "newslist", "content-list", "contentlist",
		"feed", "feed__main", "feed_main",
		"list-view", "card-list", "cardlist",
		"section__wrapper", "section_wrapper",
	}
	for _, m := range markers {
		if strings.Contains(lower, m) {
			return true
		}
	}
	return false
}

func hasArticleItemClass(class string) bool {
	if class == "" {
		return false
	}
	lower := strings.ToLower(class)
	markers := []string{
		"article-card", "articlecard", "article__card", "post-card", "postcard",
		"news-item", "newsitem", "content-item", "contentitem",
		"feed-item", "feeditem", "card-item", "carditem",
		"list-item", "listitem",
	}
	for _, m := range markers {
		if strings.Contains(lower, m) {
			return true
		}
	}
	if strings.HasPrefix(lower, "article") || strings.HasPrefix(lower, "post-") || strings.HasPrefix(lower, "news-") {
		return true
	}
	return false
}

func isListItemTag(tag string) bool {
	items := map[string]bool{
		"li": true, "tr": true, "dd": true, "dt": true,
		"div": true, "p": true, "article": true,
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

	type tagInfo struct {
		tag   string
		class string
		count int
	}
	tagStats := make(map[string]*tagInfo)
	articleCardCount := 0
	var articleCardTag, articleCardClass string

	children.Each(func(_ int, c *goquery.Selection) {
		ctag := goquery.NodeName(c)
		if _, exists := tagStats[ctag]; !exists {
			tagStats[ctag] = &tagInfo{tag: ctag, count: 0}
		}
		tagStats[ctag].count++
		if tagStats[ctag].class == "" {
			if cls, exists := c.Attr("class"); exists && cls != "" {
				classes := strings.Fields(cls)
				tagStats[ctag].class = classes[0]
			}
		}
		cclass, _ := c.Attr("class")
		if hasArticleItemClass(cclass) {
			articleCardCount++
			if articleCardTag == "" {
				articleCardTag = ctag
				classes := strings.Fields(cclass)
				for _, cls := range classes {
					if hasArticleItemClass(cls) {
						articleCardClass = cls
						break
					}
				}
			}
		}
	})

	if articleCardCount >= 2 && articleCardTag != "" && articleCardClass != "" {
		return base + " > " + articleCardTag + "." + articleCardClass
	}
	if articleCardCount >= 2 && articleCardTag != "" {
		return base + " > " + articleCardTag
	}

	var bestTag *tagInfo
	for _, info := range tagStats {
		if bestTag == nil || info.count > bestTag.count {
			bestTag = info
		}
	}
	if bestTag != nil && bestTag.count >= 2 {
		if bestTag.class != "" {
			return base + " > " + bestTag.tag + "." + bestTag.class
		}
		return base + " > " + bestTag.tag
	}
	return base
}

func detectItemPattern(sel *goquery.Selection) (tag, css string) {
	children := sel.Children()
	articleCardCount := 0
	var articleCardTag, articleCardClass string

	type tagCount struct {
		count int
		class string
	}
	tagStats := make(map[string]*tagCount)

	children.Each(func(_ int, c *goquery.Selection) {
		ctag := goquery.NodeName(c)
		cclass, _ := c.Attr("class")

		if _, exists := tagStats[ctag]; !exists {
			tagStats[ctag] = &tagCount{count: 0}
		}
		tagStats[ctag].count++
		if tagStats[ctag].class == "" {
			classes := strings.Fields(cclass)
			if len(classes) > 0 {
				tagStats[ctag].class = classes[0]
			}
		}
		if hasArticleItemClass(cclass) {
			articleCardCount++
			if articleCardTag == "" {
				articleCardTag = ctag
				classes := strings.Fields(cclass)
				for _, cls := range classes {
					if hasArticleItemClass(cls) {
						articleCardClass = cls
						break
					}
				}
			}
		}
	})

	if articleCardCount >= 2 && articleCardTag != "" && articleCardClass != "" {
		return articleCardTag, articleCardTag + "." + articleCardClass
	}
	if articleCardCount >= 2 && articleCardTag != "" {
		return articleCardTag, articleCardTag
	}

	var bestTag string
	var bestClass string
	bestCount := 0
	for t, info := range tagStats {
		if info.count > bestCount {
			bestCount = info.count
			bestTag = t
			bestClass = info.class
		}
	}
	if bestCount >= 2 && bestTag != "" {
		if bestClass != "" {
			return bestTag, bestTag + "." + bestClass
		}
		return bestTag, bestTag
	}
	return "", ""
}

func extractContainerItems(entry *scanContainerEntry) {
	entry.parent.Children().Each(func(_ int, c *goquery.Selection) {
		ctag := goquery.NodeName(c)
		cclass, _ := c.Attr("class")

		if !isListItemTag(ctag) && !hasArticleItemClass(cclass) {
			return
		}

		item := make(ExtractResult)

		// 策略1: 标题类名元素
		titleClasses := []string{
			"article__card__title", "article-card__title", "post-card__title",
			"card__title", "item__title", "news__title", "title",
			"article-title", "post-title", "entry-title",
		}
		for _, cls := range titleClasses {
			titleEl := c.Find("." + strings.ReplaceAll(cls, " ", "."))
			if titleEl.Length() > 0 {
				text := strings.TrimSpace(titleEl.Text())
				if text != "" {
					setTitle(item, text)
					link := titleEl.Find("a").First()
					if link.Length() > 0 {
						if href, exists := link.Attr("href"); exists {
							item["url"] = href
						}
					}
					break
				}
			}
		}

		// 策略2: 从链接中取标题（需用辅助函数安全判断 nil）
		if getTitle(item) == "" {
			var bestTitle, bestURL string
			c.Find("a").Each(func(_ int, a *goquery.Selection) {
				href, _ := a.Attr("href")
				text := strings.TrimSpace(a.Text())
				if text == "" {
					return
				}
				isBetter := false
				if strings.Contains(href, "/post/") || strings.Contains(href, "/article/") || strings.Contains(href, ".html") || strings.Contains(href, ".htm") || strings.Contains(href, "/item/") {
					isBetter = true
				} else if bestTitle == "" || len(text) > len(bestTitle) {
					isBetter = true
				}
				if isBetter {
					bestTitle, bestURL = text, href
				}
			})
			if bestTitle != "" {
				setTitle(item, bestTitle)
				if bestURL != "" {
					item["url"] = bestURL
				}
			}
		}

		// 策略3: 回退到子项文本
		if getTitle(item) == "" {
			text := strings.TrimSpace(c.Text())
			runes := []rune(text)
			if len(runes) > 3 {
				s := text
				if len([]byte(s)) > 80 {
					s = string(runes[:min(80, len(runes))]) + "..."
				}
				setTitle(item, s)
			}
		}

		// 提取日期
		c.Find("span, time, small, .date, .time, .meta, .publish").Each(func(_ int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if isDateLike(text) && item["date"] == "" {
				item["date"] = text
			}
		})

		if getTitle(item) != "" {
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
			ei := m[keys[i]]
			ej := m[keys[j]]
			realCountI := countArticleCards(ei.parent)
			realCountJ := countArticleCards(ej.parent)

			validI := 0
			for _, item := range ei.items {
				if getTitle(item) != "" {
					validI++
				}
			}
			validJ := 0
			for _, item := range ej.items {
				if getTitle(item) != "" {
					validJ++
				}
			}

			scoreI := realCountI*10000 + validI*100 + len(ei.items)*10 + ei.hits
			scoreJ := realCountJ*10000 + validJ*100 + len(ej.items)*10 + ej.hits

			if scoreJ > scoreI {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
}

func countArticleCards(sel *goquery.Selection) int {
	count := 0
	sel.Children().Each(func(_ int, c *goquery.Selection) {
		ctag := goquery.NodeName(c)
		cclass, _ := c.Attr("class")
		if isListItemTag(ctag) || hasArticleItemClass(cclass) {
			count++
		}
	})
	return count
}

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

	titleSelector := "a"
	urlSelector := "a"
	itemLower := strings.ToLower(itemSel)

	if strings.Contains(itemLower, "article__card") || strings.Contains(itemLower, "article-card") {
		titleSelector = ".article__card__title, .article-card__title, h3"
		urlSelector = ".article__card__link, .article-card__link"
	} else if strings.Contains(itemLower, "article") || itemLower == "article" {
		titleSelector = "h3, h2, .title, a"
	} else if strings.HasPrefix(itemLower, "div.") || strings.HasPrefix(itemLower, "li.") {
		titleSelector = "a"
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
			{Name: "title", Selector: titleSelector, Type: "text"},
			{Name: "url", Selector: urlSelector, Type: "attr", Attr: "href"},
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