package statistics

import "time"

type BacktestResult struct {
	RunID        int
	StrategyName string
	Symbol       string
	Timeframe    string

	StartDate time.Time
	EndDate   time.Time

	TotalReturn float64
	WinRate     float64
	SharpeRatio float64
	MaxDrawdown float64
}
