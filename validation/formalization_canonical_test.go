package validation

import (
	"encoding/json"
	"strings"
	"testing"
)

const canonicalTestOriginal = "1. When interest rates rise, mortgages become more expensive.\n2. The central bank has just raised interest rates.\n3. Therefore, house prices will fall. [1,2]"
const canonicalTestOutput = `{"lines":[{"line":1,"canonical":"If interest rates rise, then mortgages become more expensive."},{"line":2,"canonical":"The central bank has just raised interest rates."},{"line":3,"canonical":"Therefore, house prices will fall. [1,2]"}]}`
const canonicalTestJudge = `{"pairs":[{"line":1,"same":true,"reason":"Conditional equivalence"},{"line":2,"same":true,"reason":"Unchanged"},{"line":3,"same":true,"reason":"Unchanged"}]}`

func TestCanonicalOutputContracts(t *testing.T) {
	prompt := "<as_found_argument>" + canonicalTestOriginal + "</as_found_argument>"
	for _, tc := range []struct {
		name, output string
		pass         bool
	}{
		{"valid", canonicalTestOutput, true},
		{"fenced", "```json\n" + canonicalTestOutput + "\n```", true},
		{"preamble", "Here is the result: " + canonicalTestOutput, true},
		{"bad JSON", "{", false},
		{"count", `{"lines":[{"line":1,"canonical":"A."}]}`, false},
		{"order", strings.Replace(canonicalTestOutput, `"line":2`, `"line":1`, 1), false},
		{"citations", strings.Replace(canonicalTestOutput, "[1,2]", "[1]", 1), false},
		{"marker", strings.Replace(canonicalTestOutput, "Therefore, ", "", 1), false},
		{"premise citations", strings.Replace(canonicalTestOutput, "expensive.", "expensive. [1]", 1), false},
		{"multiple statements", strings.Replace(canonicalTestOutput, "expensive.", "expensive. Cats purr.", 1), false},
		{"missing", `{}`, false},
		{"extra field", strings.Replace(canonicalTestOutput, `"line":1`, `"line":1,"extra":true`, 1), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := ValidateCanonicalOutputWithPrompt(t.Context(), BuiltinFormalizationCanonicalizeOutput, tc.output, prompt)
			if r.Passed != tc.pass {
				t.Fatalf("%+v", r)
			}
		})
	}
	pairs := "<pairs>1 | A. | A.\n2 | B. | B.\n3 | C. | C.</pairs>"
	for _, tc := range []struct {
		output string
		pass   bool
	}{
		{canonicalTestJudge, true}, {"```json\n" + canonicalTestJudge + "\n```", true},
		{`{"pairs":[]}`, false}, {`{}`, false}, {`not JSON`, false},
		{strings.Replace(canonicalTestJudge, `"same":true`, `"same":"true"`, 1), false},
		{strings.Replace(canonicalTestJudge, `"same":true,`, "", 1), false},
		{strings.Replace(canonicalTestJudge, `"line":2`, `"line":1`, 1), false},
		{strings.Replace(canonicalTestJudge, `"reason":"Unchanged"`, `"reason":""`, 1), false},
	} {
		r := ValidateCanonicalOutputWithPrompt(t.Context(), BuiltinFormalizationLineEquivalenceOutput, tc.output, pairs)
		if r.Passed != tc.pass {
			t.Fatalf("%s: %+v", tc.output, r)
		}
	}
	for _, name := range []string{BuiltinFormalizationCanonicalizeOutput, BuiltinFormalizationLineEquivalenceOutput} {
		if _, err := Compile(LegacyKindSpec(name)); err != nil {
			t.Fatal(err)
		}
	}
}

func TestApplyCanonicalization(t *testing.T) {
	proposals, judge, err := DecodeCanonicalOutputs(canonicalTestOutput, canonicalTestJudge)
	if err != nil {
		t.Fatal(err)
	}
	result, err := ApplyCanonicalization(canonicalTestOriginal, proposals, judge)
	if err != nil || !result.Changes[0].Accepted || !strings.Contains(result.Argument, "1. If interest rates rise, then mortgages become more expensive.") {
		t.Fatalf("%+v %v", result, err)
	}
	if err := ValidateCanonicalApplication(result, canonicalTestOriginal); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, original, canonical string
		same                      bool
		accepted                  bool
	}{
		{"judge no", "When interest rates rise, mortgages become more expensive.", "If interest rates rise, then mortgages become more expensive.", false, false},
		{"negation", "The central bank will not raise interest rates.", "The central bank will raise interest rates.", true, false},
		{"new content", "The central bank will raise interest rates.", "The central bank will sharply raise interest rates.", true, false},
		{"number", "The central bank will raise interest rates by 2.5 percent.", "The central bank will raise interest rates by 25 percent.", true, false},
		{"number multiplicity", "The price is 5 plus 5.", "The price is 5.", true, false},
		{"citation", "Therefore, house prices will fall. [1,2]", "Therefore, house prices will fall. [1]", true, false},
		{"marker", "Therefore, house prices will fall. [1,2]", "House prices will fall.", true, false},
		{"lost content", "The central bank rapidly raises interest rates every month.", "The bank raises rates.", true, false},
		{"double negation", "It is not the case that banks do not raise rates.", "Banks do not raise rates.", true, false},
		{"normalized negation", "Banks don't raise interest rates.", "It is not the case that banks do raise interest rates.", true, true},
		{"no equals not", "Banks have no cash reserves.", "It is not the case that banks have cash reserves.", true, true},
		{"unchanged ignores judge", "Banks raise interest rates.", "Banks raise interest rates.", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			original := "1. " + tc.original + "\n2. Another premise."
			index := 0
			if strings.HasPrefix(tc.original, "Therefore,") {
				original = "1. A.\n2. B.\n3. " + tc.original
				index = 2
			}
			lines, parseErr := CanonicalOriginalLines(original)
			if parseErr != nil {
				t.Fatal(parseErr)
			}
			p := CanonicalizeOutput{}
			j := LineEquivalenceOutput{}
			for i, line := range lines {
				same := true
				p.Lines = append(p.Lines, CanonicalLine{Line: i + 1, Canonical: line})
				j.Pairs = append(j.Pairs, LineEquivalence{Line: i + 1, Same: &same, Reason: "test"})
			}
			p.Lines[index].Canonical = tc.canonical
			j.Pairs[index].Same = &tc.same
			got, err := ApplyCanonicalization(original, p, j)
			if err != nil {
				t.Fatal(err)
			}
			if got.Changes[index].Accepted != tc.accepted {
				t.Fatalf("%+v", got.Changes[index])
			}
			if !tc.accepted && got.Argument != original {
				t.Fatalf("rejection changed original: %s", got.Argument)
			}
		})
	}
}

func TestCanonicalContractSerialization(t *testing.T) {
	p, j, _ := DecodeCanonicalOutputs(canonicalTestOutput, canonicalTestJudge)
	got, _ := ApplyCanonicalization(canonicalTestOriginal, p, j)
	raw, _ := json.Marshal(got)
	var decoded CanonicalApplication
	if json.Unmarshal(raw, &decoded) != nil || ValidateCanonicalApplication(decoded, canonicalTestOriginal) != nil {
		t.Fatal(string(raw))
	}
	decoded.Argument = canonicalTestOriginal
	if ValidateCanonicalApplication(decoded, canonicalTestOriginal) == nil {
		t.Fatal("accepted inconsistent applied output")
	}
}
