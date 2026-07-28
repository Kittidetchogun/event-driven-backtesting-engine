package strategy
import (
	"event-driven-backtesting-engine/internal/domain"
)
type Strategy interface {
	Name() string
	OnData(candle domain.Candle)
	Initialize() error
	Reset()
}
