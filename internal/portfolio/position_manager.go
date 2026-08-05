package portfolio

import "event-driven-backtesting-engine/internal/domain"

// UpdatePosition creates, updates or removes a position after a trade.
func UpdatePosition(
	positions map[string]domain.Position,
	portfolioID int,
	trade domain.Trade,
) {

	position, exists := positions[trade.Symbol]

	// ---------- Create ----------
	if !exists && trade.Side == domain.BuyOrder {

		position = domain.NewPosition(
			1, // TODO: Generate PositionID
			portfolioID,
			trade.Symbol,
			trade.Side,
			trade.Quantity,
			trade.ExecutedPrice,
			trade.ExecutedPrice,
		)

		positions[trade.Symbol] = position
		return
	}

	// ---------- Buy ----------
	if trade.Side == domain.BuyOrder {

		totalCost :=
			position.AveragePrice*position.Quantity +
				trade.ExecutedPrice*trade.Quantity

		position.Quantity += trade.Quantity

		position.AveragePrice =
			totalCost / position.Quantity

		position.UpdatePrice(trade.ExecutedPrice)

		positions[trade.Symbol] = position
		return
	}

	// ---------- Sell ----------
	position.Quantity -= trade.Quantity

	if position.Quantity <= 0 {

		delete(positions, trade.Symbol)
		return
	}

	position.UpdatePrice(trade.ExecutedPrice)

	positions[trade.Symbol] = position
}
