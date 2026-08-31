package prometheus

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (c *Client) QueryRange(
	query string,
	start time.Time,
	end time.Time,
	step time.Duration,
) (*QueryResponse, error) {

	req, err := http.NewRequest(
		http.MethodGet,
		c.BaseURL+"/api/v1/query_range",
		nil,
	)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.Username, c.Password)

	q := req.URL.Query()
	q.Set("query", query)
	q.Set("start", start.Format(time.RFC3339))
	q.Set("end", end.Format(time.RFC3339))
	q.Set("step", step.String())

	req.URL.RawQuery = q.Encode()

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("prometheus request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"prometheus returned status %d",
			resp.StatusCode,
		)
	}

	var result QueryResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf(
			"failed to decode prometheus response: %w",
			err,
		)
	}

	if result.Status != "success" {
		return nil, fmt.Errorf(
			"prometheus query failed with status %q",
			result.Status,
		)
	}

	return &result, nil
}
