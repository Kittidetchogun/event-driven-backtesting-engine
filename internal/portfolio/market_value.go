package portfolio

import "event-driven-backtesting-engine/internal/domain"

func UpdateMarketValue(
    position *domain.Position,
) {
    position.CurrentValue = position.Quantity * position.CurrentPrice
}
