package domain

import "time"

type Portfolio struct {
	RunID int

	InitCapital float64
	Cash float64
	Equity float64

	PositionValue float64

	UnrealizedPnL float64
	RealizedPnL   float64

	UpdatedAt time.Time
}

func NewPortfolio(
	runID int,
	initialCapital float64,
) Portfolio {

	now := time.Now()

	return Portfolio{
		RunID: runID,

		InitCapital: initialCapital,
		Cash:   initialCapital,
		Equity: initialCapital,

		PositionValue: 0,

		UnrealizedPnL: 0,
		RealizedPnL:   0,

		UpdatedAt: now,
	}
}

func (p *Portfolio) UpdateTimestamp(t time.Time) {
	p.UpdatedAt = t
}

func (p *Portfolio) UpdateEquity() {
	p.Equity =
		p.Cash +
			p.PositionValue +
			p.UnrealizedPnL
}

func (p *Portfolio) Reset() {

	p.Cash = p.InitCapital
	p.Equity = p.InitCapital

	p.PositionValue = 0

	p.UnrealizedPnL = 0
	p.RealizedPnL = 0

	p.UpdatedAt = time.Now()
}
