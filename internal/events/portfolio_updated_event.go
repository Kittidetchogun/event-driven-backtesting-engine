package events

import "event-driven-backtesting-engine/internal/domain"

const PortfolioUpdatedEventType = "PortfolioUpdatedEvent"

type PortfolioUpdatedEvent struct {
	BaseEvent
	Portfolio domain.Portfolio
}

var _ Event = PortfolioUpdatedEvent{}

func NewPortfolioUpdatedEvent(
	portfolio domain.Portfolio,
) PortfolioUpdatedEvent {

	return PortfolioUpdatedEvent{
		BaseEvent: NewBaseEvent(
			PortfolioUpdatedEventType,
			portfolio.UpdatedAt,
		),
		Portfolio: portfolio,
	}
}
