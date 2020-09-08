package article

import (
	"novel_spider/db"
	"novel_spider/model"
	"novel_spider/redis"
	"novel_spider/util"
	"time"
)

var ()

type NovelWebsites interface {
	ArticleInfo(content string) (*Article, error)
	LocalArticleInfo(articleName, author string) (*model.JieqiArticle, error)
	ChapterList(content string) ([]string, []string)
	ChapterContent(chapterUrl string) (string, error)
	Consumer() (string, error)
	NewList() ([]string, error)
}

type NovelSpider struct {
	ws      NovelWebsites
	wsInfo  *NovelWebsite
	service *db.ArticleService
	redis   *redis.RedisUtil
}

type NewArticle struct {
	Url            string
	NewChapterName string
}

func NewNovelSpider(ws NovelWebsites, wsInfo *NovelWebsite, service *db.ArticleService, redis *redis.RedisUtil) *NovelSpider {
	return &NovelSpider{
		ws:      ws,
		wsInfo:  wsInfo,
		service: service,
		redis:   redis,
	}
}

func (s *NovelSpider) CanParse(articleName, author string) (bool, error) {
	return s.redis.CanParse(articleName, author)
}

func (s *NovelSpider) ParseEnd(articleName, author string) {
	s.redis.ParseEnd(articleName, author)
}

func (s *NovelSpider) Consumer() {
	c := make(chan int, s.wsInfo.Concurrent)
	for {
		if s.redis.Pause() {
			return
		}
		if len(c) < s.wsInfo.Concurrent {
			url, err := s.ws.Consumer()
			if err != nil {
				time.Sleep(time.Second * 5)
			}
			c <- 1
			go s.Process(NewArticle{
				Url:            url,
				NewChapterName: "",
			}, c)
		}
		time.Sleep(time.Second / 2)
	}
}

func (s *NovelSpider) Process(obj NewArticle, c chan int) {
	defer func() {
		<-c
		if err := recover(); err != nil {
		}
	}()

	content, err := util.Get(obj.Url, s.wsInfo.Encoding, s.wsInfo.Headers)
	if err != nil {
		return
	}
	article, err := s.ws.ArticleInfo(content)
	if err != nil || article == nil || article.ArticleName == "" || article.Author == "" {
		return
	}
	canParse, err := s.CanParse(article.ArticleName, article.Author)
	if err != nil || !canParse {
		return
	}
	defer s.ParseEnd(article.ArticleName, article.Author)

	local, err := s.ws.LocalArticleInfo(article.ArticleName, article.Author)
	if err != nil {
		return
	}
	if local.Articleid == 0 {
		newArticle := &model.JieqiArticle{
			Articlename: article.ArticleName,
			Author:      article.Author,
		}
		err := s.service.AddArticle(newArticle)
		if err != nil {
			return
		}
		local = newArticle
	}

	urls, names := s.ws.ChapterList(content)
	article.LastChapter = names[len(names)-1]
	if article.LastChapter == local.Lastchapter {
		return
	}

	start, order := false, local.Chapters
	if local.Chapters == 0 {
		start = true
	}

	for i, name := range names {
		if s.redis.Pause() {
			return
		}
		if start {
			content, err := s.ws.ChapterContent(urls[i])
			if err != nil {
				return
			}
			chapter := &model.JieqiChapter{
				Chapterorder: order + 1,
				Chaptername:  name,
				Articleid:    local.Articleid,
				Articlename:  local.Articlename,
			}
			chapter, err = s.service.AddChapter(chapter, content)
			if err != nil {
				return
			}
			order += 1
			continue
		}

		if name == local.Lastchapter {
			start = true
		}
	}

	if len(names) > 0 && obj.NewChapterName != "" && names[len(names)-1] != obj.NewChapterName {
		s.redis.PutUrlToQueue(s.wsInfo.Host, obj.Url)
	}
}

func (s *NovelSpider) NewList() {
	list, err := s.ws.NewList()
	if err != nil {
		return
	}
	for _, u := range list {
		s.redis.PutUrlToQueue(s.wsInfo.Host, u)
	}
}
