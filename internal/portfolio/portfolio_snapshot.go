package portfolio

import (
	"time"

	"event-driven-backtesting-engine/internal/domain"
)

type PortfolioSnapshot struct {
	Time time.Time

	Cash float64
	Equity float64

	PositionValue float64

	UnrealizedPnL float64
	RealizedPnL float64
}

//History ของ Portfolio
func NewPortfolioSnapshot(
	p domain.Portfolio,
) PortfolioSnapshot {

	return PortfolioSnapshot{
		Time: p.UpdatedAt,

		Cash: p.Cash,
		Equity: p.Equity,

		PositionValue: p.PositionValue,

		UnrealizedPnL: p.UnrealizedPnL,
		RealizedPnL: p.RealizedPnL,
	}
}
