package events

// Consumer is implemented by any component that can consume events
// dispatched by the EventDispatcher.
//
// Examples:
//   - PrintConsumer
//   - StrategyEngine
//   - OrderManager
//   - PortfolioEngine
//   - StatisticsEngine
type Consumer interface {
	Consume(Event) error
}