package internal

import "fmt"

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

func (s *Scheduler) Schedule(p Pipeline) error {
	g := NewGraphFromStages(p.Stages)

	for i, layer := range g.TopoSortedLayers() {
		fmt.Printf("%d: %s\n", i, layer)
	}

	return nil
}
