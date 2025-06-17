// pkg/benchmark/benchmark.go - Benchmark implementation
package benchmark

import (
	"cc_go/pkg/metrics"
	"cc_go/pkg/node"
	"cc_go/pkg/scheduler"
	"cc_go/pkg/workLoad"
	"fmt"
	"log"
	"sync"
	"time"
)

type Benchmark struct {
	scheduler       scheduler.Scheduler
	workloadGen     workLoad.WorkloadGenerator
	metricsCollector metrics.Collector
	nodes           []*node.Node
	stopChan        chan struct{}
	wg              sync.WaitGroup
}

func NewBenchmark(
	scheduler scheduler.Scheduler,
	workloadGen workLoad.WorkloadGenerator,
	collector metrics.Collector,
) *Benchmark {
	// Create a simulated cluster of nodes
	nodes := createNodes()
	
	return &Benchmark{
		scheduler:       scheduler,
		workloadGen:     workloadGen,
		metricsCollector: collector,
		nodes:           nodes,
		stopChan:        make(chan struct{}),
	}
}

func createNodes() []*node.Node {
	nodes := make([]*node.Node, 0)
	
	// Create a heterogeneous cluster
	// Small nodes
	for i := 0; i < 3; i++ {
		nodes = append(nodes, node.NewNode(
			fmt.Sprintf("small-node-%d", i),
			2.0,  // 2 CPU cores
			4096, // 4GB memory
			1000, // 1Gbps network
			5000, // 5K IOPS
		))
	}
	
	// Medium nodes
	for i := 0; i < 5; i++ {
		nodes = append(nodes, node.NewNode(
			fmt.Sprintf("medium-node-%d", i),
			4.0,   // 4 CPU cores
			8192,  // 8GB memory
			2000,  // 2Gbps network
			10000, // 10K IOPS
		))
	}
	
	// Large nodes
	for i := 0; i < 2; i++ {
		nodes = append(nodes, node.NewNode(
			fmt.Sprintf("large-node-%d", i),
			8.0,   // 8 CPU cores
			16384, // 16GB memory
			5000,  // 5Gbps network
			20000, // 20K IOPS
		))
	}
	
	return nodes
}

func (b *Benchmark) Run(duration time.Duration) {
	log.Printf("Starting benchmark with %s scheduler for %v", b.scheduler.Name(), duration)
	log.Printf("Simulating cluster with %d nodes", len(b.nodes))
	
	// Start the container scheduler
	b.wg.Add(1)
	go b.scheduleContainers()
	
	// Start the cleanup routine
	b.wg.Add(1)
	go b.cleanupContainers()
	
	// Wait for the specified duration
	time.Sleep(duration)
	
	// Signal to stop
	close(b.stopChan)
	
	// Wait for goroutines to complete
	b.wg.Wait()
	
	log.Println("Benchmark complete")
}

func (b *Benchmark) scheduleContainers() {
	defer b.wg.Done()
	
	// Rate limiting - don't flood with containers
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if !b.workloadGen.HasNext() {
				return
			}
			
			container := b.workloadGen.NextContainer()
			if container == nil {
				continue
			}
			
			startTime := time.Now()
			node, err := b.scheduler.Schedule(container, b.nodes)
			latency := time.Since(startTime)
			
			if err != nil {
				log.Printf("Failed to schedule container %s: %v", container.ID(), err)
				b.metricsCollector.RecordSchedulingEvent(container, nil, latency, false)
				continue
			}
			
			// Add container to the node
			if node.AddContainer(container) {
				log.Printf("Scheduled container %s on node %s (latency: %v)", 
					container.ID(), node.Name(), latency)
				b.metricsCollector.RecordSchedulingEvent(container, node, latency, true)
			} else {
				log.Printf("Node %s rejected container %s", node.Name(), container.ID())
				b.metricsCollector.RecordSchedulingEvent(container, node, latency, false)
			}
			
		case <-b.stopChan:
			return
		}
	}
}

func (b *Benchmark) cleanupContainers() {
	defer b.wg.Done()
	
	// Remove containers periodically to simulate completion
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			b.removeRandomContainers()
		case <-b.stopChan:
			return
		}
	}
}

func (b *Benchmark) removeRandomContainers() {
	for _, node := range b.nodes {
		containers := node.Containers()
		
		// Remove ~10% of containers from each node
		for i := 0; i < len(containers)/10+1; i++ {
			if len(containers) == 0 {
				break
			}
			
			// Remove a random container
			containerIdx := time.Now().Nanosecond() % len(containers)
			containerID := containers[containerIdx].ID()
			if node.RemoveContainer(containerID) {
				log.Printf("Removed container %s from node %s", containerID, node.Name())
			}
			
			// Update containers list
			containers = node.Containers()
		}
	}
}
