package sizing

type Sizer interface {
	Size(equity float64, price float64) float64
}
