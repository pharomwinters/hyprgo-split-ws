package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const DefaultWorkspacesPerMonitor = 10

// Config holds user preferences for hyprgo-split-ws.
type Config struct {
	// MonitorOrder overrides the default alphabetical monitor ordering.
	// If empty, monitors are sorted alphabetically by name.
	MonitorOrder []string

	// WorkspacesPerMonitor sets how many virtual workspaces each monitor gets.
	WorkspacesPerMonitor int
}

// DefaultConfig returns a config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		WorkspacesPerMonitor: DefaultWorkspacesPerMonitor,
	}
}

// configPath returns the path to the config file.
func configPath() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "hypr", "hyprgo-split-ws.conf")
}

// Load reads the config file and returns a Config.
// If the file doesn't exist, returns defaults.
func Load() (*Config, error) {
	cfg := DefaultConfig()

	path := configPath()
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to open config: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			return nil, fmt.Errorf("config line %d: missing '=' in %q", lineNum, line)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		switch key {
		case "monitor_order":
			parts := strings.Split(value, ",")
			cfg.MonitorOrder = make([]string, 0, len(parts))
			for _, p := range parts {
				name := strings.TrimSpace(p)
				if name != "" {
					cfg.MonitorOrder = append(cfg.MonitorOrder, name)
				}
			}

		case "workspaces_per_monitor":
			n, err := strconv.Atoi(value)
			if err != nil || n < 1 {
				return nil, fmt.Errorf("config line %d: workspaces_per_monitor must be a positive integer", lineNum)
			}
			cfg.WorkspacesPerMonitor = n

		default:
			return nil, fmt.Errorf("config line %d: unknown key %q", lineNum, key)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return cfg, nil
}
