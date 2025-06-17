// pkg/metrics/metrics.go - Metrics collection and reporting
package metrics

import (
	"cc_go/pkg/container"
	"cc_go/pkg/node"
	"encoding/csv"
	"os"
	"strconv"
	"time"
)

type SchedulingEvent struct {
	Timestamp           time.Time
	ContainerID         string
	ContainerType       string
	NodeID              string
	SchedulingLatency   time.Duration
	ScheduleSuccess     bool
	ResourceUtilization float64
}

type Results struct {
	ContainersScheduled   int
	SchedulingFailures    int
	AverageLatency        float64
	ResourceUtilization   float64
	Events                []SchedulingEvent
}

type Collector interface {
	RecordSchedulingEvent(container *container.Container, node *node.Node, latency time.Duration, success bool)
	GetResults() *Results
}

type MetricsCollector struct {
	events               []SchedulingEvent
	containersScheduled  int
	schedulingFailures   int
	totalLatency         time.Duration
	resourceUtilization  float64
	utilizationDatapoints int
}

func NewCollector() *MetricsCollector {
	return &MetricsCollector{
		events:              make([]SchedulingEvent, 0),
		containersScheduled: 0,
		schedulingFailures:  0,
		totalLatency:        0,
		resourceUtilization: 0,
		utilizationDatapoints: 0,
	}
}

func (c *MetricsCollector) RecordSchedulingEvent(container *container.Container, node *node.Node, latency time.Duration, success bool) {
	var nodeID string
	var utilization float64
	
	if node != nil {
		nodeID = node.ID()
		utilization = node.Utilization()
		
		// Update running average of resource utilization
		c.resourceUtilization = (c.resourceUtilization * float64(c.utilizationDatapoints) + utilization) / float64(c.utilizationDatapoints + 1)
		c.utilizationDatapoints++
	}
	
	event := SchedulingEvent{
		Timestamp:           time.Now(),
		ContainerID:         container.ID(),
		ContainerType:       container.Type(),
		NodeID:              nodeID,
		SchedulingLatency:   latency,
		ScheduleSuccess:     success,
		ResourceUtilization: utilization,
	}
	
	c.events = append(c.events, event)
	
	if success {
		c.containersScheduled++
		c.totalLatency += latency
	} else {
		c.schedulingFailures++
	}
}

func (c *MetricsCollector) GetResults() *Results {
	var avgLatency float64
	if c.containersScheduled > 0 {
		avgLatency = float64(c.totalLatency.Microseconds()) / float64(c.containersScheduled) / 1000.0 // Convert to ms
	}
	
	return &Results{
		ContainersScheduled:   c.containersScheduled,
		SchedulingFailures:    c.schedulingFailures,
		AverageLatency:        avgLatency,
		ResourceUtilization:   c.resourceUtilization,
		Events:                c.events,
	}
}

func (r *Results) SaveToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	// Write header
	header := []string{
		"Timestamp",
		"ContainerID",
		"ContainerType",
		"NodeID",
		"SchedulingLatency(ms)",
		"Success",
		"ResourceUtilization",
	}
	
	if err := writer.Write(header); err != nil {
		return err
	}
	
	// Write events
	for _, event := range r.Events {
		record := []string{
			event.Timestamp.Format(time.RFC3339),
			event.ContainerID,
			event.ContainerType,
			event.NodeID,
			strconv.FormatFloat(float64(event.SchedulingLatency.Microseconds())/1000.0, 'f', 3, 64),
			strconv.FormatBool(event.ScheduleSuccess),
			strconv.FormatFloat(event.ResourceUtilization, 'f', 3, 64),
		}
		
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	
	return nil
}
