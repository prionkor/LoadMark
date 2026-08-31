package prometheus

import (
	"fmt"
	"time"

	"github.com/prionkor/benchmark-runner/internal/model"
)

func (c *Client) CollectCPU(
	monitoring model.MonitoringResource,
	start time.Time,
	end time.Time,
	step time.Duration,
) ([]float64, error) {
	query := buildCPUQuery(monitoring)

	response, err := c.QueryRange(
		query,
		start,
		end,
		step,
	)
	if err != nil {
		return nil, err
	}

	return extractValues(response.Data.Result), nil
}

func (c *Client) CollectMemory(
	monitoring model.MonitoringResource,
	start time.Time,
	end time.Time,
	step time.Duration,
) ([]float64, error) {
	query := buildMemoryQuery(monitoring)

	response, err := c.QueryRange(query, start, end, step)
	if err != nil {
		return nil, err
	}

	return extractValues(response.Data.Result), nil
}

func (c *Client) CollectNetworkReceive(
	monitoring model.MonitoringResource,
	start time.Time,
	end time.Time,
	step time.Duration,
) ([]float64, error) {
	query := buildNetworkReceiveQuery(monitoring)

	response, err := c.QueryRange(query, start, end, step)
	if err != nil {
		return nil, err
	}

	return extractValues(response.Data.Result), nil
}

func (c *Client) CollectNetworkTransmit(
	monitoring model.MonitoringResource,
	start time.Time,
	end time.Time,
	step time.Duration,
) ([]float64, error) {
	query := buildNetworkTransmitQuery(monitoring)

	response, err := c.QueryRange(query, start, end, step)
	if err != nil {
		return nil, err
	}

	return extractValues(response.Data.Result), nil
}

func buildCPUQuery(monitoring model.MonitoringResource) string {
	return fmt.Sprintf(`
		sum by (pod) (
			rate(
				container_cpu_usage_seconds_total{
					namespace="%s",
					%s,
					container!="",
					container!="POD"
				}[1m]
			)
		)
	`, monitoring.Namespace, buildPodMatcher(monitoring.PodNamePattern))
}

func buildMemoryQuery(monitoring model.MonitoringResource) string {
	return fmt.Sprintf(`
		sum by (pod) (
			container_memory_working_set_bytes{
				namespace="%s",
				%s,
				container!="",
				container!="POD"
			}
		)
	`, monitoring.Namespace, buildPodMatcher(monitoring.PodNamePattern))
}

func buildNetworkReceiveQuery(monitoring model.MonitoringResource) string {
	return fmt.Sprintf(`
		sum by (pod) (
			rate(
				container_network_receive_bytes_total{
					namespace="%s",
					%s
				}[1m]
			)
		)
	`, monitoring.Namespace, buildPodMatcher(monitoring.PodNamePattern))
}

func buildNetworkTransmitQuery(monitoring model.MonitoringResource) string {
	return fmt.Sprintf(`
		sum by (pod) (
			rate(
				container_network_transmit_bytes_total{
					namespace="%s",
					%s
				}[1m]
			)
		)
	`, monitoring.Namespace, buildPodMatcher(monitoring.PodNamePattern))
}

func buildPodMatcher(pattern string) string {
	return fmt.Sprintf(`pod=~"%s"`, pattern)
}
