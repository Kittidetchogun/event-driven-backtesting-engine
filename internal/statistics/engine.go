package statistics

import (
	"fmt"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
	"event-driven-backtesting-engine/internal/portfolio"
)

type Engine struct {
	snapshots []portfolio.PortfolioSnapshot
	trades    []domain.Trade
}

func NewEngine() *Engine {
	return &Engine{
		snapshots: make([]portfolio.PortfolioSnapshot, 0),
		trades:    make([]domain.Trade, 0),
	}
}

// Consume receives events from Event Queue.
func (e *Engine) Consume(event events.Event) error {

	switch event := event.(type) {

	case events.TradeExecutedEvent:

		e.trades = append(
			e.trades,
			event.Trade,
		)

		return nil

	case events.PortfolioUpdatedEvent:

		snapshot := portfolio.NewPortfolioSnapshot(
			event.Portfolio,
		)

		e.snapshots = append(
			e.snapshots,
			snapshot,
		)

		return nil

	default:

		return fmt.Errorf(
			"unsupported event %T",
			event,
		)
	}
}

// Snapshots returns all portfolio snapshots.
func (e *Engine) Snapshots() []portfolio.PortfolioSnapshot {
	return e.snapshots
}

// EquityCurve returns equity history.
func (e *Engine) EquityCurve() []float64 {

	curve := make([]float64, 0, len(e.snapshots))

	for _, snapshot := range e.snapshots {
		curve = append(
			curve,
			snapshot.Equity,
		)
	}

	return curve
}

// Trades returns all executed trades.
func (e *Engine) Trades() []domain.Trade {
	return e.trades
}

// TradeStats returns aggregated trade statistics.
func (e *Engine) TradeStats() TradeStat {
	return NewTradeStat(e.trades)
}
