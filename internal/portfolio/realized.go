package portfolio

import "event-driven-backtesting-engine/internal/domain"

// UpdateRealizedPnL calculates realized profit/loss
// when a position is sold.

// Formula:
// (SellPrice - AveragePrice) × Quantity - TransactionCost
func UpdateRealizedPnL(
	position *domain.Position,
	trade domain.Trade,
) {

	if trade.Side != domain.SellOrder {
		return
	}

	position.RealizedPnL +=
		(trade.ExecutedPrice-position.AveragePrice)*
			trade.Quantity -
			trade.TransactionCost
}
