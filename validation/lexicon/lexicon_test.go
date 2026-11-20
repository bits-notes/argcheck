package lexicon

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Core coverage probe: nuclear is adjectival; dataset and quale have no
// WordNet 3.0 noun synsets (noun.exc alone lists qualia -> quale).
// Norms, money and probabilities have quantity-descended first noun senses.
func TestLexiconRequiredAndLookup(t *testing.T) {
	for _, word := range strings.Fields("subgoals humans machines robots systems agents humanity intelligence evidence water information knowledge research software people sheep deer aircraft police goods works claims results models algorithms computers networks values goals preferences outcomes risks incentives policies capabilities resources institutions governments corporations researchers scientists philosophers arguments reasons beliefs desires intentions actions consequences rights duties obligations laws experiments observations theories hypotheses data consciousness experience suffering welfare wellbeing happiness pleasure pain morality ethics justice freedom liberty safety security alignment training hardware electricity energy wealth poverty inequality education healthcare medicine vaccines drugs animals plants species organisms genes cells brains minds persons children students teachers workers employers employees citizens voters patients doctors nurses lawyers judges firms companies markets prices wages taxes profits costs benefits harms dangers threats weapons predictions credences utilities decisions choices options strategies simulations worlds universes zombies properties facts truths propositions sentences words concepts") {
		if !isAllowlistedPluralMassNoun(word) || !isAllowlistedPluralMassNoun(strings.ToUpper(word)) {
			t.Errorf("required word missing or lowercase lookup broken: %q", word)
		}
	}
	for _, word := range []string{"", "robot", "robot systems", " robots", "robots ", "#", "childs", "mans", "informations"} {
		if isAllowlistedPluralMassNoun(word) {
			t.Errorf("unexpected membership: %q", word)
		}
	}
}

func TestLexiconExclusions(t *testing.T) {
	// All explicit brief examples, including each listed singular/plural form,
	// plus quantity descendants and irregular measures.
	denied := `lot lots plenty dozen dozens hundred hundreds thousand thousands
million millions billion billions score scores bunch bunches ton tons load loads
heap heaps pile piles stack stacks majority minority bulk rest remainder most
half halves part parts portion portions share shares number numbers amount amounts
quantity quantities couple couples handful handfuls pair pairs percent percentage
percentages fraction fractions proportion proportions deal deals sum sums total
totals lack mass masses range ranges variety varieties series set sets group
groups kind kinds sort sorts type types all none both few several many much more
less least enough mile miles pound pounds dollar dollars year years hour hours
minute minutes degree degrees foot feet inch inches second seconds yen euros
rupees hectares kilograms milligrams milliseconds kilowatts trillion trillions
zero one two three four five six seven eight nine ten eleven twelve thirteen
fourteen fifteen sixteen seventeen eighteen nineteen twenty thirty forty fifty
sixty seventy eighty ninety first third quarter fourth fifth sixth seventh eighth
ninth tenth everyone something nobody anyone everything somebody anything nothing
norms money probabilities ages eons aeons droves hordes legions swathes swaths
is has does was says goes means`
	for _, word := range strings.Fields(denied) {
		if isAllowlistedPluralMassNoun(word) {
			t.Errorf("excluded word present: %q", word)
		}
	}
	// Every explicit quantity word must remain excluded, including future additions.
	quantity, err := os.ReadFile("../../scripts/gen-noun-allowlist/quantity-words.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(quantity), "\n") {
		for _, word := range strings.Fields(strings.SplitN(line, "#", 2)[0]) {
			if isAllowlistedPluralMassNoun(word) {
				t.Errorf("explicit quantity word present: %q", word)
			}
		}
	}
	// Read the actual validator declarations so later changes cannot silently
	// drift away from the generated lexicon. No production validator dependency.
	files, err := filepath.Glob("../*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"singularPartitiveWords", "termDeterminers", "quantityNumerals", "singularClosedWords"} {
		re := regexp.MustCompile(`var\s+` + name + `\s*=\s*wordSet\("([^"]*)"\)`)
		found := 0
		for _, file := range files {
			data, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			for _, match := range re.FindAllStringSubmatch(string(data), -1) {
				found++
				for _, word := range strings.Fields(match[1]) {
					if isAllowlistedPluralMassNoun(word) {
						t.Errorf("%s word present: %q; regenerate lexicon", name, word)
					}
				}
			}
		}
		if found != 1 {
			t.Errorf("expected one %s declaration, got %d", name, found)
		}
	}
}

func TestLexiconFormat(t *testing.T) {
	valid := regexp.MustCompile(`^[a-z]+$`)
	previous := ""
	count := 0
	for _, line := range strings.Split(pluralMassNouns, "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if !valid.MatchString(line) || line <= previous {
			t.Fatalf("invalid, duplicate or unsorted entry %q after %q", line, previous)
		}
		previous = line
		count++
	}
	if count < 8000 || count > 10000 || count != len(pluralMassNounSet) {
		t.Fatalf("unexpected list/map size: %d/%d", count, len(pluralMassNounSet))
	}
}
