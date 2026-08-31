package config

import (
	"fmt"
	"os"
	"time"

	"github.com/prionkor/benchmark-runner/internal/model"
	"gopkg.in/yaml.v3"
)

func Load(path string) (model.Config, error) {
	data, err := os.ReadFile("config.yml")
	if err != nil {
		return model.Config{}, err
	}

	var config model.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return model.Config{}, err
	}

	switch config.Target.Request.Method {
	case "GET", "POST", "PUT", "DELETE", "PATCH":
		// Valid method, do nothing
	default:
		return model.Config{}, fmt.Errorf(
			"unsupported HTTP method: %q",
			config.Target.Request.Method,
		)
	}

	if config.Prometheus.Auth.Type != "basic" {
		return model.Config{}, fmt.Errorf(
			"unsupported prometheus auth type: %s",
			config.Prometheus.Auth.Type,
		)
	}

	username, err := resolveValue(config.Prometheus.Auth.Username)
	if err != nil {
		return model.Config{}, fmt.Errorf("prometheus username: %w", err)
	}

	password, err := resolveValue(config.Prometheus.Auth.Password)
	if err != nil {
		return model.Config{}, fmt.Errorf("prometheus password: %w", err)
	}

	config.Prometheus.Username = username
	config.Prometheus.Password = password

	if _, err := time.ParseDuration(config.Prometheus.CollectionDelay); err != nil {
		return model.Config{}, fmt.Errorf(
			"invalid prometheus.collection_delay %q: %w",
			config.Prometheus.CollectionDelay,
			err,
		)
	}

	return config, nil
}

func resolveValue(source model.ValueSource) (string, error) {
	if source.Value != "" && source.FromEnv != "" {
		return "", fmt.Errorf("value and from_env cannot both be specified")
	}

	if source.Value != "" {
		return source.Value, nil
	}

	if source.FromEnv != "" {
		value := os.Getenv(source.FromEnv)
		if value == "" {
			return "", fmt.Errorf(
				"environment variable %q is not set",
				source.FromEnv,
			)
		}
		return value, nil
	}

	return "", fmt.Errorf("value or from_env is needed")
}
