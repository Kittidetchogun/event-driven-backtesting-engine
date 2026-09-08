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

		costBasis := trade.ExecutedPrice*trade.Quantity + trade.TransactionCost
		averagePrice := costBasis / trade.Quantity

		position = domain.NewPosition(
			1,
			portfolioID,
			trade.Symbol,
			trade.Side,
			trade.Quantity,
			averagePrice,
			trade.ExecutedPrice,
		)

		positions[trade.Symbol] = position
		return
	}

	// ---------- Buy ----------
	if trade.Side == domain.BuyOrder {

		totalCost :=
			position.AveragePrice*position.Quantity +
				trade.ExecutedPrice*trade.Quantity +
				trade.TransactionCost

		position.Quantity += trade.Quantity

		position.AveragePrice =
			totalCost / position.Quantity

		// Update Current Price
		UpdatePrice(&position, trade.ExecutedPrice)

		// Recalculate Values
		UpdateMarketValue(&position)
		UpdateUnrealizedPnL(&position)

		positions[trade.Symbol] = position
		return
	}

	// ---------- Sell ----------
	UpdateRealizedPnL(&position, trade)

	position.Quantity -= trade.Quantity

	if position.Quantity <= 0 {

		delete(positions, trade.Symbol)
		return
	}

	UpdatePrice(&position, trade.ExecutedPrice)

	UpdateMarketValue(&position)
	UpdateUnrealizedPnL(&position)

	positions[trade.Symbol] = position
}
