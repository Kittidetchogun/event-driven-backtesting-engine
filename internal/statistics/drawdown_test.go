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

func TestMaxDrawdown_ZeroAndNegativeEquity(t *testing.T) {
	if got := MaxDrawdown([]float64{100, 0, 50}); math.Abs(got-1) > 1e-12 {
		t.Fatalf("drawdown at zero = %.12f, want 1", got)
	}
	if got := MaxDrawdown([]float64{-100, -200}); got != 0 {
		t.Fatalf("negative-only drawdown = %.12f, want 0", got)
	}
	if got := MaxDrawdown([]float64{100, -10}); got != 1 {
		t.Fatalf("negative equity drawdown = %.12f, want capped 1", got)
	}
}

func TestMaxDrawdown_NonFiniteEquityDoesNotPropagate(t *testing.T) {
	for _, equity := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		got := MaxDrawdown([]float64{100, equity, 90})
		if math.IsNaN(got) || math.IsInf(got, 0) {
			t.Fatalf("MaxDrawdown with %v = %v, want finite result", equity, got)
		}
	}
}
