package domain

import (
	"testing"
	"time"
)

func TestOrderFill(t *testing.T) {
	order := NewOrder(
		1,
		"BTCUSDT",
		BuyOrder,
		1,
		50000,
		time.Now(),
	)

	filledAt := time.Now()

	if err := order.Fill(filledAt); err != nil {
		t.Fatalf("Fill() returned error: %v", err)
	}

	if order.Status != FilledOrder {
		t.Fatalf("expected status %q, got %q", FilledOrder, order.Status)
	}

	if order.FilledAt == nil {
		t.Fatal("expected FilledAt to be set")
	}

	if !order.FilledAt.Equal(filledAt) {
		t.Fatal("FilledAt timestamp mismatch")
	}
}

func TestOrderCancel(t *testing.T) {
	order := NewOrder(
		1,
		"BTCUSDT",
		BuyOrder,
		1,
		50000,
		time.Now(),
	)

	if err := order.Cancel(); err != nil {
		t.Fatalf("Cancel() returned error: %v", err)
	}

	if order.Status != CancelledOrder {
		t.Fatalf("expected status %q, got %q", CancelledOrder, order.Status)
	}

	if order.FilledAt != nil {
		t.Fatal("FilledAt should remain nil after cancel")
	}
}

func TestOrderCannotFillTwice(t *testing.T) {
	order := NewOrder(
		1,
		"BTCUSDT",
		BuyOrder,
		1,
		50000,
		time.Now(),
	)

	if err := order.Fill(time.Now()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := order.Fill(time.Now()); err == nil {
		t.Fatal("expected error when filling an already filled order")
	}
}

func TestOrderCannotCancelFilledOrder(t *testing.T) {
	order := NewOrder(
		1,
		"BTCUSDT",
		BuyOrder,
		1,
		50000,
		time.Now(),
	)

	if err := order.Fill(time.Now()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := order.Cancel(); err == nil {
		t.Fatal("expected error when cancelling a filled order")
	}
}