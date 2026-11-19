package validation

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

// attachStrictNotesExhaustive is the v42.1 notes pass: it retries every rule
// on every failing step. The fired-rules pass must attach the same notes.
func attachStrictNotesExhaustive(argument NumberedArgument, report *EntailmentGateReport) {
	if report.Passed {
		return
	}
	expanded, _, err := ExpandNumberedArgumentBridges(argument)
	if err != nil {
		return
	}
	steps := make(map[int]NumberedArgumentStep, len(expanded.Steps))
	for _, step := range expanded.Steps {
		steps[step.Number] = step
	}
	for i := range report.Steps {
		failed := &report.Steps[i]
		if failed.Passed {
			continue
		}
		step, ok := steps[failed.Step]
		if !ok || failed.NearestForm == "scope" || failed.NearestForm == "explicit-variable FOL" {
			continue
		}
		for _, note := range strictNotes {
			if allStrictRules().without(note.ID).strictStepPassesEitherReading(step, steps) {
				failed.Notes = append(failed.Notes, note)
				if len(failed.Notes) == 3 {
					break
				}
			}
		}
	}
}

func stepNotes(report EntailmentGateReport) map[int][]string {
	out := map[int][]string{}
	for _, s := range report.Steps {
		for _, n := range s.Notes {
			out[s.Step] = append(out[s.Step], n.ID)
		}
	}
	return out
}

// notesParity compares the fired-rules notes with the exhaustive pass on one
// argument and returns the two timings.
func notesParity(t *testing.T, name, text string) (fast, exhaustive time.Duration, ok bool) {
	t.Helper()
	start := time.Now()
	got, _ := CheckEntailmentGate(text, notesOn)
	fast = time.Since(start)

	argument, parsed := allStrictRules().ParseNumberedArgument(text, numberedArgumentOptions{})
	if !parsed.Passed {
		return fast, 0, true
	}
	argument = protectArgumentNameLetters(argument)
	start = time.Now()
	want, _ := CheckEntailmentGate(text)
	attachStrictNotesExhaustive(argument, &want)
	exhaustive = time.Since(start)

	if got.Passed != want.Passed {
		t.Errorf("%s: verdict changed with notes: %v vs %v", name, got.Passed, want.Passed)
		return fast, exhaustive, false
	}
	if g, w := stepNotes(got), stepNotes(want); !reflect.DeepEqual(g, w) {
		t.Errorf("%s: notes differ\n got  %v\n want %v", name, g, w)
		return fast, exhaustive, false
	}
	return fast, exhaustive, true
}

func TestStrictNotesFiredRulesMatchExhaustive(t *testing.T) {
	cases := []string{}
	for _, note := range strictNotes {
		cases = append(cases, note.AlsoRefuses...)
	}
	for i, text := range cases {
		notesParity(t, fmt.Sprintf("also_refuses[%d]", i), text)
	}
}

// ARGCHECK_NOTES_PARITY_LIST names a file of argument paths, one per line.
// ARGCHECK_NOTES_PARITY_OUT, if set, receives one JSON timing row per path.
func TestStrictNotesFiredRulesCorpusParity(t *testing.T) {
	list := os.Getenv("ARGCHECK_NOTES_PARITY_LIST")
	if list == "" {
		t.Skip("set ARGCHECK_NOTES_PARITY_LIST to compare notes on a corpus")
	}
	f, err := os.Open(list)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if p := strings.TrimSpace(sc.Text()); p != "" {
			paths = append(paths, p)
		}
	}
	f.Close()
	var out *json.Encoder
	if path := os.Getenv("ARGCHECK_NOTES_PARITY_OUT"); path != "" {
		file, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		out = json.NewEncoder(file)
	}
	type row struct {
		ID           string  `json:"id"`
		FastMS       float64 `json:"fast_ms"`
		ExhaustiveMS float64 `json:"exhaustive_ms"`
		Same         bool    `json:"same"`
	}
	var mu sync.Mutex
	jobs := make(chan string)
	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Error(err)
					continue
				}
				fast, exhaustive, same := notesParity(t, path, string(data))
				if out != nil {
					mu.Lock()
					_ = out.Encode(row{path, float64(fast.Microseconds()) / 1000, float64(exhaustive.Microseconds()) / 1000, same})
					mu.Unlock()
				}
			}
		}()
	}
	for _, p := range paths {
		jobs <- p
	}
	close(jobs)
	wg.Wait()
}
