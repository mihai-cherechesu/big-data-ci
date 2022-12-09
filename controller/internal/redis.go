package internal

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

func InitRedisClient() *redis.Client {
	log.Printf("initializing redis client")

	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	return client
}

func CheckRequestLimit(addr string, client *redis.Client) error {
	val, err := client.Get(addr).Result()
	if err != nil {
		fmt.Println(err)
	}

	if val == "" {
		err = client.Set(addr, 1, time.Second*10).Err()
		if err != nil {
			fmt.Printf("%v\n", err)
		}

	} else if v, err := strconv.ParseInt(val, 10, 64); err == nil && v < 3 {
		client.Incr(addr)

	} else {
		return errors.New("LIMIT_REACHED")
	}

	return nil
}
