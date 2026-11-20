# v46 pin audit

Counts by category (table rows): 1: 25; 2: 8; 3: 10; 4: 5.

The supplied log contains 25 top-level failures and 36 named subtest failures (61 FAIL records total), rather than 61 subtests. This table covers every leaf failure, splits the two docs examples and the two concurrent diagnostic failures into distinct rows, and collapses four identical goroutine reports per diagnostic. Top-level summaries with failing children are not double-counted.

## Category 4 — left unchanged

The five rows below represent two underlying production issues. No invalid argument was newly accepted by a changed test. “The Bob dog” is an unexpected certification across a documented grammar boundary; “Highly reliable machinery” is a production catalog/diagnostic disagreement. Both require investigation or edits outside the permitted test/doc scope.

## Complete classification

| Test / example | Category | Original argument (short) | Reason / disposition |
| --- | --- | --- | --- |
| TestFinalDescriptionAttribution | 4 | If W arrived, then W was seen; The Bob dog arrived; ∴ the Bob dog was seen. | Still expected to reject a name inside a description, but certifies. descriptionWords explicitly rejects evidenced names; the integrated result disagrees with that boundary. Left test unchanged; rule/name-evidence investigation is outside this brief. |
| TestStrictNotesConcurrentAndOwned / slot.description_shape (four goroutines) | 4 | If X is red, then X is warm; The Bob dog is red; ∴ the Bob dog is warm. | Production strictNotes.AlsoRefuses promises refusal, but this example certifies. The same description-name boundary as FinalDescriptionAttribution. Cannot replace the catalog example without editing prohibited non-test validation/strict_notes.go; test unchanged. |
| TestStrictNotesExamples/slot.description_shape | 4 | If X is red, then X is warm; The Bob dog is red; ∴ the Bob dog is warm. | Production catalog example certifies despite promising refusal. Same issue as above; test unchanged. |
| TestStrictNotesConcurrentAndOwned / opener.adverb (four goroutines) | 4 | If X is safe, then X is useful; Highly reliable machinery is safe; ∴ highly reliable machinery is useful. | Still refused by slot.value_singular, but production catalog promises opener.adverb and that note is absent. Repair requires production catalog or rule edits, both prohibited here; test unchanged. |
| TestStrictNotesExamples/opener.adverb | 4 | If X is safe, then X is useful; Highly reliable machinery is safe; ∴ highly reliable machinery is useful. | Same production catalog/diagnostic disagreement: slot.value_singular instead of opener.adverb. This tests the catalog itself; swapping its input in test code would conceal the inconsistency. Test unchanged. |
| TestAccuracyRound2ValidControls/schema | 1 | ∀x(Man(x) → Mortal(x)); Man(Socrates); ∴ Mortal(Socrates). | Instantiate the universal at Socrates and apply modus ponens. Moved to accepted controls. |
| TestAndOrGroupingPhraseLevelCoordination/adjective_or | 2 | If X results in U and U is a bad or catastrophic outcome, X is risky; corresponding Sam/Ron premise; ∴ Sam is risky. | Names now bind; mixed and/or grouping still refuses. Updated to scope.mixed_andor and grouping hint. |
| TestDocsRulesExamples / docs/rules.md:1922 property generalization | 3 | Socrates = Plato; arbitrary property P of Plato; ∴ Socrates has P; ∴ every property. | Replaced sentence-initial Socrates with unlisted Zorvan in this refusal example only; retains identity.constant_sides. |
| TestDocsRulesExamples / docs/rules.md:2806 description uniqueness | 3 | The author of Waverley is Scott; Blair is an author of Waverley; ∴ Blair is Scott. | Replaced sentence-initial Blair with unlisted Zorvan in this refusal example only; retains identity.constant_sides. |
| TestEntailmentGateCatalogV17/subject_EG | 1 | Greets(Alice,Bob); Student(Alice); ∴ ∃x(Student(x) ∧ Greets(x,Bob)). | Conjoin the two premises and introduce existential witness Alice. |
| TestCatalogV19/schematic_preserves_lowercase_constants | 1 | If X had struck x, then X would have lit; Alice struck x; ∴ Alice lit. | Under the existing centered counterfactual semantics, the rule entails ∀u(Struck(u,x) → Lit(u)). Instantiate u=Alice, keeping lowercase x fixed; modus ponens gives Lit(Alice). Restored the original counterfactual modus ponens form assertion. |
| TestCatalogV20/property_UG | 2 | Socrates = Plato; arbitrary property P of Plato; ∴ Socrates has P; ∴ every property. | Evidenced names pass constant checks; property predication still fails identity.extensional_context. Updated rule and rewrite hint. |
| TestCatalogV22/3-place_EI_slot_0 | 1 | Alice sends Parcel to Bob; Alice is a courier; ∴ some courier sends Parcel to Bob. | Conjoin Courier(Alice) with the relation and existentially quantify its subject, preserving every other argument. |
| TestCatalogV22/3-place_identity | 2 | Alice sends Parcel to Bob; Alice = Cedar; ∴ Cedar sends Parcel to Bob. | Names now pass; the non-copular relation still fails identity.extensional_context. Updated rule and hint. |
| TestCatalogV22/4-place_EI_slot_0 | 1 | Alice sends Parcel to Bob from Depot; Alice is a courier; ∴ some courier sends Parcel to Bob from Depot. | Conjoin Courier(Alice) with the relation and existentially quantify its subject, preserving every other argument. |
| TestCatalogV22/4-place_identity | 2 | Alice sends Parcel to Bob from Depot; Alice = Cedar; ∴ Cedar sends Parcel to Bob from Depot. | Names now pass; the non-copular relation still fails identity.extensional_context. Updated rule and hint. |
| TestCatalogV27Exercises/uniqueness | 1 | The author of Waverley is Scott; Blair is an author of Waverley; ∴ Blair = Scott. | The first premise gives a unique author equal to Scott. Blair satisfies the restrictor, so uniqueness gives Blair=Scott. Existing independent exhaustive domains 1–3 oracle passes. |
| TestCatalogV27Exercises/relative_three-place_substitution | 1 | The book that Ada gave Blair is Atlas; that book is red; ∴ Atlas is red. | Both definite descriptions identify the unique object satisfying the same book/gave restrictor. Equality substitution transfers red to Atlas. Existing independent exhaustive domains 1–3 oracle passes. |
| TestCatalogV28Schematic/6 | 1 | ∀x((condition0(x) ∧ … ∧ condition5(x)) → Celebrates(x)); every listed condition holds of Ada; ∴ Celebrates(Ada). | Instantiate at Ada, conjoin all supplied antecedents, then modus ponens. The inconsistent-name negative control remains false. |
| TestCatalogV28Schematic/7 | 1 | ∀x((condition0(x) ∧ … ∧ condition6(x)) → Celebrates(x)); every listed condition holds of Ada; ∴ Celebrates(Ada). | Instantiate at Ada, conjoin all supplied antecedents, then modus ponens. The inconsistent-name negative control remains false. |
| TestCatalogV28Schematic/8 | 1 | ∀x((condition0(x) ∧ … ∧ condition7(x)) → Celebrates(x)); every listed condition holds of Ada; ∴ Celebrates(Ada). | Instantiate at Ada, conjoin all supplied antecedents, then modus ponens. The inconsistent-name negative control remains false. |
| TestCatalogV28Schematic/9 | 1 | ∀x((condition0(x) ∧ … ∧ condition8(x)) → Celebrates(x)); every listed condition holds of Ada; ∴ Celebrates(Ada). | Instantiate at Ada, conjoin all supplied antecedents, then modus ponens. The inconsistent-name negative control remains false. |
| TestCatalogV28Schematic/10 | 1 | ∀x((condition0(x) ∧ … ∧ condition9(x)) → Celebrates(x)); every listed condition holds of Ada; ∴ Celebrates(Ada). | Instantiate at Ada, conjoin all supplied antecedents, then modus ponens. The inconsistent-name negative control remains false. |
| TestCatalogV28Schematic/U | 1 | If U satisfies condition0, condition1 and condition2, then U celebrates; Ada satisfies all three; ∴ Ada celebrates. | Rename the bound schema variable, instantiate at Ada, conjoin the three premises and apply modus ponens. Negative control retained. |
| TestCatalogV28Schematic/V | 1 | If V satisfies condition0, condition1 and condition2, then V celebrates; Ada satisfies all three; ∴ Ada celebrates. | Rename the bound schema variable, instantiate at Ada, conjoin the three premises and apply modus ponens. Negative control retained. |
| TestCatalogV28Schematic/W | 1 | If W satisfies condition0, condition1 and condition2, then W celebrates; Ada satisfies all three; ∴ Ada celebrates. | Rename the bound schema variable, instantiate at Ada, conjoin the three premises and apply modus ponens. Negative control retained. |
| TestCatalogV28Schematic/same_category_and_from_Y | 1 | Persons Ada, Ben, Cy; Help(Ada,Ben), Help(Ben,Cy); rule for X,Y,Z; ∴ Thanks(Ada,Cy). | Uniformly instantiate X=Ada, Y=Ben, Z=Cy. All five conjuncts are premises, so modus ponens yields Thanks(Ada,Cy). Changed-object negative remains false. |
| TestCatalogV28Diagnostics | 2 | Six-condition rule; Ada satisfies conditions0–4, Ben condition5; ∴ Ada celebrates. | Still refused: condition5(Ada) is missing. Ada now binds X before failure, so diagnostic expects binding {X="ada"}, not {}. No acceptance expectation was loosened. |
| TestCatalogV28LegacyLettersAndNegation | 1 | If X smiles then X admires U; Ada smiles; ∴ Ada admires U (U/V/W); also If U smiles then U dances; Ada smiles; ∴ Ada dances. | First schema substitutes X=Ada while preserving the consequent letter (or its arbitrary valuation). Second substitutes the antecedent letter with Ada. Both are universal instantiation plus modus ponens. All six positives now checked; negative and inherited negation checks retained. |
| TestCatalogV28WrappedLengthCompatibility | 1 | Eleven wrapped antecedent conditions imply celebration; Ada satisfies each; ∴ Ada celebrates. | Each “it is the case that” wrapper preserves its proposition. Conjoin eleven premises and use the instantiated conditional. Inconsistent-name negative retained. |
| TestCatalogV8Coverage/schematic_three | 1 | Ada smiles, sings, dances; those three predicates imply celebration; ∴ Ada celebrates. | Instantiate at Ada, conjoin three premises, apply modus ponens. |
| TestEntailmentGateGapLedger/identity_transitivity | 3 | Ada=Ben; Ben=Cy; ∴ Ada=Cy. | Preserved the refusal purpose by replacing sentence-initial Ada with unlisted Zorvan; original rule and hint retained. |
| TestEntailmentGateGapLedger/existential_generalization | 3 | Zeus is mortal; ∴ something is mortal. | Preserved the refusal purpose by replacing sentence-initial Zeus with unlisted Zorvan; original rule and hint retained. |
| TestEntailmentGateGapLedger/schematic_instantiation | 3 | If X sings then X dances; Ada sings; ∴ Ada dances. | Preserved the refusal purpose by replacing sentence-initial Ada with unlisted Zorvan; original rule and hint retained. |
| TestEntailmentGateGapLedger/v20_property_UG | 3 | Socrates=Plato; arbitrary property of Plato; ∴ Socrates has every property of Plato. | Preserved the refusal purpose by replacing sentence-initial Socrates with unlisted Zorvan; original rule and hint retained. |
| TestEntailmentGateV19GapLedger/schematic_preserves_lowercase_constants | 3 | If X had struck x then X would have lit; Alice struck x; ∴ Alice lit. | This ledger keeps its refusal purpose: local copy replaces Alice with unlisted Zorvan; the shared catalog acceptance example remains Alice. Same slot.value_singular rule. |
| TestEntailmentGatePropertyOracle | 2 | Socrates=Plato; arbitrary P of Plato; ∴ Socrates has P; ∴ every property. | Updated to identity.extensional_context; independent model checks still execute for every candidate, including this refused proof. |
| TestAcceptedFixtureStepParity/every_parcel_in_a_crate | 2 | Ada is a courier; Ada carries every parcel in a crate; corresponding X schema; ∴ Ada is ready. | Ada now binds; the embedded clause beginning with X still triggers scope.clause_start. Updated only the refusal rule/hint; frozen fixture JSON is unaffected. Invalid substitution controls retained. |
| TestAcceptedFixtureStepParity/a_parcel_with_no_label | 2 | Ada is a courier; Ada carries a parcel with no label; corresponding X schema; ∴ Ada is ready. | Ada now binds; the embedded clause beginning with X still triggers scope.clause_start. Updated only the refusal rule/hint; frozen fixture JSON is unaffected. Invalid substitution controls retained. |
| TestEntailmentGateSlotBindingScope/name | 1 | Wise(Socrates); ∀x(Wise(x) → Happy(x)); ∴ Happy(Socrates). | Instantiate at Socrates, then modus ponens. Moved from refusal list to accepted controls; invalid scope controls unchanged. |
| TestEntailmentGateSlotBindingScope/repeated_name | 1 | Wise(Socrates), Rich(Socrates); ∀x((Wise(x) ∧ Rich(x)) → Happy(x)); ∴ Happy(Socrates). | Both antecedents hold of the same constant; conjunction and instantiated modus ponens establish the conclusion. |
| TestEntailmentGateV16AcceptedCompatibility/case_13 | 1 | Ada=Ben; Ben=Cy; ∴ Ada=Cy. | Equality is transitive. Examined frozen AST and form; removing the refusal restores the existing exact step/form parity assertion. No JSON update needed. |
| TestEntailmentGateV16AcceptedCompatibility/case_16 | 1 | Ada smiles, sings, dances; schema implies celebration; ∴ Ada celebrates. | Universal instantiation, conjunction of the three premises, modus ponens. Examined frozen AST and form; removing the refusal restores the existing exact step/form parity assertion. No JSON update needed. |
| TestEntailmentGateV16AcceptedCompatibility/case_927 | 1 | Zeus is mortal; ∴ something is mortal. | Existential introduction with witness Zeus. Examined frozen AST and form; removing the refusal restores the existing exact step/form parity assertion. No JSON update needed. |
| TestEntailmentGateV16AcceptedCompatibility/case_942 | 1 | If X sings then X dances; Ada sings; ∴ Ada dances. | Universal instantiation at Ada and modus ponens. Examined frozen AST and form; removing the refusal restores the existing exact step/form parity assertion. No JSON update needed. |
| TestFinalCounterfactualAttribution | 3 | If X had struck x then X would have lit; Alice struck x; ∴ Alice lit. | Replaced Alice with unlisted, sentence-initial Zorvan to retain slot.value_singular attribution. |
| TestDocsRulesStrictNotes | 3 | Generated strictness table (catalog descriptions and example text). | Regenerated docs/rules.md through the existing -update-strict-notes mechanism. No production catalog edit. The two broken catalog examples remain documented as category 4 above. |
| TestStrictNotesV45RefusalHits | 3 | descriptionWords([muddy, wooden, cart], false). | Swapped newly allowlisted adjectives for [brick, garden, wall], preserving the same slot.description_shape refusal. Other direct guard checks unchanged. |

