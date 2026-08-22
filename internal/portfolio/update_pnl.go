package portfolio

import "event-driven-backtesting-engine/internal/domain"

// ApplyRealizedPnL updates portfolio-level realized PnL
// when a sell trade closes part or all of a position.
func ApplyRealizedPnL(
	p *domain.Portfolio,
	position domain.Position,
	trade domain.Trade,
) {
	if trade.Side != domain.SellOrder {
		return
	}

	p.RealizedPnL +=
		(trade.ExecutedPrice-position.AveragePrice)*
			trade.Quantity -
			trade.TransactionCost
}

func UpdatePortfolioUnrealizedPnL(
	p *domain.Portfolio,
	positions map[string]domain.Position,
) {
	total := 0.0

	for _, position := range positions {
		total += position.UnrealizedPnL
	}

	p.UnrealizedPnL = total
}
