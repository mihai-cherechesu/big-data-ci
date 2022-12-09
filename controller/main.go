package main

import (
	"controller/internal"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

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
	var s internal.Schema

	// Parse the request body and bind it to the Request struct
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		// If there is an error parsing the request body, return a 400 response
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the Request struct
	if err := decoder.Decode(&s, r.URL.Query()); err != nil {
		// If there is an error validating the Request struct, return a 400 response
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Create a new graph based on the pipeline stages
	g := internal.NewGraphFromStages(s.Stages)

	// Get the topologically sorted layers from the graph
	for i, layer := range g.TopoSortedLayers() {
		fmt.Printf("%d: %s\n", i, strings.Join(layer, ", "))
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
