package model

import (
	"time"

	"github.com/prionkor/benchmark-runner/internal/metrics"
)

type Config struct {
	Name        string               `yaml:"name"`
	Description string               `yaml:"description"`
	Target      Target               `yaml:"target"`
	K6          K6                   `yaml:"k6"`
	Prometheus  Prometheus           `yaml:"prometheus"`
	Monitoring  []MonitoringResource `yaml:"monitoring"`
}

type Target struct {
	BaseURL string  `yaml:"base_url"`
	Path    string  `yaml:"path"`
	Request Request `yaml:"request"`
}

type Request struct {
	Method  string            `yaml:"method"`
	Headers map[string]string `yaml:"headers"`
	Body    map[string]string `yaml:"body"`
}

type K6 struct {
	Executor         string         `yaml:"executor"`
	ExecutorSettings map[string]any `yaml:"executor_settings"`
}

type Stage struct {
	RPS      int    `yaml:"rps" json:"rps"`
	Duration string `yaml:"duration" json:"duration"`
}

type Prometheus struct {
	CollectionDelay string `yaml:"collection_delay"`
	URL             string `yaml:"url"`
	Auth            Auth   `yaml:"auth"`
	Username        string `yaml:"-"`
	Password        string `yaml:"-"`
}

type Auth struct {
	Type     string      `yaml:"type"`
	Username ValueSource `yaml:"username"`
	Password ValueSource `yaml:"password"`
}

type ValueSource struct {
	Value   string `yaml:"value"`
	FromEnv string `yaml:"from_env"`
}

type MonitoringResource struct {
	Name           string  `yaml:"name"`
	Namespace      string  `yaml:"namespace"`
	PodNamePattern string  `yaml:"pod_name_pattern"`
	Metrics        Metrics `yaml:"metrics"`
}

type Metrics struct {
	CPU      []string `yaml:"cpu"`
	Memory   []string `yaml:"memory"`
	Network  []string `yaml:"network"`
	Restarts []string `yaml:"restarts"`
}

type BenchmarkResult struct {
	Stages []StageResult
}

type StageResult struct {
	RPS       int
	Start     time.Time
	End       time.Time
	Resources []ResourceResult
}

type ResourceResult struct {
	Name    string
	Metrics map[string]metrics.Statistics
	Error   string
}
