package statistics

import (
	"errors"
	"math"
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
)

func TestPairTradesPairsBuyAndSell(t *testing.T) {
	buyTime := time.Unix(1000, 0)
	sellTime := time.Unix(2000, 0)

	trades := []domain.Trade{
		domain.NewTrade(
			1, 1, 11,
			"BTCUSDT",
			domain.BuyOrder,
			2,
			100,
			1.5,
			buyTime,
		),
		domain.NewTrade(
			2, 1, 12,
			"BTCUSDT",
			domain.SellOrder,
			2,
			112,
			1.0,
			sellTime,
		),
	}

	closed, err := pairTrades(trades)
	if err != nil {
		t.Fatal(err)
	}

	if len(closed) != 1 {
		t.Fatalf("expected 1 closed trade, got %d", len(closed))
	}

	trade := closed[0]

	if trade.Entry.Quantity != trades[0].Quantity {
		t.Fatalf(
			"expected entry quantity %.2f, got %.2f",
			trades[0].Quantity,
			trade.Entry.Quantity,
		)
	}

	if trade.Exit.TradeID != trades[1].TradeID {
		t.Fatalf(
			"expected exit trade id %d, got %d",
			trades[1].TradeID,
			trade.Exit.TradeID,
		)
	}

	// Average cost:
	//
	// Buy cost = 100 * 2 + 1.5
	// Average cost = 201.5 / 2 = 100.75
	//
	// Sell proceeds = 112 * 2 = 224
	// Sell fee = 1
	//
	// Profit = 224 - 201.5 - 1 = 21.5
	expectedProfit := 21.5

	if math.Abs(trade.Profit-expectedProfit) > 1e-9 {
		t.Fatalf(
			"expected profit %.2f, got %.2f",
			expectedProfit,
			trade.Profit,
		)
	}
}

func TestNewTradeStatAggregatesClosedTrades(t *testing.T) {
	now := time.Unix(1000, 0)

	trades := []domain.Trade{
		domain.NewTrade(
			1, 1, 11,
			"BTCUSDT",
			domain.BuyOrder,
			1,
			100,
			1,
			now,
		),
		domain.NewTrade(
			2, 1, 12,
			"BTCUSDT",
			domain.SellOrder,
			1,
			110,
			1,
			now.Add(time.Minute),
		),
		domain.NewTrade(
			3, 1, 13,
			"BTCUSDT",
			domain.BuyOrder,
			1,
			200,
			1,
			now.Add(2*time.Minute),
		),
		domain.NewTrade(
			4, 1, 14,
			"BTCUSDT",
			domain.SellOrder,
			1,
			200,
			1,
			now.Add(3*time.Minute),
		),
	}

	stat, err := NewTradeStat(trades)
	if err != nil {
		t.Fatal(err)
	}

	if stat.TotalTrades != 2 {
		t.Fatalf("expected 2 total trades, got %d", stat.TotalTrades)
	}

	if stat.ProfitableTrades != 1 {
		t.Fatalf(
			"expected 1 profitable trade, got %d",
			stat.ProfitableTrades,
		)
	}

	// Trade #1:
	// Buy 100 + fee 1 = cost 101
	// Sell 110 - fee 1 = proceeds 109
	// Profit = 8
	//
	// Trade #2:
	// Buy 200 + fee 1 = cost 201
	// Sell 200 - fee 1 = proceeds 199
	// Profit = -2
	//
	// Net = 8 + (-2) = 6
	if math.Abs(stat.NetProfit-6) > 1e-9 {
		t.Fatalf(
			"expected net profit 6, got %.2f",
			stat.NetProfit,
		)
	}

	if math.Abs(stat.Winrate-0.5) > 1e-9 {
		t.Fatalf(
			"expected winrate 0.5, got %.2f",
			stat.Winrate,
		)
	}

	if math.Abs(stat.Win_Avg-8) > 1e-9 {
		t.Fatalf(
			"expected win average 8, got %.2f",
			stat.Win_Avg,
		)
	}

	if math.Abs(stat.Loss_Avg-(-2)) > 1e-9 {
		t.Fatalf(
			"expected loss average -2, got %.2f",
			stat.Loss_Avg,
		)
	}
}

