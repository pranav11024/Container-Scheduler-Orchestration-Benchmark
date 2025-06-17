// pkg/scheduler/binpack.go - Bin-packing scheduler implementation
package scheduler

import (
	"sort"
	"cc_go/pkg/container"
	"cc_go/pkg/node"
)

type BinPackScheduler struct{}

func NewBinPackScheduler() *BinPackScheduler {
	return &BinPackScheduler{}
}

func (s *BinPackScheduler) Name() string {
	return "BinPack"
}

func (s *BinPackScheduler) Schedule(container *container.Container, nodes []*node.Node) (*node.Node, error) {
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
	
	// Sort nodes by current utilization (descending)
	sort.Slice(candidateNodes, func(i, j int) bool {
		return candidateNodes[i].Utilization() > candidateNodes[j].Utilization()
	})
	
	// Place on the node with highest utilization that can still fit the container
	return candidateNodes[0], nil
}
