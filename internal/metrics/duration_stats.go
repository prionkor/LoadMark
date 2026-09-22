package metrics

import "sort"

type DurationStats struct {
	Min float64
	Max float64
	Avg float64
	P90 float64
	P95 float64
	P99 float64
}

func calculateDurationStats(values []float64) DurationStats {
	if len(values) == 0 {
		return DurationStats{}
	}

	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)

	var sum float64

	for _, value := range sorted {
		sum += value
	}

	return DurationStats{
		Min: sorted[0],
		Max: sorted[len(sorted)-1],
		Avg: sum / float64(len(sorted)),
		P90: percentile(sorted, 0.90),
		P95: percentile(sorted, 0.95),
		P99: percentile(sorted, 0.99),
	}
}

func percentile(sorted []float64, percentile float64) float64 {
	if len(sorted) == 0 {
		return 0
	}

	index := percentile * float64(len(sorted)-1)
	lower := int(index)
	upper := lower + 1

	if upper >= len(sorted) {
		return sorted[lower]
	}

	weight := index - float64(lower)

	return sorted[lower] +
		(sorted[upper]-sorted[lower])*weight
}
