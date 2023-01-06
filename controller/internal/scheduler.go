package internal

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/google/uuid"

	"database/sql"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// A struct that represents the core scheduler for the CI system.
// Configurations:
// * maxContainers tells the scheduler the maximum number of running containers at a point in time.
type Scheduler struct {
	maxContainers int
	docker        *client.Client
	db            *sql.DB
}

// A struct to represent the elements from the depends_on list
type DependsOnMeta struct {
	Stage          string `json:"stage"`
	FetchArtifacts bool   `json:"artifacts"`
}

// Metadata for each stage. Must be an object
// with the keys "script", "depends_on", and "artifacts"
type StageMeta struct {
	Script    []string        `json:"script"`
	DependsOn []DependsOnMeta `json:"depends_on"`
	Artifacts []string        `json:"artifacts"`
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
	Name         string
	Message      string
	Status       int64
	ContainerId  string
	ArtifactUrls []string
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
			g.DependOn(k, dep.Stage)
		}
	}

	return g
}

// Creates a new Scheduler struct with configurations.
// Adds a new docker client to the new Scheduler struct
func NewScheduler(maxContainers int, dbClient *sql.DB) *Scheduler {
	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatalf("could not create docker client, %v", err)
	}

	s := &Scheduler{
		maxContainers: maxContainers,
		docker:        docker,
		db:            dbClient,
	}

	return s
}

// Function used by goroutines to run the pipeline stages
func runStage(stage string, pipeline Pipeline, stageToContainerId map[string]string, docker *client.Client, doneCh chan StageOutput) {
	ctx := context.Background()
	meta, ok := pipeline.Stages[stage]
	if !ok {
		log.Fatalf("cannot run stage %s\n", stage)
	}

	imageFQDN := strings.Split(pipeline.Image, "/")
	var registry, image string

	if len(imageFQDN) == 2 {
		registry = imageFQDN[0]
		image = imageFQDN[1]

	} else if len(imageFQDN) == 1 {
		registry = "library"
		image = imageFQDN[0]

	} else {
		log.Fatalf("image has wrong format, %s\n", imageFQDN)
	}

	reader, err := docker.ImagePull(ctx, "docker.io/"+registry+"/"+image, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	defer reader.Close()
	io.Copy(os.Stdout, reader)

	c, err := docker.ContainerCreate(ctx, &container.Config{
		Image: pipeline.Image,
		Cmd:   meta.Script,
		Tty:   false,
	}, nil, nil, nil, pipeline.Name+"-"+stage)

	if err != nil {
		log.Fatalf("could not create container for stage %s, %v\n", stage, err)
	}

	for _, d := range meta.DependsOn {
		if d.FetchArtifacts {
			for _, f := range pipeline.Stages[d.Stage].Artifacts {
				CopyFromContainerToContainer(docker, stageToContainerId[d.Stage], f, c.ID, "./")
			}
		}
	}

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
		artifactUrls := make([]string, len(meta.Artifacts))

		// Send all artifacts to S3 only if the stage finished successfully
		if status.StatusCode == 0 {
			for _, f := range meta.Artifacts {
				log.Printf("uploading artifact %s to S3\n", f)
				artifactUrls = append(artifactUrls, UploadArtifactFromContainer(docker, pipeline.Name, stage, c.ID, f))
			}
		}

		// Write logs to STDOUT
		out, err := docker.ContainerLogs(ctx, c.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
		if err != nil {
			panic(err)
		}

		outBytes, err := ioutil.ReadAll(out)
		if err != nil {
			panic(err)
		}

		stageOut := StageOutput{
			Name:         stage,
			Message:      string(ReplaceControlBytes(outBytes)),
			Status:       status.StatusCode,
			ContainerId:  c.ID,
			ArtifactUrls: artifactUrls,
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
				if states[dep.Stage] == Running || states[dep.Stage] == NotRunning {
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
	stageToContainerId := make(map[string]string)
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

		_, err := s.db.Exec("INSERT INTO stages (pipeline_id, name, status) VALUES ($1, $2, $3)",
			p.Name, stage, "RUNNING")
		if err != nil {
			log.Fatalf("Error executing query: %q", err)
		}

		go runStage(stage, p, stageToContainerId, s.docker, doneCh)
	}

	for {
		select {
		case stageOutput := <-doneCh:
			stageToContainerId[stageOutput.Name] = stageOutput.ContainerId
			log.Printf("stage %s is done with status %d and container %s\n", stageOutput.Name, stageOutput.Status, stageOutput.ContainerId)

			if stageOutput.Status != 0 {
				log.Printf("stage %s is failed, aborting pipeline\n", stageOutput.Name)

				_, err := s.db.Exec("UPDATE stages SET status = $1, message = $2 WHERE pipeline_id = $3 AND name = $4",
					"FAILED", stageOutput.Message, p.Name, stageOutput.Name)
				if err != nil {
					log.Fatalf("Error executing query: %q", err)
				}

				return errors.New("ABORT")
			}

			// Mark the stage as done, so that the stage won't run again
			states[stageOutput.Name] = Finished

			psqlArr := pq.Array(stageOutput.ArtifactUrls)

			_, err := s.db.Exec("UPDATE stages SET status = $1, message = $2, artifact_urls = $3 WHERE pipeline_id = $4 AND name = $5",
				"SUCCESS", stageOutput.Message, psqlArr, p.Name, stageOutput.Name)
			if err != nil {
				log.Fatalf("Error executing query: %q", err)
			}

			// Check if all stages finished
			if s.checkAllFinished(states) {
				log.Printf("pipeline finished successfully, closing the client\n")

				// Remove the containers
				for _, v := range stageToContainerId {
					s.docker.ContainerRemove(context.Background(), v, types.ContainerRemoveOptions{})
				}
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

					_, err := s.db.Exec("INSERT INTO stages (pipeline_id, name, status) VALUES ($1, $2, $3)",
						p.Name, n, "RUNNING")
					if err != nil {
						log.Fatalf("Error executing query: %q", err)
					}

					go runStage(n, p, stageToContainerId, s.docker, doneCh)
				}
			}

		default:
			// Waiting for any stage to finish
			time.Sleep(1 * time.Second)
		}
	}
}
