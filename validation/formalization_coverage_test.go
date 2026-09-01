package validation

import (
	"strings"
	"testing"
)

const coverageAsFound = "1. If A, then B.\n2. A.\n3. Therefore, B. [1,2]"
const coverageRepaired = "1. If A, then B.\n2. A.\n3. Therefore, B. [1,2]"

func TestValidateFormalizationCoverageCompleteMappingPasses(t *testing.T) {
	mapping := []FormalizationLineMapping{
		{RepairedLine: 1, FromAsFoundLines: []int{1}},
		{RepairedLine: 2, FromAsFoundLines: []int{2}},
		{RepairedLine: 3, FromAsFoundLines: []int{3}},
	}
	dropped, alarm, err := ValidateFormalizationCoverage(coverageAsFound, coverageRepaired, mapping, []FormalizationDroppedContent{}, nil)
	if err != nil || alarm || dropped == nil || len(dropped) != 0 {
		t.Fatalf("coverage: dropped=%#v alarm=%v err=%v", dropped, alarm, err)
	}
}

func TestValidateFormalizationCoverageRejectsOrphanAsFoundLine(t *testing.T) {
	mapping := []FormalizationLineMapping{
		{RepairedLine: 1, FromAsFoundLines: []int{1}},
		{RepairedLine: 2, FromAsFoundLines: []int{2}},
	}
	_, _, err := ValidateFormalizationCoverage(coverageAsFound, coverageRepaired, mapping, []FormalizationDroppedContent{}, nil)
	if err == nil || !strings.Contains(err.Error(), "as-found line 3 is not covered") {
		t.Fatalf("error = %v, want orphan line 3", err)
	}
}

func TestValidateFormalizationCoverageSetsDroppedConclusionAlarm(t *testing.T) {
	repaired := "1. If A, then B.\n2. A."
	mapping := []FormalizationLineMapping{
		{RepairedLine: 1, FromAsFoundLines: []int{1}},
		{RepairedLine: 2, FromAsFoundLines: []int{2}},
	}
	dropped, alarm, err := ValidateFormalizationCoverage(coverageAsFound, repaired, mapping, []FormalizationDroppedContent{{
		AsFoundLine: 3,
		Kind:        DroppedContentKindConclusion,
		Reason:      "does not follow",
	}}, nil)
	if err != nil || !alarm || len(dropped) != 1 || dropped[0].Text != "B" {
		t.Fatalf("coverage: dropped=%#v alarm=%v err=%v", dropped, alarm, err)
	}
}
