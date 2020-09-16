package redis

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"novel_spider/log"
	"sync"
	"time"
)

const (
	ParsingKey       = "parsing|%s|%s"
	NeedParseListKey = "need_parse_list_%s"
)

type RedisUtil struct {
	conn *redis.Client
	lock *sync.Mutex
}

func NewRedis() *RedisUtil {
	conn := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return &RedisUtil{conn: conn, lock: new(sync.Mutex)}
}

func (r *RedisUtil) Pause(website string) bool {
	v, _ := r.conn.Get("novel_spider_pause").Result()
	if v == "1" {
		return true
	}
	v, _ = r.conn.Get("novel_spider_pause_" + website).Result()
	return v == "1"
}

func (r *RedisUtil) CanParse(articleName, author string) (bool, error) {
	// TODO red lock
	r.lock.Lock()
	defer r.lock.Unlock()
	key := fmt.Sprintf(ParsingKey, articleName, author)
	v, err := r.conn.Incr(key).Result()
	if err != nil {
		return false, err
	}
	if v > 1 {
		return false, err
	}
	p, err := r.conn.HGet(articleName+"|"+author, "parsing").Result()
	if p == "1" {
		r.conn.Del(key)
		return false, nil
	}
	r.conn.HSet(articleName+"|"+author, "parsing", 1)
	r.conn.Expire(key, time.Minute*60*3)
	return true, nil
}

func (r *RedisUtil) ParseEnd(articleName, author string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	key := fmt.Sprintf(ParsingKey, articleName, author)
	r.conn.Del(key)
	r.conn.Del(articleName + "|" + author)
}

func (r *RedisUtil) PutUrlToQueue(website, url string) {
	log.Infof("PutUrlToQueue website: %s, url: %s", website, url)
	key := fmt.Sprintf(NeedParseListKey, website)
	v, err := r.conn.SAdd(website+"_set", url).Result()
	log.Infof("PutUrlToQueue sadd result v: %d, err: %v", v, err)
	if err != nil {
		return
	}
	if v == 1 {
		result := r.conn.LPush(key, url)
		log.Infof("lpush result %v, %v", result, result.Err())
	}
}

func (r *RedisUtil) Retry(website, url string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.PutUrlToQueue(website, url)
}

func (r *RedisUtil) GetUrlFromQueue(website string) (string, error) {
	key := fmt.Sprintf(NeedParseListKey, website)
	v, err := r.conn.BRPop(time.Second*2, key).Result()
	if err != nil {
		return "", err
	}
	if len(v) <= 0 {
		return "", errors.New("do not have some url to parse")
	}
	r.conn.SRem(website+"_set", v[1])
	return v[1], nil
}
