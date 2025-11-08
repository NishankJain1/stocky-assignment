package api

import "math"

// round rounds a float to the given number of decimal places.
func round(value float64, places int) float64 {
	factor := math.Pow(10, float64(places))
	return math.Round(value*factor) / factor
}
