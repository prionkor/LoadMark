package prometheus

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type QueryResponse struct {
	Status string    `json:"status"`
	Data   QueryData `json:"data"`
}

type QueryData struct {
	ResultType string        `json:"resultType"`
	Result     []QueryResult `json:"result"`
}

type QueryResult struct {
	Metric map[string]string `json:"metric"`
	Value  Sample            `json:"value"`
	Values []Sample          `json:"values"`
}

type Sample struct {
	Timestamp float64
	Value     float64
}

func (s *Sample) UnmarshalJSON(data []byte) error {
	var raw []any

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if len(raw) != 2 {
		return fmt.Errorf("invalid prometheus sample")
	}

	timestamp, ok := raw[0].(float64)
	if !ok {
		return fmt.Errorf("invalid prometheus timestamp")
	}

	valueString, ok := raw[1].(string)
	if !ok {
		return fmt.Errorf("invalid prometheus value")
	}

	value, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return fmt.Errorf("invalid prometheus value %q: %w", valueString, err)
	}

	s.Timestamp = timestamp
	s.Value = value

	return nil
}

func extractValues(results []QueryResult) []float64 {
	var values []float64

	for _, result := range results {
		for _, sample := range result.Values {
			values = append(values, sample.Value)
		}
	}

	return values
}
