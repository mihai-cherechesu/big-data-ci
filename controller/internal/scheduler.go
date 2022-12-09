package internal

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// A struct that represents the core scheduler for the CI system.
// Configurations:
// * maxContainers tells the scheduler the maximum number of running containers at a point in time.
// * TBD
type Scheduler struct {
	maxContainers int
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

func NewGraphFromStages(stages map[string]StageMeta) *Graph {
	g := NewGraph()

	for k, v := range stages {
		for _, dep := range v.DependsOn {
			g.DependOn(k, dep)
		}
	}

	return g
}

func NewScheduler(maxContainers int) *Scheduler {
	s := &Scheduler{
		maxContainers: maxContainers,
	}

	return s
}

func runStage(stage string, done chan string) {
	rand.Seed(time.Now().UnixNano())
	t := time.Duration(rand.Intn(6) + 5)

	fmt.Printf("sleep for %ds\n", t)
	time.Sleep(t * time.Second)

	done <- stage
}

func (s *Scheduler) Schedule(p Pipeline) error {
	docker, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatalf("could not create docker client, %v", err)
	}

	containers, err := docker.ContainerList(context.Background(), types.ContainerListOptions{})
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
	done := make(chan string, s.maxContainers)

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
	for _, stage := range first {
		fmt.Printf("starting stage %s\n", stage)
		go runStage(stage, done)
	}

	for {
		select {
		case stage := <-done:
			fmt.Printf("stage %s is done\n", stage)
		default:
			fmt.Print("no stages done, waiting...\n")
			time.Sleep(1 * time.Second)
		}
	}

	return nil
}
