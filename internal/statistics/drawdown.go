package statistics

// MaxDrawdown calculates the maximum drawdown
// from an equity curve.
func MaxDrawdown(equityCurve []float64) float64 {

	var peak float64
	var maxDrawdown float64

	for _, equity := range equityCurve {

		if equity > peak {
			peak = equity
		}

		if peak == 0 {
			continue
		}

		drawdown := (peak - equity) / peak

		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown
}
