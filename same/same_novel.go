package main

import (
	"bytes"
	"flag"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io/ioutil"
	"novel_spider/bos_utils"
	"novel_spider/db"
	"novel_spider/log"
	"novel_spider/model"
	"novel_spider/redis"
	"os"
)

const (
	localPath = "/mnt/local/local_chapter/%d/%d"
	localFile = "/mnt/local/local_chapter/%d/%d/%d.txt"
)

var (
	match = flag.String("match", "CreateBiqugeBizSpider", "website reflect")
)

func isExist(f string) bool {
	_, err := os.Stat(f)
	return err == nil || os.IsExist(err)
}

func loadPre10(item *model.JieqiArticle, service *db.ArticleService, bos *bos_utils.BosUtil) {
	log.Infof("process article id %d", item.Articleid)
	cList := service.LoadPreChapter10(item.Articleid)
	for _, c := range cList {
		fileName := fmt.Sprintf(localFile, c.Articleid/1000, c.Articleid, c.Chapterid)
		if isExist(fileName) {
			continue
		}
		b, err := bos.GetChapter(c.Articleid, c.Chapterid)
		if err != nil {
			log.Error("chapter get error aid: %d, cid: %d", c.Articleid, c.Chapterid)
			continue
		}
		r := transform.NewReader(bytes.NewReader(b), simplifiedchinese.GBK.NewDecoder())
		b, err = ioutil.ReadAll(r)
		if err != nil {
			log.Infof("trans gbk error, aid: %d, cid: %d", c.Articleid, c.Chapterid)
			return
		}
		path := fmt.Sprintf(localPath, item.Articleid/1000, item.Articleid)
		if !isExist(path) {
			err := os.MkdirAll(path, 0666)
			if err != nil {
				log.Error("local path make error %s", path)
				return
			}
		}

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

func matchSameNovel(list []*model.JieqiArticle) {
	//invalidArticle := make(map[int]bool)
	//matchList := make([]int, 0)
	//for _, item := range list {
	//	if articleId, ok := invalidArticle[item.Articleid]; !ok {
	//		for _, matchId := range matchList {
	//			p := fmt.Sprintf(localPath, item.Articleid/1000, item.Articleid)
	//
	//		}
	//	}
	//}
}

func main() {
	flag.Parse()
	dbConf := db.LoadMysqlConfig("config/conf.yaml")
	bosClient := bos_utils.NewBosClient("config/bos_conf.yaml")
	dbConn := db.New(dbConf)
	redisConn := redis.NewRedis()
	service := db.NewArticleService(dbConn, redisConn, bosClient)
	list := service.LoadAllArticle()
	log.Infof("list len is %d", len(list))
	if *match == "match" {

		return
	}
	for _, item := range list {
		loadPre10(item, service, bosClient)
	}
}
