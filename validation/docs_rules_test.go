package validation

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type rulesDocBlock struct {
	kind      string
	form      string
	startLine int
	body      string
}

func parseRulesDocBlocks(t *testing.T, text string) []rulesDocBlock {
	t.Helper()
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	var blocks []rulesDocBlock
	inFence := false
	fenceKind := ""
	fenceForm := ""
	fenceStart := 0
	var bodyLines []string

	for i, line := range lines {
		lineNo := i + 1
		trimmed := strings.TrimSpace(line)
		if !inFence {
			if !strings.HasPrefix(trimmed, "```") {
				continue
			}
			info := strings.TrimSpace(strings.TrimPrefix(trimmed, "```"))
			switch {
			case info == "argcheck-accept" || strings.HasPrefix(info, "argcheck-accept "):
				fenceKind = "accept"
				fenceForm = strings.TrimSpace(strings.TrimPrefix(info, "argcheck-accept"))
				if fenceForm == "" {
					t.Fatalf("docs/rules.md:%d: argcheck-accept fence missing form name", lineNo)
				}
			case info == "argcheck-reject" || strings.HasPrefix(info, "argcheck-reject "):
				fenceKind = "reject"
				fenceForm = strings.TrimSpace(strings.TrimPrefix(info, "argcheck-reject"))
				if fenceForm == "" {
					t.Fatalf("docs/rules.md:%d: argcheck-reject fence missing form name", lineNo)
				}
			case strings.HasPrefix(info, "argcheck"):
				t.Fatalf("docs/rules.md:%d: unrecognized argcheck fence %q", lineNo, info)
			default:
				fenceKind = "other"
				fenceForm = ""
			}
			inFence = true
			fenceStart = lineNo
			bodyLines = bodyLines[:0]
			continue
		}

		if trimmed == "```" {
			if fenceKind == "accept" || fenceKind == "reject" {
				blocks = append(blocks, rulesDocBlock{
					kind:      fenceKind,
					form:      fenceForm,
					startLine: fenceStart,
					body:      strings.Join(bodyLines, "\n"),
				})
			}
			inFence = false
			fenceKind = ""
			fenceForm = ""
			bodyLines = nil
			continue
		}
		bodyLines = append(bodyLines, line)
	}

	if inFence {
		t.Fatalf("docs/rules.md:%d: unclosed code fence", fenceStart)
	}
	return blocks
}

// checkRulesDocArgument runs the exact local check sequence used by cmd/argcheck.
func checkRulesDocArgument(argument string) (bool, EntailmentGateReport, string) {
	argument = strings.TrimSpace(argument)
	syntax := Validate(context.Background(), BuiltinSpec(BuiltinNumberedArgumentSyntax), argument)
	if !syntax.Passed {
		return false, EntailmentGateReport{}, "syntax: " + syntax.Message
	}
	report, entailment := CheckEntailmentGate(argument)
	return entailment.Passed, report, entailment.Message
}

func TestDocsRulesExamples(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "docs", "rules.md"))
	if err != nil {
		t.Fatalf("read docs/rules.md: %v", err)
	}
	blocks := parseRulesDocBlocks(t, string(raw))
	if len(blocks) == 0 {
		t.Fatal("docs/rules.md contains no argcheck-accept or argcheck-reject blocks")
	}

	catalogForms := EntailmentGateForms()
	knownForms := make(map[string]bool, len(catalogForms))
	for _, form := range catalogForms {
		knownForms[form] = true
	}

	acceptedCount := make(map[string]int, len(catalogForms))
	rejectedCount := make(map[string]int, len(catalogForms))

	for _, block := range blocks {
		if !knownForms[block.form] {
			t.Fatalf("docs/rules.md:%d: unknown catalog form %q", block.startLine, block.form)
		}
		valid, report, msg := checkRulesDocArgument(block.body)
		switch block.kind {
		case "accept":
			acceptedCount[block.form]++
			if !valid {
				t.Errorf("docs/rules.md:%d: accept block for %q was INVALID: %s\n%s", block.startLine, block.form, msg, block.body)
				continue
			}
			used := false
			for _, step := range report.Steps {
				if step.Passed && step.Form == block.form {
					used = true
					break
				}
			}
			if !used {
				t.Errorf("docs/rules.md:%d: accept block for %q passed without using %q; steps=%+v\n%s", block.startLine, block.form, block.form, report.Steps, block.body)
			}
		case "reject":
			rejectedCount[block.form]++
			if valid {
				t.Errorf("docs/rules.md:%d: reject block for %q was VALID; steps=%+v\n%s", block.startLine, block.form, report.Steps, block.body)
			}
		}
	}

	for _, form := range catalogForms {
		if acceptedCount[form] == 0 {
			t.Errorf("docs/rules.md: missing argcheck-accept block for %q", form)
		}
		if rejectedCount[form] == 0 {
			t.Errorf("docs/rules.md: missing argcheck-reject block for %q", form)
		}
	}
}
