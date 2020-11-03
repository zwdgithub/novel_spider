package db

import (
	"novel_spider/model"
	"time"
)

func (service *ArticleService) SaveNotMachLog(log *model.NovelNotMatchLog) {
	log.CreateTime = time.Now().Format("2006-01-02 15:04:05")
	log.UpdateTime = time.Now().Format("2006-01-02 15:04:05")
	log.Repair = 0
	service.db.Save(log)
}
