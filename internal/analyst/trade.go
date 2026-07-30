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