func TestNewTradeStatWithNoClosedTrades(t *testing.T) {
	stat, err := NewTradeStat(nil)
	if err != nil {
		t.Fatal(err)
	}

	if stat.TotalTrades != 0 {
		t.Fatalf("expected 0 total trades, got %d", stat.TotalTrades)
	}

	if stat.NetProfit != 0 {
		t.Fatalf("expected 0 net profit, got %.2f", stat.NetProfit)
	}

	if stat.Winrate != 0 {
		t.Fatalf("expected 0 winrate, got %.2f", stat.Winrate)
	}

	if stat.ProfitableTrades != 0 {
		t.Fatalf(
			"expected 0 profitable trades, got %d",
			stat.ProfitableTrades,
		)
	}

	if stat.Loss_Avg != 0 {
		t.Fatalf(
			"expected 0 loss average, got %.2f",
			stat.Loss_Avg,
		)
	}

	if stat.Win_Avg != 0 {
		t.Fatalf(
			"expected 0 win average, got %.2f",
			stat.Win_Avg,
		)
	}
}

func TestNewTradeStatClassifiesWinningLosingAndBreakevenTrades(t *testing.T) {
	now := time.Unix(1000, 0)
	sellPrices := []float64{120, 110, 95, 85, 100}
	trades := make([]domain.Trade, 0, len(sellPrices)*2)
	for index, sellPrice := range sellPrices {
		orderID := index*2 + 1
		trades = append(trades,
			domain.NewTrade(orderID, 1, orderID, "BTCUSDT", domain.BuyOrder, 1, 100, 0, now.Add(time.Duration(index*2)*time.Minute)),
			domain.NewTrade(orderID+1, 1, orderID+1, "BTCUSDT", domain.SellOrder, 1, sellPrice, 0, now.Add(time.Duration(index*2+1)*time.Minute)),
		)
	}

	stat, err := NewTradeStat(trades)
	if err != nil {
		t.Fatal(err)
	}
	if stat.TotalTrades != 5 || stat.ProfitableTrades != 2 || stat.LosingTrades != 2 || stat.BreakevenTrades != 1 {
		t.Fatalf("trade counts = total %d wins %d losses %d breakeven %d, want 5/2/2/1", stat.TotalTrades, stat.ProfitableTrades, stat.LosingTrades, stat.BreakevenTrades)
	}
	if math.Abs(stat.Winrate-0.4) > 1e-9 || math.Abs(stat.Win_Avg-15) > 1e-9 || math.Abs(stat.Loss_Avg-(-10)) > 1e-9 {
		t.Fatalf("trade metrics = winrate %.12f win avg %.12f loss avg %.12f, want 0.4/15/-10", stat.Winrate, stat.Win_Avg, stat.Loss_Avg)
	}
}

// Average Cost:
//
// Buy #1: 1 BTC @ 100
// Buy #2: 1 BTC @ 120
//
// Total cost = 100 + 120 = 220
// Total quantity = 2
// Average cost = 110
//
// Sell 1 BTC @ 150
//
// Profit = 150 - 110 = 40
//
// The remaining 1 BTC keeps the same average cost of 110.
func TestPairTradesUsesAverageCost(t *testing.T) {
	now := time.Unix(1000, 0)

	trades := []domain.Trade{
		domain.NewTrade(
			1, 1, 11,
			"BTCUSDT",
			domain.BuyOrder,
			1,
			100,
			0,
			now,
		),
		domain.NewTrade(
			2, 1, 12,
			"BTCUSDT",
			domain.BuyOrder,
			1,
			120,
			0,
			now.Add(time.Minute),
		),
		domain.NewTrade(
			3, 1, 13,
			"BTCUSDT",
			domain.SellOrder,
			1,
			150,
			0,
			now.Add(2*time.Minute),
		),
	}

	closed, err := pairTrades(trades)
	if err != nil {
		t.Fatal(err)
	}

	if len(closed) != 1 {
		t.Fatalf(
			"expected 1 closed trade, got %d",
			len(closed),
		)
	}

	// Average cost = (100 + 120) / 2 = 110
	// Profit = (150 - 110) * 1 = 40
	expectedProfit := 40.0

	if math.Abs(closed[0].Profit-expectedProfit) > 1e-9 {
		t.Fatalf(
			"expected profit %.2f, got %.2f",
			expectedProfit,
			closed[0].Profit,
		)
	}
}

