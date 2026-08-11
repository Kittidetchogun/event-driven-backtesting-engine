package statistics

import "time"

// Performance represents the final performance report
// of a backtest.
type Performance struct {
	NetProfit           float64
	WinRate             float64
	SharpeRatio         float64
	MaxDrawdown         float64
	TotalTrades         int
	WinningTrades       int
	LosingTrades        int
	AverageWinningTrade float64
	AverageLosingTrade  float64
}

// Performance returns the complete performance report.
func (e *Engine) Performance() Performance {

	tradeStat := e.TradeStats()
	equityCurve := e.EquityCurve()

	return Performance{
		NetProfit:           tradeStat.NetProfit,
		WinRate:             tradeStat.Winrate,
		SharpeRatio:         SharpeRatio(equityCurve),
		MaxDrawdown:         MaxDrawdown(equityCurve),
		TotalTrades:         tradeStat.TotalTrades,
		WinningTrades:       tradeStat.ProfitableTrades,
		LosingTrades:        tradeStat.TotalTrades - tradeStat.ProfitableTrades,
		AverageWinningTrade: tradeStat.Win_Avg,
		AverageLosingTrade:  tradeStat.Loss_Avg,
	}
}

func (e *Engine) NewBacktestResult(
	runID int,
	strategyName string,
	symbol string,
	timeframe string,
	startDate time.Time,
	endDate time.Time,
) BacktestResult {

	performance := e.Performance()
	equityCurve := e.EquityCurve()

	return BacktestResult{
		RunID:        runID,
		StrategyName: strategyName,
		Symbol:       symbol,
		Timeframe:    timeframe,
		StartDate:    startDate,
		EndDate:      endDate,

		TotalReturn: TotalReturn(equityCurve),
		WinRate:     performance.WinRate,
		SharpeRatio: performance.SharpeRatio,
		MaxDrawdown: performance.MaxDrawdown,
	}
}
