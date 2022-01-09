package main

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/mattheath/base62"
	"time"
)

const (
	UrlIdKey           = "next.url.id"
	ShortLinkKey       = "shortLink:%s:url"
	UrlHashKey         = "urlHash:%s:url"
	ShortLinkDetailKey = "shortLink:%s:detail"
)

type RedisCli struct {
	Cli *redis.Client
}

func (r RedisCli) Shorten(url string, exp int64) (string, error) {
	h := toSha1(url)
	d, err := r.Cli.Get(fmt.Sprintf(UrlHashKey, h)).Result()
	if err == redis.Nil {

	} else if err != nil {
		return "", err
	} else {
		if d == "{}" {

		} else {
			return d, nil
		}
	}
	err = r.Cli.Incr(UrlIdKey).Err()
	if err != nil {
		return "", err
	}

	id, err := r.Cli.Get(UrlIdKey).Int64()

	if err != nil {
		return "", err
	}

	eid := base62.EncodeInt64(id)
	err = r.Cli.Set(fmt.Sprintf(ShortLinkKey, eid), url, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}
	err = r.Cli.Set(fmt.Sprintf(UrlHashKey, h), eid, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	detail, err := json.Marshal(
		&URLDetail{
			url,
			time.Now().String(),
			time.Duration(exp),
		})

	if err != nil {
		return "", err
	}

	err = r.Cli.Set(fmt.Sprintf(ShortLinkDetailKey, eid), detail, time.Minute*time.Duration(exp)).Err()

	return eid, nil
}

func (r RedisCli) ShortLinkInfo(eid string) (interface{}, error) {
	d, err := r.Cli.Get(fmt.Sprintf(ShortLinkDetailKey, eid)).Result()
	if err == redis.Nil {
		return "", StatusError{404, errors.New("unknown short URL")}
	} else if err != nil {
		return "", nil
	}
	return d, nil
}

func (r RedisCli) UnShorten(eid string) (string, error) {
	url, err := r.Cli.Get(fmt.Sprintf(ShortLinkKey, eid)).Result()
	if err == redis.Nil {
		return "", StatusError{404, err}
	} else if err != nil {
		return "", nil
	}
	return url, nil
}

type URLDetail struct {
	URL                 string        `json:"url"`
	CreatedAt           string        `json:"created_at"`
	ExpirationInMinutes time.Duration `json:"expiration_in_minutes"`
}

func NewRedisCli(addr string, pwd string, db int) *RedisCli {
	cli := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pwd,
		DB:       db,
	})

	if _, err := cli.Ping().Result(); err != nil {
		panic(err)
	}
	return &RedisCli{cli}
}

func toSha1(str string) string {
	var (
		sha = sha1.New()
	)
	return string(sha.Sum([]byte(str)))
}
