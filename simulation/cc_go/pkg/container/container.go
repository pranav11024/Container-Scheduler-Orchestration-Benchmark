package container

import (
	"fmt"
	"time"
)

type Container struct {
	id              string
	name            string
	image           string
	cpuRequest      float64 // CPU cores requested
	memoryRequest   float64 // Memory in MB requested
	networkRequest  float64 // Network bandwidth in Mbps
	ioRequest       float64 // IO operations per second
	containerType   string  // Type of workload (e.g., "web", "database", "batch")
	creationTime    time.Time
	startupDuration time.Duration
	priority        int
}

func NewContainer(name, image string, cpuReq, memReq, netReq, ioReq float64, containerType string, priority int) *Container {
	return &Container{
		id:              fmt.Sprintf("container-%d", time.Now().UnixNano()),
		name:            name,
		image:           image,
		cpuRequest:      cpuReq,
		memoryRequest:   memReq,
		networkRequest:  netReq,
		ioRequest:       ioReq,
		containerType:   containerType,
		creationTime:    time.Now(),
		startupDuration: 0,
		priority:        priority,
	}
}

func (c *Container) ID() string {
	return c.id
}

func (c *Container) Name() string {
	return c.name
}

func (c *Container) Image() string {
	return c.image
}

func (c *Container) CPURequest() float64 {
	return c.cpuRequest
}

func (c *Container) MemoryRequest() float64 {
	return c.memoryRequest
}

func (c *Container) NetworkRequest() float64 {
	return c.networkRequest
}

func (c *Container) IORequest() float64 {
	return c.ioRequest
}

func (c *Container) Type() string {
	return c.containerType
}

func (c *Container) Priority() int {
	return c.priority
}

func (c *Container) SetStartupDuration(d time.Duration) {
	c.startupDuration = d
}

func (c *Container) StartupDuration() time.Duration {
	return c.startupDuration
}

func (c *Container) Age() time.Duration {
	return time.Since(c.creationTime)
}

func (c *Container) CPUIntensive() bool {
	return c.cpuRequest > 2.0
}

func (c *Container) MemoryIntensive() bool {
	return c.memoryRequest > 2048
}

func (c *Container) NetworkIntensive() bool {
	return c.networkRequest > 500
}

func (c *Container) IOIntensive() bool {
	return c.ioRequest > 5000
}
