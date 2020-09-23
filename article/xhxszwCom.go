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

type XhxswzCom struct {
	*NovelWebsite
	service *db.ArticleService
	redis   *redis.RedisUtil
}

func NewXhxswzCom(service *db.ArticleService, redis *redis.RedisUtil, bos *bos_utils.BosUtil) *XhxswzCom {
	website := LoadNovelWebsite("config/website.xhxswz.com.yaml")
	log.Info(website)
	website.BosUtil = bos
	c := &XhxswzCom{
		NovelWebsite: website,
		service:      service,
		redis:        redis,
	}
	c.Headers = map[string]string{
		"User-Agent": website.Headers["user-agent"],
	}
	return c
}

func (n *XhxswzCom) ArticleInfo(content string) (*Article, error) {
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

func (n *XhxswzCom) ChapterList(content string) ([]NewChapter, error) {
	newChapters := make([]NewChapter, 0)
	doc, err := htmlquery.Parse(strings.NewReader(content))
	if err != nil {
		return newChapters, err
	}
	reg := regexp.MustCompile("http://www\\.xhxswz\\.com/go/(\\d+)")
	t1 := reg.FindStringSubmatch(content)
	if len(t1) <= 1 {
		return newChapters, errors.New(fmt.Sprintf("ChapterList can not find article id"))
	}
	aid := t1[1]
	nodes := htmlquery.Find(doc, `//div[@id="list-chapterAll"]/dd/a`)
	for _, item := range nodes {
		temp := NewChapter{
			Url:         fmt.Sprintf("%s/go/%s/%s", n.Host, aid, strings.Trim(htmlquery.SelectAttr(item, "href"), " ")),
			ChapterName: util.Trim(htmlquery.InnerText(item)),
		}
		if temp.Url == "" {
			return newChapters, errors.New(fmt.Sprintf("ChapterList url or chapterName is none, url:%s, chapterName: %s", temp.Url, temp.ChapterName))
		}
		if temp.ChapterName == "" {
			temp.ChapterName = fmt.Sprintf("%d", time.Now().Unix())
		}
		newChapters = append(newChapters, temp)
	}

	return newChapters, nil
}

func (n *XhxswzCom) ChapterContent(url string) (string, error) {
	content, err := util.GetWithProxy(url, n.Encoding, n.Headers)
	if err != nil {
		return "", err
	}
	doc, err := htmlquery.Parse(strings.NewReader(content))
	if err != nil {
		return "", err
	}
	cNode, err := htmlquery.Query(doc, `//div[@class="readcontent"]`)
	if err != nil {
		return "", errors.New("")
	}
	if cNode == nil {
		return content, errors.New("content is nil ")
	}
	content = htmlquery.OutputHTML(cNode, false)
	nextHref, err := htmlquery.Query(doc, `//a[@id="linkNext"]`)
	nextText := htmlquery.InnerText(nextHref)
	nextText = strings.TrimSpace(nextText)
	next := htmlquery.SelectAttr(nextHref, "href")
	reg := regexp.MustCompile(`<center>AD4</center></div>([\w\W]*)<p class="text-danger text-center">`)
	if nextText == "下一章" {
		reg = regexp.MustCompile(`<center>AD4</center></div>([\w\W]*)`)
	}
	reg2 := regexp.MustCompile(`<center>AD4</center></div>([\w\W]*)`)
	c1 := reg.FindStringSubmatch(content)
	if len(c1) <= 1 {
		return "", errors.New(fmt.Sprintf("regex match error c1, url: %s", url))
	}
	content = c1[1]
	content = strings.ReplaceAll(content, " ", "")
	content = strings.ReplaceAll(content, "<br>", "\r\n")
	content = strings.ReplaceAll(content, "<br/>", "\r\n")
	content = strings.ReplaceAll(content, "<br >", "\r\n")
	if err != nil {
		return "", errors.New(fmt.Sprintf("regex next match error: %v, url: %s", err, url))
	}
	if nextText == "下一章" {
		return content, nil
	}
	content2, err := util.GetWithProxy(next, n.Encoding, n.Headers)
	if err != nil {
		return "", err
	}
	doc, err = htmlquery.Parse(strings.NewReader(content2))
	if err != nil {
		return "", err
	}
	cNode, err = htmlquery.Query(doc, `//div[@class="readcontent"]`)
	if err != nil {
		return "", errors.New("")
	}
	if cNode == nil {
		return content, errors.New("content is nil ")
	}
	content2 = htmlquery.OutputHTML(cNode, false)
	c1 = reg2.FindStringSubmatch(content2)
	if len(c1) <= 1 {
		return "", errors.New(fmt.Sprintf("regex match error c2, url: %s", url))
	}
	content2 = c1[1]
	content2 = strings.ReplaceAll(content2, " ", "")
	content2 = strings.ReplaceAll(content2, "<br>", "\r\n")
	content2 = strings.ReplaceAll(content2, "<br/>", "\r\n")
	content2 = strings.ReplaceAll(content2, "<br >", "\r\n")
	content += content2
	return content, err
}

func (n *XhxswzCom) Consumer() (string, error) {
	return n.redis.GetUrlFromQueue(n.Host)
}

func (n *XhxswzCom) ConsumerMany() (string, error) {
	return n.redis.GetUrlFromQueue(n.Host + "_many_chapters")
}

func (n *XhxswzCom) NewList() ([]string, error) {
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
