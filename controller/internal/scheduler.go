package internal

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// A struct that represents the core scheduler for the CI system.
// Configurations:
// * maxContainers tells the scheduler the maximum number of running containers at a point in time.
// * TBD
type Scheduler struct {
	maxContainers int
	docker        *client.Client
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
	Image string `json:"image" schema:"image"`

	// Allow for any Stages keys
	Stages map[string]StageMeta `json:"stages"`
}

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

	s := &Scheduler{
		maxContainers: maxContainers,
		docker:        docker,
	}

	return s
}

// Function used by goroutines to run the pipeline stages
func runStage(stage string, meta StageMeta, docker *client.Client, doneCh chan string) {
	// rand.Seed(time.Now().UnixNano())
	// t := time.Duration(rand.Intn(6) + 5)

	// fmt.Printf("sleep for %ds\n", t)
	// time.Sleep(t * time.Second)
	ctx := context.Background()

	c, err := docker.ContainerCreate(ctx, &container.Config{
		Image: "alpine",
		Cmd:   meta.Script,
		Tty:   false,
	}, nil, nil, nil, stage)

	if err != nil {
		log.Fatalf("could not create container for stage %s, %v\n", stage, err)
	}

	fmt.Printf("id for the %s container %s\n", stage, c.ID)

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
		fmt.Printf("received status code on wait channel %d\n", status.StatusCode)
	}

	// Write logs to STDOUT
	out, err := docker.ContainerLogs(ctx, c.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	// Remove container after it finishes
	err = docker.ContainerRemove(ctx, c.ID, types.ContainerRemoveOptions{})
	if err != nil {
		log.Fatalf("could not remove container %s for stage %s, %v\n", c.ID, stage, err)
	}

	// Send the stage name in the channel so the Scheduler can
	// traverse the graph again and schedule new stages
	doneCh <- stage
}

func (s *Scheduler) Schedule(p Pipeline) error {
	containers, err := s.docker.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		fmt.Printf("%s %s\n", container.ID[:10], container.Image)
	}

	g := NewGraphFromStages(p.Stages)
	layers := g.TopoSortedLayers()

	for i, layer := range layers {
		fmt.Printf("%d: %s\n", i, layer)
	}

	// Create a channel to receive the stage names that finished
	// Buffered with a maximum capacity of maxContainers
	doneCh := make(chan string, s.maxContainers)

	// Create a new slice to store the initial layer of stages to be executed
	var first []string
	// Number of existing stage layers
	layersLen := len(layers)

	if layersLen > 0 {
		first = layers[0]
	} else {
		log.Fatal("there are no stage layers, recheck your pipeline\n")
	}

	fmt.Printf("the first stage layer to be executed: %s\n", first)
	// Run the first stage layer by creating a goroutine per stage
	for _, stage := range first {
		fmt.Printf("starting stage %s\n", stage)
		go runStage(stage, p.Stages[stage], s.docker, doneCh)
	}

	for {
		select {
		case stage := <-doneCh:
			fmt.Printf("stage %s is done\n", stage)
		default:
			fmt.Print("no stages done, waiting...\n")
			time.Sleep(1 * time.Second)
		}
	}

	return nil
}
