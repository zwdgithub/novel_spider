package article

import (
	"fmt"
	"github.com/baidubce/bce-sdk-go/services/bos"
	"gotest/db"
	"gotest/model"
	"gotest/util"
	"regexp"
)

type BiqugeBiz struct {
	*NovelWebsite
	service    *db.ArticleService
	concurrent int
}

func NewBiqugeBiz(service *db.ArticleService, bosClient *bos.Client) *BiqugeBiz {
	c := &BiqugeBiz{
		NovelWebsite: &NovelWebsite{
			Name:       "biquge.biz",
			Host:       "biquge.biz",
			Encoding:   "GBK",
			Headers:    nil,
			Cookie:     nil,
			IProxy:     nil,
			BosClient:  bosClient,
			Concurrent: 1,
		},
		service: service,
	}
	c.Headers = map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.97 Safari/537.36",
	}
	return c
}

func (n *BiqugeBiz) ArticleInfo(content string) (*Article, error) {
	return ParseArticleInfo(content)
}

func (n *BiqugeBiz) LocalArticleInfo(articleName, author string) (*model.JieqiArticle, error) {
	return n.service.LocalArticleInfo(articleName, author)
}
func (n *BiqugeBiz) ChapterList(content string) ([]string, []string) {
	chapterUrl := make([]string, 0)
	chapterName := make([]string, 0)

	reg := regexp.MustCompile(`<dd><a href="(.+?)"  >(.+?)</a></dd>`)
	chapters := reg.FindAllString(content, -1)

	for _, v := range chapters {
		c := reg.FindStringSubmatch(v)
		fmt.Println(c[1], c[2])
		chapterUrl = append(chapterUrl, c[1])
		chapterName = append(chapterName, c[2])
	}
	return chapterUrl, chapterName
}

func (n *BiqugeBiz) ChapterContent(url string) (string, error) {
	content, err := util.Get(url, n.Encoding, n.Headers)
	if err != nil {
		return "", err
	}
	reg := regexp.MustCompile(``)
	s := reg.FindAllString(content, -1)
	return s[0], err
}

func (n *BiqugeBiz) Consumer() (string, error) {
	return "", nil
}

func (n *BiqugeBiz) NewList() {

}
