package backtest

import (
	"testing"

	"event-driven-backtesting-engine/internal/domain"
)

type mockCandlePipeline struct {
	candles []domain.Candle
	index   int
}

func (p *mockCandlePipeline) Next() (domain.Candle, bool, error) {
	if p.index >= len(p.candles) {
		return domain.Candle{}, false, nil
	}

	candle := p.candles[p.index]
	p.index++

	return candle, true, nil
}

func TestRunnerRequiresPipeline(t *testing.T) {
	_, err := NewRunner(
		RunnerConfig{
			RunID:          1,
			StrategyName:   "EMA Cross",
			Symbol:         "BTCUSDT",
			Timeframe:      "1d",
			InitialCapital: 10000,
		},
		nil,
	)

	if err == nil {
		t.Fatal("expected error when pipeline is nil")
	}
}
