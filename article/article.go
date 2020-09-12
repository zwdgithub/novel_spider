package article

import (
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"novel_spider/bos_utils"
	"strings"
)

type NovelWebsite struct {
	Name              string            `yaml:"name"`     // 网站中文名
	Host              string            `yaml:"host"`     // 域名 qidian.com
	Encoding          string            `yaml:"encoding"` // utf-8 gbk gb18030
	Headers           map[string]string `yaml:"headers"`  // header
	Category          map[string]int    `yaml:"category"` // 分类
	HasChapter        bool              `yaml:"hasChapter"`
	BosUtil           *bos_utils.BosUtil
	Concurrent        int    `yaml:"concurrent"`
	ShortContent      int    `yaml:"shortContent"`
	NewChapterListUrl string `yaml:"newChapterListUrl"`
	Proxy             bool
}

type Article struct {
	ArticleName string
	Author      string
	LastChapter string
	SortId      int
	Intro       string
	ImgUrl      string
	Category    string
}

type IProxy struct {
	ProxyRaw string
	UserName string
	PassWord string
}

func (novel *NovelWebsite) putContent(aid, cid int, content string) error {
	reader := transform.NewReader(strings.NewReader(content), simplifiedchinese.GBK.NewDecoder())
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	err = novel.BosUtil.PutChapter(aid, cid, string(b))
	return err
}
