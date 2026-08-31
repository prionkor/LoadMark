package k6

import (
	"time"

	"github.com/prionkor/benchmark-runner/internal/model"
)

type ClientResult struct {
	Start  time.Time
	End    time.Time
	Stages []model.StageResult
}
