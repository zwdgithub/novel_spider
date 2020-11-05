package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"novel_spider/bos_utils"
	"novel_spider/db"
	"novel_spider/log"
	"novel_spider/redis"
	"strconv"
)

type Controller struct {
	service *db.ArticleService
}

func (controller *Controller) LoadNotMatchLog(c *gin.Context) {
	list := controller.service.LoadNotMatchLog()
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

func main() {
	dbConf := db.LoadMysqlConfig("config/conf.yaml")
	bosClient := bos_utils.NewBosClient("config/bos_conf.yaml")
	dbConn := db.New(dbConf)
	redisConn := redis.NewRedis()
	service := db.NewArticleService(dbConn, redisConn, bosClient)
	controller := new(Controller)
	controller.service = service
	db.New(dbConf)
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	//router.SetFuncMap()
	router.LoadHTMLGlob("web/templates/*")
	router.GET("/not-match", controller.LoadNotMatchLog)
	router.GET("/delete/:id", controller.Delete)
	err := router.Run(":9999")
	if err != nil {
		log.Info("run error: %v", err)
	}
}
