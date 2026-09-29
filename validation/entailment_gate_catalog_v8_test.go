package validation

import (
	"fmt"
	"strings"
	"testing"
)

func TestCatalogV8Version(t *testing.T) {
	if EntailmentGateCatalogVersion != "v8" {
		t.Fatalf("catalog = %q, want v8", EntailmentGateCatalogVersion)
	}
}

func TestCatalogV8Coverage(t *testing.T) {
	tests := []struct {
		name, argument string
		want           bool
	}{
		{"whales", "1. All mammals breathe air.\n2. Whales are mammals.\n3. Therefore, whales breathe air. [1,2]", true},
		{"moral exact live repair", "1. Factory farming causes enormous suffering to animals that can feel pain.\n2. People can eat well without meat.\n3. If people can eat well without meat, then most of the suffering caused by factory farming is unnecessary.\n4. Therefore, most of the suffering caused by factory farming is unnecessary. [2,3]\n5. Causing unnecessary suffering is wrong.\n6. If factory farming causes enormous suffering to animals that can feel pain and most of the suffering caused by factory farming is unnecessary and causing unnecessary suffering is wrong, then we should stop supporting factory farming.\n7. Therefore, we should stop supporting factory farming. [1,4,5,6]", true},
		{"three", "1. Rain falls.\n2. Wind blows.\n3. Snow melts.\n4. If rain falls and wind blows and snow melts, then rivers rise.\n5. Therefore, rivers rise. [4,3,1,2]", true},
		{"four", "1. Rain falls.\n2. Wind blows.\n3. Snow melts.\n4. Ice cracks.\n5. If rain falls and wind blows and snow melts and ice cracks, then rivers rise.\n6. Therefore, rivers rise. [1,2,3,4,5]", true},
		{"embedded and", "1. Bread and butter are available.\n2. Tea is hot.\n3. Guests arrive.\n4. If bread and butter are available and tea is hot and guests arrive, then breakfast starts.\n5. Therefore, breakfast starts. [1,2,3,4]", true},
		{"existing class plural normalization", "1. All cats can sleep.\n2. Felix is a cat.\n3. Therefore, Felix can sleep. [1,2]", true},
		{"every exact singular", "1. Every cat purrs.\n2. Felix is a cat.\n3. Therefore, Felix purrs. [2,1]", true},
		{"an membership", "1. Every animal sleeps.\n2. Felix is an animal.\n3. Therefore, Felix sleeps. [1,2]", true},
		{"schematic three", "1. Ada smiles.\n2. Ada sings.\n3. Ada dances.\n4. If X smiles and X sings and X dances, then X celebrates.\n5. Therefore, Ada celebrates. [1,2,3,4]", true},
		{"affirming consequent", "1. If rain falls and wind blows and snow melts, then rivers rise.\n2. Rivers rise.\n3. Wind blows.\n4. Snow melts.\n5. Therefore, rain falls. [1,2,3,4]", false},
		{"extra conditional conjunct", "1. Rain falls.\n2. Wind blows.\n3. Snow melts.\n4. If rain falls and wind blows and snow melts and ice cracks, then rivers rise.\n5. Therefore, rivers rise. [1,2,3,4]", false},
		{"missing citation", "1. Rain falls.\n2. Wind blows.\n3. Snow melts.\n4. If rain falls and wind blows and snow melts, then rivers rise.\n5. Therefore, rivers rise. [1,2,4]", false},
		{"unused citation", "1. Rain falls.\n2. Wind blows.\n3. Snow melts.\n4. Ice cracks.\n5. If rain falls and wind blows and snow melts, then rivers rise.\n6. Therefore, rivers rise. [1,2,3,4,5]", false},
		{"wrong consequent", "1. Rain falls.\n2. Wind blows.\n3. Snow melts.\n4. If rain falls and wind blows and snow melts, then rivers rise.\n5. Therefore, rivers fall. [4,3,1,2]", false},
		{"wrong class", "1. All mammals breathe air.\n2. Whales are fish.\n3. Therefore, whales breathe air. [1,2]", false},
		{"noncopular membership", "1. All mammals breathe air.\n2. Whales resemble mammals.\n3. Therefore, whales breathe air. [1,2]", false},
		{"converse", "1. All mammals breathe air.\n2. Whales breathe air.\n3. Therefore, whales are mammals. [1,2]", false},
		{"predicate plural tolerance forbidden", "1. All mammals eat apples.\n2. Whales are mammals.\n3. Therefore, whales eat apple. [1,2]", false},
		{"predicate mismatch", "1. All mammals breathe air.\n2. Whales are mammals.\n3. Therefore, whales breathe water. [1,2]", false},
		{"no verb inflection", "1. All men die.\n2. Socrates is a man.\n3. Therefore, Socrates dies. [1,2]", false},
		{"attitude membership", "1. All mammals breathe air.\n2. Ada believes whales are mammals.\n3. Therefore, Ada believes whales breathe air. [1,2]", false},
		{"embedded membership", "1. All mammals breathe air.\n2. Ada thinks whales are mammals.\n3. Therefore, whales breathe air. [1,2]", false},
		{"relative class", "1. All mammals that swim breathe air.\n2. Whales are mammals.\n3. Therefore, whales breathe air. [1,2]", false},
		{"schematic inconsistent", "1. Ada smiles.\n2. Ben sings.\n3. Ada dances.\n4. If X smiles and X sings and X dances, then X celebrates.\n5. Therefore, Ada celebrates. [1,2,3,4]", false},
		{"Some", "1. Some mammals breathe air.\n2. Whales are mammals.\n3. Therefore, whales breathe air. [1,2]", false},
		{"Most", "1. Most mammals breathe air.\n2. Whales are mammals.\n3. Therefore, whales breathe air. [1,2]", false},
		{"Many", "1. Many mammals breathe air.\n2. Whales are mammals.\n3. Therefore, whales breathe air. [1,2]", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, result := CheckEntailmentGate(tt.argument)
			if report.Passed != tt.want || result.Passed != tt.want {
				t.Fatalf("want passed=%v: report=%+v result=%+v", tt.want, report, result)
			}
		})
	}
}

