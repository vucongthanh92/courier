package utils

import "math"

// AlmostEqualFloat64 checks if two float64 numbers are approximately equal within a small threshold.
func AlmostEqualFloat64(a, b float64) bool {
	const float64EqualityThreshold = 1e-6
	return math.Abs(a-b) <= float64EqualityThreshold
}

// Round rounds a float64 number to the nearest integer.
func Round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

// ToFixed rounds a float64 number to a specified number of decimal places.
func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(Round(num*output)) / output
}

// RoundDownPrice rounds down a float64 number to the nearest multiple of 10 raised to the specified precision.
// For example, RoundDownPrice(123.456, 1) returns 120.0, and RoundDownPrice(123.456, 2) returns 100.0.
func RoundDownPrice(num float64, precision int64) float64 {
	return float64(int64(num) - int64(num)%(10*(precision)))
}
