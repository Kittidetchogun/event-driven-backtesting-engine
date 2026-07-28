package domain

type PortfolioChecker interface {
	CanBuy(order Order) error
	CanSell(order Order) error
}

type DummyPortfolioChecker struct{}

func (DummyPortfolioChecker) CanBuy(order Order) error {
	return nil
}

func (DummyPortfolioChecker) CanSell(order Order) error {
	return nil
}
