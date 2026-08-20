package profile

import "testing"

func TestNormalizeCanonicalUnicode(t *testing.T) {
	got, err := Normalize("  Cafe\u0301  ", 60, true)
	if err != nil {
		t.Fatal(err)
	}
	if got != "Café" {
		t.Fatalf("got %q", got)
	}
}

func TestNormalizeRejectsControlsAndCountsGraphemes(t *testing.T) {
	if _, err := Normalize("safe\nunsafe", 60, true); err == nil {
		t.Fatal("expected control rejection")
	}
	if _, err := Normalize("👩‍💻👩‍💻", 1, true); err == nil {
		t.Fatal("expected grapheme limit")
	}
}

func TestProfileRules(t *testing.T) {
	p, err := New("Friday", "", "Help me work", "balanced", "")
	if err != nil {
		t.Fatal(err)
	}
	if p.Identity.AddressUserAs != nil {
		t.Fatal("empty address must be null")
	}
	if _, err := New("Friday", "", "Help", "custom", ""); err == nil {
		t.Fatal("custom guidance required")
	}
}
