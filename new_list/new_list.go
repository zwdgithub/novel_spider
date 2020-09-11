package main

import (
	"flag"
	"novel_spider/article"
	"novel_spider/bos_utils"
	"novel_spider/db"
	"novel_spider/log"
	"novel_spider/redis"
	"reflect"
)

var register map[string]interface{}

func init() {
	register = make(map[string]interface{})
	register["NewBiqugeBiz"] = article.NewBiqugeBiz
}

func main() {
	var w = flag.String("website", "CreateBiqugeBizSpider", "website reflect")
	dbConf := db.LoadMysqlConfig("config/conf.yaml")
	bosClient := bos_utils.NewBosClient()
	dbConn := db.New(dbConf)
	redisConn := redis.NewRedis()
	service := db.NewArticleService(dbConn, redisConn, bosClient)
	in := make([]reflect.Value, 0)
	in = append(in, reflect.ValueOf(service))
	in = append(in, reflect.ValueOf(redisConn))
	in = append(in, reflect.ValueOf(bosClient))
	callResult := reflect.ValueOf(article.GetCreateSpider(*w)).Call(in)
	if len(callResult) == 0 {
		log.Info("website: %s, call method err", *w)
		return
	}
	spider := callResult[0].Interface().(*article.NovelSpider)
	spider.NewList()
}
