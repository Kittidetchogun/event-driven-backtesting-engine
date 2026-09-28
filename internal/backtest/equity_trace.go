package backtest

import (
	"io"
	"strconv"
	"strings"
	"time"

	"event-driven-backtesting-engine/internal/domain"
	"event-driven-backtesting-engine/internal/portfolio"
)

// EquityTraceRow is the post-mark portfolio state for one processed candle.
type EquityTraceRow struct {
	Timestamp        time.Time
	Cash             float64
	PositionQuantity float64
	Close            float64
	PositionValue    float64
	RealizedPnL      float64
	UnrealizedPnL    float64
	Equity           float64
}

// EquityTrace returns one diagnostic row for each candle processed by Run.
func (r *Runner) EquityTrace() []EquityTraceRow {
	return r.equityTrace
}

// WriteEquityTraceCSV writes the diagnostic trace without recalculating any
// portfolio values.
func (r *Runner) WriteEquityTraceCSV(w io.Writer) error {
	if err := writeCSVRecord(w, []string{
		"timestamp",
		"cash",
		"position_quantity",
		"close",
		"position_value",
		"realized_pnl",
		"unrealized_pnl",
		"equity",
	}); err != nil {
		return err
	}

	for _, row := range r.equityTrace {
		if err := writeCSVRecord(w, []string{
			row.Timestamp.UTC().Format(time.RFC3339Nano),
			formatFloat(row.Cash),
			formatFloat(row.PositionQuantity),
			formatFloat(row.Close),
			formatFloat(row.PositionValue),
			formatFloat(row.RealizedPnL),
			formatFloat(row.UnrealizedPnL),
			formatFloat(row.Equity),
		}); err != nil {
			return err
		}
	}

	return nil
}

func writeCSVRecord(w io.Writer, fields []string) error {
	_, err := io.WriteString(w, strings.Join(fields, ",")+"\n")
	return err
}

func (r *Runner) recordEquityTrace(candle domain.Candle) {
	var snapshot portfolio.PortfolioSnapshot
	snapshots := r.portfolio.Snapshots()
	for index := len(snapshots) - 1; index >= 0; index-- {
		if snapshots[index].Time.Equal(candle.Timestamp) {
			snapshot = snapshots[index]
			break
		}
	}

	quantity := 0.0
	if position, ok := r.portfolio.Positions()[candle.Symbol]; ok {
		quantity = position.Quantity
	}

	r.equityTrace = append(r.equityTrace, EquityTraceRow{
		Timestamp:        candle.Timestamp,
		Cash:             snapshot.Cash,
		PositionQuantity: quantity,
		Close:            candle.Close,
		PositionValue:    snapshot.PositionValue,
		RealizedPnL:      snapshot.RealizedPnL,
		UnrealizedPnL:    snapshot.UnrealizedPnL,
		Equity:           snapshot.Equity,
	})
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'g', -1, 64)
}
