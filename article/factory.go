package article

import (
	"novel_spider/bos_utils"
	"novel_spider/db"
	"novel_spider/redis"
)

type Factory struct {
}

func (Factory *Factory) CreateBiqugeBiz(service *db.ArticleService, redisConn *redis.RedisUtil, bosClient *bos_utils.BosUtil) *NovelSpider {
	website := NewBiqugeBiz(service, redisConn, bosClient)
	return NewNovelSpider(website, website.NovelWebsite, service, redisConn)
}
