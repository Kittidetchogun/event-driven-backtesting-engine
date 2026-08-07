package statistics
import (
	"math"
)
func Mean(data []float64) float64 {
	total := 0.0
	for _, value := range data {
		total += value
	}
	return total / float64(len(data))
}
func StdDev(data []float64) float64 {
	m := Mean(data)
	variance := 0.0
	for _, value := range data {
		variance += (value - m) * (value - m)
	}
	return math.Sqrt(variance / float64(len(data)-1))
}
