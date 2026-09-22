package metrics

type DataStats struct {
	Total float64
	Rate  float64
}

type ClientResult struct {
	Requests     int
	Failed       int
	Duration     DurationStats
	DataSent     DataStats
	DataReceived DataStats
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
			result.DataSent.Total += sample.Value

		case "data_received":
			result.DataReceived.Total += sample.Value
		}
	}

	result.Duration = calculateDurationStats(durations)

	elapsed := r.Duration().Seconds()
	if elapsed > 0 {
		result.DataSent.Rate = result.DataSent.Total / elapsed
		result.DataReceived.Rate = result.DataReceived.Total / elapsed
	}

	return result
}
