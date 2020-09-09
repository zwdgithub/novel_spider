package article

import (
	"errors"
	"novel_spider/db"
	"novel_spider/model"
	"novel_spider/redis"
	"novel_spider/util"
	"time"
)

var (
	contentShortError = errors.New("content length too short")
	chapterNotMatch   = errors.New("no chapter need to update ")
)

type NovelWebsites interface {
	ArticleInfo(content string) (*Article, error)
	ChapterList(content string) ([]NewChapter, error)
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

type NewChapter struct {
	Url         string
	ChapterName string
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
		if s.redis.Pause(s.wsInfo.Host) {
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

	local, err := s.service.LocalArticleInfo(article.ArticleName, article.Author)
	if err != nil {
		return
	}
	if local.Articleid == 0 {
		newArticle := &model.JieqiArticle{
			Articlename: article.ArticleName,
			Author:      article.Author,
		}
		err := s.service.AddArticle(newArticle)
		// TODO download cover
		if err != nil {
			return
		}
		local = newArticle
	}

	allChapters, err := s.ws.ChapterList(content)
	if err != nil || len(allChapters) == 0 {
		return
	}
	targetLast := obj.NewChapterName
	if targetLast == "" {
		targetLast = allChapters[len(allChapters)-1].ChapterName
	}

	article.LastChapter = targetLast
	if article.LastChapter == local.Lastchapter {
		return
	}

	order := local.Chapters
	newChapters := make([]NewChapter, 0)
	match := false
	if local.Chapters == 0 {
		match = true
	}
	if !match {
		for _, item := range allChapters {
			if item.ChapterName == local.Lastchapter {
				match = true
			}
			if match {
				newChapters = append(newChapters)
			}
		}
	}
	if !match {
		return
	}
	if len(newChapters) == 0 {
		return
	}

	for _, item := range newChapters {
		if s.redis.Pause(s.wsInfo.Host) {
			return
		}
		content, err := s.ws.ChapterContent(item.Url)
		if err != nil {
			return
		}
		chapter := &model.JieqiChapter{
			Chapterorder: order + 1,
			Chaptername:  item.ChapterName,
			Articleid:    local.Articleid,
			Articlename:  local.Articlename,
		}
		chapter, err = s.service.AddChapter(chapter, content)
		if err != nil {
			return
		}
		order += 1
	}

	if len(newChapters) > 0 && obj.NewChapterName != "" && newChapters[len(newChapters)-1].ChapterName != obj.NewChapterName {
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
