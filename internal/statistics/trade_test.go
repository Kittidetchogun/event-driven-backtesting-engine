package statistics

import (
	"math"
	"testing"
	"time"

	"event-driven-backtesting-engine/internal/domain"
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

	closed := pairTrades(trades)

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

	stat := NewTradeStat(trades)

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
	stat := NewTradeStat(nil)

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

	closed := pairTrades(trades)

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
