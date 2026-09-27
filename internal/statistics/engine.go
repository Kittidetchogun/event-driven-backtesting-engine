package statistics

import (
	"context"
	"fmt"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/events"
	"event-driven-backtesting-engine/internal/portfolio"
)

type Engine struct {
	snapshots  []portfolio.PortfolioSnapshot
	trades     []domain.Trade
	repository BacktestResultRepository
}

func NewEngine() *Engine {
	return &Engine{
		snapshots: make([]portfolio.PortfolioSnapshot, 0),
		trades:    make([]domain.Trade, 0),
	}
}

func (e *Engine) SetRepository(
	repository BacktestResultRepository,
) {
	e.repository = repository
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

		for index := range e.snapshots {
			if e.snapshots[index].Time.Equal(snapshot.Time) {
				e.snapshots[index] = snapshot
				return nil
			}
		}

		e.snapshots = append(e.snapshots, snapshot)

		return nil

	case events.BacktestCompletedEvent:
		return e.saveBacktestResult(event)

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

func (e *Engine) saveBacktestResult(
	event events.BacktestCompletedEvent,
) error {

	if e.repository == nil {
		return fmt.Errorf(
			"backtest result repository is not configured",
		)
	}

	result := BacktestResult{
		RunID:        event.RunID,
		StrategyName: event.StrategyName,
		Symbol:       event.Symbol,
		Timeframe:    event.Timeframe,
		StartDate:    event.StartDate,
		EndDate:      event.EndDate,
		TotalReturn:  event.TotalReturn,
		WinRate:      event.WinRate,
		SharpeRatio:  event.SharpeRatio,
		MaxDrawdown:  event.MaxDrawdown,
	}

	return e.repository.SaveBacktestResult(
		context.Background(),
		result,
	)
}
