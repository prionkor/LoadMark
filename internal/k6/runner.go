package k6

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/prionkor/benchmark-runner/internal/model"
)

func Run(config model.Config) (ClientResult, error) {
	request := config.Target.Request
	url := config.Target.BaseURL + config.Target.Path

	contentType := request.Headers["Content-Type"]
	body, err := encodeBody(request.Body, contentType)
	if err != nil {
		return ClientResult{}, err
	}

	headers, err := json.Marshal(request.Headers)

	executorSettings, err := json.Marshal(config.K6.ExecutorSettings)

	cmd := exec.Command(
		"k6",
		"run",
		"scripts/request.js",
	)

	cmd.Env = append(
		os.Environ(),
		"METHOD="+config.Target.Request.Method,
		"URL="+url,
		"BODY="+body,
		"HEADERS="+string(headers),
		"EXECUTOR="+config.K6.Executor,
		"EXECUTOR_SETTINGS="+string(executorSettings),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now()

	if err := cmd.Run(); err != nil {
		return ClientResult{}, fmt.Errorf("k6 failed: %w", err)
	}
	end := time.Now()

	var settings struct {
		Stages []model.Stage `json:"stages"`
	}

	data, err := json.Marshal(config.K6.ExecutorSettings)
	if err != nil {
		return ClientResult{}, fmt.Errorf("failed to marshal executor settings: %w", err)
	}

	if err := json.Unmarshal(data, &settings); err != nil {
		return ClientResult{}, fmt.Errorf("failed to parse executor settings: %w", err)
	}

	var stages []model.StageResult
	var stageStartTime = start
	for _, st := range settings.Stages {
		duration, err := time.ParseDuration(st.Duration)
		if err != nil {
			return ClientResult{}, fmt.Errorf(
				"invalid stage duration %q: %w",
				st.Duration,
				err,
			)
		}

		stage := model.StageResult{
			RPS:   st.RPS,
			Start: stageStartTime,
			End:   stageStartTime.Add(duration),
		}

		stages = append(stages, stage)
		stageStartTime = stage.End
	}

	return ClientResult{
		Start:  start,
		End:    end,
		Stages: stages,
	}, nil
}
