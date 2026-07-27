package indicators
import (
	"event-driven-backtesting-engine/internal/domain"
)

type Indicator interface {
    Name() string
    Update(candle domain.Candle)
	Value() float64
    IsReady() bool
    Reset()
}