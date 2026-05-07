package payment

import "testing"

func TestYuanToFenRejectsSubFenPrecision(t *testing.T) {
	if _, err := YuanToFen("10.009"); err == nil {
		t.Fatalf("expected error for amount with more than 2 decimal places")
	}
}

func TestHasAtMostCents(t *testing.T) {
	if !HasAtMostCents(10.01) {
		t.Fatalf("expected 10.01 to be accepted")
	}
	if HasAtMostCents(10.009) {
		t.Fatalf("expected 10.009 to be rejected")
	}
}
