package db

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"novel_spider/bos_utils"
	"novel_spider/log"
	"novel_spider/model"
	"novel_spider/redis"
	"novel_spider/util"
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
	if article.Sortid == 0 {
		article.Sortid = 7
	}
	article.Posterid = 0
	article.Postdate = int(time.Now().Unix())
	article.Lastupdate = int(time.Now().Unix())
	err := service.db.Create(article).Error
	return err
}

func (service *ArticleService) UpdateArticleOnAddChapter(article *model.JieqiArticle) error {
	err := service.db.Model(model.JieqiArticle{}).Where("articleid = ?", article.Articleid).Update(map[string]interface{}{
		"lastupdate":    int(time.Now().Unix()),
		"lastchapter":   article.Lastchapter,
		"chapters":      article.Chapters,
		"lastchapterid": article.Lastchapterid,
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
		Articleid:     chapter.Articleid,
		Lastchapter:   chapter.Chaptername,
		Chapters:      chapter.Chapterorder,
		Lastchapterid: chapter.Chapterid,
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

func (service *ArticleService) AddErrorChapter(log model.ChapterErrorLog) {
	log.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	log.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
	service.db.Create(&log)
}

func (service *ArticleService) NeedRepairChapterList(host string, args ...interface{}) []model.ChapterErrorLog {
	var list []model.ChapterErrorLog
	a, _ := time.ParseDuration(fmt.Sprintf("-%dh", 24*7))
	n := time.Now().Add(a).Format("2006-01-02 15:04:05")
	if len(args) > 0 {
		log.Infof("repair chapter list offset: %d", args[0])
		service.db.Where("`host` = ? and retry_num < 10 and create_time > ? and repair = 0", host, n).Order("create_time").Limit("100").Offset(args[0]).Find(&list)
	} else {
		service.db.Where("`host` = ? and retry_num < 10 and create_time > ? and repair = 0", host, n).Order("create_time").Limit("100").Find(&list)
	}
	return list
}

func (service *ArticleService) UpdateErrorChapter(id, retry, repair int) {

	service.db.Model(model.ChapterErrorLog{}).Where("id = ? and repair = 0", id).Update(map[string]interface{}{
		"retry_num":   retry,
		"repair":      repair,
		"update_time": time.Now().Format("2006-01-02 15:04:05"),
	})
}

func (service *ArticleService) PutContent(aid, cid int, content string) error {
	return service.bos.PutChapter(aid, cid, content)
}

func (service *ArticleService) LastSecondChapter(articleId int) (string, error) {
	var list []model.JieqiChapter
	service.db.Where("articleid = ?", articleId).Order("chapterorder desc, chapterid desc").Limit(10).Find(&list)
	if len(list) >= 2 {
		return util.Trim(list[1].Chaptername), nil
	}
	return "", errors.New("service LastSecondChapter error")
}

func (service *ArticleService) LoadAllArticle() []*model.JieqiArticle {
	var list []*model.JieqiArticle
	service.db.Select("articleid").Order("articleid asc").Find(&list)
	return list
}

func (service *ArticleService) LoadPreChapter10(aid int) []*model.JieqiChapter {
	var list []*model.JieqiChapter
	service.db.Where("articleid = ?", aid).Order("chapterid asc").Limit(10).Find(&list)
	return list
}
