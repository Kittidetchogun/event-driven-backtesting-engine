package strategy

import (
    "errors"
    "fmt"
    "time"

    "event-driven-backtesting-engine/internal/domain"
    "event-driven-backtesting-engine/internal/events"
    "event-driven-backtesting-engine/internal/indicators"
)

var ErrInvalidEmaCrossPeriods = errors.New("fast period must be greater than zero and less than slow period")

type Signal string

const (
	NoSignal   Signal = "NONE"
	BuySignal  Signal = "BUY"
	SellSignal Signal = "SELL"
)

// EmaCross เป็นกลยุทธ์การเทรดด้วยเส้น EMA 2 เส้นตัดกัน (Fast & Slow)
type EmaCross struct {

    fastPeriod int
    slowPeriod int

    Fast *indicators.EMA
    Slow *indicators.EMA

    Signal Signal

    dispatcher *events.EventDispatcher

    initialized bool
    prevDiff    float64
    hasPrevDiff bool
}

func (e *EmaCross) SetDispatcher(d *events.EventDispatcher) {
    e.dispatcher = d
}

// NewEmaCross สร้าง instance ใหม่ของกลยุทธ์ EMA Cross
func NewEmaCross(fastPeriod, slowPeriod int) *EmaCross {
	return &EmaCross{
		fastPeriod: fastPeriod,
		slowPeriod: slowPeriod,
		Signal:     NoSignal,
	}
}

func (e *EmaCross) Name() string {
 	return "EMA Cross"
}

// Initialize ตรวจสอบ parameter และสร้าง Indicator EMA 2 เส้น
func (e *EmaCross) Initialize() error {
	if e.fastPeriod <= 0 || e.slowPeriod <= 0 || e.fastPeriod >= e.slowPeriod {
		return ErrInvalidEmaCrossPeriods
	}

	e.Fast = indicators.NewEMA(e.fastPeriod)
	e.Slow = indicators.NewEMA(e.slowPeriod)
	e.Signal = NoSignal
	e.initialized = true
	e.prevDiff = 0
	e.hasPrevDiff = false

	return nil
}

// OnData รับข้อมูลราคาแท่งเทียนเข้ามา แล้วคำนวณหาสัญญาณการเทรด (BUY/SELL/NONE)
func (e *EmaCross) OnData(candle domain.Candle) {
	if !e.initialized || e.Fast == nil || e.Slow == nil {
		return
	}

	// 1. อัปเดตราคาปิดเข้าไปใน Indicator ทั้งสองเส้น
	e.Fast.Update(candle)
	e.Slow.Update(candle)

	// 2. ถ้าข้อมูลแท่งเทียนยังสะสมไม่พอคำนวณ EMA ให้ข้ามไปก่อน
	if !e.Fast.IsReady() || !e.Slow.IsReady() {
		e.Signal = NoSignal
		return
	}

	// 3. คำนวณความต่างระหว่าง EMA Fast และ EMA Slow
	current := e.Fast.Value() - e.Slow.Value()

	// 4. กำหนดสัญญาณเริ่มต้นเป็น NoSignal
	e.Signal = NoSignal

	// 5. ตรวจสอบการตัดกัน (Crossover) เปรียบเทียบกับแท่งก่อนหน้า
	if e.hasPrevDiff {
		switch {
			case e.prevDiff <= 0 && current > 0:
				e.Signal = BuySignal
				e.dispatchSignal(
					candle.Symbol,
					candle.Timestamp,
					domain.BuyOrder,
				)

			case e.prevDiff >= 0 && current < 0:
				e.Signal = SellSignal
				e.dispatchSignal(
					candle.Symbol,
					candle.Timestamp,
					domain.SellOrder,
				)
		}
	}

	// 6. บันทึกผลต่างปัจจุบันไว้ใช้เปรียบเทียบในแท่งถัดไป
	e.prevDiff = current
	e.hasPrevDiff = true
}

// Reset ล้างค่า State และ Indicator ทั้งหมดเพื่อพร้อมสำหรับการ Backtest ใหม่
func (e *EmaCross) Reset() {
	if e.Fast != nil {
		e.Fast.Reset()
	}
	if e.Slow != nil {
		e.Slow.Reset()
	}

	e.Signal = NoSignal
	e.prevDiff = 0
	e.hasPrevDiff = false
}

func (e *EmaCross) Ready() bool {
	return e.initialized && e.Fast != nil && e.Slow != nil
}

func (e *EmaCross) CurrentSignal() Signal {
	return e.Signal
}

func (e *EmaCross) String() string {
	return fmt.Sprintf("EMA Cross(fast=%d, slow=%d)", e.fastPeriod, e.slowPeriod)
}

func (e *EmaCross) dispatchSignal(
    symbol string,
    timestamp time.Time,
    side domain.OrderSide,
) {
    if e.dispatcher == nil {
        return
    }

    event := events.NewSignalGeneratedEvent(
        1, // TODO: Replace with Backtest RunID
        symbol,
        side,
        1, // TODO: Replace with Position Sizer
        timestamp,
    )

    if err := e.dispatcher.Dispatch(event); err != nil {
        fmt.Printf("failed to dispatch signal event: %v\n", err)
    }
}