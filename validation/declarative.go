package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	maxRegexPatternBytes     = 4 * 1024
	maxLineRulePatterns      = 20
	maxLineRuleRules         = 50
	maxLineRuleLines         = 2000
	lineRuleAllLinesMatchOne = "all_lines_match_one_pattern"
	lineRuleSequentialNumber = "sequential_numbering"
	lineRuleFinalLinePattern = "final_line_pattern"
)

type regexMatchSpec struct {
	Pattern string `json:"pattern"`
}

type regexMatchValidator struct {
	pattern *regexp.Regexp
}

func compileRegexMatch(spec Spec) (OutputValidator, error) {
	if spec.Name != "" {
		return nil, fmt.Errorf("regex_match validator does not support name")
	}
	if len(spec.Options) > 0 {
		return nil, fmt.Errorf("regex_match validator does not support options")
	}
	var cfg regexMatchSpec
	if err := decodeStrictSpec(spec.Spec, &cfg, "regex_match spec"); err != nil {
		return nil, err
	}
	if cfg.Pattern == "" {
		return nil, fmt.Errorf("regex_match pattern is required")
	}
	if len(cfg.Pattern) > maxRegexPatternBytes {
		return nil, fmt.Errorf("regex_match pattern exceeds %d bytes", maxRegexPatternBytes)
	}
	compiled, err := regexp.Compile(cfg.Pattern)
	if err != nil {
		return nil, fmt.Errorf("regex_match pattern is invalid: %w", err)
	}
	return regexMatchValidator{pattern: compiled}, nil
}

func (v regexMatchValidator) Validate(ctx context.Context, output string) Result {
	ctx, cancel := withValidationContext(ctx)
	defer cancel()
	if limit := checkOutputLimit(output); !limit.Passed {
		return limit
	}
	select {
	case <-ctx.Done():
		return Result{Passed: false, Message: "validation timed out"}
	default:
	}
	match := v.pattern.FindStringIndex(output)
	if len(match) != 2 || match[0] != 0 || match[1] != len(output) {
		return Result{Passed: false, Message: "output does not match regex_match pattern"}
	}
	return Result{Passed: true, Message: "output matches regex_match"}
}

type lineRulesSpec struct {
	AllowBlankLines bool                `json:"allow_blank_lines"`
	Patterns        map[string]string   `json:"patterns"`
	Rules           []lineRulesRuleSpec `json:"rules"`
}

type lineRulesRuleSpec struct {
	Type     string   `json:"type"`
	Pattern  string   `json:"pattern,omitempty"`
	Patterns []string `json:"patterns,omitempty"`
	Capture  int      `json:"capture,omitempty"`
}

type lineRulesValidator struct {
	allowBlankLines bool
	patterns        []compiledLinePattern
	rules           []compiledLineRule
}

type compiledLinePattern struct {
	name    string
	pattern *regexp.Regexp
}

type compiledLineRule struct {
	typ      string
	pattern  string
	patterns []string
	capture  int
}

type lineMatch struct {
	patternName string
	captures    []string
}

type lineMatchSet struct {
	lineNumber int
	text       string
	matches    []lineMatch
}

func compileLineRules(spec Spec) (OutputValidator, error) {
	if spec.Name != "" {
		return nil, fmt.Errorf("line_rules validator does not support name")
	}
	if len(spec.Options) > 0 {
		return nil, fmt.Errorf("line_rules validator does not support options")
	}
	var cfg lineRulesSpec
	if err := decodeStrictSpec(spec.Spec, &cfg, "line_rules spec"); err != nil {
		return nil, err
	}
	if len(cfg.Patterns) == 0 {
		return nil, fmt.Errorf("line_rules patterns are required")
	}
	if len(cfg.Patterns) > maxLineRulePatterns {
		return nil, fmt.Errorf("line_rules patterns exceed %d entries", maxLineRulePatterns)
	}
	if len(cfg.Rules) == 0 {
		return nil, fmt.Errorf("line_rules rules are required")
	}
	if len(cfg.Rules) > maxLineRuleRules {
		return nil, fmt.Errorf("line_rules rules exceed %d entries", maxLineRuleRules)
	}

	patternNames := make([]string, 0, len(cfg.Patterns))
	for name := range cfg.Patterns {
		patternNames = append(patternNames, name)
	}
	sort.Strings(patternNames)

	patterns := make([]compiledLinePattern, 0, len(patternNames))
	knownPatterns := make(map[string]bool, len(patternNames))
	for _, name := range patternNames {
		trimmedName := strings.TrimSpace(name)
		if trimmedName == "" || trimmedName != name {
			return nil, fmt.Errorf("line_rules pattern names cannot be empty or padded")
		}
		pattern := cfg.Patterns[name]
		if pattern == "" {
			return nil, fmt.Errorf("line_rules pattern %q is empty", name)
		}
		if len(pattern) > maxRegexPatternBytes {
			return nil, fmt.Errorf("line_rules pattern %q exceeds %d bytes", name, maxRegexPatternBytes)
		}
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("line_rules pattern %q is invalid: %w", name, err)
		}
		knownPatterns[name] = true
		patterns = append(patterns, compiledLinePattern{name: name, pattern: compiled})
	}

	rules := make([]compiledLineRule, 0, len(cfg.Rules))
	for index, rule := range cfg.Rules {
		compiled, err := compileLineRule(rule, index+1, knownPatterns, patterns)
		if err != nil {
			return nil, err
		}
		rules = append(rules, compiled)
	}

	return lineRulesValidator{
		allowBlankLines: cfg.AllowBlankLines,
		patterns:        patterns,
		rules:           rules,
	}, nil
}

