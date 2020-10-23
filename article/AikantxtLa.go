package article

import (
	"errors"
	"fmt"
	"github.com/antchfx/htmlquery"
	u "net/url"
	"novel_spider/bos_utils"
	"novel_spider/db"
	"novel_spider/log"
	"novel_spider/redis"
	"novel_spider/util"
	"regexp"
	"strings"
	"time"
)

type AikantxtLa struct {
	*NovelWebsite
	service *db.ArticleService
	redis   *redis.RedisUtil
}

func NewAikantxtLa(service *db.ArticleService, redis *redis.RedisUtil, bos *bos_utils.BosUtil) *AikantxtLa {
	website := LoadNovelWebsite("config/website.aikantxt.la.yaml")
	log.Info(website)
	website.BosUtil = bos
	c := &AikantxtLa{
		NovelWebsite: website,
		service:      service,
		redis:        redis,
	}
	c.Headers = map[string]string{
		"User-Agent": website.Headers["user-agent"],
	}
	return c
}

func (n *AikantxtLa) ArticleInfo(content string) (*Article, error) {
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

func (n *AikantxtLa) ChapterList(content string) ([]NewChapter, error) {
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

func (n *AikantxtLa) ChapterContent(url string) (string, error) {
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
	titleDoc, err := htmlquery.Query(doc, `//title`)
	if err != nil {
		return "", errors.New(fmt.Sprintf("content title is nil, error: %v", err))
	}
	if titleDoc == nil {
		return "", errors.New("content title is nil")
	}
	title := htmlquery.InnerText(titleDoc)
	title = title[0:strings.Index(title, "_")]

	content = htmlquery.OutputHTML(cNode, false)
	reg := regexp.MustCompile(`([\w	\W]*)` + title + "\\(")
	c := reg.FindStringSubmatch(content)
	if len(c) <= 1 {
		return "", errors.New("reg get content error")
	}
	content = c[1]
	content = strings.ReplaceAll(content, " ", "")
	content = strings.ReplaceAll(content, "<br>", "\r\n")
	content = strings.ReplaceAll(content, "<br/>", "\r\n")
	content = strings.ReplaceAll(content, "<br >", "\r\n")

	urlReg := regexp.MustCompile("https://www.aikantxt.la/aikan(\\d+)/(\\d+).html")
	c = urlReg.FindStringSubmatch(url)
	if len(c) != 3 {
		return "", errors.New("reg get url error")
	}
	params := u.Values{}
	params.Add("nbid", c[1])
	params.Add("crid", c[2])
	params.Add("fid", "fb96549631c835eb239cd614cc6b5cb7d295121a")
	c1, err := util.PostForm("https://www.aikantxt.la/content.php", n.Encoding, params, n.Headers)
	if err != nil {
		return "", errors.New(fmt.Sprintf("get content error: %v", err))
	}
	c1 = strings.ReplaceAll(c1, " ", "")
	c1 = strings.ReplaceAll(c1, "&nbsp;", "")
	c1 = strings.ReplaceAll(c1, "<br>", "\r\n")
	c1 = strings.ReplaceAll(c1, "<br/>", "\r\n")
	c1 = strings.ReplaceAll(c1, "<br >", "\r\n")
	if strings.Contains(c1, "502 Bad Gateway") {
		return "", errors.New("get content error 502 bad gateway ")
	}
	content += c1
	return content, err
}

func (n *AikantxtLa) Consumer() (string, error) {
	return n.redis.GetUrlFromQueue(n.Host)
}

func (n *AikantxtLa) ConsumerMany() (string, error) {
	return n.redis.GetUrlFromQueue(n.Host + "_many_chapters")
}

func (n *AikantxtLa) NewList() ([]string, error) {
	r := make([]string, 0)
	return r, nil
}

func (n *AikantxtLa) HasNext() (*NewChapter, error) {
	return nil, nil
}
