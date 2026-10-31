package statistics

import (
	"errors"
	"fmt"
	"math"
)

var ErrInvalidEquityCurve = errors.New("invalid equity curve")

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
	BreakevenTrades     int
	AverageWinningTrade float64
	AverageLosingTrade  float64
}

// Performance returns the complete performance report.
func (e *Engine) Performance() (Performance, error) {
	equityCurve := e.EquityCurve()
	if err := validateEquityCurve(equityCurve); err != nil {
		return Performance{}, err
	}

	tradeStat, err := e.TradeStats()
	if err != nil {
		return Performance{}, err
	}
	return Performance{
		NetProfit:           tradeStat.NetProfit,
		WinRate:             tradeStat.Winrate,
		SharpeRatio:         SharpeRatio(equityCurve),
		MaxDrawdown:         MaxDrawdown(equityCurve),
		TotalTrades:         tradeStat.TotalTrades,
		WinningTrades:       tradeStat.ProfitableTrades,
		LosingTrades:        tradeStat.LosingTrades,
		BreakevenTrades:     tradeStat.BreakevenTrades,
		AverageWinningTrade: tradeStat.Win_Avg,
		AverageLosingTrade:  tradeStat.Loss_Avg,
	}, nil
}

func validateEquityCurve(equityCurve []float64) error {
	for index, equity := range equityCurve {
		if math.IsNaN(equity) || math.IsInf(equity, 0) {
			return fmt.Errorf("%w: non-finite equity at index %d", ErrInvalidEquityCurve, index)
		}
		if index < len(equityCurve)-1 && equity == 0 {
			return fmt.Errorf("%w: zero equity denominator at index %d", ErrInvalidEquityCurve, index)
		}
	}
	return nil
}
