package backtest

import (
	"bytes"
	"strconv"
	"strings"
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/portfolio"
)

func TestRunnerEquityTraceUsesPostMarkPortfolioState(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	candles := make([]domain.Candle, 0, 25)

	for i := 0; i < 21; i++ {
		candles = append(candles, domain.Candle{
			Symbol:    "BTCUSDT",
			Timestamp: start.AddDate(0, 0, i),
			Open:      100,
			Close:     100,
		})
	}

	for i := 21; i < 25; i++ {
		candles = append(candles, domain.Candle{
			Symbol:    "BTCUSDT",
			Timestamp: start.AddDate(0, 0, i),
			Open:      200,
			Close:     200,
		})
	}

	runner, err := NewRunner(
		RunnerConfig{
			RunID:          1,
			StrategyName:   "EMA Cross",
			Symbol:         "BTCUSDT",
			Timeframe:      "1d",
			StartDate:      candles[0].Timestamp,
			EndDate:        candles[len(candles)-1].Timestamp,
			InitialCapital: 10000,
		},
		&fakeCandlePipeline{candles: candles},
		&fakeBacktestResultRepository{},
	)
	if err != nil {
		t.Fatalf("NewRunner() error = %v", err)
	}

	if _, err := runner.Run(); err != nil {
		t.Fatalf("Runner.Run() error = %v", err)
	}

	tradesBeforeExport := len(runner.matching.Trades())
	trace := runner.EquityTrace()
	if len(trace) != len(candles) {
		t.Fatalf("trace rows = %d, want %d", len(trace), len(candles))
	}

	for index, row := range trace {
		if !row.Timestamp.Equal(candles[index].Timestamp) {
			t.Errorf("row %d timestamp = %v, want %v", index, row.Timestamp, candles[index].Timestamp)
		}
		if row.Close != candles[index].Close {
			t.Errorf("row %d close = %.12g, want %.12g", index, row.Close, candles[index].Close)
		}

		snapshot := snapshotAt(runner.portfolio.Snapshots(), row.Timestamp)
		if row.Cash != snapshot.Cash || row.PositionValue != snapshot.PositionValue ||
			row.RealizedPnL != snapshot.RealizedPnL || row.UnrealizedPnL != snapshot.UnrealizedPnL ||
			row.Equity != snapshot.Equity {
			t.Errorf("row %d does not use the matching portfolio snapshot: %+v vs %+v", index, row, snapshot)
		}
	}

	if trace[22].PositionQuantity <= 0 {
		t.Fatal("expected position quantity after the pending BUY executes on candle 22")
	}
	if trace[21].PositionQuantity != 0 {
		t.Fatal("signal candle must be traced before its order executes")
	}

	var output bytes.Buffer
	if err := runner.WriteEquityTraceCSV(&output); err != nil {
		t.Fatalf("WriteEquityTraceCSV() error = %v", err)
	}

	rows := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(rows) != len(candles)+1 {
		t.Fatalf("CSV records = %d, want %d including header", len(rows), len(candles)+1)
	}
	if got := strings.Split(rows[0], ","); len(got) != 8 || got[0] != "timestamp" || got[7] != "equity" {
		t.Fatalf("unexpected CSV header: %v", got)
	}

	last := strings.Split(rows[len(rows)-1], ",")
	if last[0] != candles[len(candles)-1].Timestamp.Format(time.RFC3339Nano) {
		t.Errorf("last CSV timestamp = %q, want %q", last[0], candles[len(candles)-1].Timestamp.Format(time.RFC3339Nano))
	}
	if equity, err := strconv.ParseFloat(last[7], 64); err != nil || equity != trace[len(trace)-1].Equity {
		t.Errorf("last CSV equity = %q, want %.12g", last[7], trace[len(trace)-1].Equity)
	}

	if len(runner.matching.Trades()) != tradesBeforeExport {
		t.Fatal("equity export changed trading behavior")
	}
}

func snapshotAt(snapshots []portfolio.PortfolioSnapshot, timestamp time.Time) portfolio.PortfolioSnapshot {
	for _, snapshot := range snapshots {
		if snapshot.Time.Equal(timestamp) {
			return snapshot
		}
	}
	return portfolio.PortfolioSnapshot{}
}
