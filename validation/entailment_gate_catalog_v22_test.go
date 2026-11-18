package validation

import (
	"fmt"
	"strings"
	"testing"
)

const v22A1 = `1. Alice gave Bob a book.
2. Bob is a child.
3. Therefore, Alice gave some child a book. [1,2]`
const v22A3 = `1. Every lab sells some model to every government.
2. OpenAI is a lab.
3. Japan is a government.
4. Therefore, OpenAI sells some model to Japan. [1,2,3]`

type nplaceExercise struct {
	name, argument string
	p              []nOracleSurface
	c              nOracleSurface
	valid          bool
}

func nplaceExercises() []nplaceExercise {
	var out []nplaceExercise
	add := func(name string, p []nOracleSurface, c nOracleSurface, valid bool) {
		texts := make([]string, len(p))
		for i, s := range p {
			texts[i] = s.text
		}
		out = append(out, nplaceExercise{name, relationArgument(texts, c.text), p, c, valid})
	}
	member := func(name, class string) nOracleSurface { return nGround(name+" is a "+class, class, name) }
	a1p := nQuantified("Alice gave Bob a book", "gave/to", []string{"Alice", "", "Bob"}, map[int]string{1: "some book"}, nil)
	a1c := nQuantified("Alice gave some child a book", "gave/to", []string{"Alice", "", ""}, map[int]string{1: "some book", 2: "some child"}, nil)
	add("a1", []nOracleSurface{a1p, member("Bob", "child")}, a1c, true)
	out[len(out)-1].argument = v22A1
	a3p := nQuantified("Every lab sells some model to every government", "sells/to", make([]string, 3), map[int]string{0: "every lab", 1: "some model", 2: "every government"}, nil)
	a3c := nQuantified("OpenAI sells some model to Japan", "sells/to", []string{"OpenAI", "", "Japan"}, map[int]string{1: "some model"}, nil)
	add("a3", []nOracleSurface{a3p, member("OpenAI", "lab"), member("Japan", "government")}, a3c, true)
	out[len(out)-1].argument = v22A3
	// Each argument position is instantiated and existentially introduced.
	for arity := 3; arity <= 4; arity++ {
		names := []string{"Alice", "Parcel", "Bob", "Depot"}[:arity]
		classes := []string{"courier", "package", "recipient", "station"}[:arity]
		render := func(ts []string) string {
			s := ts[0] + " sends " + ts[1] + " to " + ts[2]
			if len(ts) == 4 {
				s += " from " + ts[3]
			}
			return s
		}
		pred := "sends/to"
		if arity == 4 {
			pred += "/from"
		}
		ground := nGround(render(names), pred, names...)
		for slot := range names {
			ts := append([]string{}, names...)
			ts[slot] = "every " + classes[slot]
			universal := nQuantified(render(ts), pred, names, map[int]string{slot: "every " + classes[slot]}, nil)
			add(fmt.Sprintf("%d-place UI slot %d", arity, slot), []nOracleSurface{universal, member(names[slot], classes[slot])}, ground, true)
			ts[slot] = "some " + classes[slot]
			existential := nQuantified(render(ts), pred, names, map[int]string{slot: "some " + classes[slot]}, nil)
			add(fmt.Sprintf("%d-place EI slot %d", arity, slot), []nOracleSurface{ground, member(names[slot], classes[slot])}, existential, true)
		}
		// Fitch UG and EE with a fresh eigenname in the final object position.
		slot := arity - 1
		ts := append([]string{}, names...)
		ts[slot] = "every " + classes[slot]
		universal := nQuantified(render(ts), pred, names, map[int]string{slot: "every " + classes[slot]}, nil)
		ts[slot] = "z"
		ug := fmt.Sprintf("1. %s.\n| 2. Let z be an arbitrary %s.\n| 3. Therefore, %s. [1,2]\n4. Therefore, %s. [2-3]", universal.text, classes[slot], render(ts), universal.text)
		out = append(out, nplaceExercise{fmt.Sprintf("%d-place Fitch UG", arity), ug, []nOracleSurface{universal}, universal, true})
		ts[slot] = "some " + classes[slot]
		existential := nQuantified(render(ts), pred, names, map[int]string{slot: "some " + classes[slot]}, nil)
		ts[slot] = "z"
		ee := fmt.Sprintf("1. %s.\n| 2. Suppose z is a %s and %s.\n| 3. Therefore, z is a %s. [2]\n| 4. Therefore, %s. [2]\n| 5. Therefore, %s. [4,3]\n6. Therefore, %s. [1,2-5]", existential.text, classes[slot], render(ts), classes[slot], render(ts), existential.text, existential.text)
		out = append(out, nplaceExercise{fmt.Sprintf("%d-place Fitch EE", arity), ee, []nOracleSurface{existential}, existential, true})
		// v38.0: identity substitution is certified in clause subjects, so
		// the exercise renames the subject (slot 0), not an object.
		renamed := append([]string{}, names...)
		renamed[0] = "Cedar"
		add(fmt.Sprintf("%d-place identity", arity), []nOracleSurface{ground, nIdentity(names[0]+" is identical to Cedar", names[0], "Cedar")}, nGround(render(renamed), pred, renamed...), true)
	}
	alt := nQuantified("Alice gave a book to Bob", "gave/to", []string{"Alice", "", "Bob"}, map[int]string{1: "some book"}, nil)
	add("double-object to preposition", []nOracleSurface{a1p}, alt, true)
	add("preposition to double-object", []nOracleSurface{alt}, a1p, true)
	// Four binders, and a fully explicit scope containing three binders.
	four := nQuantified("Every courier sends every package to every recipient from every station", "sends/to/from", make([]string, 4), map[int]string{0: "every courier", 1: "every package", 2: "every recipient", 3: "every station"}, nil)
	fourC := nQuantified("Alice sends every package to every recipient from every station", "sends/to/from", []string{"Alice", "", "", ""}, map[int]string{1: "every package", 2: "every recipient", 3: "every station"}, nil)
	add("four quantifiers", []nOracleSurface{four, member("Alice", "courier")}, fourC, true)
	scoped := nQuantified("Some package is such that every courier is such that every recipient is such that that courier sends that package to that recipient", "sends/to", make([]string, 3), map[int]string{0: "every courier", 1: "some package", 2: "every recipient"}, []int{1, 0, 2})
	ambiguous := nQuantified("Every courier sends some package to every recipient", "sends/to", make([]string, 3), map[int]string{0: "every courier", 1: "some package", 2: "every recipient"}, nil)
	add("wide witness scope weakening", []nOracleSurface{scoped}, ambiguous, true)
	// Required counterexamples and further minimally changed look-alikes.
	swapped := nQuantified("Bob gave Alice a book", "gave/to", []string{"Bob", "", "Alice"}, map[int]string{1: "some book"}, nil)
	add("swapped roles", []nOracleSurface{a1p}, swapped, false)
	add("wrong scope", []nOracleSurface{ambiguous}, scoped, false)
	add("existential slot cannot instantiate", []nOracleSurface{nQuantified("Alice sends some package to Bob", "sends/to", []string{"Alice", "", "Bob"}, map[int]string{1: "some package"}, nil), member("Parcel", "package")}, nGround("Alice sends Parcel to Bob", "sends/to", "Alice", "Parcel", "Bob"), false)
	add("different arity", []nOracleSurface{nGround("Alice sends Parcel to Bob from Depot", "sends/to/from", "Alice", "Parcel", "Bob", "Depot")}, nGround("Alice sends Parcel to Bob", "sends/to", "Alice", "Parcel", "Bob"), false)
	add("wrong membership", []nOracleSurface{a1p, member("Bob", "adult")}, a1c, false)
	add("wrong existential name", []nOracleSurface{a1p, member("Alice", "child")}, a1c, false)
	add("wrong verb", []nOracleSurface{a1p}, nQuantified("Alice sent a book to Bob", "sent/to", []string{"Alice", "", "Bob"}, map[int]string{1: "some book"}, nil), false)
	add("wrong preposition", []nOracleSurface{alt}, nQuantified("Alice gave a book from Bob", "gave/from", []string{"Alice", "", "Bob"}, map[int]string{1: "some book"}, nil), false)
	add("active passive distinct", []nOracleSurface{a3p}, nQuantified("Some model is sold by every lab to every government", "sold-by/to", make([]string, 3), map[int]string{0: "some model", 1: "every lab", 2: "every government"}, nil), false)
	add("four-place swapped endpoints", []nOracleSurface{nGround("Alice sends Parcel to Bob from Depot", "sends/to/from", "Alice", "Parcel", "Bob", "Depot")}, nGround("Alice sends Parcel to Depot from Bob", "sends/to/from", "Alice", "Parcel", "Depot", "Bob"), false)
	add("ground ditransitive alternation", []nOracleSurface{nGround("Alice gave Bob Parcel", "gave/to", "Alice", "Parcel", "Bob")}, nGround("Alice gave Parcel to Bob", "gave/to", "Alice", "Parcel", "Bob"), true)
	noCourier := nQuantified("No courier sends Parcel to Bob", "sends/to", []string{"", "Parcel", "Bob"}, map[int]string{0: "no courier"}, nil)
	positive := nGround("Alice sends Parcel to Bob", "sends/to", "Alice", "Parcel", "Bob")
	negative := nOracleSurface{"It is not the case that " + positive.text, []*finiteFormula{foNot(positive.readings[0])}}
	add("negative three-place UI", []nOracleSurface{noCourier, member("Alice", "courier")}, negative, true)
	add("wrong negative polarity", []nOracleSurface{noCourier, member("Alice", "courier")}, positive, false)
	five := nQuantified("Alice sends Parcel to Bob from Depot with every vehicle", "sends/to/from/with", []string{"Alice", "Parcel", "Bob", "Depot", ""}, map[int]string{4: "every vehicle"}, nil)
	add("five-place UI", []nOracleSurface{five, member("Van", "vehicle")}, nGround("Alice sends Parcel to Bob from Depot with Van", "sends/to/from/with", "Alice", "Parcel", "Bob", "Depot", "Van"), true)

	staleGoal := nQuantified("Every courier sends Parcel to Bob", "sends/to", []string{"", "Parcel", "Bob"}, map[int]string{0: "every courier"}, nil)
	out = append(out, nplaceExercise{"stale Fitch eigenname", "1. Alice sends Parcel to Bob.\n| 2. Let Alice be an arbitrary courier.\n| 3. Therefore, Alice sends Parcel to Bob. [1]\n4. Therefore, Every courier sends Parcel to Bob. [2-3]", []nOracleSurface{positive}, staleGoal, false})
	add("indefinite object is not a named object", []nOracleSurface{a1p, member("Bob", "child")}, nQuantified("Alice gave some child Book", "gave/to", []string{"Alice", "Book", ""}, map[int]string{2: "some child"}, nil), false)

	repeated := nQuantified("Every courier sends every courier to every courier", "sends/to", make([]string, 3), map[int]string{0: "every courier", 1: "every courier", 2: "every courier"}, nil)
	repeatedC := nQuantified("Alice sends every courier to every courier", "sends/to", []string{"Alice", "", ""}, map[int]string{1: "every courier", 2: "every courier"}, nil)
	add("repeated class universal slots", []nOracleSurface{repeated, member("Alice", "courier")}, repeatedC, true)

	return out
}
func TestCatalogV22(t *testing.T) {
	if EntailmentGateCatalogVersion != "v39.0" {
		t.Fatal(EntailmentGateCatalogVersion)
	}
	for _, tc := range nplaceExercises() {
		t.Run(tc.name, func(t *testing.T) {
			report, result := CheckEntailmentGate(tc.argument)
			if result.Passed != tc.valid {
				t.Fatalf("want valid=%v\n%s\n%+v\n%s", tc.valid, tc.argument, report, result.Message)
			}
		})
	}
}
func TestCatalogV22QuantifierCap(t *testing.T) {
	arg := relationArgument([]string{"Every courier sends every package to every recipient from every station with some vehicle"}, "Every courier sends every package to every recipient from every station with some vehicle")
	report, result := CheckEntailmentGate(arg)
	if result.Passed || !strings.Contains(report.Steps[0].Mismatch, "at most four quantifiers") {
		t.Fatalf("%+v", report)
	}
}
