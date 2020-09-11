package main

import (
	"novel_spider/article"
	"novel_spider/bos_utils"
	"novel_spider/db"
	"novel_spider/redis"
)

func main() {
	dbConf := db.LoadMysqlConfig("config/conf.yaml")
	bosClient := bos_utils.NewBosClient()
	dbConn := db.New(dbConf)
	redisConn := redis.NewRedis()
	service := db.NewArticleService(dbConn, redisConn, bosClient)
	website := article.NewBiqugeBiz(service, redisConn, bosClient)
	spider := article.NewNovelSpider(website, website.NovelWebsite, service, redisConn)
	spider.NewList()
}
