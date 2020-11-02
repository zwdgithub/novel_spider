package redis

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"sync"
	"time"
)

const (
	ParsingKey         = "parsing|%s|%s"
	NeedParseListKey   = "need_parse_list_%s"
	RetryDelayQueueKey = "retry_delay_queue_%s"
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
	key := fmt.Sprintf(NeedParseListKey, website)
	v, err := r.conn.SAdd(website+"_set", url).Result()
	if err != nil {
		return
	}
	if v == 1 {
		r.conn.LPush(key, url)
	}
}

func (r *RedisUtil) Retry(website, url string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.conn.ZAdd(fmt.Sprintf(RetryDelayQueueKey, website), redis.Z{
		Score:  float64(time.Now().Unix() + 60), // retry in one minute,
		Member: url,
	})
}

func (r *RedisUtil) GetRetryArticle(website string) ([]redis.Z, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	key := fmt.Sprintf(RetryDelayQueueKey, website)
	result, err := r.conn.ZRangeWithScores(key, 0, 1).Result()
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, errors.New("result empty")
	}
	fmt.Println(int64(result[0].Score), time.Now().Unix())
	if time.Now().Unix() <= int64(result[0].Score) {
		return nil, errors.New("time not match")
	}
	r.conn.ZRem(key, result[0].Member)
	return result, nil
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
	r.conn.SRem("need_parse_list_"+website+"_set", v[1])
	return v[1], nil
}
