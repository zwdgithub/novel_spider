package article

import (
	"errors"
	"fmt"
	"github.com/baidubce/bce-sdk-go/services/bos"
	xhttp "github.com/zwdgithub/simple_http"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"gotest/model"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	EncodingGBK  = "GBK"
	EncodingUTF8 = "UTF-8"
)

type Spider struct {
	client *http.Client
}

func (spider *Spider) Get(url string, headers map[string]string, encoding string) (string, error) {
	h := xhttp.NewHttpUtil()
	h.Get(url).SetHeader(headers).Do()

	if h.Error() != nil {
		return "", h.Error()
	}
	response := h.Response()
	defer response.Body.Close()
	var reader io.Reader
	reader = response.Body
	if encoding == EncodingGBK {
		reader = transform.NewReader(response.Body, simplifiedchinese.GBK.NewDecoder())
	}
	content, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", nil
	}
	// fmt.Println(string(content))
	return string(content), nil
}

type NovelWebsite struct {
	Name     string                 // 网站中文名
	Host     string                 // 域名 qidian.com
	Encoding string                 // utf-8 gbk gb18030
	Spider   Spider                 // spider
	Headers  map[string]string      // header
	Cookie   http.CookieJar         // cookie
	Category map[string]interface{} // 分类
	IProxy   *IProxy
	//DB           *gorm.DB
	HasChapter   bool
	BosClient    *bos.Client
	ChapterList  func(content string) ([]string, []string)
	ArticleInfo  func(content string) (*Article, error)
	ParseContent func(content string) (string, error)
}

func (novel *NovelWebsite) Init() {
	fmt.Println("init ...")

	novel.ChapterList = func(content string) ([]string, []string) {
		chapterUrl := make([]string, 0)
		chapterName := make([]string, 0)
		return chapterUrl, chapterName
	}

	novel.ArticleInfo = func(content string) (*Article, error) {
		info, err := novel.ParseArticleInfo(content)
		if err != nil {
			return nil, err
		}
		return info, nil
	}
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

func (novel *NovelWebsite) NewSpider() {
	// 代理写死了
	var proxy *url.URL
	if novel.IProxy != nil {
		proxyStr := fmt.Sprintf("http://%s:%s@%s",
			novel.IProxy.UserName,
			novel.IProxy.PassWord,
			novel.IProxy.ProxyRaw)
		proxy, _ = url.Parse(proxyStr)
	}
	novel.Spider = Spider{
		client: &http.Client{
			Transport: &http.Transport{Proxy: http.ProxyURL(proxy)},
		},
	}
}

func (novel *NovelWebsite) BuildInfoUrl(novelId string) string {
	return ""
}

func (novel *NovelWebsite) BuildChapterUrl(novelId, chapterId string) string {
	return ""
}

func (novel *NovelWebsite) LocalArticleInfo(articleName, author string) (model.JieqiArticle, error) {
	var info model.JieqiArticle
	//err := novel.DB.Select("articleid, articlename, author, lastchapter, chapters").
	//	Where("articlename = ? and author = ?", articleName, author).Find(&info).Error
	//if err != nil && err.Error() == "record not found" {
	//	err = nil
	//}
	return info, nil

}

func (novel *NovelWebsite) ParseArticleInfo(content string) (*Article, error) {
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
		// info.ArticleName = info.ArticleName[0:index]
	}
	fmt.Println(info)
	return &info, nil
}

func (novel *NovelWebsite) ChapterContent(chapterUrl string) (string, error) {
	content, err := novel.Spider.Get(chapterUrl, novel.Headers, novel.Encoding)
	if err != nil {
		return "", nil
	}
	return content, err
}

func (novel *NovelWebsite) addChapter(chapter model.JieqiChapter) (model.JieqiChapter, error) {
	//err := novel.DB.Create(&chapter).Error
	return chapter, nil
}

func (novel *NovelWebsite) addArticle(article model.JieqiArticle) error {
	//err := novel.DB.Create(&article).Error
	return nil
}

func (novel *NovelWebsite) DownloadCover(url string, articleId int) {

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

func (novel *NovelWebsite) Run(url string) {
	content, err := novel.Spider.Get(url, novel.Headers, novel.Encoding)
	if err != nil {
		return
	}
	article, err := novel.ArticleInfo(content)
	if err != nil || article == nil || article.ArticleName == "" || article.Author == "" {
		return
	}
	local, err := novel.LocalArticleInfo(article.ArticleName, article.Author)
	if err != nil {
		fmt.Println(err)
		return
	}
	if local.Articleid == 0 {
		newArticle := model.JieqiArticle{
			Articlename: article.ArticleName,
			Author:      article.Author,
		}
		err := novel.addArticle(newArticle)
		if err != nil {
			return
		}
	}
	urls, names := novel.ChapterList(content)
	fmt.Println(names[len(names)-1])
	article.LastChapter = names[len(names)-1]
	if article.LastChapter == local.Lastchapter {
		return
	}
	start, order := false, local.Chapters
	if local.Chapters == 0 {
		start = true
	}
	for i, name := range names {
		if start {
			content, err := novel.ChapterContent(urls[i])
			if err != nil {
				return
			}
			chapter := model.JieqiChapter{
				Chapterorder: order + 1,
			}
			chapter, err = novel.addChapter(chapter)
			if err != nil {
				return
			}
			return
			err = novel.putContent(local.Articleid, chapter.Chapterid, content)
			if err != nil {
				return
			}
			order += 1
			continue
		}

		if name == local.Lastchapter {
			start = true
			continue
		}
	}
}
