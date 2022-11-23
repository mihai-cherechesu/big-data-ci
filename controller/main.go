package main

import (
	"controller/internal"
	"log"
	"net"
	"net/http"

	"github.com/go-redis/redis"
)

var (
	redisClient *redis.Client
)

func handler(w http.ResponseWriter, r *http.Request) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Fatalf("could not parse remote address, %v", err)
	}
	err = internal.CheckRequestLimit(ip, redisClient)
	if err != nil {
		log.Printf("requests limit reached, %v", err)
	}
}

func main() {
	redisClient = internal.InitRedisClient()
	http.HandleFunc("/", handler)

	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatalf("could not listen, %v", err)
	}
}
