package statistics

import "event-driven-backtesting-engine/internal/domain"

type TradeStat struct {
	TotalTrades      int
	NetProfit        float64
	Winrate          float64
	ProfitableTrades int
	Loss_Avg         float64
	Win_Avg          float64
}

type ClosedTrade struct {
	Entry  domain.Trade
	Exit   domain.Trade
	Profit float64
}

type openPosition struct {
	quantity    float64
	averageCost float64
}

func pairTrades(trades []domain.Trade) []ClosedTrade {
	var result []ClosedTrade

	positions := make(map[string]openPosition)

	for _, trade := range trades {

		// -------------------------
		// Buy
		// -------------------------
		if trade.Side == domain.BuyOrder {
			position := positions[trade.Symbol]

			totalCost :=
				position.averageCost*position.quantity +
					trade.ExecutedPrice*trade.Quantity +
					trade.TransactionCost

			position.quantity += trade.Quantity

			if position.quantity > 0 {
				position.averageCost =
					totalCost / position.quantity
			}

			positions[trade.Symbol] = position

			continue
		}

		// -------------------------
		// Sell
		// -------------------------
		if trade.Side != domain.SellOrder {
			continue
		}

		position, ok := positions[trade.Symbol]
		if !ok || position.quantity <= 0 {
			continue
		}

		matchQuantity := trade.Quantity

		// ป้องกันกรณีขายมากกว่า position ที่มี
		if matchQuantity > position.quantity {
			matchQuantity = position.quantity
		}

		// Cost basis ของ quantity ที่ขาย
		costBasis := position.averageCost * matchQuantity

		// รายรับจากการขาย
		proceeds := trade.ExecutedPrice * matchQuantity

		// BUY fee ถูกฝังอยู่ใน averageCost แล้ว
		// เหลือหักเฉพาะ SELL fee
		exitCost := 0.0

		if trade.Quantity > 0 {
			exitCost =
				trade.TransactionCost *
					(matchQuantity / trade.Quantity)
		}

		profit :=
			proceeds -
				costBasis -
				exitCost

		result = append(result, ClosedTrade{
			Entry: domain.Trade{
				Symbol:         trade.Symbol,
				Side:           domain.BuyOrder,
				Quantity:       matchQuantity,
				ExecutedPrice:  position.averageCost,
				TransactionCost: 0,
				ExecutedTime:  trade.ExecutedTime,
			},
			Exit:   trade,
			Profit: profit,
		})

		// ลด position
		position.quantity -= matchQuantity

		if position.quantity <= 1e-12 {
			delete(positions, trade.Symbol)
		} else {
			positions[trade.Symbol] = position
		}
	}

	return result
}

func netProfit(closedTrades []ClosedTrade) float64 {
	total := 0.0

	for _, trade := range closedTrades {
		total += trade.Profit
	}

	return total
}

func winrate(closedTrades []ClosedTrade) float64 {
	if len(closedTrades) == 0 {
		return 0
	}

	win := 0

	for _, trade := range closedTrades {
		if trade.Profit > 0 {
			win++
		}
	}

	return float64(win) / float64(len(closedTrades))
}

func profitTrades(closedTrades []ClosedTrade) int {
	count := 0

	for _, trade := range closedTrades {
		if trade.Profit > 0 {
			count++
		}
	}

	return count
}

func lossAvg(closedTrades []ClosedTrade) float64 {
	total := 0.0
	count := 0

	for _, trade := range closedTrades {
		if trade.Profit < 0 {
			total += trade.Profit
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return total / float64(count)
}

func winAvg(closedTrades []ClosedTrade) float64 {
	total := 0.0
	count := 0

	for _, trade := range closedTrades {
		if trade.Profit > 0 {
			total += trade.Profit
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return total / float64(count)
}

func NewTradeStat(trades []domain.Trade) TradeStat {
	closedTrades := pairTrades(trades)

	return TradeStat{
		TotalTrades:      len(closedTrades),
		NetProfit:        netProfit(closedTrades),
		Winrate:          winrate(closedTrades),
		ProfitableTrades: profitTrades(closedTrades),
		Loss_Avg:         lossAvg(closedTrades),
		Win_Avg:          winAvg(closedTrades),
	}
}
