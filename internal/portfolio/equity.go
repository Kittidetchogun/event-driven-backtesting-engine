package portfolio

import "event-driven-backtesting-engine/internal/domain"

// Equity = Cash + Position Value
func UpdateEquity(
	portfolio *domain.Portfolio,
) {

	portfolio.Equity =
		portfolio.Cash +
			portfolio.PositionValue
}
