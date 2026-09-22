package metrics

import (
	"encoding/json"
	"fmt"
)

func decodeContainer(raw json.RawMessage) ([]sample, error) {
	// Shape B: array of samples
	var samples []sample

	if err := json.Unmarshal(raw, &samples); err == nil {
		return samples, nil
	}

	// Shapes A and C: object containing nested Samples
	var container struct {
		Samples []sample `json:"Samples"`
	}

	if err := json.Unmarshal(raw, &container); err == nil {
		if container.Samples != nil {
			return container.Samples, nil
		}
	}

	// Shape D: a single unwrapped sample
	var singleSample sample

	if err := json.Unmarshal(raw, &singleSample); err == nil {
		if singleSample.Metric.Name != "" {
			return []sample{singleSample}, nil
		}
	}

	return nil, fmt.Errorf("unknown sample container")
}
