package portfolio

import "event-driven-backtesting-engine/internal/domain"

func UpdatePrice(
	position *domain.Position,
	price float64,
) {
	position.CurrentPrice = price
}
