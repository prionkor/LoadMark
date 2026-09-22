package metrics

type DataStats struct {
	Total float64
	Rate  float64
}

type GaugeStats struct {
	Current float64
	Min     float64
	Max     float64
}

type ClientResult struct {
	Requests     int
	Failed       int
	Duration     DurationStats
	DataSent     DataStats
	DataReceived DataStats
	Iterations   struct {
		Total    int
		Duration DurationStats
		Dropped  int
	}
	VUs    GaugeStats
	VUsMax GaugeStats
}

func (r *result) ClientResult() ClientResult {
	var result ClientResult
	var durations []float64
	var iterationDurations []float64

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
			result.DataSent.Total += sample.Value

		case "data_received":
			result.DataReceived.Total += sample.Value

		case "iterations":
			result.Iterations.Total += int(sample.Value)

		case "iteration_duration":
			iterationDurations = append(iterationDurations, sample.Value)

		case "dropped_iterations":
			result.Iterations.Dropped += int(sample.Value)
		}
	}

	result.Duration = calculateDurationStats(durations)
	result.Iterations.Duration = calculateDurationStats(iterationDurations)

	elapsed := r.Duration().Seconds()
	if elapsed > 0 {
		result.DataSent.Rate = result.DataSent.Total / elapsed
		result.DataReceived.Rate = result.DataReceived.Total / elapsed
	}

	return result
}
