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
	"github.com/lib/pq"
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
	PipelineId   string
	Name         string
	Messages     []string
	Status       string
	ArtifactUrls []string
}

type StageSubrecord struct {
	PipelineId string
	Name       string
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

func handleStages(w http.ResponseWriter, r *http.Request) {
	var ids []string

	if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
		// If there is an error parsing the request body, return a 400 response
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	q := "SELECT s.pipeline_id, s.name, s.status, FROM stages s INNER JOIN pipelines p ON p.id = s.pipeline_id WHERE p.id = ANY($1)"
	rows, err := dbClient.Query(q, pq.Array(ids))
	if err != nil {
		log.Fatalf("Error executing query: %q", err)
	}
	defer rows.Close()

	var stageSubrecords []StageSubrecord

	for rows.Next() {
		var pipelineId string
		var name string
		var status string

		err = rows.Scan(&pipelineId, &name, &status)
		if err != nil {
			log.Fatalf("Error scanning rows: %q", err)
		}

		r := StageSubrecord{
			PipelineId: pipelineId,
			Name:       name,
			Status:     status,
		}

		stageSubrecords = append(stageSubrecords, r)
	}

	err = rows.Err()
	if err != nil {
		log.Fatalf("Error: %q", err)
	}

	formatted := make(map[string][]map[string]string)
	for _, s := range stageSubrecords {
		_, ok := formatted[s.PipelineId]
		if !ok {
			formatted[s.PipelineId] = make([]map[string]string, 0)
		}

		ns := map[string]string{
			"name":   s.Name,
			"status": s.Status,
		}

		formatted[s.PipelineId] = append(formatted[s.PipelineId], ns)
	}

	response, err := json.Marshal(formatted)
	if err != nil {
		log.Fatalf("could not marshal list of records, %v", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
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
		rows, err := dbClient.Query("SELECT s.pipeline_id, s.name, s.message, s.status, s.artifact_urls FROM stages s INNER JOIN pipelines p ON p.id = s.pipeline_id WHERE p.user_id = $1 AND p.id = $2", ip, id)
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
			var artifactUrls pq.StringArray

			err = rows.Scan(&pipelineId, &name, &message, &status, &artifactUrls)
			if err != nil {
				log.Fatalf("Error scanning rows: %q", err)
			}

			// Filter evil urls?
			urls := []string(artifactUrls)
			for i, u := range urls {
				if len(u) < 80 {
					urls = append(urls[:i], urls[i+1:]...)
				}
			}

			messages := strings.Split(strings.Trim(message, "\n"), "\n")

			r := StageRecord{
				PipelineId:   pipelineId,
				Name:         name,
				Messages:     messages,
				Status:       status,
				ArtifactUrls: urls,
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
	http.HandleFunc("/stages", handleStages)

	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		log.Fatalf("could not listen, %v", err)
	}
}
