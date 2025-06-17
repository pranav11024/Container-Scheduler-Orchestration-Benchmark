// pkg/scheduler/scheduler.go - Scheduler interface
package scheduler

import (
	"cc_go/pkg/container"
	"cc_go/pkg/node"
)

type Scheduler interface {
	// Schedule attempts to schedule a container on available nodes
	Schedule(container *container.Container, nodes []*node.Node) (*node.Node, error)
	
	// Name returns the name of the scheduler
	Name() string
}
