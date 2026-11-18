// Command formcoverage prints docs/test-coverage-by-rule.md from a coverage
// JSON file. english-arg-validator --forms prints validation.EntailmentGateForms; this
// uses that same list, in that order.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/bits-notes/english-arg-validator/validation"
)

type coverageCount struct {
	Positive int `json:"positive"`
	Negative int `json:"negative"`
}

const regenerateCommands = "ARGCHECK_FORM_COVERAGE=$PWD/tmp/form-coverage.json go test ./validation/...\n" +
	"go run ./tools/formcoverage tmp/form-coverage.json > docs/test-coverage-by-rule.md\n"

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: formcoverage <form-coverage.json>")
		os.Exit(2)
	}
	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "formcoverage:", err)
		os.Exit(1)
	}
	counts := map[string]coverageCount{}
	if len(bytes.TrimSpace(raw)) > 0 {
		if err := json.Unmarshal(raw, &counts); err != nil {
			fmt.Fprintln(os.Stderr, "formcoverage:", err)
			os.Exit(1)
		}
	}

	forms := validation.EntailmentGateForms()
	known := make(map[string]bool, len(forms))
	for _, name := range forms {
		known[name] = true
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# Test coverage by inference form\n\n")
	fmt.Fprintf(&b, "Positive is the number of gated steps the validation tests accepted with that form. Negative is the number of gated steps they rejected with that form as `nearest_form`.\n\n")
	fmt.Fprintf(&b, "Catalog version: `%s`.\n\n", validation.EntailmentGateCatalogVersion)
	fmt.Fprintf(&b, "The table lists every form `english-arg-validator --forms` prints (`validation.EntailmentGateForms`), in that order. Regenerate it with the commands below.\n\n")
	fmt.Fprintf(&b, "## Regenerate\n\n```\n%s```\n\n", regenerateCommands)
	fmt.Fprintf(&b, "| Form | Positive | Negative |\n")
	fmt.Fprintf(&b, "| --- | ---: | ---: |\n")

	noPositive, noNegative := 0, 0
	for _, name := range forms {
		count := counts[name]
		if count.Positive == 0 {
			noPositive++
		}
		if count.Negative == 0 {
			noNegative++
		}
		fmt.Fprintf(&b, "| %s | %d | %d |\n", formCell(name), count.Positive, count.Negative)
	}

	fmt.Fprintf(&b, "\n## Summary\n\n")
	fmt.Fprintf(&b, "%d forms have 0 positive. %d forms have 0 negative.\n\n", noPositive, noNegative)
	fmt.Fprintf(&b, "## Forms counted outside `--forms`\n\n")
	fmt.Fprintf(&b, "A name in this list was reported for a gated step and is not printed by `english-arg-validator --forms`. Reported names can differ from catalog names by a suffix such as ` (necessitation)`.\n\n")

	var extra []string
	for name := range counts {
		if !known[name] {
			extra = append(extra, name)
		}
	}
	sort.Strings(extra)
	if len(extra) == 0 {
		fmt.Fprintf(&b, "None.\n")
	} else {
		for _, name := range extra {
			count := counts[name]
			fmt.Fprintf(&b, "- %s: positive %d, negative %d\n", formCell(name), count.Positive, count.Negative)
		}
	}

	fmt.Fprint(os.Stdout, b.String())
}

func formCell(name string) string {
	name = strings.ReplaceAll(name, "|", "\\|")
	name = strings.ReplaceAll(name, "\n", " ")
	name = strings.ReplaceAll(name, "`", "'")
	return "`" + name + "`"
}
