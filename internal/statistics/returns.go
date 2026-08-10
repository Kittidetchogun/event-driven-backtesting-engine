package statistics

// Returns calculates period-to-period returns from an equity curve.
func Returns(equityCurve []float64) []float64 {

	if len(equityCurve) < 2 {
		return []float64{}
	}

	returns := make([]float64, 0, len(equityCurve)-1)

	for i := 1; i < len(equityCurve); i++ {

		if equityCurve[i-1] == 0 {
			return []float64{}
		}

		ret :=
			(equityCurve[i] - equityCurve[i-1]) /
				equityCurve[i-1]

		returns = append(returns, ret)
	}

	return returns
}

// TotalReturn calculates the total return of an equity curve.
func TotalReturn(equityCurve []float64) float64 {

	if len(equityCurve) < 2 {
		return 0
	}

	if equityCurve[0] == 0 {
		return 0
	}

	return (equityCurve[len(equityCurve)-1] - equityCurve[0]) /
		equityCurve[0]
}
