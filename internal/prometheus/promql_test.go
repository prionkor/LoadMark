package prometheus

import (
	"strings"
	"testing"

	"github.com/prionkor/benchmark-runner/internal/model"
)

func TestBuildCPUQuery(t *testing.T) {
	monitoring := model.MonitoringResource{
		Namespace:      "leadhog",
		PodNamePattern: "webhooks-.*",
	}

	query := buildCPUQuery(monitoring)

	expected := []string{
		`container_cpu_usage_seconds_total`,
		`namespace="leadhog"`,
		`pod=~"webhooks-.*"`,
		`container!=""`,
		`container!="POD"`,
		`rate(`,
		`sum by (pod)`,
	}

	for _, value := range expected {
		if !strings.Contains(query, value) {
			t.Errorf("query does not contain %q:\n%s", value, query)
		}
	}
}