func compileLineRule(rule lineRulesRuleSpec, ruleNumber int, knownPatterns map[string]bool, patterns []compiledLinePattern) (compiledLineRule, error) {
	if strings.TrimSpace(rule.Type) == "" {
		return compiledLineRule{}, fmt.Errorf("line_rules rule %d type is required", ruleNumber)
	}
	switch rule.Type {
	case lineRuleAllLinesMatchOne:
		if rule.Pattern != "" || len(rule.Patterns) > 0 || rule.Capture != 0 {
			return compiledLineRule{}, fmt.Errorf("line_rules rule %d %q does not support pattern, patterns, or capture", ruleNumber, rule.Type)
		}
		return compiledLineRule{typ: rule.Type}, nil
	case lineRuleSequentialNumber:
		if rule.Capture <= 0 {
			return compiledLineRule{}, fmt.Errorf("line_rules rule %d sequential_numbering requires positive capture", ruleNumber)
		}
		selected, err := validateLineRulePatterns(rule, ruleNumber, knownPatterns, true)
		if err != nil {
			return compiledLineRule{}, err
		}
		for _, pattern := range patterns {
			if len(selected) > 0 && !containsString(selected, pattern.name) {
				continue
			}
			if pattern.pattern.NumSubexp() < rule.Capture {
				return compiledLineRule{}, fmt.Errorf("line_rules rule %d pattern %q does not have capture %d", ruleNumber, pattern.name, rule.Capture)
			}
		}
		return compiledLineRule{typ: rule.Type, patterns: selected, capture: rule.Capture}, nil
	case lineRuleFinalLinePattern:
		selected, err := validateLineRulePatterns(rule, ruleNumber, knownPatterns, false)
		if err != nil {
			return compiledLineRule{}, err
		}
		if len(selected) != 1 {
			return compiledLineRule{}, fmt.Errorf("line_rules rule %d final_line_pattern requires exactly one pattern", ruleNumber)
		}
		if rule.Capture != 0 {
			return compiledLineRule{}, fmt.Errorf("line_rules rule %d final_line_pattern does not support capture", ruleNumber)
		}
		return compiledLineRule{typ: rule.Type, pattern: selected[0]}, nil
	default:
		return compiledLineRule{}, fmt.Errorf("line_rules rule %d has unknown type %q", ruleNumber, rule.Type)
	}
}

func validateLineRulePatterns(rule lineRulesRuleSpec, ruleNumber int, knownPatterns map[string]bool, allowEmpty bool) ([]string, error) {
	if rule.Pattern != "" && len(rule.Patterns) > 0 {
		return nil, fmt.Errorf("line_rules rule %d cannot set both pattern and patterns", ruleNumber)
	}
	selected := make([]string, 0, len(rule.Patterns)+1)
	if rule.Pattern != "" {
		selected = append(selected, rule.Pattern)
	}
	selected = append(selected, rule.Patterns...)
	if len(selected) == 0 {
		if allowEmpty {
			return nil, nil
		}
		return nil, fmt.Errorf("line_rules rule %d pattern is required", ruleNumber)
	}
	seen := map[string]bool{}
	for _, name := range selected {
		if !knownPatterns[name] {
			return nil, fmt.Errorf("line_rules rule %d references unknown pattern %q", ruleNumber, name)
		}
		if seen[name] {
			return nil, fmt.Errorf("line_rules rule %d repeats pattern %q", ruleNumber, name)
		}
		seen[name] = true
	}
	sort.Strings(selected)
	return selected, nil
}

