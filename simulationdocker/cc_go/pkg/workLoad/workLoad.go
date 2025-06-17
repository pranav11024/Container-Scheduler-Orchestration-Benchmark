package workLoad

import (
	"cc_go/pkg/container"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"time"
)

type WorkloadGenerator interface {
	// NextContainer generates the next container to be scheduled
	NextContainer() *container.Container
	
	// HasNext returns true if there are more containers to generate
	HasNext() bool
}

type ContainerTemplate struct {
	Name           string  `json:"name"`
	Image          string  `json:"image"`
	CPUMin         float64 `json:"cpu_min"`
	CPUMax         float64 `json:"cpu_max"`
	MemoryMin      float64 `json:"memory_min"`
	MemoryMax      float64 `json:"memory_max"`
	NetworkMin     float64 `json:"network_min"`
	NetworkMax     float64 `json:"network_max"`
	IOMin          float64 `json:"io_min"`
	IOMax          float64 `json:"io_max"`
	Type           string  `json:"type"`
	Priority       int     `json:"priority"`
	Weight         int     `json:"weight"`
}

type WorkloadDefinition struct {
	Templates []ContainerTemplate `json:"templates"`
}

type FileWorkloadGenerator struct {
	definition WorkloadDefinition
	templates  []ContainerTemplate
	weights    []int
	totalWeight int
	count      int
	maxCount   int
}

func NewWorkloadFromFile(filename string) (*FileWorkloadGenerator, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	
	var definition WorkloadDefinition
	if err := json.Unmarshal(data, &definition); err != nil {
		return nil, err
	}
	
	templates := definition.Templates
	weights := make([]int, len(templates))
	totalWeight := 0
	
	for i, template := range templates {
		weights[i] = template.Weight
		totalWeight += template.Weight
	}
	
	rand.Seed(time.Now().UnixNano())
	
	return &FileWorkloadGenerator{
		definition:  definition,
		templates:   templates,
		weights:     weights,
		totalWeight: totalWeight,
		count:       0,
		maxCount:    10000, // Large number as default
	}, nil
}

func (g *FileWorkloadGenerator) SetMaxCount(count int) {
	g.maxCount = count
}

func (g *FileWorkloadGenerator) HasNext() bool {
	return g.count < g.maxCount
}

func (g *FileWorkloadGenerator) NextContainer() *container.Container {
	if !g.HasNext() {
		return nil
	}
	
	g.count++
	
	// Select a template based on weights
	r := rand.Intn(g.totalWeight)
	templateIndex := 0
	for i, weight := range g.weights {
		r -= weight
		if r < 0 {
			templateIndex = i
			break
		}
	}
	
	template := g.templates[templateIndex]
	
	// Generate random values within the template's ranges
	cpu := template.CPUMin + rand.Float64()*(template.CPUMax-template.CPUMin)
	memory := template.MemoryMin + rand.Float64()*(template.MemoryMax-template.MemoryMin)
	network := template.NetworkMin + rand.Float64()*(template.NetworkMax-template.NetworkMin)
	io := template.IOMin + rand.Float64()*(template.IOMax-template.IOMin)
	
	return container.NewContainer(
		template.Name,
		template.Image,
		cpu,
		memory,
		network,
		io,
		template.Type,
		template.Priority,
	)
}