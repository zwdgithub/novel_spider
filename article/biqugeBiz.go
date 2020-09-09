package article

import (
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"novel_spider/db"
	"novel_spider/util"
	"regexp"
	"strings"
)

type BiqugeBiz struct {
	*NovelWebsite
}

func NewBiqugeBiz(service *db.ArticleService) *BiqugeBiz {
	c := &BiqugeBiz{
		NovelWebsite: &NovelWebsite{
			Name:       "biquge.biz",
			Host:       "biquge.biz",
			Encoding:   "GBK",
			Headers:    nil,
			Cookie:     nil,
			IProxy:     nil,
			Concurrent: 1,
		},
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
		fmt.Println(c[1], c[2])
		newChapters = append(newChapters, NewChapter{
			Url:         c[1],
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
	return "", nil
}

func (n *BiqugeBiz) NewList() ([]string, error) {
	r := make([]string, 0)
	content, err := util.Get("https://www.biquge.biz/", n.Encoding, n.Headers)
	if err != nil {
		return r, err
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return r, err
	}
	doc.Find(".")
	return r, nil
}
