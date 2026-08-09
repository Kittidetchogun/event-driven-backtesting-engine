package statistics

import (
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
)

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

	for _, equity := range values {

		portfolio := domain.NewPortfolio(
			1,
			100000,
		)

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

func TestStatisticsEngineEquityCurve(t *testing.T) {

	engine := NewEngine()

	values := []float64{
		100000,
		101000,
		99500,
		105000,
	}

	for _, equity := range values {

		portfolio := domain.NewPortfolio(
			1,
			100000,
		)

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
