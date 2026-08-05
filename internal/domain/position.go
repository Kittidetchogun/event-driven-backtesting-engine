package domain

// Position represents an open position in the portfolio.
type Position struct {
	PositionID int
	PortfolioID int

	Symbol string
	Side   OrderSide

	Quantity float64

	AveragePrice float64
	CurrentPrice float64

	CurrentValue float64

	UnrealizedPnL float64
	RealizedPnL   float64
}

// NewPosition creates a new position.
func NewPosition(
	positionID int,
	portfolioID int,
	symbol string,
	side OrderSide,
	quantity float64,
	averagePrice float64,
	currentPrice float64,
) Position {

	position := Position{
		PositionID:  positionID,
		PortfolioID: portfolioID,
		Symbol:      symbol,
		Side:        side,
		Quantity:    quantity,
		AveragePrice: averagePrice,
		CurrentPrice: currentPrice,
	}

	position.UpdateMarketValue()

	return position
}

// UpdatePrice updates the market price and recalculates values.
func (p *Position) UpdatePrice(price float64) {
	p.CurrentPrice = price
	p.UpdateMarketValue()
}

// UpdateMarketValue recalculates market value and unrealized PnL.
func (p *Position) UpdateMarketValue() {
	p.CurrentValue = p.Quantity * p.CurrentPrice

	p.UnrealizedPnL =
	(p.CurrentPrice - p.AveragePrice) * p.Quantity
}

// Reset clears the position.
func (p *Position) Reset() {
	p.Quantity = 0
	p.AveragePrice = 0
	p.CurrentPrice = 0
	p.CurrentValue = 0
	p.UnrealizedPnL = 0
	p.RealizedPnL = 0
}
