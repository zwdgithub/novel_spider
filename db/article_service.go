package db

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
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

func (service *ArticleService) UpdateArticleOnAddChapter(article *model.JieqiArticle, size int) error {
	err := service.db.Model(model.JieqiArticle{}).Where("articleid = ?", article.Articleid).Update(map[string]interface{}{
		"lastupdate":    int(time.Now().Unix()),
		"lastchapter":   article.Lastchapter,
		"chapters":      article.Chapters,
		"lastchapterid": article.Lastchapterid,
		"size":          gorm.Expr("size + ?", size),
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
	err = service.UpdateArticleOnAddChapter(article, len(content))
	if err != nil {
		return nil, err
	}
	return chapter, err
}

/**
更新 chapter
*/
func (service *ArticleService) UpdateChapter(chapter *model.JieqiChapter, content string) {
	service.db.Model(model.JieqiChapter{}).Where("chapterid = ?", chapter.Chapterid).
		Updates(map[string]interface{}{
			"size": len(content),
		})
	_ = service.bos.PutChapter(chapter.Articleid, chapter.Chapterid, content)
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

func (service *ArticleService) LoadArticleInfoById(id int) (*model.JieqiArticle, error) {
	var info model.JieqiArticle
	err := service.db.
		Where("articleid = ?", id).Find(&info).Error
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

func (service *ArticleService) NeedRepairChapterListQuick(host string, args ...interface{}) []model.ChapterErrorLog {
	var list []model.ChapterErrorLog
	if len(args) > 0 {
		log.Infof("repair chapter list offset: %d", args[0])
		service.db.Where("`host` = ?  and repair = 0", host).Order("create_time desc").Limit("100").Offset(args[0]).Find(&list)
	} else {
		service.db.Where("`host` = ?  and repair = 0", host).Order("create_time desc").Limit("100").Find(&list)
	}
	return list
}

func (service *ArticleService) ErrorChapter(id int) model.ChapterErrorLog {
	var item model.ChapterErrorLog
	service.db.Where("id = ?", id).Find(&item)
	return item
}

func (service *ArticleService) UpdateErrorChapter(id, retry, repair int, chapter model.JieqiChapter) {
	service.db.Model(model.ChapterErrorLog{}).Where("id = ? and repair = 0", id).Update(map[string]interface{}{
		"retry_num":   retry,
		"repair":      repair,
		"update_time": time.Now().Format("2006-01-02 15:04:05"),
	})
	if repair == 1 {
		service.db.Model(&model.JieqiChapter{}).Where("chapterid = ?", chapter.Chapterid).Updates(map[string]interface{}{
			"size": chapter.Size,
		})
	}
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
	service.db.Where("articleid = ? and chaptertype = 0", aid).Order("chapterid asc").Limit(10).Find(&list)
	return list
}

func (service *ArticleService) RepairSyncSameAll(articleId int) {
	var sameList []model.SameArticle
	service.db.Where("from_article_id = ?", articleId).Find(&sameList)
	for _, item := range sameList {
		var blankChapters []model.JieqiChapter
		service.db.Where("articleid = ? and size < 500 and chaptertype = 0", item.ArticleId).Find(&blankChapters)
		for _, c := range blankChapters {
			var chapter model.JieqiChapter
			service.db.Where("articleid = ? and chaptername = ?", articleId, c.Chaptername).Find(&chapter)
			if chapter.Chapterid != 0 {
				b, err := service.bos.GetChapter(articleId, chapter.Chapterid)
				if err != nil {
					continue
				}
				r := transform.NewReader(bytes.NewReader(b), simplifiedchinese.GBK.NewDecoder())
				b, err = ioutil.ReadAll(r)
				if err != nil {
					log.Infof("trans gbk error, aid: %d, cid: %d", c.Articleid, c.Chapterid)
					return
				}
				content := string(b)
				err = service.bos.PutChapter(c.Articleid, c.Chapterid, content)
				if err == nil {
					service.db.Model(&model.JieqiChapter{}).Where("chapterid = ?", c.Chapterid).Updates(map[string]interface{}{
						"size": len(content),
					})
				}
				log.Infof("repair article: %d, sync article: %d, chapter: %d", articleId, c.Articleid, c.Chapterid)
			}
		}
	}
}

func (service *ArticleService) LoadShortChapter(articleId int, host string) []*model.JieqiChapter {
	var chapters []*model.JieqiChapter
	service.db.Where("articleid = ? and size < 500", articleId).Find(&chapters)
	return chapters
}

func (service *ArticleService) LastChapterList(articleId int, num int) []*model.JieqiChapter {
	var list []*model.JieqiChapter
	service.db.Where("articleid = ?", articleId).Order("chapterorder desc, chapterid desc").Limit(num).Find(&list)
	return list
}

func (service *ArticleService) AllChapterList(articleId int) []*model.JieqiChapter {
	var list []*model.JieqiChapter
	service.db.Where("articleid = ?", articleId).Order("chapterorder, chapterid").Find(&list)
	return list
}

func (service *ArticleService) ChapterCount(articleId int) int {
	var count int
	service.db.Model(model.JieqiChapter{}).Where("articleid = ?", articleId).Count(&count)
	return count
}

func (service *ArticleService) GetLocalContent(articleId, chapterId int) (string, error) {
	b, err := service.bos.GetChapter(articleId, chapterId)
	if err != nil {
		return "", err
	}
	r := transform.NewReader(bytes.NewReader(b), simplifiedchinese.GBK.NewDecoder())
	b, err = ioutil.ReadAll(r)
	if err != nil {
		log.Infof("trans gbk error, aid: %d, cid: %d", articleId, chapterId)
		return "", err
	}
	return string(b), nil
}

func (service *ArticleService) ChapterNameExists(articleId int, chapterName string) bool {
	var chapter model.JieqiChapter
	service.db.Where("articleid = ? and chaptername = ?", articleId, chapterName).Find(&chapter)
	return chapter.Chapterid != 0
}

func (service *ArticleService) SetLastChapter(articleId int, lastChapter string) {
	service.db.Model(model.JieqiArticle{}).Where("articleid = ?", articleId).Updates(map[string]interface{}{
		"lastchapter": lastChapter,
	})
}
