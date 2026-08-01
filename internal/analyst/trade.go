package main

import "event-driven-backtesting-engine/internal/domain"

// type Trade struct {
// 	InitCapital float64
// 	Trades      []domain.Trade
// 	Equity      float64
// }

type TradeStat struct {
	TotalTrades int
	NetProfit   float64
	Winrate     float64
	ProfitableTrades int
	Loss_Avg float64
	Win_Avg float64
}

// func NewTrade(initCapital float64) *TradeStat {
// 	return &TradeStat{
// 		InitCapital: initCapital,
// 		Tradestat:      make([]domain.Trade, 0),
// 	}
// }

// func NetProfit(trades []domain.Trade) float64 {
// 	total := 0.0
// 	for _,t := range trades {
// 		total += domain.Profit(t)
// 	}
// 	return total
// }

func winrate(trades []domain.Trade) float64 {
	if len(trades) == 0 {
		return 0.0
	} else {
		win := 0
		for _, t := range trades {
			if domain.Profit(t) > 0 {
				win++
			}
		}
		return float64(win) / float64(len(trades))
	}
}

func profitTrades(trades []domain.Trade) int {
	profit := 0
	for _, t := range trades {
		if domain.Profit(t) > 0 {
			profit++
		}
	}
	return profit
}	

func lossAvg(trades []domain.Trade) float64 {
	Loss := 0.0
	LossCount := 0
	for _, t := range trades {
		if domain.Profit(t) < 0 {
			Loss += domain.Profit(t)
			LossCount++
		}
	}
	if LossCount == 0 {
		return 0.0
	}
	return Loss / float64(LossCount)
}

func winAvg(trades []domain.Trade) float64 {
	Win := 0.0
	WinCount := 0
	for _, t := range trades {
		if domain.Profit(t) > 0 {
			Win += domain.Profit(t)
			WinCount++
		}
	}
	if WinCount == 0 {
		return 0.0
	}
	return Win / float64(WinCount)
}
