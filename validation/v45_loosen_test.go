package validation

import (
	"strings"
	"testing"
)

func TestV45Loosen(t *testing.T) {
	for _, v := range []string{"the game of chess", "the game of ordinary chess", "Carlsmith's APS concept", "the first ultraintelligent machine", "the group of robots", "the set of primes"} {
		text := "1. If X is solved, then X is boring.\n2. " + v + " is solved.\n3. Therefore, " + v + " is boring. [1,2]"
		t.Run(v, func(t *testing.T) {
			if r, res := CheckEntailmentGate(text); !r.Passed {
				t.Fatal(res.Message)
			}
		})
	}
	for _, pred := range []string{"informational or causal", "possible and financially feasible"} {
		text := "1. If X is " + pred + ", then X is a link.\n2. The channel is " + pred + ".\n3. Therefore, the channel is a link. [1,2]"
		t.Run(pred, func(t *testing.T) {
			if r, res := CheckEntailmentGate(text); !r.Passed {
				t.Fatal(res.Message)
			}
		})
	}
}
func TestV45Refuse(t *testing.T) {
	for _, v := range []string{"subgoals", "humanity", "the majority of voters", "the bulk of the evidence", "the rest of the robots", "the game of some chess", "the cost of the war of 1812", "the man who left", "the two robots", "Alice and Bob's dog", "the dog sees cat", "the number of robots"} {
		t.Run(v, func(t *testing.T) {
			text := "1. If X is dangerous, then X should be studied.\n2. " + v + " is dangerous.\n3. Therefore, " + v + " should be studied. [1,2]"
			if r, _ := CheckEntailmentGate(text); r.Passed {
				t.Fatal("certified", text)
			}
		})
	}
	p := allStrictRules()
	for _, s := range []string{"Every robot is safe or useful", "Robots are safe or useful", "The machine is possible and the plan is feasible", "The machine is possible and Bob is happy", "The machine is safe or fish", "The machine is safe or a criminal", "The machine is safe or they", "The machine is safe or two", "The machine is safe or general", "The machine is safe and useful or risky"} {
		if p.coordinatedPredicate(s) {
			t.Errorf("recognized unsafe coordination: %s", s)
		}
	}
	for _, text := range []string{
		"1. Every robot is safe or useful.\n2. The robot is a robot.\n3. Therefore, either the robot is safe or useful. [1,2]",
		"1. If X is safe or useful, then X is valuable.\n2. Robots are safe or useful.\n3. Therefore, robots are valuable. [1,2]",
	} {
		if r, _ := CheckEntailmentGate(text); r.Passed {
			t.Error("certified", text)
		}
	}
}
func TestV45MessagesAndCopula(t *testing.T) {
	text := "1. If X is dangerous, then X should be studied.\n2. Subgoals are dangerous.\n3. Therefore, subgoals should be studied. [1,2]"
	if r, res := CheckEntailmentGate(text); r.Passed || !strings.Contains(res.Message, "slot.value_singular") {
		t.Error(res.Message)
	}
	why := allStrictRules().scopeOperandFailure("perhaps the lamp is lit")
	if !strings.Contains(why, `"perhaps the lamp is lit"`) || !strings.Contains(why, `"perhaps"`) {
		t.Error(why)
	}
	for _, prefix := range []string{"the machine is", "the machine needs to be", "the machine need to be", "the machine has to be", "the machine have to be", "the machine had to be", "the machine must be", "the machine is being"} {
		if !afterCopula(prefix) {
			t.Error(prefix)
		}
	}
	for _, pred := range []string{"is being", "needs to be", "has to be", "must be"} {
		text := "1. If the machine " + pred + " X, then the channel is X.\n2. The machine " + pred + " ultraparallel.\n3. Therefore, the channel is ultraparallel. [1,2]"
		if r, res := CheckEntailmentGate(text); !r.Passed {
			t.Error(res.Message)
		}
	}
}

func TestV45Boundaries(t *testing.T) {
	p := allStrictRules()
	for w := range singularPartitiveWords {
		if p.singularValue("the "+w) || p.singularValue("the "+w+" of chess") {
			t.Error("quantity head accepted", w)
		}
	}
	for _, copula := range []string{"is", "are", "was", "were"} {
		if !p.coordinatedPredicate("the machine " + copula + " possible and financially feasible") {
			t.Error(copula)
		}
	}
	for _, text := range []string{
		"1. If X is possible and feasible, then X is useful.\n2. The machine is possible and the plan is feasible.\n3. Therefore, the machine is useful. [1,2]",
		"1. If the machine is being X, then X is useful.\n2. The machine is being ultraparallel.\n3. Therefore, ultraparallel is useful. [1,2]",
	} {
		if r, _ := CheckEntailmentGate(text); r.Passed {
			t.Error("unsafe accepted", text)
		}
	}
}
