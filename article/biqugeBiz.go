package article

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/antchfx/htmlquery"
	"novel_spider/db"
	"novel_spider/redis"
	"novel_spider/util"
	"regexp"
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
			Concurrent:        2,
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
	reg := regexp.MustCompile(`<dd><a href="(.+?)"  >(.+?)</a></dd>`)
	chapters := reg.FindAllString(content, -1)

	for _, v := range chapters {
		c := reg.FindStringSubmatch(v)
		newChapters = append(newChapters, NewChapter{
			Url:         n.Host + c[1],
			ChapterName: c[2],
		})
	}
	return newChapters, nil
}

func (n *BiqugeBiz) ChapterContent(url string) (string, error) {
	content, err := util.Get(url, n.Encoding, n.Headers)
	if err != nil {
		return "", err
	}
	reg := regexp.MustCompile(`<div id="content">(.+?)</div>`)
	c := reg.FindStringSubmatch(content)
	if len(c) <= 1 {
		return "", errors.New(fmt.Sprintf("chapter content regex error, err:%v, url: %s", err, url))
	}
	c[1] = strings.ReplaceAll(c[1], "&nbsp;", "")
	c[1] = strings.ReplaceAll(c[1], "<br>", "\r\n")
	c[1] = strings.ReplaceAll(c[1], "<br/>", "\r\n")
	c[1] = strings.ReplaceAll(c[1], "<br >", "\r\n")
	if len(c[1]) < n.ShortContent {
		return "", errors.New(fmt.Sprintf("short content, url: %s", url))
	}
	return c[1], err
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
		r = append(r, s)
	}
	fmt.Println(r)
	return r, nil
}
