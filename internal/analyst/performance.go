package analyst

import (
	"event-driven-backtesting-engine/internal/domain"
)

type Performance struct {
	SharpeRatio float64
	// SortinoRatio float64
	MaxDrawdown  float64
}