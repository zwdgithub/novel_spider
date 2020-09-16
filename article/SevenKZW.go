package article

import (
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

type SevenKZW struct {
	*NovelWebsite
	service *db.ArticleService
	redis   *redis.RedisUtil
}

func NewSevenKZW(service *db.ArticleService, redis *redis.RedisUtil, bos *bos_utils.BosUtil) *SevenKZW {
	website := LoadNovelWebsite("config/website.7kzw.com.yaml")
	log.Info(website)
	website.BosUtil = bos
	c := &SevenKZW{
		NovelWebsite: website,
		service:      service,
		redis:        redis,
	}
	c.Headers = map[string]string{
		"User-Agent": website.Headers["user-agent"],
	}
	return c
}

func (n *SevenKZW) ArticleInfo(content string) (*Article, error) {
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

func (n *SevenKZW) ChapterList(content string) ([]NewChapter, error) {
	newChapters := make([]NewChapter, 0)
	doc, err := htmlquery.Parse(strings.NewReader(content))
	if err != nil {
		return newChapters, err
	}
	nodes := htmlquery.Find(doc, `//div[@id="list"]/dl/dt[2]/following-sibling::dd/a`)
	for _, item := range nodes {
		temp := NewChapter{
			Url:         n.Host + strings.Trim(htmlquery.SelectAttr(item, "href"), " "),
			ChapterName: util.Trim(htmlquery.InnerText(item)),
		}
		fmt.Println(temp.ChapterName)
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

func (n *SevenKZW) ChapterContent(url string) (string, error) {
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
		return "", errors.New("content is nil ")
	}
	content = htmlquery.OutputHTML(cNode, false)
	reg := regexp.MustCompile(`</p>([\w\W]*)<div align="center">`)
	c := reg.FindStringSubmatch(content)
	if len(c) <= 1 {
		return "", errors.New("reg get content error")
	}
	content = c[1]
	content = strings.ReplaceAll(content, " ", "")
	content = strings.ReplaceAll(content, "<br>", "\r\n")
	content = strings.ReplaceAll(content, "<br/>", "\r\n")
	content = strings.ReplaceAll(content, "<br >", "\r\n")
	return content, err
}

func (n *SevenKZW) Consumer() (string, error) {
	return n.redis.GetUrlFromQueue(n.Host)
}

func (n *SevenKZW) NewList() ([]string, error) {
	r := make([]string, 0)
	return r, nil
}
