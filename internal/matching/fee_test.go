package matching

import "testing"

func TestCalculateFee_Valid(t *testing.T) {
	fee := CalculateFee(100, 2, 0.001)

	expected := 0.2

	if fee != expected {
		t.Fatalf("expected fee %.4f, got %.4f", expected, fee)
	}
}

func TestCalculateFee_ZeroCommission(t *testing.T) {
	fee := CalculateFee(100, 2, 0)

	if fee != 0 {
		t.Fatalf("expected fee 0, got %.4f", fee)
	}
}

func TestCalculateFee_ZeroQuantity(t *testing.T) {
	fee := CalculateFee(100, 0, 0.001)

	if fee != 0 {
		t.Fatalf("expected fee 0, got %.4f", fee)
	}
}

func TestCalculateFee_ZeroPrice(t *testing.T) {
	fee := CalculateFee(0, 2, 0.001)

	if fee != 0 {
		t.Fatalf("expected fee 0, got %.4f", fee)
	}
}

func TestCalculateFee_NegativePrice(t *testing.T) {
	fee := CalculateFee(-100, 2, 0.001)

	if fee != 0 {
		t.Fatalf("expected fee 0, got %.4f", fee)
	}
}

func TestCalculateFee_NegativeQuantity(t *testing.T) {
	fee := CalculateFee(100, -2, 0.001)

	if fee != 0 {
		t.Fatalf("expected fee 0, got %.4f", fee)
	}
}

func TestCalculateFee_NegativeCommission(t *testing.T) {
	fee := CalculateFee(100, 2, -0.001)

	if fee != 0 {
		t.Fatalf("expected fee 0, got %.4f", fee)
	}
}
