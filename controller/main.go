package main

import (
	"controller/internal"
	"encoding/json"
	"log"
	"net"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/gorilla/schema"
)

var (
	redisClient *redis.Client
)

func handler(w http.ResponseWriter, r *http.Request) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "could not parse remote address, %v", http.StatusInternalServerError)
		return
	}

	err = internal.CheckRequestLimit(ip, redisClient)
	if err != nil {
		http.Error(w, "requests limit reached, %v", http.StatusTooManyRequests)
		return
	}

	// Create a new schema decoder
	decoder := schema.NewDecoder()

	// Create a new Pipeline struct
	var p internal.Pipeline

	// Parse the request body and bind it to the Pipeline struct
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		// If there is an error parsing the request body, return a 400 response
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the Request struct
	if err := decoder.Decode(&p, r.URL.Query()); err != nil {
		// If there is an error validating the Pipeline struct, return a 400 response
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	s := internal.NewScheduler(20)
	s.Schedule(p)
}

func main() {
	redisClient = internal.InitRedisClient()
	http.HandleFunc("/execute", handler)

	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatalf("could not listen, %v", err)
	}
}
