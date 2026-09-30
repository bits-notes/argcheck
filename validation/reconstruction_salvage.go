package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// SalvagedReconstruction identifies deterministic assumptions separately from
// writer-supplied premises. FailingSteps uses the writer's original numbering.
type SalvagedReconstruction struct {
	Reconstruction FormalizedReconstructionOutput
	AssumedLines   []int
	FailingSteps   []int
}

// SalvageReconstruction keeps a structurally sound, sentence-linked reply and
// supplies at most three explicit connecting assumptions for isolated failures.
func SalvageReconstruction(reply, source string) (SalvagedReconstruction, error) {
	n := len(SegmentSourceSentences(source))
	checked := validateReconstructionChecks(context.Background(), reply, &n, false, false)
	if !checked.Passed {
		return SalvagedReconstruction{}, fmt.Errorf("reconstruction structure: %s", checked.Message)
	}
	var original FormalizedReconstructionOutput
	if err := json.Unmarshal(checked.ParsedOutput, &original); err != nil {
		return SalvagedReconstruction{}, err
	}
	argument, _ := ParseNumberedArgumentAllowingNoConclusions(original.Argument)
	mappings := []FormalizationLineMapping{}
	conclusionCount := 0
	for _, line := range original.Lines {
		if len(line.Sentences) > 0 {
			mappings = append(mappings, FormalizationLineMapping{RepairedLine: line.Line, FromAsFoundLines: []int{line.Line}})
		}
	}
	removed := map[int]bool{}
	for _, step := range repairAssociatedConditionalSteps(original.Argument, mappings) {
		removed[step.Number] = true
	}
	// Expand inline references before renumbering, so no stale numeric references
	// remain inside statements when padding is removed or bridges are inserted.
	argument, _, err := ExpandNumberedArgumentBridges(argument)
	if err != nil {
		return SalvagedReconstruction{}, err
	}
	for _, step := range argument.Steps {
		if step.Kind == "conclusion" {
			conclusionCount++
		}
	}
	result := SalvagedReconstruction{AssumedLines: []int{}, FailingSteps: []int{}}
	built := NumberedArgument{}
	byNumber := map[int]NumberedArgumentStep{}
	renumber := map[int]int{}
	for _, step := range argument.Steps {
		old := step.Number
		if removed[old] {
			continue
		}
		step.Number = len(built.Steps) + 1
		refs := []int{}
		for _, ref := range step.Citations {
			if !removed[ref] {
				refs = append(refs, renumber[ref])
			}
		}
		step.Citations = refs
		if step.Kind == "conclusion" && !gateConclusionStep(step, byNumber).Passed {
			// One bridge is always allowed: a short argument whose only inference
			// fails is shown as "doesn't follow" on its faithful reconstruction.
			bridges := len(result.AssumedLines) + 1
			if bridges > 3 || (bridges > 1 && bridges*2 > conclusionCount) {
				return SalvagedReconstruction{}, fmt.Errorf("too many deterministic bridges: maximum three, and more than one only up to half of conclusion steps")
			}
			if len(refs) == 0 {
				return SalvagedReconstruction{}, fmt.Errorf("step %d has no remaining antecedents", old)
			}
			antecedents := make([]string, 0, len(refs))
			for _, ref := range refs {
				// The gate compares clauses case-insensitively; lower-case each
				// conjunct's ordinary leading word so the bridge reads as prose.
				antecedents = append(antecedents, lowercaseLeadingOrdinaryWord(byNumber[ref].Statement, original.Argument))
			}
			bridge, _ := FallbackBridge(strings.Join(antecedents, " and "), step.Statement, original.Argument)
			assumption := NumberedArgumentStep{Number: step.Number, Kind: "premise", Statement: bridge}
			built.Steps = append(built.Steps, assumption)
			byNumber[assumption.Number] = assumption
			result.Reconstruction.Lines = append(result.Reconstruction.Lines, ReconstructionLine{Line: assumption.Number, Sentences: []int{}, Assumed: true})
			result.AssumedLines = append(result.AssumedLines, assumption.Number)
			result.FailingSteps = append(result.FailingSteps, old)
			step.Citations = append(step.Citations, assumption.Number)
			step.Number++
		}
		renumber[old] = step.Number
		built.Steps = append(built.Steps, step)
		byNumber[step.Number] = step
		line := original.Lines[old-1]
		line.Line = step.Number
		result.Reconstruction.Lines = append(result.Reconstruction.Lines, line)
	}
	lines := make([]string, 0, len(built.Steps))
	for _, step := range built.Steps {
		if step.Kind == "premise" {
			lines = append(lines, fmt.Sprintf("%d. %s.", step.Number, step.Statement))
			continue
		}
		refs := make([]string, len(step.Citations))
		for i, ref := range step.Citations {
			refs[i] = strconv.Itoa(ref)
		}
		lines = append(lines, fmt.Sprintf("%d. Therefore, %s. [%s]", step.Number, step.Statement, strings.Join(refs, ",")))
	}
	result.Reconstruction.Argument = strings.Join(lines, "\n")
	if report, detail := CheckEntailmentGate(result.Reconstruction.Argument); !report.Passed {
		return SalvagedReconstruction{}, fmt.Errorf("salvaged reconstruction failed gate: %s", detail.Message)
	}
	return result, nil
}
