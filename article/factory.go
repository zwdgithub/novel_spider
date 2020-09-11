package article

import (
	"novel_spider/bos_utils"
	"novel_spider/db"
	"novel_spider/redis"
)

var (
	methods = make(map[string]func(s *db.ArticleService, r *redis.RedisUtil, b *bos_utils.BosUtil) *NovelSpider)
)

func init() {
	methods["CreateBiqugeBizSpider"] = CreateBiqugeBizSpider
}

func CreateBiqugeBizSpider(service *db.ArticleService, redisConn *redis.RedisUtil, bosClient *bos_utils.BosUtil) *NovelSpider {
	website := NewBiqugeBiz(service, redisConn, bosClient)
	return NewNovelSpider(website, website.NovelWebsite, service, redisConn)
}

func GetCreateSpider(funcName string) func(s *db.ArticleService, r *redis.RedisUtil, b *bos_utils.BosUtil) *NovelSpider {
	return methods[funcName]
}
