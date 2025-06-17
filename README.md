# OGSim: Container Cluster Simulator with Adaptive Scheduling

OGSim is a modular and extensible container cluster simulation framework written in Go, designed to benchmark and evaluate container scheduling strategies over a heterogeneous node infrastructure. It integrates real Docker containers and collects fine-grained metrics to analyze scheduler performance under diverse workloads.

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
