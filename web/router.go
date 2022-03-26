package main

import "github.com/gin-gonic/gin"

func Router(router *gin.Engine, controller *Controller) {
	router.GET("/not-match", controller.LoadNotMatchLog)
	router.GET("/delete/:id", controller.Delete)
	router.GET("/load/:id", controller.Load)
	router.GET("/chapter-list", controller.LoadChapterList)
	router.GET("/set-last-chapter", controller.SetLastChapter)

	api := router.Group("/api")
	api.POST("/put")
	api.GET("/get-proxy")
}
