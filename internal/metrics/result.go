package metrics

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
