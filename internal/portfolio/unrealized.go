package portfolio

import "event-driven-backtesting-engine/internal/domain"

func UpdateUnrealizedPnL(
    position *domain.Position,
) {
    position.UnrealizedPnL =
        (position.CurrentPrice-position.AveragePrice) * position.Quantity
}
