package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Remote settings. EAV_* are the current names; ARGCHECK_* are the names from
// before the rename to english-arg-validator and are still read as fallbacks.
const (
	settingRemoteURL  = "EAV_REMOTE_URL"
	settingRemoteTool = "EAV_REMOTE_TOOL"
)

var legacySettingNames = map[string]string{
	settingRemoteURL:  "ARGCHECK_REMOTE_URL",
	settingRemoteTool: "ARGCHECK_REMOTE_TOOL",
}

// remoteSetting returns key from the environment, or else from the config
// file: $XDG_CONFIG_HOME/english-arg-validator/config (default
// ~/.config/english-arg-validator/config), one KEY=VALUE per line, with #
// comments. The environment always wins. The pre-rename names (ARGCHECK_* and
// .../argcheck/config) are read when the current ones are unset.
func remoteSetting(key string) string {
	names := []string{key}
	if legacy, ok := legacySettingNames[key]; ok {
		names = append(names, legacy)
	}
	for _, name := range names {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}
	for _, path := range configPaths() {
		for _, name := range names {
			if value := settingFromFile(path, name); value != "" {
				return value
			}
		}
	}
	return ""
}

func settingFromFile(path, key string) string {
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

func configDir() string {
	if dir := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, ".config")
}

// configPaths lists the current config file, then the pre-rename one.
func configPaths() []string {
	dir := configDir()
	if dir == "" {
		return nil
	}
	return []string{
		filepath.Join(dir, "english-arg-validator", "config"),
		filepath.Join(dir, "argcheck", "config"),
	}
}

func configPathHint() string {
	if paths := configPaths(); len(paths) > 0 {
		return paths[0]
	}
	return "$XDG_CONFIG_HOME/english-arg-validator/config"
}
