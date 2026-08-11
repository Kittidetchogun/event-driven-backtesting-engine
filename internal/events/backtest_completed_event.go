package events

import "time"

const BacktestCompletedEventType = "BacktestCompletedEvent"

type BacktestCompletedEvent struct {
	BaseEvent

	RunID        int
	StrategyName string
	Symbol       string
	Timeframe    string

	StartDate time.Time
	EndDate   time.Time

	TotalReturn float64
	WinRate     float64
	SharpeRatio float64
	MaxDrawdown float64
}

func NewBacktestCompletedEvent(
	runID int,
	strategyName string,
	symbol string,
	timeframe string,
	startDate time.Time,
	endDate time.Time,
	totalReturn float64,
	winRate float64,
	sharpeRatio float64,
	maxDrawdown float64,
) BacktestCompletedEvent {
	return BacktestCompletedEvent{
		BaseEvent: NewBaseEvent(
			BacktestCompletedEventType,
			endDate,
		),

		RunID:        runID,
		StrategyName: strategyName,
		Symbol:       symbol,
		Timeframe:    timeframe,

		StartDate: startDate,
		EndDate:   endDate,

		TotalReturn: totalReturn,
		WinRate:     winRate,
		SharpeRatio: sharpeRatio,
		MaxDrawdown: maxDrawdown,
	}
}

var _ Event = BacktestCompletedEvent{}
