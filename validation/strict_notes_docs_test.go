package validation

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
)

var updateStrictNotes = flag.Bool("update-strict-notes", false, "regenerate the strictness notes table in docs/rules.md")

func TestDocsRulesStrictNotes(t *testing.T) {
	const start = "<!-- strictness-notes:start -->"
	const end = "<!-- strictness-notes:end -->"
	path := "../docs/rules.md"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	a, b := strings.Index(text, start), strings.Index(text, end)
	if a < 0 || b < a {
		t.Fatal("strictness notes table markers missing")
	}
	var table strings.Builder
	table.WriteString(start + "\n| rule id | name | since | why | also refuses | fix |\n| --- | --- | --- | --- | --- | --- |\n")
	cell := func(s string) string {
		return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", "|", "&#124;", "\n", "<br>").Replace(s)
	}
	for _, n := range strictNotes {
		fmt.Fprintf(&table, "| `%s` | %s | %s | %s | %s | %s |\n", n.ID, cell(n.Name), cell(n.Since), cell(n.Why), cell(strings.Join(n.AlsoRefuses, "\n\n")), cell(strings.Join(n.Fix, "\n\n")))
	}
	table.WriteString(end)
	want := text[:a] + table.String() + text[b+len(end):]
	if *updateStrictNotes {
		if err := os.WriteFile(path, []byte(want), 0644); err != nil {
			t.Fatal(err)
		}
	} else if text != want {
		t.Fatal("strictness notes table is stale; run go test ./validation -run '^TestDocsRulesStrictNotes$' -update-strict-notes")
	}
}
