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

	// 4. 按匹配数排序，并提高"有实际标题"容器的优先级
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

// findBestContainer 寻找最适合作为容器的祖先元素
// 使用评分制：遍历所有祖先，找出最精确的那个容器
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
			hasLinks := 0

			children.Each(func(_ int, c *goquery.Selection) {
				ctag := goquery.NodeName(c)
				cclass, _ := c.Attr("class")

				if isListItemTag(ctag) || hasArticleItemClass(cclass) {
					childCount++
				}

				if c.Find("a").Length() > 0 {
					linkCount++
				}

				if c.Find("a").Find("img").Length() > 0 || c.Find("img").Length() > 0 {
					hasLinks++
				}

				if hasArticleItemClass(cclass) {
					articleCardCount++
				}
			})

			// 评分系统
			score := 0

			// article-card 模式是最明确的信号
			score += articleCardCount * 100

			// 子项数量
			score += childCount * 10

			// 链接数量很重要
			score += linkCount * 5

			// 传统列表模式
			if childCount >= 2 && linkCount >= childCount/2 {
				score += 50
			}

			// 深度加分：更深 = 更精确
			score += depth * 2

			// 排除页脚/导航等低质量容器
			lowerClass := strings.ToLower(class)
			if strings.Contains(lowerClass, "footer") || strings.Contains(lowerClass, "header") ||
				strings.Contains(lowerClass, "nav") || strings.Contains(lowerClass, "sidebar") ||
				strings.Contains(lowerClass, "menu") || strings.Contains(lowerClass, "banner") {
				score = 0
			}

			// 排除 section__wrapper 中不是 feed 的
			if strings.Contains(lowerClass, "section__wrapper") {
				if !strings.Contains(lowerClass, "feed") && articleCardCount == 0 {
					score -= 20
				}
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

	// 如果找到有意义的容器，返回它
	if best.score >= 20 {
		return best.sel
	}

	// 回退：查找最近的 ul/ol 列表
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
		"article": true, "main": true, "nav": false,
	}
	return containers[tag]
}

func hasArticleClass(class string) bool {
	// 检测常见的文章列表容器类名
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
	// 也检测以 article 或 post 开头的特定 class 模式
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

	// 统计每种标签的出现次数，以及是否有 article-card 类
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

		// 记录第一个子项的 class
		if tagStats[ctag].class == "" {
			if cls, exists := c.Attr("class"); exists && cls != "" {
				classes := strings.Fields(cls)
				tagStats[ctag].class = classes[0]
			}
		}

		// 检测是否 article-card 模式
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

	// 优先：如果检测到 article-card 模式，用 article-card 的标签和类
	if articleCardCount >= 2 && articleCardTag != "" && articleCardClass != "" {
		return base + " > " + articleCardTag + "." + articleCardClass
	}
	if articleCardCount >= 2 && articleCardTag != "" {
		return base + " > " + articleCardTag
	}

	// 找到最常见的子标签
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

	// 优先检测 article-card 模式
	children.Each(func(_ int, c *goquery.Selection) {
		cclass, _ := c.Attr("class")
		if hasArticleItemClass(cclass) {
			articleCardCount++
			ctag := goquery.NodeName(c)
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
	return "", ""
}

func extractContainerItems(entry *scanContainerEntry) {
	entry.parent.Children().Each(func(_ int, c *goquery.Selection) {
		ctag := goquery.NodeName(c)
		cclass, _ := c.Attr("class")

		// 接受：列表项标签 或 article-card 类
		if !isListItemTag(ctag) && !hasArticleItemClass(cclass) {
			return
		}

		item := make(ExtractResult)

		// 策略 1: 先找标题类名元素（最精确）
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
					item["title"] = text
					// 从标题元素内找链接
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

		// 策略 2: 从链接中找标题（优先找文章链接，跳过作者/头像链接）
		if item["title"] == "" {
			var bestTitle, bestURL string
			c.Find("a").Each(func(_ int, a *goquery.Selection) {
				href, _ := a.Attr("href")
				text := strings.TrimSpace(a.Text())
				if text == "" {
					return
				}
				// 优先选择：URL 包含 /post/ 的链接，其次是标题较长的链接
				isBetter := false
				if strings.Contains(href, "/post/") || strings.Contains(href, "/article/") || strings.Contains(href, "/item/") {
					isBetter = true
				} else if bestTitle == "" || len(text) > len(bestTitle) {
					isBetter = true
				}
				if isBetter {
					bestTitle, bestURL = text, href
				}
			})
			if bestTitle != "" {
				item["title"] = bestTitle
				if bestURL != "" {
					item["url"] = bestURL
				}
			}
		}

		// 如果还是没有标题，回退到卡片文本并截取前 80 字
		if item["title"] == "" {
			text := strings.TrimSpace(c.Text())
			if text != "" {
				if len(text) > 80 {
					text = text[:80] + "..."
				}
				item["title"] = text
			}
		}

		// 提取日期
		c.Find("span, time, small, .date, .time, .meta, .publish").Each(func(_ int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if isDateLike(text) && item["date"] == "" {
				item["date"] = text
			}
		})

		if len(item) > 0 && item["title"] != "" {
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

			// 计算容器内 article-card 实际数量（不依赖提取结果）
			realCountI := countArticleCards(ei.parent)
			realCountJ := countArticleCards(ej.parent)

			// 计算有效标题数（非空标题）
			validI := 0
			for _, item := range ei.items {
				if title, ok := item["title"]; ok && title != "" {
					validI++
				}
			}
			validJ := 0
			for _, item := range ej.items {
				if title, ok := item["title"]; ok && title != "" {
					validJ++
				}
			}

			// 排序优先级：article-card 子项数 > 有效标题数 > 总条数 > 关键词命中
			scoreI := realCountI*10000 + validI*100 + len(ei.items)*10 + ei.hits
			scoreJ := realCountJ*10000 + validJ*100 + len(ej.items)*10 + ej.hits

			if scoreJ > scoreI {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
}

// countArticleCards 统计容器内 article-card 子项数量
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

	// 智能选择字段选择器
	// 根据 item 类型选择合适的标题和链接提取方式
	titleSelector := "a"
	urlSelector := "a"
	itemLower := strings.ToLower(itemSel)

	if strings.Contains(itemLower, "article__card") || strings.Contains(itemLower, "article-card") {
		// article-card 模式：标题在 .article__card__title 或 h3 中
		// URL 在 .article__card__link 中
		titleSelector = ".article__card__title, .article-card__title, h3"
		urlSelector = ".article__card__link, .article-card__link"
	} else if strings.Contains(itemLower, "article") || itemLower == "article" {
		// 通用 article 模式：用类名精确匹配
		titleSelector = "h3, h2, .title, a"
	} else if strings.HasPrefix(itemLower, "div.") || strings.HasPrefix(itemLower, "li.") {
		// 带 class 的 div/li：优先用 class 内的链接
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
