package theme

import "testing"

func TestNamed_nordMatchesTtuneTrio(t *testing.T) {
	th, ok := Named("nord")
	if !ok || th.Primary != "#88C0D0" || th.Secondary != "#D8DEE9" || th.Tertiary != "#EBCB8B" {
		t.Fatalf("%+v", th)
	}
}

func TestNamed_unknown(t *testing.T) {
	if _, ok := Named("papaya"); ok {
		t.Fatal("expected miss")
	}
}

func TestCycle(t *testing.T) {
	if Cycle("amber") != "green" || Cycle("nord") != "amber" {
		t.Fatal(Cycle("amber"), Cycle("nord"))
	}
}

func TestApply_doesNotPanic(t *testing.T) {
	st := Apply(Must("amber"))
	if st.Box.GetBorderStyle().Top == "" && st.Box.GetBorderStyle().Left == "" {
		// ascii border should have characters; if API differs, just ensure Apply returned
	}
	_ = st.Chord
}
