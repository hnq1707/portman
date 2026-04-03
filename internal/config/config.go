package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/hnq1707/portman/internal/port"
)

// Config represents a .portman.yml project config.
type Config struct {
	Name  string      `yaml:"name"`
	Ports []PortSpec  `yaml:"ports"`
}

// PortSpec defines an expected port state.
type PortSpec struct {
	Port   int    `yaml:"port"`
	Name   string `yaml:"name"`
	Expect string `yaml:"expect"` // "listening" or "free"
}

// CheckResult holds the result of checking one port spec.
type CheckResult struct {
	Spec    PortSpec
	OK      bool
	Message string
}

// LoadConfig reads .portman.yml from the given directory.
func LoadConfig(dir string) (*Config, error) {
	path := dir + "/.portman.yml"

	data, err := os.ReadFile(path)
	if err != nil {
		// Try .portman.yaml too
		path = dir + "/.portman.yaml"
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("no .portman.yml or .portman.yaml found in %s", dir)
		}
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Default expect to "listening"
	for i := range cfg.Ports {
		if cfg.Ports[i].Expect == "" {
			cfg.Ports[i].Expect = "listening"
		}
	}

	return &cfg, nil
}

// CheckConfig verifies all ports against their expected state.
func CheckConfig(cfg *Config) ([]CheckResult, error) {
	currentPorts, err := port.ScanPorts()
	if err != nil {
		return nil, err
	}

	// Build lookup
	activeMap := make(map[int]port.PortInfo)
	for _, p := range currentPorts {
		activeMap[p.Port] = p
	}

	var results []CheckResult
	for _, spec := range cfg.Ports {
		result := CheckResult{Spec: spec}

		current, isListening := activeMap[spec.Port]

		switch spec.Expect {
		case "listening":
			if isListening {
				result.OK = true
				result.Message = fmt.Sprintf("listening (%s)", current.ProcessName)
			} else {
				result.OK = false
				result.Message = "not listening"
			}
		case "free":
			if !isListening {
				result.OK = true
				result.Message = "free"
			} else {
				result.OK = false
				result.Message = fmt.Sprintf("occupied by %s (PID %d)", current.ProcessName, current.PID)
			}
		default:
			result.OK = false
			result.Message = fmt.Sprintf("unknown expect: %s", spec.Expect)
		}

		results = append(results, result)
	}

	return results, nil
}
