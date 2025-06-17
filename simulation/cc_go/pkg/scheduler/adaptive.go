// pkg/scheduler/adaptive.go - Novel adaptive scheduler implementation
package scheduler

import (
	"cc_go/pkg/container"
	"cc_go/pkg/node"
	"math"
	"sort"
	"time"
)

type AdaptiveScheduler struct {
	// Historical data for performance tracking
	containerHistory    map[string][]float64 // container type to resource usage patterns
	nodeHistory         map[string][]float64 // node ID to performance metrics
	schedulingStartTime time.Time
	schedulerPhase      int // 0: startup, 1: normal, 2: high-load
	
	// Resource score weights (dynamically adjusted)
	cpuWeight    float64
	memoryWeight float64
	networkWeight float64
	ioWeight     float64
}

func NewAdaptiveScheduler() *AdaptiveScheduler {
	return &AdaptiveScheduler{
		containerHistory:    make(map[string][]float64),
		nodeHistory:         make(map[string][]float64),
		schedulingStartTime: time.Now(),
		schedulerPhase:      0,
		cpuWeight:           0.25,
		memoryWeight:        0.25,
		networkWeight:       0.25,
		ioWeight:            0.25,
	}
}

func (s *AdaptiveScheduler) Name() string {
	return "Adaptive"
}

func (s *AdaptiveScheduler) Schedule(container *container.Container, nodes []*node.Node) (*node.Node, error) {
	candidateNodes := make([]*node.Node, 0)
	
	// Update scheduler phase based on runtime
	s.updateSchedulerPhase()
	
	// Filter nodes that can accommodate the container
	for _, n := range nodes {
		if n.CanFit(container) {
			candidateNodes = append(candidateNodes, n)
		}
	}
	
	if len(candidateNodes) == 0 {
		return nil, ErrNoSuitableNode
	}
	
	// Calculate fitness scores for each candidate node
	nodeScores := make(map[*node.Node]float64)
	for _, n := range candidateNodes {
		nodeScores[n] = s.calculateFitnessScore(container, n)
	}
	
	// Sort by fitness score (higher is better)
	sort.Slice(candidateNodes, func(i, j int) bool {
		return nodeScores[candidateNodes[i]] > nodeScores[candidateNodes[j]]
	})
	
	// Take container type into account for resource prediction
	containerType := container.Type()
	if _, exists := s.containerHistory[containerType]; exists {
		// Update weights based on historical performance of this container type
		s.adjustWeightsForContainer(containerType)
	}
	
	// Record placement decision for future reference
	bestNode := candidateNodes[0]
	s.recordPlacement(container, bestNode)
	
	return bestNode, nil
}

func (s *AdaptiveScheduler) calculateFitnessScore(container *container.Container, n *node.Node) float64 {
	// Base score is weighted sum of normalized resource availability
	cpuScore := (n.AvailableCPU() - container.CPURequest()) / n.TotalCPU()
	memScore := (n.AvailableMemory() - container.MemoryRequest()) / n.TotalMemory()
	netScore := (n.AvailableNetwork() - container.NetworkRequest()) / n.TotalNetwork()
	ioScore := (n.AvailableIO() - container.IORequest()) / n.TotalIO()
	
	// Apply current weights (these are dynamically adjusted)
	baseScore := (cpuScore * s.cpuWeight) + 
		(memScore * s.memoryWeight) + 
		(netScore * s.networkWeight) + 
		(ioScore * s.ioWeight)
	
	// Consider interference score based on container affinity/anti-affinity
	interferenceScore := s.calculateInterferenceScore(container, n)
	
	// Consider node health and historical performance
	nodeHealthScore := s.calculateNodeHealthScore(n)
	
	// Combine all factors
	finalScore := baseScore * 0.6 + interferenceScore * 0.2 + nodeHealthScore * 0.2
	return finalScore
}

