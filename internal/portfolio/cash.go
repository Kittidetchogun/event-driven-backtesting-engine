package portfolio

import "event-driven-backtesting-engine/internal/domain"

// ApplyCashUpdate updates portfolio cash after a trade.
func ApplyCashUpdate(
	p *domain.Portfolio,
	trade domain.Trade,
) {

	switch trade.Side {

	case domain.BuyOrder:
		p.Cash -=
			trade.ExecutedPrice*trade.Quantity +
				trade.TransactionCost

	case domain.SellOrder:
		p.Cash +=
			trade.ExecutedPrice*trade.Quantity -
				trade.TransactionCost
	}
}
