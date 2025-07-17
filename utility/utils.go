package utility

import "math"

func Round[T ~float32 | ~float64](x T, places int) T {
	factor := math.Pow(10, float64(places))
	rounded := math.Round(float64(x)*factor) / factor
	return T(rounded)
}
