package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/prionkor/benchmark-runner/internal/config"
	"github.com/prionkor/benchmark-runner/internal/k6"
	"github.com/prionkor/benchmark-runner/internal/metrics"
	"github.com/prionkor/benchmark-runner/internal/model"
	"github.com/prionkor/benchmark-runner/internal/prometheus"
)

func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatalf("Failed to load .env: %v", err)
	}
	config, err := config.Load("config.yml")

	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	collectionDelay, err := time.ParseDuration(
		config.Prometheus.CollectionDelay,
	)
	if err != nil {
		log.Fatalf("Invalid Prometheus collection delay: %v", err)
	}

	// start k6
	clientResult, err := k6.Run(config)
	if err != nil {
		log.Fatalf("Failed to run benchmark: %v", err)
	}

	fmt.Printf(
		"Waiting %s for Prometheus to collect metrics...\n",
		collectionDelay,
	)

	time.Sleep(collectionDelay)

	promClient := &prometheus.Client{
		BaseURL:  config.Prometheus.URL,
		Username: config.Prometheus.Username,
		Password: config.Prometheus.Password,
		HTTP:     &http.Client{},
	}

	benchmarkResult := model.BenchmarkResult{}

	for _, stage := range clientResult.Stages {
		stage := collectStageMetrics(
			stage,
			config.Monitoring,
			promClient,
			5*time.Second,
		)

		benchmarkResult.Stages = append(
			benchmarkResult.Stages,
			stage,
		)
	}

	printBenchmarkResult(benchmarkResult)

}

func collectStageMetrics(
	stage model.StageResult,
	resources []model.MonitoringResource,
	promClient *prometheus.Client,
	step time.Duration,
) model.StageResult {
	stage.Resources = make([]model.ResourceResult, 0, len(resources))

	for _, resource := range resources {
		result, err := collectResourceMetrics(
			resource,
			promClient,
			stage.Start,
			stage.End,
			step,
		)
		if err != nil {
			stage.Resources = append(stage.Resources, model.ResourceResult{
				Name:  resource.Name,
				Error: err.Error(),
			})
			continue
		}

		stage.Resources = append(stage.Resources, result)
	}

	return stage
}

func collectResourceMetrics(
	resource model.MonitoringResource,
	promClient *prometheus.Client,
	start time.Time,
	end time.Time,
	step time.Duration,
) (model.ResourceResult, error) {
	result := model.ResourceResult{
		Name:    resource.Name,
		Metrics: make(map[string]metrics.Statistics),
	}

	if len(resource.Metrics.CPU) > 0 {
		values, err := promClient.CollectCPU(resource, start, end, step)
		if err != nil {
			return result, fmt.Errorf("failed to collect CPU metrics: %w", err)
		}

		result.Metrics["cpu"] = metrics.Calculate(values)
	}

	if len(resource.Metrics.Memory) > 0 {
		values, err := promClient.CollectMemory(resource, start, end, step)
		if err != nil {
			return result, fmt.Errorf("failed to collect memory metrics: %w", err)
		}

		result.Metrics["memory"] = metrics.Calculate(values)
	}

	if len(resource.Metrics.Network) > 0 {
		values, err := promClient.CollectNetworkReceive(resource, start, end, step)
		if err != nil {
			return result, fmt.Errorf("failed to collect network receive metrics: %w", err)
		}

		result.Metrics["network_receive"] = metrics.Calculate(values)

		values, err = promClient.CollectNetworkTransmit(resource, start, end, step)
		if err != nil {
			return result, fmt.Errorf("failed to collect network transmit metrics: %w", err)
		}

		result.Metrics["network_transmit"] = metrics.Calculate(values)
	}

	return result, nil
}

func printBenchmarkResult(result model.BenchmarkResult) {
	fmt.Println()
	fmt.Println("========== BENCHMARK RESULTS ==========")
	fmt.Println()

	metricOrder := []string{
		"cpu",
		"memory",
		"network_receive",
		"network_transmit",
	}

	for i, stage := range result.Stages {
		fmt.Printf("Stage %d — %d RPS\n", i+1, stage.RPS)

		fmt.Printf(
			"  %s → %s\n",
			stage.Start.Format(time.RFC3339),
			stage.End.Format(time.RFC3339),
		)

		fmt.Println()

		for _, resource := range stage.Resources {
			fmt.Println("  " + resource.Name)

			if resource.Error != "" {
				fmt.Printf("    ERROR: %s\n", resource.Error)
				fmt.Println()
				continue
			}

			for _, name := range metricOrder {
				stats, exists := resource.Metrics[name]
				if !exists {
					continue
				}

				switch name {
				case "cpu":
					fmt.Println("    CPU")
					fmt.Printf("      Average: %.6f cores\n", stats.Average)
					fmt.Printf("      Peak:    %.6f cores\n", stats.Peak)

				case "memory":
					fmt.Println("    Memory")
					fmt.Printf("      Average: %s\n", formatBytes(stats.Average))
					fmt.Printf("      Peak:    %s\n", formatBytes(stats.Peak))

				case "network_receive":
					fmt.Println("    Network Receive")
					fmt.Printf("      Average: %.2f B/s\n", stats.Average)
					fmt.Printf("      Peak:    %.2f B/s\n", stats.Peak)

				case "network_transmit":
					fmt.Println("    Network Transmit")
					fmt.Printf("      Average: %.2f B/s\n", stats.Average)
					fmt.Printf("      Peak:    %.2f B/s\n", stats.Peak)
				}
			}

			fmt.Println()
		}

		fmt.Println("----------------------------------------")
		fmt.Println()
	}
}

func formatBytes(bytes float64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", bytes/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", bytes/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", bytes/KB)
	default:
		return fmt.Sprintf("%.0f B", bytes)
	}
}