func (s *AdaptiveScheduler) calculateInterferenceScore(container *container.Container, n *node.Node) float64 {
	// Higher score means less interference
	score := 1.0
	
	// Check for anti-affinity with containers already on this node
	existingContainers := n.Containers()
	
	for _, existing := range existingContainers {
		// Containers of same type might interfere
		if existing.Type() == container.Type() {
			score -= 0.1
		}
		
		// Adjust for specific resource competition
		if existing.CPUIntensive() && container.CPUIntensive() {
			score -= 0.15
		}
		
		if existing.MemoryIntensive() && container.MemoryIntensive() {
			score -= 0.15
		}
		
		if existing.IOIntensive() && container.IOIntensive() {
			score -= 0.15
		}
		
		if existing.NetworkIntensive() && container.NetworkIntensive() {
			score -= 0.15
		}
	}
	
	// Ensure score doesn't go negative
	return math.Max(0.1, score)
}

func (s *AdaptiveScheduler) calculateNodeHealthScore(n *node.Node) float64 {
	// Higher score means healthier node
	baseScore := 1.0
	
	// Consider historical failures or performance issues
	if history, exists := s.nodeHistory[n.ID()]; exists {
		// Last entry is the most recent health score
		if len(history) > 0 {
			baseScore = history[len(history)-1]
		}
	}
	
	// Consider load variance (unstable nodes get lower scores)
	loadVariance := n.LoadVariance()
	variancePenalty := loadVariance * 0.2
	
	// Consider node uptime and reliability
	uptimeScore := math.Min(1.0, n.UptimeHours()/24.0) * 0.1
	
	return baseScore - variancePenalty + uptimeScore
}

func (s *AdaptiveScheduler) updateSchedulerPhase() {
	elapsedTime := time.Since(s.schedulingStartTime).Minutes()
	
	if elapsedTime < 1 {
		// Startup phase - prefer spreading out containers
		s.schedulerPhase = 0
	} else if elapsedTime > 10 {
		// High-load phase - focus on efficient packing
		s.schedulerPhase = 2
	} else {
		// Normal operation - balanced approach
		s.schedulerPhase = 1
	}
	
	// Adjust weights based on phase
	switch s.schedulerPhase {
	case 0: // Startup
		s.cpuWeight = 0.2
		s.memoryWeight = 0.2
		s.networkWeight = 0.3
		s.ioWeight = 0.3
	case 1: // Normal
		s.cpuWeight = 0.25
		s.memoryWeight = 0.25
		s.networkWeight = 0.25
		s.ioWeight = 0.25
	case 2: // High-load
		s.cpuWeight = 0.3
		s.memoryWeight = 0.3
		s.networkWeight = 0.2
		s.ioWeight = 0.2
	}
}

func (s *AdaptiveScheduler) adjustWeightsForContainer(containerType string) {
	history := s.containerHistory[containerType]
	if len(history) < 4 {
		return
	}
	
	// history order: [cpu, memory, network, io]
	cpuUsage := history[0]
	memUsage := history[1]
	netUsage := history[2]
	ioUsage := history[3]
	
	// Normalize to make sure weights sum to 1.0
	total := cpuUsage + memUsage + netUsage + ioUsage
	if total > 0 {
		s.cpuWeight = 0.1 + (cpuUsage / total * 0.6)
		s.memoryWeight = 0.1 + (memUsage / total * 0.6)
		s.networkWeight = 0.1 + (netUsage / total * 0.6)
		s.ioWeight = 0.1 + (ioUsage / total * 0.6)
	}
}

func (s *AdaptiveScheduler) recordPlacement(container *container.Container, n *node.Node) {
	// Record container resource pattern
	containerType := container.Type()
	s.containerHistory[containerType] = []float64{
		container.CPURequest(),
		container.MemoryRequest(),
		container.NetworkRequest(),
		container.IORequest(),
	}
	
	// Update node history
	s.nodeHistory[n.ID()] = []float64{
		n.HealthScore(),
	}
}