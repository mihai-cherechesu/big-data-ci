package main

import (
	"controller/internal"
	"database/sql"
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
	scheduler   *internal.Scheduler
	dbClient    *sql.DB
)

type PipelineRecord struct {
	Id     string
	UserId string
}

type StageRecord struct {
	PipelineId string
	Name       string
	Message    string
	Status     string
}

func handleExecute(w http.ResponseWriter, r *http.Request) {
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

	go scheduler.Schedule(p, ip)
}

func handlePipelines(w http.ResponseWriter, r *http.Request) {
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		http.Error(w, "could not parse remote address, %v", http.StatusInternalServerError)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/pipelines/")

	if id == "" {
		rows, err := dbClient.Query("SELECT * FROM pipelines WHERE user_id = $1", ip)
		if err != nil {
			log.Fatalf("Error executing query: %q", err)
		}
		defer rows.Close()

		var pipelineRecords []PipelineRecord

		for rows.Next() {
			var id string
			var userId string

			err = rows.Scan(&id, &userId)
			if err != nil {
				log.Fatalf("Error scanning rows: %q", err)
			}
			fmt.Printf("ID: %s, Name: %s\n", id, userId)

			r := PipelineRecord{
				Id:     id,
				UserId: userId,
			}

			pipelineRecords = append(pipelineRecords, r)
		}

		err = rows.Err()
		if err != nil {
			log.Fatalf("Error: %q", err)
		}

		response, err := json.Marshal(pipelineRecords)
		if err != nil {
			log.Fatalf("could not marshal list of records, %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)

		// Get all stages for a pipeline id
	} else {
		rows, err := dbClient.Query("SELECT s.pipeline_id, s.name, s.message, s.status FROM stages s INNER JOIN pipelines p ON p.id = s.pipeline_id WHERE p.user_id = $1 AND p.id = $2", ip, id)
		if err != nil {
			log.Fatalf("Error executing query: %q", err)
		}
		defer rows.Close()

		var stageRecords []StageRecord

		for rows.Next() {
			var pipelineId string
			var name string
			var message string
			var status string
			err = rows.Scan(&pipelineId, &name, &message, &status)
			if err != nil {
				log.Fatalf("Error scanning rows: %q", err)
			}
			fmt.Printf("ID: %s, Name: %s, message: %s, status: %s\n", pipelineId, name, message, status)

			r := StageRecord{
				PipelineId: pipelineId,
				Name:       name,
				Message:    message,
				Status:     status,
			}

			stageRecords = append(stageRecords, r)
		}

		err = rows.Err()
		if err != nil {
			log.Fatalf("Error: %q", err)
		}

		response, err := json.Marshal(stageRecords)
		if err != nil {
			log.Fatalf("could not marshal list of records, %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	}
}

func main() {
	redisClient = internal.InitRedisClient()
	dbClient = internal.InitDBConn()
	scheduler = internal.NewScheduler(20, dbClient)

	http.HandleFunc("/execute", handleExecute)
	http.HandleFunc("/pipelines/", handlePipelines)

	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatalf("could not listen, %v", err)
	}
}
