package metrics

import "time"

type sample struct {
	Metric struct {
		Name string `json:"name"`
	} `json:"Metric"`

	Time  time.Time         `json:"Time"`
	Value float64           `json:"Value"`
	Tags  map[string]string `json:"Tags"`
}
