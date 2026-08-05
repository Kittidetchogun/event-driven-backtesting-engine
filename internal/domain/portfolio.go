package domain

// Portfolio represents the current portfolio state during a backtest.
type Portfolio struct {
	RunID int
	InitCapital float64
	Cash float64
	Equity float64
	PositionValue float64
	UnrealizedPnL float64
	RealizedPnL float64
}

// NewPortfolio creates a new portfolio with the specified initial capital.
func NewPortfolio(
	runID int,
	initCapital float64,
) Portfolio {
	return Portfolio{
		RunID:          runID,
		InitCapital:    initCapital,
		Cash:           initCapital,
		Equity:         initCapital,
		PositionValue:  0,
		UnrealizedPnL:  0,
		RealizedPnL:    0,
	}
}

// Reset resets the portfolio back to its initial state.
func (p *Portfolio) Reset() {
	p.Cash = p.InitCapital
	p.Equity = p.InitCapital
	p.PositionValue = 0
	p.UnrealizedPnL = 0
	p.RealizedPnL = 0
}