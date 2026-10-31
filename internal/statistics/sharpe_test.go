package statistics

import (
	"math"
	"reflect"
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

func TestReturnsStopsAtInvalidIntervalsWithoutDiscardingPrefix(t *testing.T) {
	if got := Returns([]float64{100, 110, 0, 120}); !reflect.DeepEqual(got, []float64{0.1, -1}) {
		t.Fatalf("Returns() = %#v, want [0.1 -1]", got)
	}
	if got := Returns([]float64{100, -50, -25}); math.Abs(got[0]-(-1.5)) > 1e-12 || math.Abs(got[1]-(-0.5)) > 1e-12 {
		t.Fatalf("negative equity returns = %#v, want [-1.5 -0.5]", got)
	}
	for _, curve := range [][]float64{{100, math.NaN(), 120}, {100, math.Inf(1), 120}, {100, math.Inf(-1), 120}} {
		got := Returns(curve)
		if len(got) != 0 || (len(got) > 0 && (math.IsNaN(got[0]) || math.IsInf(got[0], 0))) {
			t.Fatalf("Returns(%v) = %#v, want no invalid returns", curve, got)
		}
	}
}

func TestSharpeRatioInvalidInputsRemainFinite(t *testing.T) {
	for _, curve := range [][]float64{{100, math.NaN(), 120}, {100, math.Inf(1), 120}, {100, math.Inf(-1), 120}} {
		if got := SharpeRatio(curve); math.IsNaN(got) || math.IsInf(got, 0) {
			t.Fatalf("SharpeRatio(%v) = %v, want finite result", curve, got)
		}
	}
}
