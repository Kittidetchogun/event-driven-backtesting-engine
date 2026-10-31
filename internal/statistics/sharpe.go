package statistics

import "math"

// SharpeRatio calculates the Sharpe Ratio from an equity curve.
// Assumes a risk-free rate of 0.
func SharpeRatio(equityCurve []float64) float64 {

	returns := Returns(equityCurve)

	if len(returns) < 2 {
		return 0
	}

	stdDev := StdDev(returns)

	if stdDev == 0 || math.IsNaN(stdDev) || math.IsInf(stdDev, 0) {
		return 0
	}

	sharpe := Mean(returns) / stdDev
	if math.IsNaN(sharpe) || math.IsInf(sharpe, 0) {
		return 0
	}
	return sharpe
}
