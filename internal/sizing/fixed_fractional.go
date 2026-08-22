package sizing

type FixedFractional struct {
	fraction float64
}

func NewFixedFractional(fraction float64) *FixedFractional {
	return &FixedFractional{
		fraction: fraction,
	}
}

func (s *FixedFractional) Size(
	equity float64,
	price float64,
) float64 {
	if equity <= 0 {
		return 0
	}

	if price <= 0 {
		return 0
	}

	if s.fraction <= 0 {
		return 0
	}

	capital := equity * s.fraction

	return capital / price
}