func TestPlainAndPartitionBound(t *testing.T) {
	for _, count := range []int{12, 13} {
		parts := make([]string, count+1)
		for i := range parts {
			parts[i] = fmt.Sprintf("clause %d holds", i)
		}
		got := plainAndPartitions(strings.Join(parts, " and "), len(parts))
		if count == 12 && len(got) != 1 {
			t.Fatalf("12 boundaries: %d partitions", len(got))
		}
		if count == 13 && len(got) != 0 {
			t.Fatalf("over cap: %d partitions", len(got))
		}
	}
}

func TestPlainAndGateBound(t *testing.T) {
	for _, count := range []int{12, 13} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			var lines, clauses, citations []string
			for i := 0; i <= count; i++ {
				clause := fmt.Sprintf("Clause %d holds", i)
				clauses = append(clauses, clause)
				lines = append(lines, fmt.Sprintf("%d. %s.", i+1, clause))
				citations = append(citations, fmt.Sprint(i+1))
			}
			lines = append(lines, fmt.Sprintf("%d. If %s, then the result holds.", count+2, strings.Join(clauses, " and ")))
			citations = append(citations, fmt.Sprint(count+2))
			lines = append(lines, fmt.Sprintf("%d. Therefore, the result holds. [%s]", count+3, strings.Join(citations, ",")))
			report, result := CheckEntailmentGate(strings.Join(lines, "\n"))
			if report.Passed != (count == 12) || result.Passed != (count == 12) {
				t.Fatalf("unexpected cap result: %+v %+v", report, result)
			}
		})
	}
}
