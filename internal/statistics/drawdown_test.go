package statistics

import (
	"math"
	"testing"
)

func TestMaxDrawdown(t *testing.T) {
	equityCurve := []float64{
		100,
		120,
		90,
		150,
	}

	got := MaxDrawdown(equityCurve)

	expected := 0.25

	if math.Abs(got-expected) > 1e-12 {
		t.Fatalf(
			"expected max drawdown %.2f, got %.15f",
			expected,
			got,
		)
	}
}

func TestMaxDrawdown_NoDrawdown(t *testing.T) {
	equityCurve := []float64{
		100,
		110,
		120,
		130,
	}

	got := MaxDrawdown(equityCurve)

	expected := 0.0

	if math.Abs(got-expected) > 1e-12 {
		t.Fatalf(
			"expected max drawdown %.2f, got %.15f",
			expected,
			got,
		)
	}
}

func TestMaxDrawdown_DecliningEquity(t *testing.T) {
	equityCurve := []float64{
		100,
		80,
		60,
	}

	got := MaxDrawdown(equityCurve)

	expected := 0.40

	if math.Abs(got-expected) > 1e-12 {
		t.Fatalf(
			"expected max drawdown %.2f, got %.15f",
			expected,
			got,
		)
	}
}

func TestMaxDrawdown_EmptyCurve(t *testing.T) {
	equityCurve := []float64{}

	got := MaxDrawdown(equityCurve)

	expected := 0.0

	if math.Abs(got-expected) > 1e-12 {
		t.Fatalf(
			"expected max drawdown %.2f, got %.15f",
			expected,
			got,
		)
	}
}
