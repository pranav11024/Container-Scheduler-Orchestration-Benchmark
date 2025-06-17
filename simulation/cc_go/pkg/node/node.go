// pkg/node/node.go - Node representation
package node

import (
	"cc_go/pkg/container"
	"fmt"
	"math"
	"time"
)

type Node struct {
	id              string
	name            string
	totalCPU        float64
	totalMemory     float64
	totalNetwork    float64
	totalIO         float64
	usedCPU         float64
	usedMemory      float64
	usedNetwork     float64
	usedIO          float64
	containers      []*container.Container
	creationTime    time.Time
	loadHistory     []float64
	healthScore     float64
}

func NewNode(name string, cpu, memory, network, io float64) *Node {
	return &Node{
		id:           fmt.Sprintf("node-%d", time.Now().UnixNano()),
		name:         name,
		totalCPU:     cpu,
		totalMemory:  memory,
		totalNetwork: network,
		totalIO:      io,
		usedCPU:      0,
		usedMemory:   0,
		usedNetwork:  0,
		usedIO:       0,
		containers:   make([]*container.Container, 0),
		creationTime: time.Now(),
		loadHistory:  make([]float64, 0),
		healthScore:  1.0,
	}
}

func (n *Node) ID() string {
	return n.id
}

func (n *Node) Name() string {
	return n.name
}

func (n *Node) TotalCPU() float64 {
	return n.totalCPU
}

func (n *Node) TotalMemory() float64 {
	return n.totalMemory
}

func (n *Node) TotalNetwork() float64 {
	return n.totalNetwork
}

func (n *Node) TotalIO() float64 {
	return n.totalIO
}

func (n *Node) AvailableCPU() float64 {
	return n.totalCPU - n.usedCPU
}

func (n *Node) AvailableMemory() float64 {
	return n.totalMemory - n.usedMemory
}

func (n *Node) AvailableNetwork() float64 {
	return n.totalNetwork - n.usedNetwork
}

func (n *Node) AvailableIO() float64 {
	return n.totalIO - n.usedIO
}

func (n *Node) Utilization() float64 {
	cpuUtil := n.usedCPU / n.totalCPU
	memUtil := n.usedMemory / n.totalMemory
	netUtil := n.usedNetwork / n.totalNetwork
	ioUtil := n.usedIO / n.totalIO
	
	return (cpuUtil + memUtil + netUtil + ioUtil) / 4.0
}

func (n *Node) CanFit(c *container.Container) bool {
	return c.CPURequest() <= n.AvailableCPU() &&
		c.MemoryRequest() <= n.AvailableMemory() &&
		c.NetworkRequest() <= n.AvailableNetwork() &&
		c.IORequest() <= n.AvailableIO()
}

func (n *Node) AddContainer(c *container.Container) bool {
	if !n.CanFit(c) {
		return false
	}
	
	n.usedCPU += c.CPURequest()
	n.usedMemory += c.MemoryRequest()
	n.usedNetwork += c.NetworkRequest()
	n.usedIO += c.IORequest()
	n.containers = append(n.containers, c)
	
	// Update load history
	n.loadHistory = append(n.loadHistory, n.Utilization())
	if len(n.loadHistory) > 10 {
		// Keep only the last 10 entries
		n.loadHistory = n.loadHistory[1:]
	}
	
	return true
}

func (n *Node) RemoveContainer(containerID string) bool {
	for i, c := range n.containers {
		if c.ID() == containerID {
			n.usedCPU -= c.CPURequest()
			n.usedMemory -= c.MemoryRequest()
			n.usedNetwork -= c.NetworkRequest()
			n.usedIO -= c.IORequest()
			
			// Remove the container from the slice
			n.containers = append(n.containers[:i], n.containers[i+1:]...)
			
			// Update load history
			n.loadHistory = append(n.loadHistory, n.Utilization())
			if len(n.loadHistory) > 10 {
				// Keep only the last 10 entries
				n.loadHistory = n.loadHistory[1:]
			}
			
			return true
		}
	}
	
	return false
}

func (n *Node) Containers() []*container.Container {
	return n.containers
}

func (n *Node) ContainerCount() int {
	return len(n.containers)
}

func (n *Node) UptimeHours() float64 {
	return time.Since(n.creationTime).Hours()
}

func (n *Node) LoadVariance() float64 {
	if len(n.loadHistory) < 2 {
		return 0.0
	}
	
	// Calculate variance of the load history
	mean := 0.0
	for _, load := range n.loadHistory {
		mean += load
	}
	mean /= float64(len(n.loadHistory))
	
	variance := 0.0
	for _, load := range n.loadHistory {
		diff := load - mean
		variance += diff * diff
	}
	variance /= float64(len(n.loadHistory))
	
	return math.Sqrt(variance)
}

func (n *Node) HealthScore() float64 {
	return n.healthScore
}

func (n *Node) UpdateHealthScore(score float64) {
	n.healthScore = math.Max(0.0, math.Min(1.0, score))
}
