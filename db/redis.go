package db

import (
	"fmt"
	"log"
	"pmon/config"

	"github.com/gomodule/redigo/redis"
)

var redisPool *redis.Pool

func initRedis() {
	cfg := config.GetConfig().DB.Redis
	if cfg == nil {
		return
	}

	log.Println("Init Redis  ...")

	server := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	log.Printf("Connecting %s ...", server)

	redisPool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			return c, err
		},
	}
}
