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
	// อาจพิจารณาย้าย RealizedPnL ไปเก็บใน Portfolio แทนทีหลัง
	// เพราะ Position จะถูกลบเมื่อปิด Position ทั้งหมด
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

	return Position{
        PositionID:      positionID,
        PortfolioID:     portfolioID,
        Symbol:          symbol,
        Side:            side,
        Quantity:        quantity,
        AveragePrice:    averagePrice,
        CurrentPrice:    currentPrice,
        CurrentValue:    quantity * currentPrice,
        UnrealizedPnL:   (currentPrice - averagePrice) * quantity,
        RealizedPnL:     0,
    }
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