func TestPairTradesRejectsInvalidTradeValues(t *testing.T) {
	invalidTrades := []domain.Trade{
		domain.NewTrade(1, 1, 1, "BTCUSDT", domain.BuyOrder, 0, 100, 0, time.Now()),
		domain.NewTrade(2, 1, 2, "BTCUSDT", domain.BuyOrder, -1, 100, 0, time.Now()),
		domain.NewTrade(3, 1, 3, "BTCUSDT", domain.BuyOrder, math.NaN(), 100, 0, time.Now()),
		domain.NewTrade(4, 1, 4, "BTCUSDT", domain.BuyOrder, 1, 0, 0, time.Now()),
		domain.NewTrade(5, 1, 5, "BTCUSDT", domain.BuyOrder, 1, math.Inf(1), 0, time.Now()),
		domain.NewTrade(6, 1, 6, "BTCUSDT", domain.BuyOrder, 1, 100, -1, time.Now()),
		domain.NewTrade(7, 1, 7, "BTCUSDT", domain.BuyOrder, 1, 100, math.NaN(), time.Now()),
		domain.NewTrade(8, 1, 8, "BTCUSDT", domain.OrderSide("INVALID"), 1, 100, 0, time.Now()),
	}
	for _, trade := range invalidTrades {
		if _, err := pairTrades([]domain.Trade{trade}); !errors.Is(err, ErrInvalidTrade) {
			t.Fatalf("trade %+v error = %v, want ErrInvalidTrade", trade, err)
		}
	}
}

func TestStatisticsRejectsInvalidSellWithoutMutatingTradeState(t *testing.T) {
	engine := NewEngine()
	now := time.Unix(1000, 0)
	buy := domain.NewTrade(1, 1, 1, "BTCUSDT", domain.BuyOrder, 1, 100, 0, now)
	if err := engine.Consume(events.NewTradeExecutedEvent(buy)); err != nil {
		t.Fatal(err)
	}

	for _, sell := range []domain.Trade{
		domain.NewTrade(2, 1, 2, "BTCUSDT", domain.SellOrder, 2, 110, 0, now.Add(time.Minute)),
		domain.NewTrade(3, 1, 3, "ETHUSDT", domain.SellOrder, 1, 110, 0, now.Add(time.Minute)),
	} {
		if err := engine.Consume(events.NewTradeExecutedEvent(sell)); err == nil {
			t.Fatalf("expected invalid sell to be rejected: %+v", sell)
		}
		if len(engine.Trades()) != 1 {
			t.Fatalf("invalid sell mutated stored trades: %d", len(engine.Trades()))
		}
	}
}

func TestPairTradesPreservesPartialSellFeeProrating(t *testing.T) {
	now := time.Unix(1000, 0)
	trades := []domain.Trade{
		domain.NewTrade(1, 1, 1, "BTCUSDT", domain.BuyOrder, 2, 100, 2, now),
		domain.NewTrade(2, 1, 2, "BTCUSDT", domain.SellOrder, 1, 110, 1, now.Add(time.Minute)),
	}
	closed, err := pairTrades(trades)
	if err != nil {
		t.Fatal(err)
	}
	// Average cost = 101; the valid partial position sale charges its sell fee.
	want := 110.0 - 101.0 - 1.0
	if len(closed) != 1 || math.Abs(closed[0].Profit-want) > 1e-12 {
		t.Fatalf("closed trades = %+v, want one profit %.12f", closed, want)
	}
}
