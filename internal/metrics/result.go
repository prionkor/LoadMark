package metrics

import "time"

type result struct {
	Samples []sample
}

func (r result) Metrics() map[string][]sample {
	metrics := make(map[string][]sample)

	for _, sample := range r.Samples {
		name := sample.Metric.Name
		metrics[name] = append(metrics[name], sample)
	}

	return metrics
}

func (r result) Duration() time.Duration {
	if len(r.Samples) < 2 {
		return 0
	}

	start := r.Samples[0].Time
	end := r.Samples[0].Time

	for _, sample := range r.Samples[1:] {
		if sample.Time.Before(start) {
			start = sample.Time
		}

		if sample.Time.After(end) {
			end = sample.Time
		}
	}

	return end.Sub(start)
}
