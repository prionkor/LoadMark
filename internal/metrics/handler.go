package metrics

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Metrics WebSocket connection attempt")

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade failed:", err)
		return
	}

	fmt.Println("Metrics WebSocket connected")
	defer conn.Close()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("WebSocket read failed:", err)
			return
		}

		var containers []json.RawMessage

		if err := json.Unmarshal(data, &containers); err != nil {
			fmt.Println("Failed to decode metrics:", err)
			continue
		}

		for _, raw := range containers {
			samples, err := decodeContainer(raw)
			if err != nil {
				fmt.Println("Failed to decode container:", err)
				continue
			}

			for _, sample := range samples {
				s.result.Samples = append(s.result.Samples, sample)
			}
		}
	}
}
