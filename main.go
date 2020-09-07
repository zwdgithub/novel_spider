package main

import (
	"gotest/article"
	"gotest/bos_utils"
	"gotest/db"
	"gotest/redis"
)

func main() {
	dbConf := db.LoadMysqlConfig("config/conf.yaml")
	bosClient := bos_utils.NewBosClient()
	dbConn := db.New(dbConf)
	redisConn := redis.NewRedis()
	service := db.NewArticleService(dbConn)
	website := article.NewBiqugeBiz(service, bosClient)
	spider := article.NewNovelSpider(website, website.NovelWebsite, service, redisConn)
	spider.Process("https://www.biquge.biz/39_39082/", nil)

}
