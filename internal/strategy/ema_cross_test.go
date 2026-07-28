package strategy

import (
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
)

func TestEmaCrossInitialize(t *testing.T) {
	strategy := NewEmaCross(3, 5)

	if err := strategy.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v, want nil", err)
	}

	if !strategy.Ready() {
		t.Fatal("Ready() = false, want true after initialize with valid periods")
	}

	if got := strategy.CurrentSignal(); got != NoSignal {
		t.Fatalf("CurrentSignal() = %q, want %q", got, NoSignal)
	}
}

func TestEmaCrossInitializeInvalidPeriods(t *testing.T) {
	strategy := NewEmaCross(5, 3)

	if err := strategy.Initialize(); err == nil {
		t.Fatal("Initialize() error = nil, want error")
	}
}

func TestEmaCrossSignalCrossUpAndDown(t *testing.T) {
	strategy := NewEmaCross(2, 3)
	if err := strategy.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	candles := []domain.Candle{
		newCandle(1),
		newCandle(1),
		newCandle(1),
		newCandle(10),
		newCandle(1),
		newCandle(1),
	}

	var gotBuy, gotSell bool
	for _, candle := range candles {
		strategy.OnData(candle)
		switch strategy.CurrentSignal() {
		case BuySignal:
			gotBuy = true
		case SellSignal:
			gotSell = true
		}
	}

	if !gotBuy {
		t.Fatal("expected BuySignal from bullish crossover")
	}

	if !gotSell {
		t.Fatal("expected SellSignal from bearish crossover")
	}
}

func TestEmaCrossReset(t *testing.T) {
	strategy := NewEmaCross(2, 3)
	if err := strategy.Initialize(); err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	strategy.OnData(newCandle(1))
	strategy.OnData(newCandle(2))
	strategy.Reset()

	if got := strategy.CurrentSignal(); got != NoSignal {
		t.Fatalf("CurrentSignal() = %q, want %q", got, NoSignal)
	}

	if strategy.Fast == nil || strategy.Slow == nil {
		t.Fatal("expected EMAs to remain allocated after Reset")
	}

	if got := len(strategy.Fast.Values); got != 0 {
		t.Fatalf("len(Fast.Values) = %d, want 0", got)
	}

	if got := len(strategy.Slow.Values); got != 0 {
		t.Fatalf("len(Slow.Values) = %d, want 0", got)
	}
}

func newCandle(close float64) domain.Candle {
	return domain.Candle{
		Timestamp: time.Unix(0, 0),
		Close:     close,
	}
}
