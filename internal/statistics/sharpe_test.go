package statistics

import (
	"math"
	"testing"
)

func TestSharpeRatio(t *testing.T) {

	equityCurve := []float64{
		100,
		120,
		108,
	}

	got := SharpeRatio(equityCurve)

	expected := 0.23570226039551584

	if math.Abs(got-expected) > 1e-12 {
		t.Fatalf(
			"expected sharpe ratio %.15f, got %.15f",
			expected,
			got,
		)
	}
}

func TestSharpeRatio_InsufficientData(t *testing.T) {

	equityCurve := []float64{
		100,
	}

	got := SharpeRatio(equityCurve)

	expected := 0.0

	if got != expected {
		t.Fatalf(
			"expected sharpe ratio %.2f, got %.15f",
			expected,
			got,
		)
	}
}

func TestSharpeRatio_ZeroStandardDeviation(t *testing.T) {

	equityCurve := []float64{
		100,
		110,
		121,
	}

	got := SharpeRatio(equityCurve)

	expected := 0.0

	if got != expected {
		t.Fatalf(
			"expected sharpe ratio %.2f, got %.15f",
			expected,
			got,
		)
	}
}

func TestSharpeRatio_ZeroPreviousEquity(t *testing.T) {

	equityCurve := []float64{
		0,
		100,
	}

	got := SharpeRatio(equityCurve)

	expected := 0.0

	if got != expected {
		t.Fatalf(
			"expected sharpe ratio %.2f, got %.15f",
			expected,
			got,
		)
	}
}
