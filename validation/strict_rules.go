package validation

// strictRules belongs to one evaluation. Diagnostic retries copy it; no mutable
// package state or locks are needed, including for concurrent library callers.
type strictRules struct {
	isolatedPronounPrefix           string // private replay spelling; never user vocabulary
	term_pronoun_cross_statement    bool
	term_possessive_cross_statement bool
	slot_plural_mass_value          bool
	slot_plural_witness             bool
	pluralWitnessBlocked            [256]bool
	slot_description_shape          bool
	slot_one_type                   bool
	eg_subject_only                 bool
	eg_name_witness                 bool
	term_numeral                    bool
	identity_constant_sides         bool
	identity_extensional_context    bool
	slot_letter_in_name             bool
	term_inflection_verbatim        bool
	termSlots                       [256]bool
	slot_value_singular             bool
	slotNames                       *map[string]bool
	opener_frame                    bool
	opener_adverb                   bool
	coordination_clause             bool
	comma_exact                     bool
	identity_safe_position          bool
	slot_indefinite_object          bool
	slot_indefinite_written         bool
	slot_indefinite_value           bool
	slot_modified_quantifier        bool
	slot_quantity                   bool
	term_clause_inside              bool
	noun_pluralia_tantum            bool
	conditional_postposed_if        bool
	relation_object_extensional     bool
	unless_anaphor                  bool
	scope_mixed_andor               bool
	scope_clause_start              bool
	opener_ing_frame                bool
	scope_elided_copula             bool
	scope_description_negation      bool
	opener_ly_exception             bool
	// notes asks for strictness notes on failing steps (opt-in; diagnostic
	// retries never set it).
	notes bool
	// fired, when non-nil, records each strict rule whose refusal changed a
	// result during this evaluation. Each step and diagnostic retry uses its
	// own copy, so no other evaluation sees it.
	fired *firedRules
}

// firedRules is owned by one step evaluation on one goroutine.
type firedRules struct{ ids map[string]bool }

// hit records that rule id refused something and returns true, so it can end
// the condition that makes the refusal.
func (rules strictRules) hit(id string) bool {
	if rules.fired != nil {
		rules.fired.ids[id] = true
	}
	return true
}

// recording returns a copy of rules that records the strict rules that fire.
func (rules strictRules) recording() (strictRules, *firedRules) {
	fired := &firedRules{ids: map[string]bool{}}
	rules.fired = fired
	return rules, fired
}

func allStrictRules() strictRules {
	return strictRules{
		term_pronoun_cross_statement:    true,
		term_possessive_cross_statement: true,
		slot_plural_mass_value:          false,
		slot_plural_witness:             true,
		slot_description_shape:          true,
		slot_one_type:                   true,
		eg_subject_only:                 true,
		eg_name_witness:                 true,
		term_numeral:                    true,
		identity_constant_sides:         true,
		identity_extensional_context:    true,
		slot_letter_in_name:             true,
		term_inflection_verbatim:        true,
		slot_value_singular:             true,
		opener_frame:                    true,
		opener_adverb:                   true,
		coordination_clause:             true,
		comma_exact:                     true,
		identity_safe_position:          true,
		slot_indefinite_object:          true,
		slot_indefinite_written:         true,
		slot_indefinite_value:           true,
		slot_modified_quantifier:        true,
		slot_quantity:                   true,
		term_clause_inside:              true,
		noun_pluralia_tantum:            true,
		conditional_postposed_if:        true,
		relation_object_extensional:     true,
		unless_anaphor:                  true,
		scope_mixed_andor:               true,
		scope_clause_start:              true,
		opener_ing_frame:                true,
		scope_elided_copula:             true,
		scope_description_negation:      true,
		opener_ly_exception:             true,
	}
}

func (rules strictRules) without(id string) strictRules {
	switch id {
	case "term.pronoun_cross_statement":
		rules.term_pronoun_cross_statement = false
	case "term.possessive_cross_statement":
		rules.term_possessive_cross_statement = false
	case "slot.plural_mass_value":
		rules.slot_plural_mass_value = false
	case "slot.plural_witness":
		rules.slot_plural_witness = false
	case "slot.description_shape":
		rules.slot_description_shape = false
	case "slot.one_type":
		rules.slot_one_type = false
	case "eg.subject_only":
		rules.eg_subject_only = false
	case "eg.name_witness":
		rules.eg_name_witness = false
	case "term.numeral":
		rules.term_numeral = false
	case "identity.constant_sides":
		rules.identity_constant_sides = false
	case "identity.extensional_context":
		rules.identity_extensional_context = false
	case "slot.letter_in_name":
		rules.slot_letter_in_name = false
	case "term.inflection_verbatim":
		rules.term_inflection_verbatim = false
	case "scope.mixed_andor":
		rules.scope_mixed_andor = false
	case "scope.clause_start":
		rules.scope_clause_start = false
	case "opener.ing_frame":
		rules.opener_ing_frame = false
	case "scope.elided_copula":
		rules.scope_elided_copula = false
	case "scope.description_negation":
		rules.scope_description_negation = false
	case "opener.ly_exception":
		rules.opener_ly_exception = false
	case "slot.value_singular":
		rules.slot_value_singular = false
	case "opener.frame":
		rules.opener_frame = false
	case "opener.adverb":
		rules.opener_adverb = false
	case "coordination.clause":
		rules.coordination_clause = false
	case "comma.exact":
		rules.comma_exact = false
	case "identity.safe_position":
		rules.identity_safe_position = false
	case "slot.indefinite_object":
		rules.slot_indefinite_object = false
	case "slot.indefinite_written":
		rules.slot_indefinite_written = false
	case "slot.indefinite_value":
		rules.slot_indefinite_value = false
	case "slot.modified_quantifier":
		rules.slot_modified_quantifier = false
	case "slot.quantity":
		rules.slot_quantity = false
	case "term.clause_inside":
		rules.term_clause_inside = false
	case "noun.pluralia_tantum":
		rules.noun_pluralia_tantum = false
	case "conditional.postposed_if":
		rules.conditional_postposed_if = false
	case "relation.object_extensional":
		rules.relation_object_extensional = false
	case "unless.anaphor":
		rules.unless_anaphor = false
	}
	return rules
}
