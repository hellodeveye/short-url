package main

import (
	"log"
	"os"
	"strconv"
)

type Env struct {
	S Storage
}

func getEnv() *Env {
	addr := os.Getenv("APP_REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	pwd := os.Getenv("APP_REDIS_PASSWD")
	db := os.Getenv("APP_REDIS_DB")
	if db == "" {
		db = "0"
	}
	d, err := strconv.Atoi(db)
	if err != nil {
		log.Fatal(err)
	}

	r := NewRedisCli(addr, pwd, d)
	return &Env{S: r}
}

func getEnvConfig() *Env {
	r := NewRedisCli(":6379", "redis-pass", 0)
	return &Env{S: r}
}
