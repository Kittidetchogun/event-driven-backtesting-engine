package statistics

import (
	"fmt"
	"event-driven-backtesting-engine/internal/events"
	"event-driven-backtesting-engine/internal/portfolio"
)

type Engine struct {
	snapshots []portfolio.PortfolioSnapshot
}

func NewEngine() *Engine {
	return &Engine{
		snapshots: make([]portfolio.PortfolioSnapshot, 0),
	}
}

// Consume receives PortfolioUpdatedEvent from Event Queue.
func (e *Engine) Consume(event events.Event) error {

	// Event ต้องเป็น PortfolioUpdatedEvent
	portfolioEvent, ok := event.(events.PortfolioUpdatedEvent)
	if !ok {
		return fmt.Errorf("unsupported event %T", event)
	}

	// สร้าง Snapshot
	snapshot := portfolio.NewPortfolioSnapshot(
		portfolioEvent.Portfolio,
	)

	// เก็บ History
	e.snapshots = append(
		e.snapshots,
		snapshot,
	)

	return nil
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