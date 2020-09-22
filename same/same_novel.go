package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"novel_spider/bos_utils"
	"novel_spider/db"
	"novel_spider/log"
	"novel_spider/model"
	"novel_spider/redis"
	"os"
)

const (
	localPath = "/home/data/local_chapter/%d"
	localFile = "/home/data/local_chapter/%d/%d.txt"
)

func isExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

func loadPre10(item *model.JieqiArticle, service *db.ArticleService, bos *bos_utils.BosUtil) {
	log.Info("process article id %d", item.Articleid)
	cList := service.LoadPreChapter10(item.Articleid)
	for _, c := range cList {
		b, err := bos.GetChapter(c.Articleid, c.Chapterid)
		if err != nil {
			log.Error("chapter get error aid: %d, cid: %d", c.Articleid, c.Chapterid)
			continue
		}
		path := fmt.Sprintf(localPath, item.Articleid)
		if !isExist(path) {
			err := os.MkdirAll(path, 0666)
			if err != nil {
				log.Error("local path make error %s", path)
				return
			}
		}
		fileName := fmt.Sprintf(localFile, c.Articleid, c.Chapterid)
		if !isExist(fileName) {
			f, err := os.Create(fileName)
			if err != nil {
				log.Error("local file make error %s", f)
			}
			f.Write(b)
			f.Close()
			continue
		}
		ioutil.WriteFile(fileName, b, 0666)
	}
}

func main() {
	flag.Parse()
	dbConf := db.LoadMysqlConfig("config/conf.yaml")
	bosClient := bos_utils.NewBosClient("config/bos_conf.yaml")
	dbConn := db.New(dbConf)
	redisConn := redis.NewRedis()
	service := db.NewArticleService(dbConn, redisConn, bosClient)
	list := service.LoadAllArticle()
	log.Info("list len is %d", len(list))
	for _, item := range list {
		loadPre10(item, service, bosClient)
		return
	}
}
