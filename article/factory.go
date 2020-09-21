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
	methods["CreateXsbiqugeComSpider"] = CreateXsbiqugeComSpider
	methods["CreateKanshuLaSpider"] = CreateKanshuLaSpider
	methods["CreateSevenKZWComSpider"] = CreateSevenKZWComSpider
	methods["CreateAikantxtLaSpider"] = CreateAikantxtLaSpider
}

func CreateBiqugeBizSpider(service *db.ArticleService, redisConn *redis.RedisUtil, bosClient *bos_utils.BosUtil) *NovelSpider {
	website := NewBiqugeBiz(service, redisConn, bosClient)
	return NewNovelSpider(website, website.NovelWebsite, service, redisConn)
}

func CreateXsbiqugeComSpider(service *db.ArticleService, redisConn *redis.RedisUtil, bosClient *bos_utils.BosUtil) *NovelSpider {
	website := NewXsbiqugeCom(service, redisConn, bosClient)
	return NewNovelSpider(website, website.NovelWebsite, service, redisConn)
}

func CreateKanshuLaSpider(service *db.ArticleService, redisConn *redis.RedisUtil, bosClient *bos_utils.BosUtil) *NovelSpider {
	website := NewKanshuLa(service, redisConn, bosClient)
	return NewNovelSpider(website, website.NovelWebsite, service, redisConn)
}

func CreateSevenKZWComSpider(service *db.ArticleService, redisConn *redis.RedisUtil, bosClient *bos_utils.BosUtil) *NovelSpider {
	website := NewSevenKZW(service, redisConn, bosClient)
	return NewNovelSpider(website, website.NovelWebsite, service, redisConn)
}

func CreateAikantxtLaSpider(service *db.ArticleService, redisConn *redis.RedisUtil, bosClient *bos_utils.BosUtil) *NovelSpider {
	website := NewAikantxtLa(service, redisConn, bosClient)
	return NewNovelSpider(website, website.NovelWebsite, service, redisConn)
}

func GetCreateSpider(funcName string) func(s *db.ArticleService, r *redis.RedisUtil, b *bos_utils.BosUtil) *NovelSpider {
	return methods[funcName]
}
