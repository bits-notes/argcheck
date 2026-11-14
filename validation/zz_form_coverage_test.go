package validation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

type formCoverageCount struct {
	Positive int `json:"positive"`
	Negative int `json:"negative"`
}

func TestMain(m *testing.M) {
	path := os.Getenv("ARGCHECK_FORM_COVERAGE")
	if path == "" {
		os.Exit(m.Run())
	}

	var mu sync.Mutex
	counts := map[string]formCoverageCount{}
	gateStepObserver = func(form string, passed bool) {
		mu.Lock()
		defer mu.Unlock()
		entry := counts[form]
		if passed {
			entry.Positive++
		} else {
			entry.Negative++
		}
		counts[form] = entry
	}

	code := m.Run()
	if err := writeFormCoverage(path, counts); err != nil {
		fmt.Fprintf(os.Stderr, "form coverage: %v\n", err)
		if code == 0 {
			code = 1
		}
	}
	os.Exit(code)
}

func writeFormCoverage(path string, counts map[string]formCoverageCount) error {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	data, err := json.MarshalIndent(counts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}
