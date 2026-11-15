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
	t.Setenv("ARGCHECK_REMOTE_URL", "")
	t.Setenv("ARGCHECK_REMOTE_TOOL", "")
	if got := remoteSetting("ARGCHECK_REMOTE_URL"); got != "http://example.test/mcp" {
		t.Fatalf("url from file = %q", got)
	}
	if got := remoteSetting("ARGCHECK_REMOTE_TOOL"); got != "formalizer.validate" {
		t.Fatalf("tool from file = %q", got)
	}
	t.Setenv("ARGCHECK_REMOTE_URL", "http://env.test/mcp")
	if got := remoteSetting("ARGCHECK_REMOTE_URL"); got != "http://env.test/mcp" {
		t.Fatalf("env should win, got %q", got)
	}
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "missing"))
	t.Setenv("ARGCHECK_REMOTE_URL", "")
	if got := remoteSetting("ARGCHECK_REMOTE_URL"); got != "" {
		t.Fatalf("missing config should give empty, got %q", got)
	}
}
