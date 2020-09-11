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

func main() {
	var w = flag.String("website", "CreateBiqugeBiz", "website reflect")
	flag.Parse()

	dbConf := db.LoadMysqlConfig("config/conf.yaml")
	bosClient := bos_utils.NewBosClient()
	dbConn := db.New(dbConf)
	redisConn := redis.NewRedis()
	service := db.NewArticleService(dbConn, redisConn, bosClient)
	factory := new(article.Factory)
	in := make([]reflect.Value, 0)
	in = append(in, reflect.ValueOf(service))
	in = append(in, reflect.ValueOf(redisConn))
	in = append(in, reflect.ValueOf(bosClient))
	callResult := reflect.ValueOf(factory).MethodByName(*w).Call(in)
	if len(callResult) == 0 {
		log.Info("website: %s, call method err", *w)
		return
	}
	spider := callResult[0].Interface().(*article.NovelSpider)
	spider.Consumer()
}