func (v lineRulesValidator) Validate(ctx context.Context, output string) Result {
	ctx, cancel := withValidationContext(ctx)
	defer cancel()
	if limit := checkOutputLimit(output); !limit.Passed {
		return limit
	}
	select {
	case <-ctx.Done():
		return Result{Passed: false, Message: "validation timed out"}
	default:
	}

	lineSets, result := v.matchLines(output)
	if !result.Passed {
		return result
	}
	for _, rule := range v.rules {
		if result := v.evaluateRule(rule, lineSets); !result.Passed {
			return result
		}
	}
	return Result{
		Passed:  true,
		Message: "output matches line_rules",
		Details: map[string]any{
			"line_count": len(lineSets),
		},
	}
}

func (v lineRulesValidator) matchLines(output string) ([]lineMatchSet, Result) {
	normalized := strings.ReplaceAll(output, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	normalized = strings.TrimSpace(normalized)
	if normalized == "" {
		return nil, Result{Passed: false, Message: "output is empty"}
	}

	lines := strings.Split(normalized, "\n")
	if len(lines) > maxLineRuleLines {
		return nil, Result{
			Passed:  false,
			Message: fmt.Sprintf("line_rules output exceeds %d line validation limit", maxLineRuleLines),
		}
	}

	lineSets := make([]lineMatchSet, 0, len(lines))
	for index, line := range lines {
		lineNumber := index + 1
		if strings.TrimSpace(line) == "" {
			if v.allowBlankLines {
				continue
			}
			return nil, Result{
				Passed:  false,
				Message: fmt.Sprintf("line %d is blank", lineNumber),
				Details: map[string]any{"line": lineNumber},
			}
		}
		matches := make([]lineMatch, 0, len(v.patterns))
		for _, pattern := range v.patterns {
			if captures, ok := fullLineSubmatches(pattern.pattern, line); ok {
				matches = append(matches, lineMatch{patternName: pattern.name, captures: captures})
			}
		}
		lineSets = append(lineSets, lineMatchSet{
			lineNumber: lineNumber,
			text:       line,
			matches:    matches,
		})
	}
	if len(lineSets) == 0 {
		return nil, Result{Passed: false, Message: "output contains no non-blank lines"}
	}
	return lineSets, Result{Passed: true}
}

func (v lineRulesValidator) evaluateRule(rule compiledLineRule, lines []lineMatchSet) Result {
	switch rule.typ {
	case lineRuleAllLinesMatchOne:
		for _, line := range lines {
			if len(line.matches) != 1 {
				return Result{
					Passed: false,
					Message: fmt.Sprintf(
						"line %d matches %d patterns; expected exactly one",
						line.lineNumber,
						len(line.matches),
					),
					Details: map[string]any{"line": line.lineNumber, "match_count": len(line.matches)},
				}
			}
		}
	case lineRuleSequentialNumber:
		expected := 1
		for _, line := range lines {
			match, ok := firstAllowedLineMatch(line.matches, rule.patterns)
			if !ok {
				return Result{
					Passed:  false,
					Message: fmt.Sprintf("line %d does not match a numbered pattern", line.lineNumber),
					Details: map[string]any{"line": line.lineNumber},
				}
			}
			value, err := strconv.Atoi(match.captures[rule.capture])
			if err != nil {
				return Result{
					Passed:  false,
					Message: fmt.Sprintf("line %d capture %d is not an integer", line.lineNumber, rule.capture),
					Details: map[string]any{"line": line.lineNumber, "capture": rule.capture},
				}
			}
			if value != expected {
				return Result{
					Passed:  false,
					Message: fmt.Sprintf("line_rules numbering jumps from %d to %d", expected-1, value),
					Details: map[string]any{"line": line.lineNumber, "expected": expected, "actual": value},
				}
			}
			expected++
		}
	case lineRuleFinalLinePattern:
		final := lines[len(lines)-1]
		if _, ok := firstAllowedLineMatch(final.matches, []string{rule.pattern}); !ok {
			return Result{
				Passed:  false,
				Message: fmt.Sprintf("final line must match pattern %q", rule.pattern),
				Details: map[string]any{"line": final.lineNumber, "pattern": rule.pattern},
			}
		}
	}
	return Result{Passed: true}
}

func fullLineSubmatches(pattern *regexp.Regexp, line string) ([]string, bool) {
	indices := pattern.FindStringSubmatchIndex(line)
	if len(indices) < 2 || indices[0] != 0 || indices[1] != len(line) {
		return nil, false
	}
	return pattern.FindStringSubmatch(line), true
}

func firstAllowedLineMatch(matches []lineMatch, allowed []string) (lineMatch, bool) {
	for _, match := range matches {
		if len(allowed) == 0 || containsString(allowed, match.patternName) {
			return match, true
		}
	}
	return lineMatch{}, false
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func decodeStrictSpec(raw json.RawMessage, target any, label string) error {
	if len(raw) == 0 {
		return fmt.Errorf("%s is required", label)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("invalid %s: %w", label, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return fmt.Errorf("invalid %s: must contain exactly one JSON object", label)
	}
	return nil
}
