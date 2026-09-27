package statistics

import (
	"context"
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
)

type mockBacktestResultRepository struct {
	result *BacktestResult
}

func (m *mockBacktestResultRepository) SaveBacktestResult(
	ctx context.Context,
	result BacktestResult,
) error {
	m.result = &result
	return nil
}

func TestNewStatisticsEngine(t *testing.T) {

	engine := NewEngine()

	if engine == nil {
		t.Fatal("expected engine")
	}

	if len(engine.Snapshots()) != 0 {
		t.Fatalf(
			"expected 0 snapshots, got %d",
			len(engine.Snapshots()),
		)
	}
}

func TestStatisticsEngineConsume(t *testing.T) {

	engine := NewEngine()

	portfolio := domain.NewPortfolio(
		1,
		100000,
	)

	event := events.NewPortfolioUpdatedEvent(
		portfolio,
	)

	if err := engine.Consume(event); err != nil {
		t.Fatalf(
			"unexpected error %v",
			err,
		)
	}

	if len(engine.Snapshots()) != 1 {
		t.Fatalf(
			"expected 1 snapshot, got %d",
			len(engine.Snapshots()),
		)
	}

	snapshot := engine.Snapshots()[0]

	if snapshot.Cash != portfolio.Cash {
		t.Fatalf(
			"expected cash %.2f, got %.2f",
			portfolio.Cash,
			snapshot.Cash,
		)
	}

	if snapshot.Equity != portfolio.Equity {
		t.Fatalf(
			"expected equity %.2f, got %.2f",
			portfolio.Equity,
			snapshot.Equity,
		)
	}
}

func TestStatisticsEngineConsume_InvalidEvent(t *testing.T) {

	engine := NewEngine()

	candle := domain.Candle{
		Timestamp: time.Now(),
		Symbol:    "BTCUSDT",
		Timeframe: "1m",
		Open:      100,
		High:      110,
		Low:       90,
		Close:     105,
		Volume:    1000,
	}

	event := events.NewCandleReceivedEvent(candle)

	err := engine.Consume(event)

	if err == nil {
		t.Fatal("expected error")
	}
}

func TestStatisticsEngineAppendMultipleSnapshots(t *testing.T) {

	engine := NewEngine()

	values := []float64{
		100000,
		101000,
		99000,
	}

	for index, equity := range values {

		portfolio := domain.NewPortfolio(
			1,
			100000,
		)

		portfolio.UpdatedAt = time.Unix(int64(index), 0).UTC()
		portfolio.Equity = equity

		event := events.NewPortfolioUpdatedEvent(
			portfolio,
		)

		if err := engine.Consume(event); err != nil {
			t.Fatal(err)
		}
	}

	if len(engine.Snapshots()) != len(values) {
		t.Fatalf(
			"expected %d snapshots, got %d",
			len(values),
			len(engine.Snapshots()),
		)
	}
}

func TestStatisticsEngineSnapshots_ReplaceDuplicateAndPreserveOrder(t *testing.T) {
	engine := NewEngine()
	first := time.Unix(1000, 0).UTC()
	second := first.Add(time.Minute)

	portfolios := []struct {
		timestamp time.Time
		equity    float64
	}{
		{first, 100},
		{second, 200},
		{first, 300},
	}

	for _, item := range portfolios {
		portfolio := domain.NewPortfolio(1, 1000)
		portfolio.UpdatedAt = item.timestamp
		portfolio.Equity = item.equity
		if err := engine.Consume(events.NewPortfolioUpdatedEvent(portfolio)); err != nil {
			t.Fatal(err)
		}
	}

	snapshots := engine.Snapshots()
	if len(snapshots) != 2 {
		t.Fatalf("expected 2 snapshots, got %d", len(snapshots))
	}
	if !snapshots[0].Time.Equal(first) || snapshots[0].Equity != 300 {
		t.Fatalf("duplicate snapshot was not replaced with latest values: %+v", snapshots[0])
	}
	if !snapshots[1].Time.Equal(second) || snapshots[1].Equity != 200 {
		t.Fatalf("unique snapshot order or values changed: %+v", snapshots[1])
	}
}

