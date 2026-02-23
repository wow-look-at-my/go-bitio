package bitio

import "math"

// float32FromBits converts uint32 to float32.
func float32FromBits(b uint32) float32 {
	return math.Float32frombits(b)
}

// float64FromBits converts uint64 to float64.
func float64FromBits(b uint64) float64 {
	return math.Float64frombits(b)
}
