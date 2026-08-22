package sizing

import (
	"math"
	"testing"
)

func TestFixedFractional_Size(t *testing.T) {
	sizer := NewFixedFractional(0.10)

	got := sizer.Size(10_000, 4_000)

	want := 0.25

	if math.Abs(got-want) > 1e-9 {
		t.Errorf(
			"Size() = %.10f, want %.10f",
			got,
			want,
		)
	}
}

func TestFixedFractional_Size_InvalidEquity(t *testing.T) {
	sizer := NewFixedFractional(0.10)

	got := sizer.Size(0, 4_000)

	if got != 0 {
		t.Errorf(
			"Size() = %.10f, want 0",
			got,
		)
	}
}

func TestFixedFractional_Size_InvalidPrice(t *testing.T) {
	sizer := NewFixedFractional(0.10)

	got := sizer.Size(10_000, 0)

	if got != 0 {
		t.Errorf(
			"Size() = %.10f, want 0",
			got,
		)
	}
}

func TestFixedFractional_Size_InvalidFraction(t *testing.T) {
	sizer := NewFixedFractional(0)

	got := sizer.Size(10_000, 4_000)

	if got != 0 {
		t.Errorf(
			"Size() = %.10f, want 0",
			got,
		)
	}
}
