package internal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/google/uuid"

	"database/sql"

	_ "github.com/lib/pq"
)

const (
	host     = "postgres"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "big-data-ci"
)

// A struct that represents the core scheduler for the CI system.
// Configurations:
// * maxContainers tells the scheduler the maximum number of running containers at a point in time.
// * TBD
type Scheduler struct {
	maxContainers int
	docker        *client.Client
	db            *sql.DB
}

// Metadata for each stage. Must be an object
// with the keys "script", "depends_on", and "artifacts"
type StageMeta struct {
	Script    []string `json:"script"`
	DependsOn []string `json:"depends_on"`
	Artifacts []string `json:"artifacts"`
}

// A struct to represent the JSON schema
type Pipeline struct {
	// TODO: Add a pipeline unique name
	// * Save all the pipelines associated to an ip in Redis
	// * This will allow the CLI to inspect the current running pipelines (e.g. cili pipelines ls)
	// * On each pipeline run, lookup redis to check if ip already has a pipeline with similar name
	// * This also adds unicity to the containers (not sure if they can be created with the same name, check ContainerCreate)
	Name string

	Image string `json:"image" schema:"image"`

	// Allow for any Stages keys
	Stages map[string]StageMeta `json:"stages"`
}

type StageOutput struct {
	Name    string
	Message string
	Status  int64
}

// Used to create an enum for the state of stages
type StageState int64

// Enum for stage state
const (
	NotRunning StageState = iota
	Running
	Finished
)

// Creates a directed graph by iterating through
// the list of dependencies found for each stage at "depends_on" key.
// Cycles are detected automatically while adding the edges
func NewGraphFromStages(stages map[string]StageMeta) *Graph {
	g := NewGraph()

	for k, v := range stages {
		for _, dep := range v.DependsOn {
			g.DependOn(k, dep)
		}
	}

	return g
}

// Creates a new Scheduler struct with configurations.
// Adds a new docker client to the new Scheduler struct
func NewScheduler(maxContainers int) *Scheduler {
	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatalf("could not create docker client, %v", err)
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error pinging database: %q", err)
	}

	fmt.Println("Successfully connected to database!")

	s := &Scheduler{
		maxContainers: maxContainers,
		docker:        docker,
		db:            db,
	}

	return s
}

// Function used by goroutines to run the pipeline stages
func runStage(stage string, meta StageMeta, docker *client.Client, doneCh chan StageOutput) {
	ctx := context.Background()

	reader, err := docker.ImagePull(ctx, "docker.io/library/alpine", types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	defer reader.Close()
	io.Copy(os.Stdout, reader)

	c, err := docker.ContainerCreate(ctx, &container.Config{
		Image: "alpine",
		Cmd:   meta.Script,
		Tty:   false,
	}, nil, nil, nil, stage)

	// Remove the container, similar to the flag --rm passed to docker run
	defer docker.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{})

	if err != nil {
		log.Fatalf("could not create container for stage %s, %v\n", stage, err)
	}

	log.Printf("id for the %s container %s\n", stage, c.ID)

	err = docker.ContainerStart(ctx, c.ID, types.ContainerStartOptions{})
	if err != nil {
		log.Fatalf("could not start container, %v\n", err)
	}

	statusCh, errCh := docker.ContainerWait(ctx, c.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			log.Fatalf("could not wait for container, %v\n", err)
		}
	case status := <-statusCh:
		log.Printf("received status code on wait channel %d\n", status.StatusCode)

		// Write logs to STDOUT
		out, err := docker.ContainerLogs(ctx, c.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
		if err != nil {
			panic(err)
		}

		outBytes, err := ioutil.ReadAll(out)
		if err != nil {
			panic(err)
		}

		log.Printf("outBytes before trim: %v\n", outBytes)
		outBytes = bytes.ReplaceAll(outBytes, []byte{'\x00'}, []byte{})
		log.Printf("outBytes after trim: %v\n", outBytes)

		outBuffer := string(outBytes)
		log.Printf("outBuffer %s\n", outBuffer)

		stageOut := StageOutput{
			Name:    stage,
			Message: outBuffer,
			Status:  status.StatusCode,
		}

		// Send the stage name in the channel so the Scheduler can
		// traverse the graph again and schedule new stages
		doneCh <- stageOut
	}
}

