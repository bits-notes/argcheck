package validation

import "testing"

func TestSegmentSourceSentencesFullWidthTerminators(t *testing.T) {
	source := "すべての研究員は入館証を持っている。葵は研究員である。したがって、葵は入館証を持っている。"
	got := SegmentSourceSentences(source)
	if len(got) != 3 {
		t.Fatalf("got %d sentences, want 3: %+v", len(got), got)
	}
	runes := []rune(source)
	if want := "葵は研究員である。"; string(runes[got[1].Start:got[1].End]) != want {
		t.Fatalf("sentence 2 = %q, want %q", string(runes[got[1].Start:got[1].End]), want)
	}
	if n := len(SegmentSourceSentences("「行く。」次の文。")); n != 2 {
		t.Fatalf("closing bracket stays with its sentence: got %d sentences, want 2", n)
	}
	if n := len(SegmentSourceSentences("The rate is 2.1 today. It rose.")); n != 2 {
		t.Fatalf("ASCII behaviour changed: got %d sentences", n)
	}
}
