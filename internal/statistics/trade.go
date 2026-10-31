package statistics

import (
	"errors"
	"fmt"
	"math"

	"event-driven-backtesting-engine/internal/domain"
)

var (
	ErrInvalidTrade          = errors.New("invalid trade")
	ErrNoTrackedPosition     = errors.New("no tracked position")
	ErrExcessiveSellQuantity = errors.New("sell quantity exceeds tracked position")
)

type TradeStat struct {
	TotalTrades      int
	NetProfit        float64
	Winrate          float64
	ProfitableTrades int
	LosingTrades     int
	BreakevenTrades  int
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

func validateTrade(trade domain.Trade) error {
	if trade.Symbol == "" || trade.Quantity <= 0 || math.IsNaN(trade.Quantity) || math.IsInf(trade.Quantity, 0) ||
		trade.ExecutedPrice <= 0 || math.IsNaN(trade.ExecutedPrice) || math.IsInf(trade.ExecutedPrice, 0) ||
		trade.TransactionCost < 0 || math.IsNaN(trade.TransactionCost) || math.IsInf(trade.TransactionCost, 0) {
		return fmt.Errorf("%w: invalid numeric data", ErrInvalidTrade)
	}
	if trade.Side != domain.BuyOrder && trade.Side != domain.SellOrder {
		return fmt.Errorf("%w: invalid side %q", ErrInvalidTrade, trade.Side)
	}
	return nil
}

func pairTrades(trades []domain.Trade) ([]ClosedTrade, error) {
	var result []ClosedTrade

	positions := make(map[string]openPosition)

	for _, trade := range trades {
		if err := validateTrade(trade); err != nil {
			return nil, err
		}

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
			return nil, fmt.Errorf("%w: symbol=%s", ErrNoTrackedPosition, trade.Symbol)
		}

		if trade.Quantity > position.quantity {
			return nil, fmt.Errorf("%w: symbol=%s required=%.8f available=%.8f", ErrExcessiveSellQuantity, trade.Symbol, trade.Quantity, position.quantity)
		}

		matchQuantity := trade.Quantity

		// ป้องกันกรณีขายมากกว่า position ที่มี
		if matchQuantity > position.quantity {
			return nil, fmt.Errorf("%w: symbol=%s required=%.8f available=%.8f", ErrExcessiveSellQuantity, trade.Symbol, trade.Quantity, position.quantity)
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
				Symbol:          trade.Symbol,
				Side:            domain.BuyOrder,
				Quantity:        matchQuantity,
				ExecutedPrice:   position.averageCost,
				TransactionCost: 0,
				ExecutedTime:    trade.ExecutedTime,
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

	return result, nil
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

func lossTrades(closedTrades []ClosedTrade) int {
	count := 0

	for _, trade := range closedTrades {
		if trade.Profit < 0 {
			count++
		}
	}

	return count
}

func breakevenTrades(closedTrades []ClosedTrade) int {
	count := 0

	for _, trade := range closedTrades {
		if trade.Profit == 0 {
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

func NewTradeStat(trades []domain.Trade) (TradeStat, error) {
	closedTrades, err := pairTrades(trades)
	if err != nil {
		return TradeStat{}, err
	}

	return TradeStat{
		TotalTrades:      len(closedTrades),
		NetProfit:        netProfit(closedTrades),
		Winrate:          winrate(closedTrades),
		ProfitableTrades: profitTrades(closedTrades),
		LosingTrades:     lossTrades(closedTrades),
		BreakevenTrades:  breakevenTrades(closedTrades),
		Loss_Avg:         lossAvg(closedTrades),
		Win_Avg:          winAvg(closedTrades),
	}, nil
}
