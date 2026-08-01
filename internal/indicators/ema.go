package indicators

import (
	"event-driven-backtesting-engine/internal/domain"
)

type EMA struct {
	Period int
	Values []float64

}

func NewEMA(period int) *EMA {
	return &EMA{
		Period: period,
		Values: make([]float64, 0),
	}
}

func (e *EMA) Name() string {
	return "EMA"
}


func (e *EMA) Value() float64 {
	if len(e.Values) == 0 {
		return 0
	}

	return e.Values[len(e.Values)-1]
}

func (e *EMA) IsReady() bool {
	return e.Period > 0 && len(e.Values) >= e.Period
}

func (e *EMA) calculateEMA(price float64) float64 {
	if len(e.Values) == 0 {
		return price
	}
	smoothingFactor := 2.0 / float64(e.Period+1)
	previousEMA := e.Values[len(e.Values)-1]
	return (price-previousEMA)*smoothingFactor + previousEMA
}

func (e *EMA) Update(candle domain.Candle) {
	ema := e.calculateEMA(candle.Close)
	e.Values = append(e.Values, ema)
}


func (e *EMA) Reset() {
	e.Values = e.Values[:0]
}