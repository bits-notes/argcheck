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

func TestValidateFormalizationCoverageSourceAnchorsMeasuredDroppedLines(t *testing.T) {
	source := "1. If it rains, then the ground is wet.\n2. It rains.\n3. Therefore, the ground is wet. [1,2]\n4. If the sun shines, then the sky is bright.\n5. The sun shines.\n6. Therefore, the sky is bright. [4,5]"
	repaired := "1. If it rains, then the ground is wet.\n2. It rains.\n3. Therefore, the ground is wet. [1,2]"
	mapping := []FormalizationLineMapping{
		{RepairedLine: 1, FromAsFoundLines: []int{1}},
		{RepairedLine: 2, FromAsFoundLines: []int{2}},
		{RepairedLine: 3, FromAsFoundLines: []int{3}},
	}
	dropped, alarm, err := ValidateFormalizationCoverageAnchored(FormalizationCoverageAnchorSource, source, repaired, mapping, []FormalizationDroppedContent{
		{AsFoundLine: 4, Kind: DroppedContentKindPremise, Reason: "separate sub-argument"},
		{AsFoundLine: 5, Kind: DroppedContentKindPremise, Reason: "separate sub-argument"},
		{AsFoundLine: 6, Kind: DroppedContentKindConclusion, Reason: "separate sub-argument"},
	}, nil)
	if err != nil || !alarm || len(dropped) != 3 || dropped[0].Text != "If the sun shines, then the sky is bright" {
		t.Fatalf("source coverage: dropped=%#v alarm=%v err=%v", dropped, alarm, err)
	}
}

func TestValidateFormalizationCoverageSourceAnchorClassifiesDeadPremiseAndInertCitation(t *testing.T) {
	source := "1. If it rains, then the ground is wet.\n2. It rains.\n3. The moon orbits the Earth.\n4. Therefore, the ground is wet. [1,2]"
	repaired := "1. If it rains, then the ground is wet.\n2. It rains.\n3. Therefore, the ground is wet. [1,2]"
	mapping := []FormalizationLineMapping{
		{RepairedLine: 1, FromAsFoundLines: []int{1}},
		{RepairedLine: 2, FromAsFoundLines: []int{2}},
		{RepairedLine: 3, FromAsFoundLines: []int{4}},
	}
	dropped, alarm, err := ValidateFormalizationCoverageAnchored(FormalizationCoverageAnchorSource, source, repaired, mapping, []FormalizationDroppedContent{{
		AsFoundLine: 3, Kind: DroppedContentKindPremise, Reason: "dead premise",
	}}, nil)
	if err != nil || alarm || len(dropped) != 1 || dropped[0].Text != "The moon orbits the Earth" {
		t.Fatalf("dead-premise coverage: dropped=%#v alarm=%v err=%v", dropped, alarm, err)
	}
}

