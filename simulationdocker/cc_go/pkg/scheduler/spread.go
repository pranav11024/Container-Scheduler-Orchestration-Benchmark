// pkg/scheduler/spread.go - Resource spreading scheduler implementation
package scheduler

import (
	"sort"
	"cc_go/pkg/container"
	"cc_go/pkg/node"
)

type SpreadScheduler struct{}

func NewSpreadScheduler() *SpreadScheduler {
	return &SpreadScheduler{}
}

func (s *SpreadScheduler) Name() string {
	return "Spread"
}

func (s *SpreadScheduler) Schedule(container *container.Container, nodes []*node.Node) (*node.Node, error) {
	candidateNodes := make([]*node.Node, 0)
	
	// Filter nodes that can accommodate the container
	for _, n := range nodes {
		if n.CanFit(container) {
			candidateNodes = append(candidateNodes, n)
		}
	}
	
	if len(candidateNodes) == 0 {
		return nil, ErrNoSuitableNode
	}
	
	// Sort nodes by current utilization (ascending)
	sort.Slice(candidateNodes, func(i, j int) bool {
		return candidateNodes[i].Utilization() < candidateNodes[j].Utilization()
	})
	
	// Place on the node with lowest utilization
	return candidateNodes[0], nil
}