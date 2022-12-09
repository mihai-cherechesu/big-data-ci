package internal

// Metadata for each stage. Must be an object
// with the keys "script", "depends_on", and "artifacts"
type StageMeta struct {
	Script    []string `json:"script"`
	DependsOn []string `json:"depends_on"`
	Artifacts []string `json:"artifacts"`
}

// A struct to represent the JSON schema
type Schema struct {
	Image string `json:"image" schema:"image"`

	// Allow for any Stages keys
	Stages map[string]StageMeta `json:"stages"`
}

func NewGraphFromStages(stages map[string]StageMeta) *Graph {
	g := New()

	for k, v := range stages {
		for _, dep := range v.DependsOn {
			g.DependOn(k, dep)
		}
	}

	return g
}