func TestStatisticsEngineEquityCurve(t *testing.T) {

	engine := NewEngine()

	values := []float64{
		100000,
		101000,
		99500,
		105000,
	}

	for index, equity := range values {

		portfolio := domain.NewPortfolio(
			1,
			100000,
		)

		portfolio.UpdatedAt = time.Unix(int64(index), 0).UTC()
		portfolio.Equity = equity

		event := events.NewPortfolioUpdatedEvent(
			portfolio,
		)

		if err := engine.Consume(event); err != nil {
			t.Fatal(err)
		}
	}

	curve := engine.EquityCurve()

	if len(curve) != len(values) {
		t.Fatalf(
			"expected %d values, got %d",
			len(values),
			len(curve),
		)
	}

	for i := range values {

		if curve[i] != values[i] {
			t.Fatalf(
				"expected %.2f, got %.2f",
				values[i],
				curve[i],
			)
		}
	}
}

func TestStatisticsEngineAppendTrade(t *testing.T) {
	engine := NewEngine()

	now := time.Unix(1000, 0)

	trade := domain.NewTrade(
		1,
		1,
		11,
		"BTCUSDT",
		domain.BuyOrder,
		1,
		100,
		1,
		now,
	)

	event := events.NewTradeExecutedEvent(trade)

	err := engine.Consume(event)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	trades := engine.Trades()

	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}

	if trades[0].TradeID != trade.TradeID {
		t.Fatalf(
			"expected trade id %d, got %d",
			trade.TradeID,
			trades[0].TradeID,
		)
	}
}

func TestStatisticsEngineTradeStats(t *testing.T) {
	engine := NewEngine()

	now := time.Unix(1000, 0)

	buy := domain.NewTrade(
		1,
		1,
		11,
		"BTCUSDT",
		domain.BuyOrder,
		1,
		100,
		1,
		now,
	)

	sell := domain.NewTrade(
		2,
		1,
		12,
		"BTCUSDT",
		domain.SellOrder,
		1,
		110,
		1,
		now.Add(time.Minute),
	)

	if err := engine.Consume(events.NewTradeExecutedEvent(buy)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := engine.Consume(events.NewTradeExecutedEvent(sell)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stat := engine.TradeStats()

	if stat.TotalTrades != 1 {
		t.Fatalf(
			"expected 1 total trade, got %d",
			stat.TotalTrades,
		)
	}

	if stat.ProfitableTrades != 1 {
		t.Fatalf(
			"expected 1 profitable trade, got %d",
			stat.ProfitableTrades,
		)
	}

	if stat.NetProfit != 8 {
		t.Fatalf(
			"expected net profit 8, got %.2f",
			stat.NetProfit,
		)
	}

	if stat.Winrate != 1 {
		t.Fatalf(
			"expected win rate 1, got %.2f",
			stat.Winrate,
		)
	}

	if stat.Win_Avg != 8 {
		t.Fatalf(
			"expected average win 8, got %.2f",
			stat.Win_Avg,
		)
	}
}

func TestStatisticsEngineConsume_BacktestCompletedEvent_SavesResult(t *testing.T) {
	repository := &mockBacktestResultRepository{}

	engine := NewEngine()
	engine.SetRepository(repository)

	startDate := time.Date(
		2024, 1, 1,
		0, 0, 0, 0,
		time.UTC,
	)

	endDate := time.Date(
		2024, 1, 31,
		0, 0, 0, 0,
		time.UTC,
	)

	event := events.NewBacktestCompletedEvent(
		1,
		"EMA Cross",
		"BTCUSDT",
		"1h",
		startDate,
		endDate,
		0.15,
		0.60,
		1.25,
		-0.10,
	)

	err := engine.Consume(event)
	if err != nil {
		t.Fatalf("Consume() error = %v", err)
	}

	if repository.result == nil {
		t.Fatal("expected backtest result to be saved")
	}

	if repository.result.RunID != 1 {
		t.Errorf(
			"RunID = %d, expected 1",
			repository.result.RunID,
		)
	}

	if repository.result.TotalReturn != 0.15 {
		t.Errorf(
			"TotalReturn = %f, expected 0.15",
			repository.result.TotalReturn,
		)
	}
}

func TestStatisticsEngineConsume_BacktestCompletedEvent_RequiresRepository(
	t *testing.T,
) {
	engine := NewEngine()

	event := events.NewBacktestCompletedEvent(
		1,
		"EMA Cross",
		"BTCUSDT",
		"1h",
		time.Now(),
		time.Now(),
		0.10,
		0.50,
		1.20,
		-0.05,
	)

	err := engine.Consume(event)

	if err == nil {
		t.Fatal("expected error when repository is not configured")
	}
}
