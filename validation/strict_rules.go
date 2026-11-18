package validation

// strictRules belongs to one evaluation. Diagnostic retries copy it; no mutable
// package state or locks are needed, including for concurrent library callers.
type strictRules struct {
	opener_frame                bool
	opener_adverb               bool
	coordination_clause         bool
	comma_exact                 bool
	identity_safe_position      bool
	slot_indefinite_object      bool
	slot_modified_quantifier    bool
	slot_quantity               bool
	term_clause_inside          bool
	noun_pluralia_tantum        bool
	conditional_postposed_if    bool
	relation_object_extensional bool
	unless_anaphor              bool
	// notes asks for strictness notes on failing steps (opt-in; diagnostic
	// retries never set it).
	notes bool
}

func allStrictRules() strictRules {
	return strictRules{
		opener_frame:                true,
		opener_adverb:               true,
		coordination_clause:         true,
		comma_exact:                 true,
		identity_safe_position:      true,
		slot_indefinite_object:      true,
		slot_modified_quantifier:    true,
		slot_quantity:               true,
		term_clause_inside:          true,
		noun_pluralia_tantum:        true,
		conditional_postposed_if:    true,
		relation_object_extensional: true,
		unless_anaphor:              true,
	}
}

func (rules strictRules) without(id string) strictRules {
	switch id {
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
