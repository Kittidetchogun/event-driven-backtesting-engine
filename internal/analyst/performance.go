package analyst
import (
	"event-driven-backtesting-engine/internal/statistics"
)

type Performance struct {
	SharpeRatio float64
	MaxDrawdown  float64
	TotalReturn   float64
	stat *statistics.Engine

}
func (p *Performance) returns() []float64 {
	var returns []float64
	var snapshots = p.stat.Snapshots()
	for i := 1; i < len(snapshots); i++ {
		prevEquity := snapshots[i-1].Equity
		currEquity := snapshots[i].Equity
		returns = append(returns, (currEquity-prevEquity)/prevEquity)
	}
	return returns
}
func (p *Performance) sharpeRatio() float64 {
	returns := p.returns()
	if len(returns) == 0 {
		return 0
	}
	return statistics.Mean(returns) / statistics.StdDev(returns)
}
func (p *Performance) maxDrawdown() float64 {
	var maxDrawdown float64
	var peak float64
	var snapshots = p.stat.Snapshots()
	for _, snapshot := range snapshots {
		if snapshot.Equity > peak {
			peak = snapshot.Equity
		}
		drawdown := (peak - snapshot.Equity) / peak
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}
	return maxDrawdown
}