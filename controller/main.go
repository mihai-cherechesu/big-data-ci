package main

import (
	"controller/internal"
	"encoding/json"
	"fmt"
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
		log.Fatalf("requests limit reached, %v", err)
	}

	var data map[string]interface{}

	err = json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Fatalf("could not decode body, %v", err)
	}

	for k, _ := range data {
		fmt.Printf("%s\n", k)
	}
}

func main() {
	redisClient = internal.InitRedisClient()
	http.HandleFunc("/execute", handler)

	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatalf("could not listen, %v", err)
	}
}
