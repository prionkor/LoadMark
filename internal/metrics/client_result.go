package metrics

type ClientResult struct {
	Requests     int
	Failed       int
	Duration     DurationStats
	DataSent     float64
	DataReceived float64
}

func (r *result) ClientResult() ClientResult {
	var result ClientResult
	var durations []float64

	for _, sample := range r.Samples {
		switch sample.Metric.Name {
		case "http_reqs":
			result.Requests += int(sample.Value)

		case "http_req_failed":
			if sample.Value > 0 {
				result.Failed++
			}

		case "http_req_duration":
			durations = append(durations, sample.Value)

		case "data_sent":
			result.DataSent += sample.Value

		case "data_received":
			result.DataReceived += sample.Value
		}
	}

	result.Duration = calculateDurationStats(durations)

	return result
}
