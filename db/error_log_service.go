package db

import (
	"fmt"
	"novel_spider/model"
	"time"
)

func (service *ArticleService) SaveNotMachLog(log *model.NovelNotMatchLog) {
	log.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	log.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
	log.Repair = 0
	service.db.Save(log)
}

func (service *ArticleService) LoadNotMatchLogList() []*model.NovelNotMatchLog {
	var list []*model.NovelNotMatchLog
	fmt.Println(service.db.Find(&list).Error)
	fmt.Println(len(list))
	return list
}

func (service *ArticleService) LoadNotMatchLog(id int) *model.NovelNotMatchLog {
	var notMatchLog model.NovelNotMatchLog
	service.db.Where("id = ?", id).Find(&notMatchLog)
	return &notMatchLog
}

func (service *ArticleService) DeleteNotMatchLogByArticleId(id int) {
	service.db.Unscoped().Where("local_article_id = ?", id).Delete(&model.NovelNotMatchLog{})
}

func (service *ArticleService) DeleteNotMatchLog(id int) {
	service.db.Unscoped().Where("id = ?", id).Delete(&model.NovelNotMatchLog{})
}
