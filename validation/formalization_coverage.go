package validation

import (
	"fmt"
	"strings"
)

const (
	DroppedContentKindPremise                  = "premise"
	DroppedContentKindConclusion               = "conclusion"
	DroppedContentKindCitation                 = "citation"
	FormalizationCoverageAnchorSourceSentences = "source_sentences"
	FormalizationCoverageAnchorSource          = "source"
	FormalizationCoverageAnchorAsFound         = "as_found"
)

type FormalizationLineMapping struct {
	RepairedLine        int     `json:"repaired_line"`
	FromAsFoundLines    []int   `json:"from_as_found_lines"`
	FromSourceQuote     *string `json:"from_source_quote,omitempty"`
	FromSourceSentences []int   `json:"from_source_sentences,omitempty"`
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
	return ValidateFormalizationCoverageAnchored(
		FormalizationCoverageAnchorAsFound,
		asFoundArgument,
		repairedArgument,
		lineMapping,
		dropped,
		suppliedPremiseLines,
	)
}

// ValidateFormalizationCoverageAnchored validates the repair declaration
// against either the parsed source argument or, when no numbered source is
// available, the as-found argument. The JSON field names retain their Phase 2
// legacy spelling (from_as_found_lines/as_found_line), but their indices refer
// to the declared anchor. Optional sourceMaterial is the original passage,
// required whenever a mapping declares from_source_quote.
func ValidateFormalizationCoverageAnchored(
	anchorName string,
	anchorArgument string,
	repairedArgument string,
	lineMapping []FormalizationLineMapping,
	dropped []FormalizationDroppedContent,
	suppliedPremiseLines []int,
	sourceMaterial ...string,
) ([]FormalizationDroppedContent, bool, error) {
	if anchorName != FormalizationCoverageAnchorSource && anchorName != FormalizationCoverageAnchorAsFound {
		return nil, false, fmt.Errorf("unsupported coverage anchor %q", anchorName)
	}
	anchor, parsed := ParseNumberedArgumentAllowingNoConclusions(anchorArgument)
	if !parsed.Passed {
		return nil, false, fmt.Errorf("parse %s argument: %s", formalizationCoverageAnchorLabel(anchorName), parsed.Message)
	}
	repaired, parsed := ParseNumberedArgumentAllowingNoConclusions(repairedArgument)
	if !parsed.Passed {
		return nil, false, fmt.Errorf("parse repaired argument: %s", parsed.Message)
	}

	source := ""
	if len(sourceMaterial) > 0 {
		source = sourceMaterial[0]
	}
	if err := ValidateRestoredLineMappings(source, repairedArgument, lineMapping); err != nil {
		return nil, false, err
	}

	anchorCoverage := make([]int, len(anchor.Steps)+1)
	repairedMapped := make([]bool, len(repaired.Steps)+1)
	seenRepairedMapping := make(map[int]bool, len(lineMapping))
	restoredLines := make(map[int]bool)
	for index, mapping := range lineMapping {
		if mapping.RepairedLine < 1 || mapping.RepairedLine > len(repaired.Steps) {
			return nil, false, fmt.Errorf("line_mapping[%d] references repaired line %d outside 1..%d", index, mapping.RepairedLine, len(repaired.Steps))
		}
		if seenRepairedMapping[mapping.RepairedLine] {
			return nil, false, fmt.Errorf("repaired line %d has duplicate line_mapping entries", mapping.RepairedLine)
		}
		seenRepairedMapping[mapping.RepairedLine] = true
		if len(mapping.FromAsFoundLines) == 0 && mapping.FromSourceQuote == nil {
			return nil, false, fmt.Errorf("repaired line %d maps from no as-found lines", mapping.RepairedLine)
		}
		seenSources := make(map[int]bool, len(mapping.FromAsFoundLines))
		for _, sourceLine := range mapping.FromAsFoundLines {
			if sourceLine < 1 || sourceLine > len(anchor.Steps) {
				return nil, false, fmt.Errorf("repaired line %d references %s line %d outside 1..%d", mapping.RepairedLine, formalizationCoverageAnchorLabel(anchorName), sourceLine, len(anchor.Steps))
			}
			if seenSources[sourceLine] {
				return nil, false, fmt.Errorf("repaired line %d repeats %s line %d", mapping.RepairedLine, formalizationCoverageAnchorLabel(anchorName), sourceLine)
			}
			seenSources[sourceLine] = true
			anchorCoverage[sourceLine]++
		}
		repairedMapped[mapping.RepairedLine] = true
		restoredLines[mapping.RepairedLine] = mapping.FromSourceQuote != nil
	}

	enrichedDropped := make([]FormalizationDroppedContent, len(dropped))
	droppedConclusion := false
	for index, flag := range dropped {
		if flag.AsFoundLine < 1 || flag.AsFoundLine > len(anchor.Steps) {
			return nil, false, fmt.Errorf("dropped[%d] references %s line %d outside 1..%d", index, formalizationCoverageAnchorLabel(anchorName), flag.AsFoundLine, len(anchor.Steps))
		}
		step := anchor.Steps[flag.AsFoundLine-1]
		switch flag.Kind {
		case DroppedContentKindPremise:
			if step.Kind != DroppedContentKindPremise {
				return nil, false, fmt.Errorf("dropped %s line %d is a %s, not a premise", formalizationCoverageAnchorLabel(anchorName), flag.AsFoundLine, step.Kind)
			}
		case DroppedContentKindConclusion:
			if step.Kind != DroppedContentKindConclusion {
				return nil, false, fmt.Errorf("dropped %s line %d is a %s, not a conclusion", formalizationCoverageAnchorLabel(anchorName), flag.AsFoundLine, step.Kind)
			}
			droppedConclusion = true
		case DroppedContentKindCitation:
			if len(step.Citations) == 0 {
				return nil, false, fmt.Errorf("dropped %s line %d has no citation", formalizationCoverageAnchorLabel(anchorName), flag.AsFoundLine)
			}
		default:
			return nil, false, fmt.Errorf("dropped[%d] has unsupported kind %q", index, flag.Kind)
		}
		if strings.TrimSpace(flag.Reason) == "" {
			return nil, false, fmt.Errorf("dropped %s line %d requires a reason", formalizationCoverageAnchorLabel(anchorName), flag.AsFoundLine)
		}
		anchorCoverage[flag.AsFoundLine]++
		flag.Text = step.Statement
		enrichedDropped[index] = flag
	}

	for line := 1; line <= len(anchor.Steps); line++ {
		switch anchorCoverage[line] {
		case 0:
			return nil, false, fmt.Errorf("%s line %d is not covered by line_mapping or dropped", formalizationCoverageAnchorLabel(anchorName), line)
		case 1:
		default:
			// A compound source premise may legitimately be split into several
			// repaired premises. Conclusions must remain single-origin, and the
			// as-found fallback retains Phase 2's exact-once rule.
			if anchorName == FormalizationCoverageAnchorAsFound || anchor.Steps[line-1].Kind == DroppedContentKindConclusion {
				return nil, false, fmt.Errorf("%s line %d is covered %d times; want exactly once", formalizationCoverageAnchorLabel(anchorName), line, anchorCoverage[line])
			}
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
		if restoredLines[line] {
			return nil, false, fmt.Errorf("restored repaired line %d must not carry a supplied-premise flag", line)
		}
		supplied[line] = true
	}
	coveredRepaired := make([]bool, len(repaired.Steps)+1)
	for line := 1; line <= len(repaired.Steps); line++ {
		coveredRepaired[line] = repairedMapped[line] || supplied[line]
		if !coveredRepaired[line] && repaired.Steps[line-1].Kind == DroppedContentKindConclusion {
			derived := len(repaired.Steps[line-1].Citations) > 0
			for _, citation := range repaired.Steps[line-1].Citations {
				if citation < 1 || citation >= line || !coveredRepaired[citation] {
					derived = false
					break
				}
			}
			coveredRepaired[line] = derived
		}
		if !coveredRepaired[line] {
			return nil, false, fmt.Errorf("repaired line %d is neither mapped from %s content, covered by a supplied-premise flag, nor recursively derived", line, formalizationCoverageAnchorLabel(anchorName))
		}
	}
	return enrichedDropped, droppedConclusion, nil
}

func formalizationCoverageAnchorLabel(anchorName string) string {
	if anchorName == FormalizationCoverageAnchorSource {
		return "source"
	}
	return "as-found"
}
