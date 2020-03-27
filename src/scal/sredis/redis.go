package sredis

import (
	"errors"
	"log"
	"time"

	"github.com/go-redsync/redsync"
	"github.com/gomodule/redigo/redis"
)

// RedisPool defines redis pooling
type RedisPool struct {
	pool *redis.Pool
}

// NewRedisClient will initialize redis
// redisAddr is {host}:{port}
// FIXME: use 6379 as default port
func NewRedisClient(redisAddr string, redisPassword string, redisNamespace int) (*RedisPool, error) {
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisAddr, redis.DialPassword(redisPassword))
			if err != nil {
				return nil, err
			}

			if _, err := c.Do("SELECT", redisNamespace); err != nil {
				c.Close()
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	_, err := pool.Get().Do("PING")
	if err != nil {
		return nil, err
	}

	return &RedisPool{
		pool: pool,
	}, nil
}

// GetPool will get redis pool
func (r *RedisPool) GetPool() *redis.Pool {
	if r.pool == nil {
		log.Fatalln(errors.New("error get redis pool"))
	}
	return r.pool
}

// GetConn will get redis connection
func (r *RedisPool) GetConn() redis.Conn {
	if r.pool == nil {
		log.Fatalln(errors.New("error get redis pool"))
	}
	return r.pool.Get()
}

// GetPoolLocker gets redis mutex locker
func (r *RedisPool) GetPoolLocker() *redsync.Redsync {
	var arrPool []redsync.Pool
	arrPool = append(arrPool, r.GetPool())

	return redsync.New(arrPool)
}
