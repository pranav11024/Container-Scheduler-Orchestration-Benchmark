# DockerSim: Container Cluster Simulator with Adaptive Scheduling

DockerSim is a modular and extensible container cluster simulation framework written in Go, designed to benchmark and evaluate container scheduling strategies over a heterogeneous node infrastructure. It integrates real Docker containers and collects fine-grained metrics to analyze scheduler performance under diverse workloads.

The two folders are two versions of the porject one with real (containerssimulationdocker/cc_go  | [simulation](https://github.com/pranav11024/Container-Scheduler-Orchestration-Benchmark/tree/main/simulation/cc_go) |), and with simulated containers(simulation/cc_go  | [simulation](https://github.com/pranav11024/Container-Scheduler-Orchestration-Benchmark/tree/main/simulationdocker/cc_go) |)

## Features

- ⚙️ **Pluggable Schedulers**: Supports Adaptive, BinPack, and Spread schedulers with dynamic phase-aware logic.
- 🐳 **Docker Integration**: Runs real containers with configurable resource limits (CPU, Memory, I/O, Network).
- 🧠 **Adaptive Strategy**: Learns and adapts to workload patterns, node health, and load variance over time.
- 📊 **Metrics Collection**: Tracks latency, success rates, and resource utilization for each scheduling event.
- 🧪 **Workload Generator**: Generates stochastic workloads from customizable JSON templates.
- 🔁 **Automated Benchmarking**: Runs full simulations with cleanup and teardown logic.
- 📦 **Dockerized Deployment**: Uses Docker Compose to orchestrate multiple simulation runs.

---

## Architecture

```bash
+------------------+
| Workload Gen     | ---> Generates Containers
+------------------+
         |
         v
+------------------+        +-------------------+
| Scheduler        | -----> | Node Manager      |
| (Adaptive, etc.) |        | (Docker Backend)  |
+------------------+        +-------------------+
         |
         v
+------------------+
| Metrics Collector| ---> Stores CSV results
+------------------+
```
Prerequisites
Docker (exposed via tcp://host.docker.internal:2375)
Go 1.23+
Docker Compose

Build the Simulator
```bash
docker-compose build
```
Run a Benchmark
```bash
docker-compose up scheduler
```
This runs the Adaptive scheduler on a mixed workload for 300 seconds and outputs results to:
results/adaptive_results.csv

You can also run other schedulers:
```bash
docker-compose up scheduler-binpack
docker-compose up scheduler-spread
```
Workload Configuration
Edit or create your own JSON workload files inside the workloads/ directory. Example:
```json
{
  "templates": [
    {
      "name": "web-service",
      "image": "nginx",
      "cpu_min": 0.5,
      "cpu_max": 1.5,
      "memory_min": 512,
      "memory_max": 1024,
      "network_min": 100,
      "network_max": 300,
      "io_min": 1000,
      "io_max": 5000,
      "type": "frontend",
      "priority": 1,
      "weight": 3
    }
  ]
}
```
