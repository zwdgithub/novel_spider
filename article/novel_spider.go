package article

import (
	"gotest/db"
	"gotest/model"
	"gotest/redis"
	"gotest/util"
)

type NovelWebsites interface {
	ArticleInfo(content string) (*Article, error)
	LocalArticleInfo(articleName, author string) (*model.JieqiArticle, error)
	ChapterList(content string) ([]string, []string)
	ChapterContent(chapterUrl string) (string, error)
}

type NovelSpider struct {
	ws      NovelWebsites
	wsInfo  *NovelWebsite
	service *db.ArticleService
	redis   *redis.RedisUtil
}

func NewNovelSpider(ws NovelWebsites, wsInfo *NovelWebsite, service *db.ArticleService, redis *redis.RedisUtil) *NovelSpider {
	return &NovelSpider{
		ws:      ws,
		wsInfo:  wsInfo,
		service: service,
		redis:   redis,
	}
}

func (s *NovelSpider) isParsing(articleName, author string) {

}

func (s *NovelSpider) Process(url string) {
	content, err := util.Get(url, s.wsInfo.Headers, s.wsInfo.Encoding)
	if err != nil {
		return
	}
	article, err := s.ws.ArticleInfo(content)
	if err != nil || article == nil || article.ArticleName == "" || article.Author == "" {
		return
	}
	canParse, err := s.redis.CanParse(article.ArticleName, article.Author)
	if err != nil || !canParse {
		return
	}
	defer s.redis.ParseEnd(article.ArticleName, article.Author)

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
		if start {
			content, err := s.ws.ChapterContent(urls[i])
			if err != nil {
				return
			}
			chapter := &model.JieqiChapter{
				Chapterorder: order + 1,
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
}
