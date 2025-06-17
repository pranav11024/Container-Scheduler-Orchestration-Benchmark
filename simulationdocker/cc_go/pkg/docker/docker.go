package docker

import (
	"context"
	"cc_go/pkg/container"
	"fmt"
	"io"
	"os"
	"time"
	"encoding/json"
	"github.com/docker/docker/api/types"
	dockercontainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type DockerManager struct {
	client  *client.Client
	ctx     context.Context
}


func NewDockerManager() (*DockerManager, error) {
	ctx := context.Background()
	
	cli, err := client.NewClientWithOpts(
		client.WithHost("tcp://host.docker.internal:2375"),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %v", err)
	}
	
	_, err = cli.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Docker daemon: %v", err)
	}
	
	return &DockerManager{
		client: cli,
		ctx:    ctx,
	}, nil
}


func (m *DockerManager) Close() {
	m.client.Close()
}

func (m *DockerManager) RunContainer(c *container.Container) (string, error) {
	// Pull the image
	reader, err := m.client.ImagePull(m.ctx, c.Image(), types.ImagePullOptions{})
	if err != nil {
		return "", err
	}
	io.Copy(os.Stdout, reader)
	
	// Create container
	resp, err := m.client.ContainerCreate(m.ctx, &dockercontainer.Config{
		Image: c.Image(),
		Env: []string{
			fmt.Sprintf("CPU_LIMIT=%f", c.CPURequest()),
			fmt.Sprintf("MEMORY_LIMIT=%f", c.MemoryRequest()),
		},
		Hostname: c.Name(),
	}, &dockercontainer.HostConfig{
		Resources: dockercontainer.Resources{
			CPUPeriod:  100000,
			CPUQuota:   int64(c.CPURequest() * 100000),
			Memory:     int64(c.MemoryRequest() * 1024 * 1024),
			MemorySwap: -1,
		},
	}, nil, nil, c.Name())
	
	if err != nil {
		return "", err
	}
	
	// Start container
	if err := m.client.ContainerStart(m.ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}
	
	return resp.ID, nil
}

func (m *DockerManager) StopContainer(containerID string) error {
	timeout := 10 * time.Second
	if err := m.client.ContainerStop(m.ctx, containerID, &timeout); err != nil {
		return err
	}
	
	return m.client.ContainerRemove(m.ctx, containerID, types.ContainerRemoveOptions{})
}

func (m *DockerManager) GetContainerStats(containerID string) (float64, float64, error) {
	stats, err := m.client.ContainerStats(m.ctx, containerID, false)
	if err != nil {
		return 0, 0, err
	}
	defer stats.Body.Close()
	
	// Parse stats
	var statsJSON types.StatsJSON
	if err := json.NewDecoder(stats.Body).Decode(&statsJSON); err != nil {
		return 0, 0, err
	}
	
	// Calculate CPU usage percentage
	cpuDelta := float64(statsJSON.CPUStats.CPUUsage.TotalUsage - statsJSON.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(statsJSON.CPUStats.SystemUsage - statsJSON.PreCPUStats.SystemUsage)
	cpuPercent := 0.0
	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(statsJSON.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	
	// Memory usage in MB
	memoryUsageMB := float64(statsJSON.MemoryStats.Usage) / 1024.0 / 1024.0
	
	return cpuPercent, memoryUsageMB, nil
}
