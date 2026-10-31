package statistics

import "math"

// MaxDrawdown calculates the maximum drawdown
// from an equity curve. Negative-only curves return zero because
// there is no positive peak against which to define a percentage drawdown.
func MaxDrawdown(equityCurve []float64) float64 {

	var peak float64
	var maxDrawdown float64

	for _, equity := range equityCurve {
		if math.IsNaN(equity) || math.IsInf(equity, 0) {
			break
		}

		if equity > peak {
			peak = equity
		}

		if peak == 0 {
			continue
		}
		if equity < 0 {
			if maxDrawdown < 1 {
				maxDrawdown = 1
			}
			continue
		}

		drawdown := (peak - equity) / peak

		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown
}
