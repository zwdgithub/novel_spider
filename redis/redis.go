package redis

import (
	"fmt"
	"github.com/go-redis/redis"
	"sync"
	"time"
)

const (
	ParsingKey = "parsing|%s|%s"
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
	return &RedisUtil{conn: conn}
}

func (r *RedisUtil) CanParse(articleName, author string) (bool, error) {
	// TODO red lock
	r.lock.Lock()
	defer r.lock.Unlock()
	key := fmt.Sprintf(ParsingKey, articleName, author)
	v, err := r.conn.Get(key).Result()
	if err != nil {
		return false, err
	}
	if v == "1" {
		return false, err
	}
	r.conn.Set(key, "1", time.Minute*60)
	return true, nil
}

func (r *RedisUtil) ParseEnd(articleName, author string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	key := fmt.Sprintf(ParsingKey, articleName, author)
	r.conn.Del(key)
}
