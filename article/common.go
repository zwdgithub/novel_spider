package article

import (
	"errors"
	xhttp "github.com/zwdgithub/simple_http"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

func ParseArticleInfo(content string) (*Article, error) {
	var info Article
	reg := regexp.MustCompile(`<meta property="og:title" content="(.+?)"`)
	title := reg.FindStringSubmatch(content)
	if len(title) <= 1 {
		return nil, errors.New("title find error")
	}

	reg = regexp.MustCompile(`<meta property="og:novel:author" content="(.+?)"`)
	author := reg.FindStringSubmatch(content)
	if len(title) <= 1 {
		return nil, errors.New("author find error")
	}
	info.ArticleName = title[1]
	info.Author = author[1]
	index := strings.Index(info.ArticleName, "（")
	if index != -1 {
		info.ArticleName = info.ArticleName[0:index]
	}
	index = strings.Index(info.ArticleName, "(")
	if index != -1 {
		info.ArticleName = info.ArticleName[0:index]
	}
	reg = regexp.MustCompile(`<meta property="og:novel:category" content="(.+?)"`)
	category := reg.FindStringSubmatch(content)
	if len(category) <= 1 {
		return nil, errors.New("category find error")
	}
	info.Category = category[1]
	reg = regexp.MustCompile(`<meta property="og:image" content="(.+?)"`)
	img := reg.FindStringSubmatch(content)
	if len(img) <= 1 {
		return nil, errors.New("category find error")
	}
	info.ImgUrl = img[1]
	reg = regexp.MustCompile(`<meta property="og:description" content="(.+?)"`)
	intro := reg.FindStringSubmatch(content)
	if len(intro) <= 1 {
		info.Intro = ""
	} else {
		info.Intro = intro[1]
	}
	return &info, nil
}

func Get(url string, customClient func(client *http.Client) *http.Client) (string, error) {
	h := xhttp.NewHttpUtil().Get(url)
	if customClient == nil {
		return h.RContent()
	}
	return h.CustomClient(customClient).RContent()
}

func LoadNovelWebsite(fileName string) *NovelWebsite {
	var dst *NovelWebsite
	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(file, &dst)
	if err != nil {
		log.Fatal(err)
	}
	return dst
}
