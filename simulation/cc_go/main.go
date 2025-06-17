//main.go - Entry point for the scheduler
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"cc_go/pkg/benchmark"
	"cc_go/pkg/metrics"
	"cc_go/pkg/scheduler"
	"cc_go/pkg/workLoad"
)

func main() {
	schedulerType := flag.String("scheduler", "adaptive", "Scheduler type: 'binpack', 'spread', or 'adaptive'")
	workloadFile := flag.String("workload", "workloads/mixed_workload.json", "Path to workload definition file")
	outputFile := flag.String("output", "results.csv", "Path to output results file")
	duration := flag.Int("duration", 300, "Duration of simulation in seconds")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	if *verbose {
		log.SetOutput(os.Stdout)
	} else {
		logFile, err := os.Create("scheduler.log")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create log file: %v\n", err)
			os.Exit(1)
		}
		defer logFile.Close()
		log.SetOutput(logFile)
	}

	log.Printf("Starting container scheduler with %s algorithm", *schedulerType)
	log.Printf("Using workload file: %s", *workloadFile)
	log.Printf("Running on %d CPU cores", runtime.NumCPU())

	// Initialize the workload generator
	workloadGen, err := workLoad.NewWorkloadFromFile(*workloadFile)
	if err != nil {
		log.Fatalf("Failed to initialize workload: %v", err)
	}

	// Initialize the chosen scheduler
	var sched scheduler.Scheduler
	switch *schedulerType {
	case "binpack":
		sched = scheduler.NewBinPackScheduler()
	case "spread":
		sched = scheduler.NewSpreadScheduler()
	case "adaptive":
		sched = scheduler.NewAdaptiveScheduler()
	default:
		log.Fatalf("Unknown scheduler type: %s", *schedulerType)
	}

	// Create metrics collector
	collector := metrics.NewCollector()

	// Run benchmark
	benchmark := benchmark.NewBenchmark(sched, workloadGen, collector)
	fmt.Printf("Starting benchmark for %d seconds...\n", *duration)
	benchmark.Run(time.Duration(*duration) * time.Second)

	// Output results
	results := collector.GetResults()
	fmt.Printf("Benchmark complete. Saving results to %s\n", *outputFile)
	err = results.SaveToFile(*outputFile)
	if err != nil {
		log.Fatalf("Failed to save results: %v", err)
	}

	fmt.Println("Summary of results:")
	fmt.Printf("  Scheduler type: %s\n", *schedulerType)
	fmt.Printf("  Containers scheduled: %d\n", results.ContainersScheduled)
	fmt.Printf("  Average scheduling latency: %.2fms\n", results.AverageLatency)
	fmt.Printf("  Resource utilization: %.2f%%\n", results.ResourceUtilization*100)
	fmt.Printf("  Scheduling failures: %d\n", results.SchedulingFailures)
}
