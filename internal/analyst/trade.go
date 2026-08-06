package analyst

import "event-driven-backtesting-engine/internal/domain"

type TradeStat struct {
	TotalTrades      int
	NetProfit        float64
	Winrate          float64
	ProfitableTrades int
	Loss_Avg         float64
	Win_Avg          float64
}

func profit(trades []domain.Trade, position domain.Position) float64 {
	calProfit := 0.0

	for _, trade := range trades {
		calProfit += profitTrade(trade)
	}

	return calProfit
}

func netProfit(trades []domain.Trade) float64 {
	total := 0.0
	for _, t := range trades {
		total += profitTrade(t)
	}
	return total
}

func winrate(trades []domain.Trade) float64 {
	if len(trades) == 0 {
		return 0.0
	}

	win := 0
	for _, t := range trades {
		if profitTrade(t) > 0 {
			win++
		}
	}

	return float64(win) / float64(len(trades))
}

func profitTrades(trades []domain.Trade) int {
	profitCount := 0
	for _, t := range trades {
		if profitTrade(t) > 0 {
			profitCount++
		}
	}
	return profitCount
}

func lossAvg(trades []domain.Trade) float64 {
	loss := 0.0
	lossCount := 0
	for _, t := range trades {
		tradeProfit := profitTrade(t)
		if tradeProfit < 0 {
			loss += tradeProfit
			lossCount++
		}
	}
	if lossCount == 0 {
		return 0.0
	}
	return loss / float64(lossCount)
}

func winAvg(trades []domain.Trade) float64 {
	win := 0.0
	winCount := 0
	for _, t := range trades {
		tradeProfit := profitTrade(t)
		if tradeProfit > 0 {
			win += tradeProfit
			winCount++
		}
	}
	if winCount == 0 {
		return 0.0
	}
	return win / float64(winCount)
}

func profitTrade(trade domain.Trade) float64 {
	profit := trade.ExecutedPrice * trade.Quantity
	if trade.Side == domain.BuyOrder {
		profit = -profit
	}

	return profit - trade.TransactionCost
}
