package lexicon

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestLexiconV46Membership(t *testing.T) {
	for _, w := range strings.Fields("pump valve dog cat child death concept proof argument agent system machine robot model firm company government king queen man woman person world universe simulation process policy rule law right reason cause effect outcome risk goal value") {
		if !IsSingularNoun(w) || !IsSingularNoun(strings.ToUpper(w)) {
			t.Error(w)
		}
	}
	for _, w := range strings.Fields("registered aligned informed trained aps ultraintelligent parking reliable happy red causal informational dangerous") {
		if !IsAdjective(w) || !IsAdjective(strings.ToUpper(w)) {
			t.Error(w)
		}
	}
	for _, w := range strings.Fields("lot bulk majority amount group set plenty dozens everyone something is has does") {
		if IsSingularNoun(w) {
			t.Error("singular admitted", w)
		}
	}
	excluded, err := os.ReadFile("../../scripts/gen-noun-allowlist/adjective-exclusions.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(excluded), "\n") {
		for _, w := range strings.Fields(strings.Split(line, "#")[0]) {
			if IsAdjective(w) {
				t.Error("adjective admitted", w)
			}
		}
	}
	for _, w := range []string{"", " robot", "robot ", "red blue", "went", "was", "quokkas"} {
		if IsAdjective(w) || IsSingularNoun(w) {
			t.Error("unexpected", w)
		}
	}
	if !IsMassNoun("water") || !IsMassNoun("humanity") || IsMassNoun("robots") {
		t.Fatal("mass agreement classification")
	}
}
func TestLexiconV46Format(t *testing.T) {
	for _, tc := range []struct {
		data     string
		min, max int
	}{{singularNouns, 12000, 16000}, {adjectives, 4000, 6000}, {massNouns, 1, 1000}} {
		previous := ""
		count := 0
		for _, line := range strings.Split(tc.data, "\n") {
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if !regexp.MustCompile(`^[a-z]+$`).MatchString(line) || line <= previous {
				t.Fatal(line)
			}
			previous = line
			count++
		}
		if count < tc.min || count > tc.max {
			t.Fatal(count)
		}
	}
}
func TestLexiconDeterministicRegeneration(t *testing.T) {
	cache, err := filepath.Abs("../../.cache/noun-allowlist")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(cache, "wordnet.zip")); err != nil {
		t.Skip("populate pinned source cache to verify regeneration")
	}
	out := t.TempDir()
	for run := 0; run < 2; run++ {
		cmd := exec.Command("python3", "../../scripts/gen-noun-allowlist/generate.py", "--cache", cache, "--output-dir", out)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("generator: %v: %s", err, output)
		}
		for _, name := range []string{"singular_nouns.txt", "adjectives.txt", "mass_nouns.txt", "plural_mass_nouns.txt", "EXCLUDED.md"} {
			got, err := os.ReadFile(filepath.Join(out, name))
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("run %d: regenerate %s", run, name)
			}
		}
	}
}
