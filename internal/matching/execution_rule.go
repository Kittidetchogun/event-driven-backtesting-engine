package matching

import "event-driven-backtesting-engine/internal/domain"

// executionPrice returns the execution price based on the
// execution rule of the backtest.
//
// Prototype Rule:
//
//   Market Order
//          ↓
//   Fill Immediately
//          ↓
//   Executed Price = Current Candle Close
//          ↓
//   Full Fill
//
// In this prototype, Order.Price stores the current candle's
// closing price, so that value is used as the execution price.
// ยังมี Look-ahead Bias อยู่ แต่สามารถแก้ได้ทีหลัง เพราะมีวิธีนี้ทำง่ายแต่มี Bias
func executionPrice(order domain.Order) float64 {
	return order.Price
}