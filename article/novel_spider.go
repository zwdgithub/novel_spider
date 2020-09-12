package article

import (
	"encoding/json"
	"errors"
	"novel_spider/db"
	"novel_spider/log"
	"novel_spider/model"
	"novel_spider/redis"
	"novel_spider/util"
	"runtime"
	"strings"
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
			log.Infof("%s, spider stop", s.wsInfo.Host)
			break
		}
		if len(c) < s.wsInfo.Concurrent {
			content, err := s.ws.Consumer()
			if err != nil {
				time.Sleep(time.Second * 5)
			}
			var obj NewArticle
			err = json.Unmarshal([]byte(content), &obj)
			if err != nil {
				continue
			}
			c <- 1
			go s.Process(obj, c)
		}
		time.Sleep(time.Second / 2)
	}
	for len(c) > 0 {

	}
	log.Infof("%s, stop success", s.wsInfo.Host)
}

func (s *NovelSpider) Process(obj NewArticle, c chan int) {
	defer func() {
		<-c
		if err := recover(); err != nil {
			log.Errorf("process %s, err: %v", obj.Url, err)
			stack := make([]byte, 1024*8)
			stack = stack[:runtime.Stack(stack, false)]
			log.Error(string(stack))
		}
		log.Infof("process %s, end", obj.Url)
	}()
	log.Infof("process %s, start", obj.Url)
	content, err := util.Get(obj.Url, s.wsInfo.Encoding, s.wsInfo.Headers)
	if err != nil {
		log.Infof("process %s, http get error: %v", obj.Url, err)
		return
	}
	article, err := s.ws.ArticleInfo(content)
	if err != nil || article == nil || article.ArticleName == "" || article.Author == "" {
		log.Infof("process %s, parse article info error, ", obj.Url)
		return
	}
	canParse, err := s.CanParse(article.ArticleName, article.Author)
	if err != nil {
		log.Infof("process url: %s, can not parse now, error: %v", obj.Url, err)
		return
	}
	if !canParse {
		log.Infof("process url: %s, can not parse now,", obj.Url)
		return
	}
	defer s.ParseEnd(article.ArticleName, article.Author)
	local, err := s.service.LocalArticleInfo(article.ArticleName, article.Author)
	if err != nil {
		log.Infof("process %s, get local info error: %v ", obj.Url, err)
		return
	}
	if local.Articleid == 0 {
		newArticle := &model.JieqiArticle{
			Articlename: article.ArticleName,
			Author:      article.Author,
			Intro:       article.Intro,
			Sortid:      article.SortId,
		}
		err := s.service.AddArticle(newArticle)
		_ = s.wsInfo.BosUtil.PutCover(article.ImgUrl, newArticle.Articleid)
		if err != nil {
			log.Infof("process %s, add new article error %v", obj.Url, err)
			return
		}
		local = newArticle
	}

	allChapters, err := s.ws.ChapterList(content)
	if err != nil || len(allChapters) == 0 {
		log.Infof("process %s, parse chapter list error: %v", obj.Url, err)
		return
	}
	targetLast := obj.NewChapterName
	if targetLast == "" {
		targetLast = allChapters[len(allChapters)-1].ChapterName
	}

	article.LastChapter = targetLast
	if article.LastChapter == local.Lastchapter {
		log.Infof("process %s, need not update", obj.Url)
		return
	}

	order := local.Chapters
	newChapters := make([]NewChapter, 0)
	match := false
	if local.Chapters == 0 {
		match = true
	}
	for _, item := range allChapters {
		if strings.Trim(item.ChapterName, " ") == strings.Trim(local.Lastchapter, " ") {
			match = true
		}
		if match {
			newChapters = append(newChapters, item)
		}
	}
	if !match {
		log.Infof("process %s, no chapter match, info: %s, %s, %s, %s", obj.Url, local.Articlename, local.Author, newChapters[len(newChapters)-1].ChapterName, local.Lastchapter)
		return
	}

	log.Infof("process %s, need crawl chapter %d", obj.Url, len(newChapters))
	if len(newChapters) == 0 {
		log.Infof("process %s, new chapters none, info: name:%s, author:%s, last:%s", obj.Url, article.ArticleName, article.Author, article.LastChapter)
		return
	}

	retry := true
	if obj.NewChapterName != "" {
		retry = false
	}
	addChapterNum := 0
	for _, item := range newChapters {
		if s.redis.Pause(s.wsInfo.Host) {
			log.Infof("process %s stop", obj.Url)
			return
		}
		content, err := s.ws.ChapterContent(item.Url)
		if err != nil {
			log.Infof("process %s get content error: %v", obj.Url, err)
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
			s.redis.Retry(s.wsInfo.Host, obj.Url)
			log.Infof("process %s add chapter error: %v", obj.Url, err)
			return
		}
		addChapterNum++
		order += 1
		if obj.NewChapterName != "" && obj.NewChapterName == item.ChapterName {
			retry = false
		}
	}
	log.Infof("process %s, success, add %d chapter", obj.Url, addChapterNum)

	if retry {
		log.Infof("process %s need retry, new: %s, old:%s", obj.Url, obj.NewChapterName, newChapters[len(newChapters)-1].ChapterName)
		s.redis.Retry(s.wsInfo.Host, obj.Url)
	}
	return
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