// Find next stages to run by traversing the stage layers.
// This function is called after one dependency stage finishes
func (s *Scheduler) findNextStages(p Pipeline, states map[string]StageState, layers [][]string) []string {
	var nextStages []string

	// The layers array is assured to have at least 2 elements
	for i := 1; i < len(layers); i++ {
		for _, stage := range layers[i] {
			// Stages already running are skipped
			if states[stage] == Running || states[stage] == Finished {
				continue
			}

			// Iterate through the stage dependencies and check
			// their statuses
			allDone := true
			for _, dep := range p.Stages[stage].DependsOn {
				// If any stage dependency is in either NotRunning or Running states,
				// the stage is not considered in the next run
				if states[dep] == Running || states[dep] == NotRunning {
					allDone = false
					break
				}
			}

			// Add the stage to the next run if dependencies finished
			if allDone {
				nextStages = append(nextStages, stage)
			}
		}
	}

	return nextStages
}

// Check if all stages have finished
func (s *Scheduler) checkAllFinished(states map[string]StageState) bool {
	for _, state := range states {
		if state == NotRunning || state == Running {
			return false
		}
	}

	return true
}

func (s *Scheduler) Schedule(p Pipeline, ip string) error {
	p.Name = uuid.New().String()

	_, err := s.db.Exec("INSERT INTO pipelines (id, user_id) VALUES ($1, $2)", p.Name, ip)
	if err != nil {
		log.Fatalf("Error executing query: %q", err)
	}

	g := NewGraphFromStages(p.Stages)
	layers := g.TopoSortedLayers()

	for i, layer := range layers {
		log.Printf("%d: %s\n", i, layer)
	}

	// Create a channel to receive the stage names that finished
	// Buffered with a maximum capacity of maxContainers
	doneCh := make(chan StageOutput, s.maxContainers)

	// Create a map to hold the stages that finished
	states := make(map[string]StageState)
	// Initialize the map with false values for all stages
	for stage := range p.Stages {
		states[stage] = NotRunning
	}

	// Create a new slice to store the initial layer of stages to be executed
	var first []string
	// Number of existing stage layers
	layersLen := len(layers)

	if layersLen > 0 {
		first = layers[0]

		// Case when the pipeline has only one stage
	} else if len(p.Stages) == 1 {
		first = make([]string, 1)
		for k := range p.Stages {
			first[0] = k
		}

		// Case when the pipeline has no stages
	} else {
		log.Fatal("there are no stage layers, recheck your pipeline\n")
	}

	log.Printf("the first stage layer to be executed: %s\n", first)
	// Run the first stage layer by creating a goroutine per stage
	for _, stage := range first {
		log.Printf("starting stage %s\n", stage)
		states[stage] = Running

		_, err := s.db.Exec("INSERT INTO stages (pipeline_id, name, status) VALUES ($1, $2, $3)", p.Name, stage, "RUNNING")
		if err != nil {
			log.Fatalf("Error executing query: %q", err)
		}

		go runStage(stage, p.Stages[stage], s.docker, doneCh)
	}

	for {
		select {
		case stageOutput := <-doneCh:
			log.Printf("stage %s is done with status %d\n", stageOutput.Name, stageOutput.Status)

			if stageOutput.Status != 0 {
				log.Printf("stage %s is failed, aborting pipeline\n", stageOutput.Name)

				_, err := s.db.Exec("UPDATE stages SET status = $1, message = $2 WHERE pipeline_id = $3 AND name = $4", "FAILED", stageOutput.Message, p.Name, stageOutput.Name)
				if err != nil {
					log.Fatalf("Error executing query: %q", err)
				}

				return errors.New("ABORT")
			}

			// Mark the stage as done, so that the stage won't run again
			states[stageOutput.Name] = Finished

			_, err := s.db.Exec("UPDATE stages SET status = $1, message = $2 WHERE pipeline_id = $3 AND name = $4", "SUCCESS", stageOutput.Message, p.Name, stageOutput.Name)
			if err != nil {
				log.Fatalf("Error executing query: %q", err)
			}

			// Check if all stages finished
			if s.checkAllFinished(states) {
				log.Printf("pipeline finished successfully, closing the client\n")
				return nil
			}

			// 1 layer means all stages have no dependencies
			// > 1 layers means there are dependencies
			// Look for stages starting with the second layer
			if len(layers) > 1 {
				log.Printf("looking for other stages to run...\n")
				nextStages := s.findNextStages(p, states, layers)
				log.Printf("found next stages: %s\n", nextStages)

				// Run the next stages and set their status to Running
				for _, n := range nextStages {
					states[n] = Running

					_, err := s.db.Exec("INSERT INTO stages (pipeline_id, name, status) VALUES ($1, $2, $3)", p.Name, n, "RUNNING")
					if err != nil {
						log.Fatalf("Error executing query: %q", err)
					}

					go runStage(n, p.Stages[n], s.docker, doneCh)
				}
			}

		default:
			// Waiting for any stage to finish
			time.Sleep(1 * time.Second)
		}
	}
}
