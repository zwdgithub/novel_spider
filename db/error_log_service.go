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

func (service *ArticleService) LoadNotMatchLog() []*model.NovelNotMatchLog {
	var list []*model.NovelNotMatchLog
	fmt.Println(service.db.Find(&list).Error)
	fmt.Println(len(list))
	return list
}

func (service *ArticleService) DeleteNotMatchLog(id int) {
	service.db.Unscoped().Where("id = ?", id).Delete(&model.NovelNotMatchLog{})
}
