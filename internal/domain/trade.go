package domain
import "time"
type Trade struct {
    ID         int64
    BacktestID int64
    Side       string
    EntryTime  time.Time
    ExitTime   time.Time
    EntryPrice float64
    ExitPrice  float64
    Quantity   float64
    Fee        float64
}

func Profit(t Trade) float64 {
	switch t.Side {
	case "BUY":
		return (t.ExitPrice - t.EntryPrice) * t.Quantity - t.Fee
	case "SELL":
		return (t.EntryPrice - t.ExitPrice) * t.Quantity - t.Fee
	default:
		return 0
	}
}
