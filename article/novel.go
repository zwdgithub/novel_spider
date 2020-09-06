package article

import "gotest/model"

type Novel interface {
	NewSpider()
	BuildInfoUrl(novelId string) string
	BuildChapterUrl(novelId, chapterId string) string
	LocalArticleInfo(articleName, author string) (model.JieqiArticle, error)
	ArticleInfo(content string) (*Article, error)
	ParseArticleInfo(content string) (*Article, error)
	ChapterList(content string) ([]string, []string)
	ChapterContent(chapterUrl string) (string, error)
	addChapter(chapter model.JieqiChapter) (model.JieqiChapter, error)
	addArticle(article model.JieqiArticle) error
	putContent(aid, cid int, content string) error
	Run(url string)
}
