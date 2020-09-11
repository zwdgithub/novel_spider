package article

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/antchfx/htmlquery"
	"novel_spider/db"
	"novel_spider/log"
	"novel_spider/redis"
	"novel_spider/util"
	"strings"
)

type BiqugeBiz struct {
	*NovelWebsite
	service *db.ArticleService
	redis   *redis.RedisUtil
}

func NewBiqugeBiz(service *db.ArticleService, redis *redis.RedisUtil) *BiqugeBiz {
	c := &BiqugeBiz{
		NovelWebsite: &NovelWebsite{
			Name:              "biquge.biz",
			Host:              "https://www.biquge.biz",
			Encoding:          "GBK",
			Headers:           nil,
			Cookie:            nil,
			IProxy:            nil,
			Concurrent:        3,
			NewChapterListUrl: "https://www.biquge.biz",
		},
		service: service,
		redis:   redis,
	}
	c.Headers = map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.97 Safari/537.36",
	}
	return c
}

func (n *BiqugeBiz) ArticleInfo(content string) (*Article, error) {
	return ParseArticleInfo(content)
}

func (n *BiqugeBiz) ChapterList(content string) ([]NewChapter, error) {
	newChapters := make([]NewChapter, 0)
	doc, err := htmlquery.Parse(strings.NewReader(content))
	if err != nil {
		return newChapters, err
	}
	nodes := htmlquery.Find(doc, `//div[@id="list"]/dl/dd/a`)
	for _, item := range nodes {
		temp := NewChapter{
			Url:         n.Host + strings.Trim(htmlquery.SelectAttr(item, "href"), " "),
			ChapterName: strings.Trim(htmlquery.InnerText(item), " "),
		}
		if temp.Url == "" || temp.ChapterName == "" {
			return newChapters, errors.New(fmt.Sprintf("url or chapterName is none, url:%s, chapterName: %s", temp.Url, temp.ChapterName))
		}
		newChapters = append(newChapters, temp)
	}

	return newChapters, nil
}

func (n *BiqugeBiz) ChapterContent(url string) (string, error) {
	content, err := util.Get(url, n.Encoding, n.Headers)
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
	content = htmlquery.OutputHTML(cNode, false)
	content = strings.ReplaceAll(content, "&nbsp;", "")
	content = strings.ReplaceAll(content, "<br>", "\r\n")
	content = strings.ReplaceAll(content, "<br/>", "\r\n")
	content = strings.ReplaceAll(content, "<br >", "\r\n")
	if len(content) < n.ShortContent {
		return "", errors.New(fmt.Sprintf("short content, url: %s", url))
	}
	log.Info(content)
	return content, err
}

func (n *BiqugeBiz) Consumer() (string, error) {
	return n.redis.GetUrlFromQueue(n.Host)
}

func (n *BiqugeBiz) NewList() ([]string, error) {
	r := make([]string, 0)
	content, err := util.Get(n.NewChapterListUrl, n.Encoding, n.Headers)
	if err != nil {
		return r, err
	}
	doc, err := htmlquery.Parse(strings.NewReader(content))
	if err != nil {
		return r, err
	}
	liList := htmlquery.Find(doc, `//div[@id="newscontent"]/div[1]/ul/li`)
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
