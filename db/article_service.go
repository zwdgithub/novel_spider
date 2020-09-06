package db

import (
	"github.com/jinzhu/gorm"
	"gotest/model"
)

type ArticleService struct {
	db *gorm.DB
}

func NewArticleService(db *gorm.DB) *ArticleService {
	return &ArticleService{
		db: db,
	}
}

func (service *ArticleService) AddArticle(article *model.JieqiArticle) error {
	err := service.db.Create(article).Error
	return err
}

func (service *ArticleService) AddChapter(chapter *model.JieqiChapter, content string) (*model.JieqiChapter, error) {
	err := service.db.Create(&chapter).Error
	// TODO put content
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
