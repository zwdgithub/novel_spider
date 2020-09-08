package article

import (
	"fmt"
	"github.com/baidubce/bce-sdk-go/services/bos"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"net/http"
	"strings"
)

type NovelWebsite struct {
	Name         string                 // 网站中文名
	Host         string                 // 域名 qidian.com
	Encoding     string                 // utf-8 gbk gb18030
	Headers      map[string]string      // header
	Cookie       http.CookieJar         // cookie
	Category     map[string]interface{} // 分类
	IProxy       *IProxy
	HasChapter   bool
	BosClient    *bos.Client
	Concurrent   int
	ShortContent int
}

type Article struct {
	ArticleName string
	Author      string
	LastChapter string
	SortId      int
	Intro       string
}

type IProxy struct {
	ProxyRaw string
	UserName string
	PassWord string
}

func (novel *NovelWebsite) putContent(aid, cid int, content string) error {
	reader := transform.NewReader(strings.NewReader(content), simplifiedchinese.GBK.NewDecoder())
	fileName := fmt.Sprintf("%d%d%d", aid, aid, cid)
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}
	_, err = novel.BosClient.PutObjectFromString("bucket", fileName, string(b), nil)
	return err
}
