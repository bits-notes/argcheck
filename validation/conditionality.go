package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type conditionalityPreservationOptions struct {
	Conditionals []sourceConditional `json:"conditionals"`
}

type sourceConditional struct {
	Antecedent string `json:"antecedent"`
	Consequent string `json:"consequent"`
}

type conditionalityPreservationValidator struct {
	options conditionalityPreservationOptions
}

func parseConditionalityPreservationOptions(raw json.RawMessage) (conditionalityPreservationOptions, error) {
	if len(raw) == 0 {
		return conditionalityPreservationOptions{}, fmt.Errorf("conditionality_preservation requires conditionals options")
	}
	var options conditionalityPreservationOptions
	if err := json.Unmarshal(raw, &options); err != nil {
		return conditionalityPreservationOptions{}, fmt.Errorf("invalid conditionality_preservation options: %w", err)
	}
	if len(options.Conditionals) == 0 {
		return conditionalityPreservationOptions{}, fmt.Errorf("conditionality_preservation requires at least one conditional")
	}
	for i, conditional := range options.Conditionals {
		if strings.TrimSpace(conditional.Antecedent) == "" {
			return conditionalityPreservationOptions{}, fmt.Errorf("conditional %d antecedent is required", i)
		}
		if strings.TrimSpace(conditional.Consequent) == "" {
			return conditionalityPreservationOptions{}, fmt.Errorf("conditional %d consequent is required", i)
		}
	}
	return options, nil
}

func (v conditionalityPreservationValidator) Validate(ctx context.Context, output string) Result {
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

	normalizedOutput := normalizeConditionalityText(output)
	drifts := make([]map[string]any, 0)
	for _, conditional := range v.options.Conditionals {
		antecedent := normalizeConditionalityText(conditional.Antecedent)
		consequent := normalizeConditionalityText(conditional.Consequent)
		if consequent == "" || !strings.Contains(normalizedOutput, consequent) {
			continue
		}
		claims := conditionalityClaims(output)
		if antecedent != "" && (hasPreservedConditional(claims, antecedent, consequent) || hasEstablishedAntecedent(claims, antecedent, consequent)) {
			continue
		}
		drifts = append(drifts, map[string]any{
			"antecedent": conditional.Antecedent,
			"consequent": conditional.Consequent,
		})
	}
	if len(drifts) > 0 {
		return Result{
			Passed:  false,
			Message: "conditionality drift detected: candidate uses a source-supported consequent without preserving or establishing its antecedent",
			Details: map[string]any{
				"drifted_conditionals": drifts,
			},
		}
	}
	return Result{
		Passed:  true,
		Message: "output preserves configured source conditionals",
		Details: map[string]any{
			"conditional_count": len(v.options.Conditionals),
		},
	}
}

func normalizeConditionalityText(value string) string {
	value = strings.ToLower(value)
	replacer := strings.NewReplacer(
		"\r\n", " ",
		"\r", " ",
		"\n", " ",
		".", "",
		",", "",
		";", "",
		":", "",
		"(", "",
		")", "",
		"[", "",
		"]", "",
	)
	value = replacer.Replace(value)
	return strings.Join(strings.Fields(value), " ")
}

func conditionalityClaims(output string) []string {
	normalized := strings.ReplaceAll(output, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	fields := strings.FieldsFunc(normalized, func(r rune) bool {
		return r == '\n' || r == '.'
	})
	claims := make([]string, 0, len(fields))
	for _, field := range fields {
		claim := normalizeConditionalityText(field)
		if claim != "" {
			claims = append(claims, claim)
		}
	}
	return claims
}

func hasPreservedConditional(claims []string, antecedent string, consequent string) bool {
	for _, claim := range claims {
		if !strings.Contains(claim, antecedent) || !strings.Contains(claim, consequent) {
			continue
		}
		if hasNegatedAntecedentContext(claim, antecedent) {
			continue
		}
		if strings.Contains(claim, " if ") ||
			strings.HasPrefix(claim, "if ") ||
			strings.Contains(claim, " under ") ||
			strings.Contains(claim, " assuming ") ||
			strings.Contains(claim, " provided ") ||
			strings.Contains(claim, " when ") {
			return true
		}
	}
	return false
}

func hasEstablishedAntecedent(claims []string, antecedent string, consequent string) bool {
	for _, claim := range claims {
		if !strings.Contains(claim, antecedent) || strings.Contains(claim, consequent) {
			continue
		}
		if isNegatedOrHedgedAntecedentMention(claim, antecedent) {
			continue
		}
		return true
	}
	return false
}

func hasNegatedAntecedentContext(claim string, antecedent string) bool {
	index := strings.Index(claim, antecedent)
	if index < 0 {
		return true
	}
	prefix := strings.TrimSpace(claim[:index])
	blockedPhrases := []string{
		"not",
		"not the case that",
		"false that",
		"no evidence that",
		"possible that",
		"it is possible that",
		"may be",
		"might be",
		"could be",
		"possibly",
		"whether",
		"unclear whether",
	}
	if hasBlockedAntecedentPrefix(prefix, blockedPhrases) {
		return true
	}
	return hasBlockedAntecedentSuffix(claim[index+len(antecedent):])
}

func isNegatedOrHedgedAntecedentMention(claim string, antecedent string) bool {
	index := strings.Index(claim, antecedent)
	if index < 0 {
		return true
	}
	prefix := strings.TrimSpace(claim[:index])
	blockedPhrases := []string{
		"not",
		"not the case that",
		"false that",
		"no evidence that",
		"whether",
		"if",
		"assuming",
		"under",
		"provided",
		"may be",
		"might be",
		"could be",
		"possibly",
		"does not establish that",
		"fails to establish that",
	}
	if hasBlockedAntecedentPrefix(prefix, blockedPhrases) {
		return true
	}
	return hasBlockedAntecedentSuffix(claim[index+len(antecedent):])
}

func hasBlockedAntecedentPrefix(prefix string, blockedPhrases []string) bool {
	if prefix == "" {
		return false
	}
	for _, phrase := range blockedPhrases {
		if strings.HasSuffix(prefix, phrase) || strings.Contains(prefix, phrase+" ") {
			return true
		}
	}
	return false
}

func hasBlockedAntecedentSuffix(suffix string) bool {
	suffix = strings.TrimSpace(suffix)
	blockedPhrases := []string{
		"is false",
		"are false",
		"is not true",
		"are not true",
		"does not hold",
		"do not hold",
		"is unresolved",
		"are unresolved",
		"is not established",
		"are not established",
		"is unsupported",
		"are unsupported",
	}
	for _, phrase := range blockedPhrases {
		if strings.HasPrefix(suffix, phrase) || strings.Contains(suffix, " "+phrase) {
			return true
		}
	}
	return false
}