## Validation

Used the requested isolated Go caches, local file GOPROXY, GOSUMDB=off, GOMAXPROCS=2, and GOFLAGS="-p=2 -modcacherw". Ran only the original failing top-level tests by an anchored name pattern; no ./... suite. Docs regeneration used the existing targeted update flag. Zorvan and Quellin were checked absent from names.txt. No production Go or fixture JSON changed.

Final targeted log: `workspace/pins-targeted-final.log`. The only remaining failures are the unchanged category-4 tests. A separate targeted verification excluding the three category-4 top-level tests is recorded in `workspace/pins-safe-final.log`.

Final command (from this worktree):

```sh
sh workspace/run-pins.sh -run "$(cat workspace/failing-pattern.txt)" -skip '^(TestFinalDescriptionAttribution|TestStrictNotesConcurrentAndOwned|TestStrictNotesExamples)$'
```

Last targeted run tail:

```text
ok  	github.com/bits-notes/english-arg-validator/validation	44.678s
```

The preceding all-original-failures run completed in 43.270s and failed only TestFinalDescriptionAttribution, TestStrictNotesConcurrentAndOwned, and TestStrictNotesExamples/{opener.adverb,slot.description_shape}. All 36 named failing subtests and 10 other leaf failure records are represented above. `git diff --check` passed.
