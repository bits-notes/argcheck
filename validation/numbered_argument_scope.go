package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var arbitraryNamePattern = regexp.MustCompile(`^Let ([A-Za-z][A-Za-z0-9_]*) be an arbitrary (.+)$`)

func (arg NumberedArgument) HasSubproofs() bool {
	for _, s := range arg.Steps {
		if s.Depth != 0 || len(s.Ranges) > 0 || s.Kind == numberedArgumentStepAssumption || s.Kind == numberedArgumentStepArbitrary {
			return true
		}
	}
	return false
}

// ParseFlatNumberedArgumentAllowingNoConclusions is for legacy transformations
// that renumber/reconstruct lines without carrying scope or discharge edges.
func ParseFlatNumberedArgumentAllowingNoConclusions(output string) (NumberedArgument, Result) {
	return ParseNumberedArgumentAllowingNoConclusions(output)
}

func parseTerminalCitations(raw string) ([]int, [][2]int, error) {
	var singles []int
	var ranges [][2]int
	for _, part := range strings.Split(raw, ",") {
		bounds := strings.Split(part, "-")
		start, err := strconv.Atoi(bounds[0])
		if err != nil {
			return nil, nil, err
		}
		if len(bounds) == 1 {
			singles = append(singles, start)
			continue
		}
		end, err := strconv.Atoi(bounds[1])
		if err != nil {
			return nil, nil, err
		}
		if start > end {
			return nil, nil, fmt.Errorf("range %d-%d is reversed", start, end)
		}
		ranges = append(ranges, [2]int{start, end})
	}
	return singles, ranges, nil
}

type numberedSubproof struct{ start, end, depth, parent int }

// Scope identity, not depth alone, determines accessibility: sibling blocks
// have the same depth but cannot see each other's lines.
func checkNumberedArgumentScope(arg NumberedArgument) Result {
	if !arg.HasSubproofs() {
		return Result{Passed: true}
	} // retain the legacy flat contract
	fail := func(format string, args ...any) Result {
		return Result{Passed: false, Message: fmt.Sprintf(format, args...)}
	}
	blocks := map[int]numberedSubproof{}
	owners := map[int]int{}
	active := []int{0}
	for index, s := range arg.Steps {
		if s.Number != index+1 {
			return fail("step %d is not sequential (expected %d)", s.Number, index+1)
		}
		assumption := s.Kind == numberedArgumentStepAssumption || s.Kind == numberedArgumentStepArbitrary
		if s.Depth < 0 || s.Depth > len(active) {
			return fail("step %d increases depth by more than one", s.Number)
		}
		// A new assumption at the same depth starts a sibling subproof.
		for len(active)-1 > s.Depth || (assumption && s.Depth > 0 && len(active)-1 == s.Depth) {
			id := active[len(active)-1]
			b := blocks[id]
			b.end = s.Number - 1
			blocks[id] = b
			active = active[:len(active)-1]
		}
		if assumption {
			if s.Depth == 0 || s.Depth != len(active) {
				return fail("step %d assumption must open a subproof one depth below its parent", s.Number)
			}
			if len(s.Citations)+len(s.Ranges) > 0 {
				return fail("step %d assumption cannot have citations", s.Number)
			}
			blocks[s.Number] = numberedSubproof{start: s.Number, depth: s.Depth, parent: active[len(active)-1]}
			active = append(active, s.Number)
		} else if s.Depth != len(active)-1 {
			return fail("step %d increases depth without an assumption", s.Number)
		}
		if s.Kind == numberedArgumentStepPremise {
			if s.Depth != 0 {
				return fail("step %d premise must be at depth 0", s.Number)
			}
		}
		owners[s.Number] = active[len(active)-1]
	}
	if len(active) > 1 {
		return fail("final step %d must be at depth 0; subproof %d is still open", len(arg.Steps), active[len(active)-1])
	}
	for _, s := range arg.Steps {
		citations := append([]int(nil), s.Citations...)
		for _, m := range bracketReferencePattern.FindAllStringSubmatch(s.Statement, -1) {
			refs, _ := parseCitations(m[1])
			citations = append(citations, refs...)
		}
		for _, c := range citations {
			owner, exists := owners[c]
			if !exists || c >= s.Number {
				return fail("step %d cites unavailable prior step %d", s.Number, c)
			}
			visible := owner == 0
			for id := owners[s.Number]; id != 0; id = blocks[id].parent {
				if id == owner {
					visible = true
					break
				}
			}
			if !visible {
				b := blocks[owner]
				return fail("step %d cites %d, which is inside the closed subproof %d-%d", s.Number, c, b.start, b.end)
			}
		}
		for _, r := range s.Ranges {
			b, ok := blocks[r[0]]
			if !ok || b.end != r[1] || b.end >= s.Number || b.depth != s.Depth+1 || b.parent != owners[s.Number] {
				return fail("step %d range %d-%d is not a complete closed subproof immediately subordinate to depth %d", s.Number, r[0], r[1], s.Depth)
			}
			if arg.Steps[r[1]-1].Depth != b.depth {
				return fail("step %d range %d-%d ends inside a nested subproof", s.Number, r[0], r[1])
			}
		}
	}
	return Result{Passed: true}
}
