package db

import (
	"github.com/gomodule/redigo/redis"
	"github.com/jinzhu/gorm"
)

func Mysql(name string) *gorm.DB {
	db, ok := mysqlDBs[name]
	if ok {
		return db
	}
	panic("mysql db not configured")
}

func Mssql(name string) *gorm.DB {
	db, ok := mssqlDBs[name]
	if ok {
		return db
	}
	panic("mysql db not configured")
}

func Redis() redis.Conn {
	if redisPool == nil {
		panic("redis db not configured")
	}
	return redisPool.Get()
}

func Init() {
	initMssql()
	initMysql()
	initRedis()
}
