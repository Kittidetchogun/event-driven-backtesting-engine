package analyst

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
		domain.NewTrade(1, 1, 11, "BTCUSDT", domain.BuyOrder, 2, 100, 1.5, buyTime),
		domain.NewTrade(2, 1, 12, "BTCUSDT", domain.SellOrder, 2, 112, 1.0, sellTime),
	}

	closed := pairTrades(trades)

	if len(closed) != 1 {
		t.Fatalf("expected 1 closed trade, got %d", len(closed))
	}

	trade := closed[0]
	if trade.Entry.TradeID != trades[0].TradeID {
		t.Fatalf("expected entry trade id %d, got %d", trades[0].TradeID, trade.Entry.TradeID)
	}
	if trade.Exit.TradeID != trades[1].TradeID {
		t.Fatalf("expected exit trade id %d, got %d", trades[1].TradeID, trade.Exit.TradeID)
	}

	expectedProfit := (112.0-100.0)*2 - 1.5 - 1.0
	if trade.Profit != expectedProfit {
		t.Fatalf("expected profit %.2f, got %.2f", expectedProfit, trade.Profit)
	}
}

func TestNewTradeStatAggregatesClosedTrades(t *testing.T) {
	now := time.Unix(1000, 0)

	trades := []domain.Trade{
		domain.NewTrade(1, 1, 11, "BTCUSDT", domain.BuyOrder, 1, 100, 1, now),
		domain.NewTrade(2, 1, 12, "BTCUSDT", domain.SellOrder, 1, 110, 1, now.Add(time.Minute)),
		domain.NewTrade(3, 1, 13, "BTCUSDT", domain.BuyOrder, 1, 200, 1, now.Add(2*time.Minute)),
		domain.NewTrade(4, 1, 14, "BTCUSDT", domain.SellOrder, 1, 190, 1, now.Add(3*time.Minute)),
	}

	stat := NewTradeStat(trades)

	if stat.TotalTrades != 2 {
		t.Fatalf("expected 2 total trades, got %d", stat.TotalTrades)
	}

	if stat.ProfitableTrades != 1 {
		t.Fatalf("expected 1 profitable trade, got %d", stat.ProfitableTrades)
	}

	if stat.NetProfit != 6 {
		t.Fatalf("expected net profit 6, got %.2f", stat.NetProfit)
	}

	if stat.Winrate != 0.5 {
		t.Fatalf("expected winrate 0.5, got %.2f", stat.Winrate)
	}

	if stat.Win_Avg != 8 {
		t.Fatalf("expected win average 8, got %.2f", stat.Win_Avg)
	}

	if stat.Loss_Avg != -2 {
		t.Fatalf("expected loss average -2, got %.2f", stat.Loss_Avg)
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
		t.Fatalf("expected 0 profitable trades, got %d", stat.ProfitableTrades)
	}
	if stat.Loss_Avg != 0 {
		t.Fatalf("expected 0 loss average, got %.2f", stat.Loss_Avg)
	}
	if stat.Win_Avg != 0 {
		t.Fatalf("expected 0 win average, got %.2f", stat.Win_Avg)
	}
}

func TestPairTradesUsesLatestBuyBeforeSell(t *testing.T) {
	now := time.Unix(1000, 0)

	trades := []domain.Trade{
		domain.NewTrade(1, 1, 11, "BTCUSDT", domain.BuyOrder, 1, 100, 0, now),
		domain.NewTrade(2, 1, 12, "BTCUSDT", domain.BuyOrder, 1, 120, 0, now.Add(time.Minute)),
		domain.NewTrade(3, 1, 13, "BTCUSDT", domain.SellOrder, 1, 150, 0, now.Add(2*time.Minute)),
	}

	closed := pairTrades(trades)

	if len(closed) != 1 {
		t.Fatalf("expected 1 closed trade, got %d", len(closed))
	}

	if closed[0].Entry.TradeID != 2 {
		t.Fatalf("expected latest buy to be used as entry, got trade id %d", closed[0].Entry.TradeID)
	}

	if math.Abs(closed[0].Profit-30) > 1e-9 {
		t.Fatalf("expected profit 30, got %.2f", closed[0].Profit)
	}
}