func TestValidateFormalizationCoverageAllowsMeasuredDerivedConclusions(t *testing.T) {
	tests := []struct {
		name, source, repaired string
		mapping                []FormalizationLineMapping
		supplied               []int
	}{
		{
			name: "pyrexia",
			source: "1. Two calibrated oral thermometers recorded the patient's temperature as 38.4 degrees Celsius within five minutes of admission.\n" +
				"2. Neither measurement was taken immediately after hot or cold fluid intake, and the two readings differed by less than the devices' stated measurement tolerance.\n" +
				"3. Therefore, the patient had pyrexia at admission. [1,2]",
			repaired: "1. Two calibrated oral thermometers recorded the patient's temperature as 38.4 degrees Celsius within five minutes of admission.\n" +
				"2. Neither measurement was taken immediately after hot or cold fluid intake.\n" +
				"3. The two readings differed by less than the devices' stated measurement tolerance.\n" +
				"4. A patient has pyrexia when their temperature is recorded as above 38.0 degrees Celsius by calibrated instruments under valid measurement conditions.\n" +
				"5. If two calibrated oral thermometers recorded the patient's temperature as 38.4 degrees Celsius within five minutes of admission, and neither measurement was taken immediately after hot or cold fluid intake, and the two readings differed by less than the devices' stated measurement tolerance, then the patient's temperature was recorded as above 38.0 degrees Celsius by calibrated instruments under valid measurement conditions.\n" +
				"6. Therefore, the patient's temperature was recorded as above 38.0 degrees Celsius by calibrated instruments under valid measurement conditions. [1,2,3,5]\n" +
				"7. Therefore, the patient had pyrexia at admission. [4,6]",
			mapping:  []FormalizationLineMapping{{RepairedLine: 1, FromAsFoundLines: []int{1}}, {RepairedLine: 2, FromAsFoundLines: []int{2}}, {RepairedLine: 3, FromAsFoundLines: []int{2}}, {RepairedLine: 7, FromAsFoundLines: []int{3}}},
			supplied: []int{4, 5},
		},
		{
			name: "standard-atmosphere",
			source: "1. The calibrated pressure transducer reported a stable absolute pressure of 101.325 kilopascals.\n" +
				"2. The calibration certificate was current, the transducer remained within its stated operating range, and no pressure transient occurred.\n" +
				"3. Therefore, the chamber was at one standard atmosphere of pressure. [1,2]",
			repaired: "1. The calibrated pressure transducer reported a stable absolute pressure of 101.325 kilopascals.\n" +
				"2. The calibration certificate was current.\n3. The transducer remained within its stated operating range.\n4. No pressure transient occurred.\n" +
				"5. If the transducer reported 101.325 kilopascals under valid conditions, then the chamber's actual pressure was 101.325 kilopascals.\n" +
				"6. Therefore, the chamber's actual pressure was 101.325 kilopascals. [1,2,3,4,5]\n" +
				"7. One standard atmosphere is 101.325 kilopascals.\n8. If the chamber's actual pressure was 101.325 kilopascals and one standard atmosphere is 101.325 kilopascals, then the chamber was at one standard atmosphere.\n" +
				"9. Therefore, the chamber was at one standard atmosphere of pressure. [6,7,8]",
			mapping:  []FormalizationLineMapping{{RepairedLine: 1, FromAsFoundLines: []int{1}}, {RepairedLine: 2, FromAsFoundLines: []int{2}}, {RepairedLine: 3, FromAsFoundLines: []int{2}}, {RepairedLine: 4, FromAsFoundLines: []int{2}}, {RepairedLine: 9, FromAsFoundLines: []int{3}}},
			supplied: []int{5, 7, 8},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			_, _, err := ValidateFormalizationCoverageAnchored(FormalizationCoverageAnchorSource, testCase.source, testCase.repaired, testCase.mapping, []FormalizationDroppedContent{}, testCase.supplied)
			if err != nil {
				t.Fatalf("derived-conclusion coverage failed: %v", err)
			}
		})
	}
}

func TestValidateFormalizationCoverageWTVAllowsSourcePremiseSplitButRejectsDuplicateConclusion(t *testing.T) {
	source := "1. A.\n2. B and C.\n3. Therefore, D. [1,2]"
	repaired := "1. A.\n2. B.\n3. C.\n4. Therefore, E. [1,2,3]\n5. Therefore, D. [4]"
	base := []FormalizationLineMapping{
		{RepairedLine: 1, FromAsFoundLines: []int{1}},
		{RepairedLine: 2, FromAsFoundLines: []int{2}},
		{RepairedLine: 3, FromAsFoundLines: []int{2}},
		{RepairedLine: 5, FromAsFoundLines: []int{3}},
	}
	if _, _, err := ValidateFormalizationCoverageAnchored(FormalizationCoverageAnchorSource, source, repaired, base, []FormalizationDroppedContent{}, nil); err != nil {
		t.Fatalf("source premise split failed: %v", err)
	}
	duplicateConclusion := append(append([]FormalizationLineMapping(nil), base...), FormalizationLineMapping{RepairedLine: 4, FromAsFoundLines: []int{3}})
	if _, _, err := ValidateFormalizationCoverageAnchored(FormalizationCoverageAnchorSource, source, repaired, duplicateConclusion, []FormalizationDroppedContent{}, nil); err == nil || !strings.Contains(err.Error(), "source line 3 is covered 2 times") {
		t.Fatalf("duplicate source conclusion error = %v", err)
	}
}
