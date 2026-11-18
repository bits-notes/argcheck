package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoteSettingReadsConfigFileAndEnvWins(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	if err := os.MkdirAll(filepath.Join(dir, "argcheck"), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "# remote checker\nARGCHECK_REMOTE_URL = \"http://example.test/mcp\"\nARGCHECK_REMOTE_TOOL=formalizer.validate\n"
	if err := os.WriteFile(filepath.Join(dir, "argcheck", "config"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EAV_REMOTE_URL", "")
	t.Setenv("EAV_REMOTE_TOOL", "")
	t.Setenv("ARGCHECK_REMOTE_URL", "")
	t.Setenv("ARGCHECK_REMOTE_TOOL", "")
	// The pre-rename file and names are read as a fallback.
	if got := remoteSetting(settingRemoteURL); got != "http://example.test/mcp" {
		t.Fatalf("url from file = %q", got)
	}
	if got := remoteSetting(settingRemoteTool); got != "formalizer.validate" {
		t.Fatalf("tool from file = %q", got)
	}
	t.Setenv("ARGCHECK_REMOTE_URL", "http://env.test/mcp")
	if got := remoteSetting(settingRemoteURL); got != "http://env.test/mcp" {
		t.Fatalf("env should win, got %q", got)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "missing"))
	t.Setenv("ARGCHECK_REMOTE_URL", "")
	if got := remoteSetting(settingRemoteURL); got != "" {
		t.Fatalf("missing config should give empty, got %q", got)
	}
}

func TestRemoteSettingPrefersCurrentNames(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	for _, sub := range []string{"english-arg-validator", "argcheck"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "english-arg-validator", "config"), []byte("EAV_REMOTE_URL=http://new.test/mcp\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "argcheck", "config"), []byte("ARGCHECK_REMOTE_URL=http://old.test/mcp\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("EAV_REMOTE_URL", "")
	t.Setenv("ARGCHECK_REMOTE_URL", "")
	if got := remoteSetting(settingRemoteURL); got != "http://new.test/mcp" {
		t.Fatalf("current config should win over the pre-rename one, got %q", got)
	}
	t.Setenv("ARGCHECK_REMOTE_URL", "http://legacy-env.test/mcp")
	if got := remoteSetting(settingRemoteURL); got != "http://legacy-env.test/mcp" {
		t.Fatalf("an environment variable (even the pre-rename name) beats any file, got %q", got)
	}
	t.Setenv("EAV_REMOTE_URL", "http://env.test/mcp")
	if got := remoteSetting(settingRemoteURL); got != "http://env.test/mcp" {
		t.Fatalf("EAV_REMOTE_URL should win, got %q", got)
	}
}
