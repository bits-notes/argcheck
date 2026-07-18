package validation

import (
	"encoding/json"
	"os"
	"testing"
)

// The gate-aware benchmark contract (2026-07-17): transparency is decided by the
// deterministic entailment gate, not by the LLM judge. Every gold
// transparently_valid case must gate-pass, and every non-transparent case
// (has_opaque_entailments / opaquely_valid / invalid) must gate-fail. Cases
// flagged catalog_gap are the documented exceptions — transparent to a reader
// but outside the current schema catalog (e.g. relational/variable-binding
// instantiation), so they legitimately gate-fail while staying gold-transparent.
// This test enforces the contract in CI without any LLM call, so a mislabeled
// gold cannot be frozen.

type goldGateCase struct {
	ID               string `json:"id"`
	GoldStatus       string `json:"gold_status"`
	CatalogGap       string `json:"catalog_gap"`
	StructuralDefect string `json:"structural_defect"`
	Argument         string `json:"argument"`
	Input            string `json:"input"`
}

func (c goldGateCase) text() string {
	if c.Argument != "" {
		return c.Argument
	}
	return c.Input
}

func loadGoldGateCases(t *testing.T, path string) []goldGateCase {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	// battery.json is {cases:[...]}; cases.json is also {cases:[...]}.
	var wrapper struct {
		Cases []goldGateCase `json:"cases"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	if len(wrapper.Cases) == 0 {
		t.Fatalf("%s has no cases", path)
	}
	return wrapper.Cases
}

func assertGoldGateConsistent(t *testing.T, path string) {
	t.Helper()
	for _, c := range loadGoldGateCases(t, path) {
		c := c
		t.Run(c.ID, func(t *testing.T) {
			parsed, res := ParseNumberedArgument(c.text(), numberedArgumentOptions{})
			if !res.Passed {
				t.Fatalf("case %s does not parse: %s", c.ID, res.Message)
			}
			unused := NumberedArgumentUnusedSteps(parsed)
			if c.StructuralDefect == "unused_premise" {
				// Connectivity defect is orthogonal to the gate (its conclusions may
				// still gate-pass); the structural check is what must reject it.
				if len(unused) == 0 {
					t.Fatalf("structural_defect case %s must have an uncited premise but none was found", c.ID)
				}
				return
			}
			if len(unused) > 0 {
				t.Fatalf("gold case %s has dead premises %v; every battery argument must be connectivity-clean", c.ID, unused)
			}

			report, _ := CheckEntailmentGate(c.text())
			transparent := c.GoldStatus == "transparently_valid"
			switch {
			case c.CatalogGap != "":
				if !transparent {
					t.Fatalf("catalog_gap case %s must stay gold transparently_valid, got %q", c.ID, c.GoldStatus)
				}
				if report.Passed {
					t.Fatalf("catalog_gap case %s (%s) is expected to gate-FAIL (documents the gap) but gate-passed", c.ID, c.CatalogGap)
				}
			case transparent:
				if !report.Passed {
					t.Fatalf("gold transparently_valid case %s must gate-pass; gate failed. Relabel it, decompose it, or flag catalog_gap.", c.ID)
				}
			default:
				if report.Passed {
					t.Fatalf("gold %s case %s must gate-fail (gate must never certify a non-transparent step) but gate-passed", c.GoldStatus, c.ID)
				}
			}
		})
	}
}

func TestBatteryGoldIsGateConsistent(t *testing.T) {
	assertGoldGateConsistent(t, "../../bench/argbench/battery.json")
}

func TestAtomicVerificationGoldIsGateConsistent(t *testing.T) {
	assertGoldGateConsistent(t, "../../bench/suites/argument-verification-atomic/cases.json")
}
