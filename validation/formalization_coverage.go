package validation

import (
	"fmt"
	"strings"
)

const (
	DroppedContentKindPremise    = "premise"
	DroppedContentKindConclusion = "conclusion"
	DroppedContentKindCitation   = "citation"
)

type FormalizationLineMapping struct {
	RepairedLine     int   `json:"repaired_line"`
	FromAsFoundLines []int `json:"from_as_found_lines"`
}

type FormalizationDroppedContent struct {
	AsFoundLine int    `json:"as_found_line"`
	Kind        string `json:"kind"`
	Text        string `json:"text"`
	Reason      string `json:"reason"`
}

// ValidateFormalizationCoverage proves that every as-found line is accounted
// for exactly once and every repaired line is either source-derived or visibly
// supplied. It also replaces dropped-content text with the parsed as-found
// statement so persisted flags never rely on model-authored copies.
func ValidateFormalizationCoverage(
	asFoundArgument string,
	repairedArgument string,
	lineMapping []FormalizationLineMapping,
	dropped []FormalizationDroppedContent,
	suppliedPremiseLines []int,
) ([]FormalizationDroppedContent, bool, error) {
	asFound, parsed := ParseNumberedArgumentAllowingNoConclusions(asFoundArgument)
	if !parsed.Passed {
		return nil, false, fmt.Errorf("parse as-found argument: %s", parsed.Message)
	}
	repaired, parsed := ParseNumberedArgumentAllowingNoConclusions(repairedArgument)
	if !parsed.Passed {
		return nil, false, fmt.Errorf("parse repaired argument: %s", parsed.Message)
	}

	asFoundCoverage := make([]int, len(asFound.Steps)+1)
	repairedMapped := make([]bool, len(repaired.Steps)+1)
	seenRepairedMapping := make(map[int]bool, len(lineMapping))
	for index, mapping := range lineMapping {
		if mapping.RepairedLine < 1 || mapping.RepairedLine > len(repaired.Steps) {
			return nil, false, fmt.Errorf("line_mapping[%d] references repaired line %d outside 1..%d", index, mapping.RepairedLine, len(repaired.Steps))
		}
		if seenRepairedMapping[mapping.RepairedLine] {
			return nil, false, fmt.Errorf("repaired line %d has duplicate line_mapping entries", mapping.RepairedLine)
		}
		seenRepairedMapping[mapping.RepairedLine] = true
		if len(mapping.FromAsFoundLines) == 0 {
			return nil, false, fmt.Errorf("repaired line %d maps from no as-found lines", mapping.RepairedLine)
		}
		seenSources := make(map[int]bool, len(mapping.FromAsFoundLines))
		for _, sourceLine := range mapping.FromAsFoundLines {
			if sourceLine < 1 || sourceLine > len(asFound.Steps) {
				return nil, false, fmt.Errorf("repaired line %d references as-found line %d outside 1..%d", mapping.RepairedLine, sourceLine, len(asFound.Steps))
			}
			if seenSources[sourceLine] {
				return nil, false, fmt.Errorf("repaired line %d repeats as-found line %d", mapping.RepairedLine, sourceLine)
			}
			seenSources[sourceLine] = true
			asFoundCoverage[sourceLine]++
		}
		repairedMapped[mapping.RepairedLine] = true
	}

	enrichedDropped := make([]FormalizationDroppedContent, len(dropped))
	droppedConclusion := false
	for index, flag := range dropped {
		if flag.AsFoundLine < 1 || flag.AsFoundLine > len(asFound.Steps) {
			return nil, false, fmt.Errorf("dropped[%d] references as-found line %d outside 1..%d", index, flag.AsFoundLine, len(asFound.Steps))
		}
		step := asFound.Steps[flag.AsFoundLine-1]
		switch flag.Kind {
		case DroppedContentKindPremise:
			if step.Kind != DroppedContentKindPremise {
				return nil, false, fmt.Errorf("dropped as-found line %d is a %s, not a premise", flag.AsFoundLine, step.Kind)
			}
		case DroppedContentKindConclusion:
			if step.Kind != DroppedContentKindConclusion {
				return nil, false, fmt.Errorf("dropped as-found line %d is a %s, not a conclusion", flag.AsFoundLine, step.Kind)
			}
			droppedConclusion = true
		case DroppedContentKindCitation:
			if len(step.Citations) == 0 {
				return nil, false, fmt.Errorf("dropped as-found line %d has no citation", flag.AsFoundLine)
			}
		default:
			return nil, false, fmt.Errorf("dropped[%d] has unsupported kind %q", index, flag.Kind)
		}
		if strings.TrimSpace(flag.Reason) == "" {
			return nil, false, fmt.Errorf("dropped as-found line %d requires a reason", flag.AsFoundLine)
		}
		asFoundCoverage[flag.AsFoundLine]++
		flag.Text = step.Statement
		enrichedDropped[index] = flag
	}

	for line := 1; line <= len(asFound.Steps); line++ {
		switch asFoundCoverage[line] {
		case 0:
			return nil, false, fmt.Errorf("as-found line %d is not covered by line_mapping or dropped", line)
		case 1:
		default:
			return nil, false, fmt.Errorf("as-found line %d is covered %d times; want exactly once", line, asFoundCoverage[line])
		}
	}

	supplied := make(map[int]bool, len(suppliedPremiseLines))
	for _, line := range suppliedPremiseLines {
		if line < 1 || line > len(repaired.Steps) {
			return nil, false, fmt.Errorf("supplied premise references repaired line %d outside 1..%d", line, len(repaired.Steps))
		}
		if repaired.Steps[line-1].Kind != DroppedContentKindPremise {
			return nil, false, fmt.Errorf("supplied premise line %d is a %s", line, repaired.Steps[line-1].Kind)
		}
		supplied[line] = true
	}
	for line := 1; line <= len(repaired.Steps); line++ {
		if !repairedMapped[line] && !supplied[line] {
			return nil, false, fmt.Errorf("repaired line %d is neither mapped from as-found content nor covered by a supplied-premise flag", line)
		}
	}
	return enrichedDropped, droppedConclusion, nil
}
