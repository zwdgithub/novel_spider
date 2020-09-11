package article

import (
	"errors"
	"fmt"
	xhttp "github.com/zwdgithub/simple_http"
	"net/http"
	"regexp"
	"strings"
)

func ParseArticleInfo(content string) (*Article, error) {
	var info Article
	regexpTitle := regexp.MustCompile(`<meta property="og:title" content="(.+?)"/>`)
	title := regexpTitle.FindStringSubmatch(content)
	if len(title) <= 1 {
		return nil, errors.New("title find error")
	}

	regexpAuthor := regexp.MustCompile(`<meta property="og:novel:author" content="(.+?)"/>`)
	author := regexpAuthor.FindStringSubmatch(content)
	if len(title) <= 1 {
		return nil, errors.New("author find error")
	}
	info.ArticleName = title[1]
	info.Author = author[1]
	index := strings.Index(info.ArticleName, "（")
	if index != -1 {
		info.ArticleName = info.ArticleName[0:index]
	}
	fmt.Println(info)
	return &info, nil
}

func Get(url string, customClient func(client *http.Client) *http.Client) (string, error) {
	h := xhttp.NewHttpUtil().Get(url)
	if customClient == nil {
		return h.RContent()
	}
	return h.CustomClient(customClient).RContent()
}
