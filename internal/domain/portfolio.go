package domain

type Portfolio struct {
	InitCapital float64
	Equity      float64
	Trades      []Trade
	Cash		float64
}

// func NewPortfolio(initCapital float64) *Portfolio {
// 	return &Portfolio{
// 		InitCapital: initCapital,
// 		Equity:      initCapital,
// 		Cash:        initCapital,
// 		Trades:      make([]Trade, 0),
// 	}
// }

func (p *Portfolio) UpdateEquity() {
	total := p.Cash
	for _, trade := range p.Trades {
		total += Profit(trade)
	}
	p.Equity = total
}

func (p *Portfolio) ApplyTrade(trade Trade) {
	profit := Profit(trade)
	p.Cash += profit
	p.UpdateEquity()
	p.Trades = append(p.Trades, trade)
}

func (p *Portfolio) Reset() {
	p.Cash = p.InitCapital
	p.Equity = p.InitCapital
	p.Trades = make([]Trade, 0)
}


