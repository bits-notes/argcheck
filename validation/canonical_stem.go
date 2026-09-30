package validation

import "strings"

// canonicalContentStems deliberately wraps the surface-word helper rather than
// changing it: quote verification and its diagnostics must remain exact. Values
// retain a deterministic surface spelling for rejection diagnostics.
func canonicalContentStems(text string) map[string]string {
	result := map[string]string{}
	for word := range sourceContentWords(text) {
		if canonicalFormWords[word] || word == "nothing" {
			continue // nothing/anything/something equivalence; polarity checked first
		}
		stem := canonicalEnglishStem(word)
		if previous, ok := result[stem]; !ok || word < previous {
			result[stem] = word
		}
	}
	return result
}

// canonicalEnglishStem is a small deterministic suffix stripper, not a
// lemmatizer. On lowercase ASCII words it removes plural -s/-es (-ies -> -y),
// then one of -ing/-ed/-er/-ers, -ity/-ities/-ility/-ability/-ibility,
// -able/-ible, or -ly (-ility -> -il). It folds doubled consonants after
// inflectional suffixes and drops final silent e (so proposed/propose and
// making/make agree).
// Eligibility and eligible both become elig via -ibility/-ible. At least three
// letters must remain after suffix removal. Non-ASCII and numeric tokens are
// untouched. These deliberately limited lexical heuristics are not a semantic
// equivalence test; canonical apply still requires the equivalence judge.
func canonicalEnglishStem(word string) string {
	if canonicalPolarity(word) != 0 {
		return word
	}
	for _, c := range word {
		if c < 'a' || c > 'z' {
			return word
		}
	}
	if strings.HasSuffix(word, "ies") && len(word) > 5 {
		word = strings.TrimSuffix(word, "ies") + "y"
	} else if len(word) > 5 && (strings.HasSuffix(word, "sses") || strings.HasSuffix(word, "shes") || strings.HasSuffix(word, "ches") || strings.HasSuffix(word, "xes") || strings.HasSuffix(word, "zes")) {
		word = strings.TrimSuffix(word, "es")
	} else if len(word) > 3 && strings.HasSuffix(word, "s") && !strings.HasSuffix(word, "ss") && !strings.HasSuffix(word, "us") && !strings.HasSuffix(word, "is") {
		word = strings.TrimSuffix(word, "s")
	}
	for _, suffix := range []string{"ability", "ibility", "ility", "able", "ible", "ing", "ity", "ed", "er", "ly"} {
		if !strings.HasSuffix(word, suffix) || len(word)-len(suffix) < 3 {
			continue
		}
		word = strings.TrimSuffix(word, suffix)
		if suffix == "ility" {
			word += "il"
		}
		if suffix == "ing" || suffix == "ed" || suffix == "er" {
			n := len(word)
			if n > 3 && word[n-1] == word[n-2] && strings.ContainsRune("bdgmnprt", rune(word[n-1])) {
				word = word[:n-1]
			}
		}
		break
	}
	if len(word) > 3 && strings.HasSuffix(word, "e") && !strings.HasSuffix(word, "ee") {
		word = strings.TrimSuffix(word, "e")
	}
	return word
}
