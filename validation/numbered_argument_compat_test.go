package validation

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Golden digests were captured with the v13 parser before Fitch support. They
// cover every parseable string in the existing battery/cases/testdata JSON.
func TestNumberedArgumentFlatJSONCompatibility(t *testing.T) {
	got := map[string]string{}
	var walkValue func(any)
	walkValue = func(v any) {
		switch v := v.(type) {
		case string:
			if !strings.Contains(v, "\n") {
				return
			}
			a, r := ParseNumberedArgument(v, numberedArgumentOptions{})
			if !r.Passed {
				return
			}
			b, _ := json.Marshal(a)
			got[fmt.Sprintf("%x", sha256.Sum256([]byte(v)))] = fmt.Sprintf("%x", sha256.Sum256(b))
		case []any:
			for _, item := range v {
				walkValue(item)
			}
		case map[string]any:
			for _, item := range v {
				walkValue(item)
			}
		}
	}
	err := filepath.WalkDir("testdata", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "workspace" || d.Name() == "node_modules" || d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var v any
		if json.Unmarshal(b, &v) == nil {
			walkValue(v)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	const golden = "testdata/flat_parser_v13_hashes.json"
	if os.Getenv("UPDATE_FLAT_PARSER_GOLDEN") == "1" {
		b, _ := json.MarshalIndent(got, "", "  ")
		if err := os.WriteFile(golden, append(b, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
		return
	}
	b, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	var want map[string]string
	if err := json.Unmarshal(b, &want); err != nil {
		t.Fatal(err)
	}
	for input, digest := range want {
		if got[input] != digest {
			t.Errorf("flat fixture %s changed or no longer parses", input)
		}
	}
	if len(want) == 0 {
		t.Fatal("empty baseline")
	}
	t.Logf("%d v13 fixture strings retain byte-identical JSON", len(want))
}
