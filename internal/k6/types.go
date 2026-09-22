package k6

import (
	"time"

	"github.com/prionkor/loadmark/internal/model"
)

type ClientResult struct {
	Start  time.Time
	End    time.Time
	Stages []model.StageResult
}
