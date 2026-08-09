package statistics

// SharpeRatio calculates the Sharpe Ratio from an equity curve.
// Assumes a risk-free rate of 0.
func SharpeRatio(equityCurve []float64) float64 {

	returns := Returns(equityCurve)

	if len(returns) == 0 {
		return 0
	}

	stdDev := StdDev(returns)

	if stdDev == 0 {
		return 0
	}

	return Mean(returns) / stdDev
}
