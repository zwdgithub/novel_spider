package article

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/antchfx/htmlquery"
	"novel_spider/bos_utils"
	"novel_spider/db"
	"novel_spider/log"
	"novel_spider/redis"
	"novel_spider/util"
	"regexp"
	"strings"
	"time"
)

type BiqumoCom struct {
	*NovelWebsite
	service *db.ArticleService
	redis   *redis.RedisUtil
}

func NewBiqumoCom(service *db.ArticleService, redis *redis.RedisUtil, bos *bos_utils.BosUtil) *BiqumoCom {
	website := LoadNovelWebsite("config/website.biqumo.com.yaml")
	log.Info(website)
	website.BosUtil = bos
	c := &BiqumoCom{
		NovelWebsite: website,
		service:      service,
		redis:        redis,
	}
	c.Headers = map[string]string{
		"User-Agent": website.Headers["user-agent"],
	}
	return c
}

func (n *BiqumoCom) ArticleInfo(content string) (*Article, error) {
	article, err := ParseArticleInfo(content)
	if err != nil {
		return nil, err
	}
	if v, ok := n.Category[article.Category]; ok {
		article.SortId = v
	}
	if article.SortId == 0 {
		article.SortId = 7
	}
	log.Infof("article info :%v", article)
	return article, err
}

func (n *BiqumoCom) ChapterList(content string) ([]NewChapter, error) {
	newChapters := make([]NewChapter, 0)
	doc, err := htmlquery.Parse(strings.NewReader(content))
	if err != nil {
		return newChapters, err
	}
	nodes := htmlquery.Find(doc, `//div[@class="listmain"]/dl/dt[2]/following-sibling::dd/a`)
	for _, item := range nodes {
		temp := NewChapter{
			Url:         n.Host + strings.Trim(htmlquery.SelectAttr(item, "href"), " "),
			ChapterName: util.Trim(htmlquery.InnerText(item)),
		}
		if temp.Url == "" {
			return newChapters, errors.New(fmt.Sprintf("url or chapterName is none, url:%s, chapterName: %s", temp.Url, temp.ChapterName))
		}
		if temp.ChapterName == "" {
			temp.ChapterName = fmt.Sprintf("%d", time.Now().Unix())
		}
		newChapters = append(newChapters, temp)
	}

	return newChapters, nil
}

func (n *BiqumoCom) ChapterContent(url string) (string, error) {
	content, err := util.GetWithProxy(url, n.Encoding, n.Headers)
	if err != nil {
		return "", err
	}
	doc, err := htmlquery.Parse(strings.NewReader(content))
	if err != nil {
		return "", err
	}
	cNode, err := htmlquery.Query(doc, `//div[@id="content"]`)
	if err != nil {
		return "", errors.New("")
	}
	if cNode == nil {
		return content, errors.New("content is nil ")
	}
	content = htmlquery.OutputHTML(cNode, false)
	reg := regexp.MustCompile(`<script>app2\(\);</script>([\w\W]*)<script>[a-zA-Z0-9]{1,10}\(\)`)
	c1 := reg.FindStringSubmatch(content)
	if len(c1) <= 1 {
		return "", errors.New(fmt.Sprintf("regex match error c1, url: %s", url))
	}
	content = c1[1]
	content = strings.ReplaceAll(content, "readx()", "")
	content = strings.ReplaceAll(content, " ", "")
	content = strings.ReplaceAll(content, "<br>", "\r\n")
	content = strings.ReplaceAll(content, "<br/>", "\r\n")
	content = strings.ReplaceAll(content, "<br >", "\r\n")
	return content, err
}

func (n *BiqumoCom) Consumer() (string, error) {
	return n.redis.GetUrlFromQueue(n.Host)
}

func (n *BiqumoCom) ConsumerMany() (string, error) {
	return n.redis.GetUrlFromQueue(n.Host + "_many_chapters")
}

func (n *BiqumoCom) NewList() ([]string, error) {
	r := make([]string, 0)
	content, err := util.Get(n.NewChapterListUrl, n.Encoding, n.Headers)
	if err != nil {
		log.Infof("new list get content error %v", err)
		return r, err
	}
	doc, err := htmlquery.Parse(strings.NewReader(content))
	if err != nil {
		log.Infof("new list parse content error %v", err)
		return r, err
	}
	liList := htmlquery.Find(doc, `//div[@class="l bd"]/ul/li`)
	log.Infof("lilist len is %d", len(liList))
	if len(liList) == 0 {
		log.Infof(content)
	}
	for _, item := range liList {
		articleInfo := htmlquery.Find(item, `./span[@class="s2"]/a`)
		newChapter := htmlquery.Find(item, `./span[@class="s3"]/a`)
		authorInfo := htmlquery.Find(item, `./span[@class="s4"]`)
		if len(articleInfo) == 0 || len(newChapter) == 0 || len(authorInfo) == 0 {
			return r, errors.New("new list find article or chapter is none")
		}
		href := htmlquery.SelectAttr(articleInfo[0], "href")
		if strings.Contains(href, "id") {
			continue
		}
		articleName := htmlquery.InnerText(articleInfo[0])
		author := htmlquery.InnerText(authorInfo[0])
		newChapterName := htmlquery.InnerText(newChapter[0])
		if href == "" || newChapterName == "" {
			return r, errors.New("new list find articleName or chapterName is blank")
		}
		exists, err := n.service.LocalArticleInfo(articleName, author)
		if err != nil {
			return r, err
		}
		if exists.Lastchapter == newChapterName {
			continue
		}
		b, _ := json.Marshal(NewArticle{
			Url:            n.Host + href,
			NewChapterName: newChapterName,
		})
		s := string(b)
		log.Infof("%s, need update %s", n.Host, s)
		r = append(r, s)
	}
	return r, nil
}

func (n *BiqumoCom) HasNext(url string) *NewChapter {
	return nil
}
