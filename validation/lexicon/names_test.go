package lexicon

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestNameDeterministicRegeneration(t *testing.T) {
	cache, err := filepath.Abs("../../workspace/name-cache")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"names.zip", "surnames.csv", "census-LICENSE", "census-README", "wordnet.zip", "word_count.txt", "frequency-LICENSE"} {
		if _, err := os.Stat(filepath.Join(cache, f)); err != nil {
			t.Skip("populate pinned name cache to verify regeneration")
		}
	}
	out := t.TempDir()
	for run := 0; run < 2; run++ {
		cmd := exec.Command("python3", "../../scripts/gen-name-allowlist/generate.py", "--cache", cache, "--output-dir", out)
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("generator: %v: %s", err, output)
		}
		for _, name := range []string{"names.txt", "name_forbidden.txt", "EXCLUDED-names.md", "LICENSE-names.txt", "LICENSE-census-names.txt"} {
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
