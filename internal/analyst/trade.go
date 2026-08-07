package analyst

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
    Entry domain.Trade
    Exit  domain.Trade
    Profit float64
}

func pairTrades(trades []domain.Trade) []ClosedTrade {
    var result []ClosedTrade
    var entryTrade *domain.Trade

    for _, t := range trades {
        if t.Side == domain.BuyOrder {
            tmp := t
            entryTrade = &tmp
        } else if t.Side == domain.SellOrder && entryTrade != nil {

            profit := (t.ExecutedPrice-entryTrade.ExecutedPrice)*
                t.Quantity -
                entryTrade.TransactionCost -
                t.TransactionCost

            result = append(result, ClosedTrade{
                Entry:  *entryTrade,
                Exit:   t,
                Profit: profit,
            })

            entryTrade = nil
        }
    }

    return result
}


func netProfit(close []ClosedTrade) float64 {
	total := 0.0
	for _, t := range close {
		total += t.Profit
	}
	return total
}

func winrate(close []ClosedTrade) float64 {
	if len(close) == 0 {
		return 0.0
	}

	win := 0
	for _, t := range close {
		if t.Profit > 0 {
			win++
		}
	}

	return float64(win) / float64(len(close))
}

func profitTrades(close []ClosedTrade) int {
	profitCount := 0
	for _, t := range close {
		if t.Profit > 0 {
			profitCount++
		}
	}
	return profitCount
}

func lossAvg(close []ClosedTrade) float64 {
	loss := 0.0
	lossCount := 0
	for _, t := range close {
		if t.Profit < 0 {
			loss += t.Profit
			lossCount++
		}
	}
	if lossCount == 0 {
		return 0.0
	}
	return loss / float64(lossCount)
}

func winAvg(close []ClosedTrade) float64 {
	win := 0.0
	winCount := 0
	for _, t := range close {
		if t.Profit > 0 {
			win += t.Profit
			winCount++
		}
	}
	if winCount == 0 {
		return 0.0
	}
	return win / float64(winCount)
}

func NewTradeStat(trades []domain.Trade) TradeStat {
	closedTrades := pairTrades(trades)
	return TradeStat{
		TotalTrades:      len(closedTrades),
		NetProfit:        netProfit(closedTrades),
		Winrate:            winrate(closedTrades),
		ProfitableTrades: profitTrades(closedTrades),
		Loss_Avg:           lossAvg(closedTrades),
		Win_Avg:            winAvg(closedTrades),
	}
}
