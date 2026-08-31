package metrics

import "math"

func Average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64

	for _, value := range values {
		sum += value
	}

	return sum / float64(len(values))
}

func Peak(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	peak := math.Inf(-1)

	for _, value := range values {
		if value > peak {
			peak = value
		}
	}

	return peak
}

type Statistics struct {
	Average float64
	Peak    float64
}

func Calculate(values []float64) Statistics {
	if len(values) == 0 {
		return Statistics{}
	}

	var sum float64
	peak := values[0]

	for _, value := range values {
		sum += value

		if value > peak {
			peak = value
		}
	}

	return Statistics{
		Average: sum / float64(len(values)),
		Peak:    peak,
	}
}
