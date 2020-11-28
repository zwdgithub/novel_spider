package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"net/http"
	"novel_spider/article"
	"novel_spider/bos_utils"
	"novel_spider/db"
	"novel_spider/log"
	"novel_spider/redis"
	"novel_spider/util"
	"reflect"
	"strconv"
)

type Controller struct {
	service *db.ArticleService
}

var (
	dbConf    *db.MysqlConfig
	bosClient *bos_utils.BosUtil
	dbConn    *gorm.DB
	redisConn *redis.RedisUtil
	service   *db.ArticleService
	ws        = make(map[string]article.NovelWebsites)
	wsInfo    = make(map[string]*article.NovelWebsite)
)

func init() {
	dbConf = db.LoadMysqlConfig("config/conf.yaml")
	bosClient = bos_utils.NewBosClient("config/bos_conf.yaml")
	dbConn = db.New(dbConf)
	redisConn = redis.NewRedis()
	service = db.NewArticleService(dbConn, redisConn, bosClient)
	ws["https://www.xsbiquge.com"] = article.NewXsbiqugeCom(service, redisConn, bosClient)
	ws["https://www.biquge.biz"] = article.NewBiqugeBiz(service, redisConn, bosClient)
	ws["https://www.kanshu5.la"] = article.NewKanshuLa(service, redisConn, bosClient)
	ws["https://www.aikantxt.la"] = article.NewAikantxtLa(service, redisConn, bosClient)
	wsInfo["https://www.xsbiquge.com"] = article.NewXsbiqugeCom(service, redisConn, bosClient).NovelWebsite
	wsInfo["https://www.biquge.biz"] = article.NewBiqugeBiz(service, redisConn, bosClient).NovelWebsite
	wsInfo["https://www.kanshu5.la"] = article.NewKanshuLa(service, redisConn, bosClient).NovelWebsite
	wsInfo["https://www.aikantxt.la"] = article.NewAikantxtLa(service, redisConn, bosClient).NovelWebsite
}

func (controller *Controller) LoadNotMatchLog(c *gin.Context) {
	list := controller.service.LoadNotMatchLogList()
	c.HTML(http.StatusOK, "not-match.tmpl", gin.H{
		"list": list,
	})
}

func (controller *Controller) Delete(c *gin.Context) {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
		})
		return
	}
	controller.service.DeleteNotMatchLog(i)
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
	})
}

func (controller *Controller) Load(c *gin.Context) {
	id := c.Param("id")
	i, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 0,
		})
		return
	}
	notMatchLog := controller.service.LoadNotMatchLog(i)
	info, _ := controller.service.LoadArticleInfoById(notMatchLog.LocalArticleId)
	chapterList := controller.service.AllChapterList(notMatchLog.LocalArticleId)
	c.HTML(http.StatusOK, "info.tmpl", gin.H{
		"notMatchLog": notMatchLog,
		"chapterList": chapterList,
		"info":        info,
	})
}

func (controller *Controller) LoadChapterList(c *gin.Context) {
	url := c.Query("url")
	host := c.Query("host")
	website := ws[host]
	content, err := util.Get(url, wsInfo[host].Encoding, wsInfo[host].Headers)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
		})
		return
	}
	list, err := website.ChapterList(content)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"list": list,
		},
	})
}

func main() {
	dbConf := db.LoadMysqlConfig("config/conf.yaml")
	bosClient := bos_utils.NewBosClient("config/bos_conf.yaml")
	dbConn := db.New(dbConf)
	redisConn := redis.NewRedis()
	service := db.NewArticleService(dbConn, redisConn, bosClient)
	In := make([]reflect.Value, 0)
	In = append(In, reflect.ValueOf(service))
	In = append(In, reflect.ValueOf(redisConn))
	In = append(In, reflect.ValueOf(bosClient))

	controller := new(Controller)
	controller.service = service
	db.New(dbConf)
	gin.SetMode(gin.DebugMode)
	router := gin.Default()
	//router.SetFuncMap()
	router.LoadHTMLGlob("web/templates/*")
	router.GET("/not-match", controller.LoadNotMatchLog)
	router.GET("/delete/:id", controller.Delete)
	router.GET("/load/:id", controller.Load)
	router.GET("/chapter-list", controller.LoadChapterList)
	err := router.Run(":9999")
	if err != nil {
		log.Info("run error: %v", err)
	}
}
