package portfolio

import "event-driven-backtesting-engine/internal/domain"

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
