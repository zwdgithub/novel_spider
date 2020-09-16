package main

import (
	"flag"
	"novel_spider/article"
	"novel_spider/bos_utils"
	"novel_spider/db"
	"novel_spider/log"
	"novel_spider/redis"
	"reflect"
	"time"
)

func main() {
	var w = flag.String("website", "CreateBiqugeBizSpider", "website reflect")
	var u = flag.String("url", "", "website reflect")
	flag.Parse()
	log.Infof("website: %s", *w)
	dbConf := db.LoadMysqlConfig("config/conf.yaml")
	bosClient := bos_utils.NewBosClient("config/bos_conf.yaml")
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
	if *u != "" {
		c := make(chan int, 1)
		c <- 1
		spider.Process(article.NewArticle{
			Url:            *u,
			NewChapterName: "",
		}, c)
		return
	}
	go spider.Repair()

	spider.Consumer()

	time.Sleep(time.Second * 10)
}
