package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// remoteSetting returns key from the environment, or else from the argcheck
// config file: $XDG_CONFIG_HOME/argcheck/config (default ~/.config/argcheck/config),
// one KEY=VALUE per line, with # comments. The environment always wins.
func remoteSetting(key string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	path := configPath()
	if path == "" {
		return ""
	}
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name, value, ok := strings.Cut(line, "=")
		if ok && strings.TrimSpace(name) == key {
			return strings.Trim(strings.TrimSpace(value), `"'`)
		}
	}
	return ""
}

func configPath() string {
	if dir := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); dir != "" {
		return filepath.Join(dir, "argcheck", "config")
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, ".config", "argcheck", "config")
}

func configPathHint() string {
	if path := configPath(); path != "" {
		return path
	}
	return "$XDG_CONFIG_HOME/argcheck/config"
}
