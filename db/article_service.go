package db

import (
	"github.com/jinzhu/gorm"
	"novel_spider/bos_utils"
	"novel_spider/model"
	"novel_spider/redis"
	"time"
)

type ArticleService struct {
	db    *gorm.DB
	bos   *bos_utils.BosUtil
	redis *redis.RedisUtil
}

func NewArticleService(db *gorm.DB, redis *redis.RedisUtil, bos *bos_utils.BosUtil) *ArticleService {
	return &ArticleService{
		db:    db,
		bos:   bos,
		redis: redis,
	}
}

func (service *ArticleService) GetRedis() *redis.RedisUtil {
	return service.redis
}

func (service *ArticleService) AddArticle(article *model.JieqiArticle) error {
	article.Posterid = 0
	article.Postdate = int(time.Now().Unix())
	article.Lastupdate = int(time.Now().Unix())
	err := service.db.Create(article).Error
	return err
}

func (service *ArticleService) UpdateArticleOnAddChapter(article *model.JieqiArticle) error {
	err := service.db.Model(model.JieqiArticle{}).Where("articleid = ?", article.Articleid).Update(map[string]interface{}{
		"lastupdate":  int(time.Now().Unix()),
		"lastchapter": article.Lastchapter,
		"chapters":    article.Chapters,
	}).Error
	return err
}

func (service *ArticleService) AddChapter(chapter *model.JieqiChapter, content string) (*model.JieqiChapter, error) {
	chapter.Postdate = int(time.Now().Unix())
	chapter.Lastupdate = int(time.Now().Unix())
	chapter.Posterid = 0
	chapter.Poster = "a"
	chapter.Size = len(content)
	err := service.db.Create(chapter).Error
	if err != nil {
		return chapter, err
	}

	err = service.bos.PutChapter(chapter.Articleid, chapter.Chapterid, content)
	if err != nil {
		service.db.Unscoped().Where("chapterid = ?", chapter.Chapterid).Delete(&chapter)
		return nil, err
	}
	article := &model.JieqiArticle{
		Articleid:   chapter.Articleid,
		Lastchapter: chapter.Chaptername,
		Chapters:    chapter.Chapterorder,
	}
	err = service.UpdateArticleOnAddChapter(article)
	if err != nil {
		return nil, err
	}
	return chapter, err
}

func (service *ArticleService) LocalArticleInfo(articleName, author string) (*model.JieqiArticle, error) {
	var info model.JieqiArticle
	err := service.db.Select("articleid, articlename, author, lastchapter, chapters").
		Where("articlename = ? and author = ?", articleName, author).Find(&info).Error
	if err != nil && err.Error() == "record not found" {
		err = nil
	}
	return &info, err

}
