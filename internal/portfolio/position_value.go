package portfolio

import "event-driven-backtesting-engine/internal/domain"

func UpdatePositionValue(
	portfolio *domain.Portfolio,
	positions map[string]domain.Position,
) {

	total := 0.0

	for _, position := range positions {
		total += position.CurrentValue
	}

	portfolio.PositionValue = total
}
