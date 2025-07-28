package float8

import (
	"math"
)

// Mathematical functions for Float8

// Sqrt returns the square root of the Float8 value.
//
// Special cases are:
//
//	Sqrt(+0) = +0
//	Sqrt(-0) = -0
//	Sqrt(+Inf) = +Inf
//	Sqrt(x) = NaN if x < 0 (including -Inf)
//	Sqrt(NaN) = NaN
//
// For finite x ≥ 0, the result is the greatest Float8 value y such that y² ≤ x.
// The result is rounded to the nearest representable Float8 value.
func Sqrt(f Float8) Float8 {
	if f == PositiveZero || f == NegativeZero {
		return PositiveZero
	}
	if f == PositiveInfinity {
		return PositiveInfinity
	}
	if f.Sign() < 0 {
		// Square root of negative number - return zero (NaN equivalent)
		return PositiveZero
	}

	f32 := f.ToFloat32()
	result := float32(math.Sqrt(float64(f32)))
	return ToFloat8(result)
}

// Pow returns f raised to the power of exp.
//
// Special cases are:
//
//	Pow(±0, exp) = ±0 for exp > 0
//	Pow(±0, exp) = +Inf for exp < 0
//	Pow(1, exp) = 1 for any exp (even NaN)
//	Pow(f, 0) = 1 for any f (including NaN, +Inf, -Inf)
//	Pow(f, 1) = f for any f
//	Pow(NaN, exp) = NaN
//	Pow(f, NaN) = NaN
//	Pow(±0, -Inf) = +Inf
//	Pow(±0, +Inf) = +0
//	Pow(+Inf, exp) = +Inf for exp > 0
//	Pow(+Inf, exp) = +0 for exp < 0
//	Pow(-Inf, exp) = -0 for exp a negative odd integer
//	Pow(-Inf, exp) = +0 for exp a negative non-odd integer
//	Pow(-Inf, exp) = -Inf for exp a positive odd integer
//	Pow(-Inf, exp) = +Inf for exp a positive non-odd integer
//	Pow(-1, ±Inf) = 1
//	Pow(f, +Inf) = +Inf for |f| > 1
//	Pow(f, -Inf) = +0 for |f| > 1
//	Pow(f, +Inf) = +0 for |f| < 1
//	Pow(f, -Inf) = +Inf for |f| < 1
//
// The result is rounded to the nearest representable Float8 value.
func Pow(f, exp Float8) Float8 {
	// Handle special cases
	if f == PositiveZero || f == NegativeZero {
		if exp.Sign() > 0 {
			return PositiveZero
		}
		if exp.Sign() < 0 {
			return PositiveInfinity
		}
		return ToFloat8(1.0) // 0^0 = 1
	}

	if f == PositiveInfinity {
		if exp.Sign() > 0 {
			return PositiveInfinity
		}
		if exp.Sign() < 0 {
			return PositiveZero
		}
		return ToFloat8(1.0) // inf^0 = 1
	}

	f32 := f.ToFloat32()
	exp32 := exp.ToFloat32()
	result := float32(math.Pow(float64(f32), float64(exp32)))
	return ToFloat8(result)
}

// Exp returns e^f
func Exp(f Float8) Float8 {
	if f == PositiveZero || f == NegativeZero {
		return ToFloat8(1.0)
	}
	if f == PositiveInfinity {
		return PositiveInfinity
	}
	if f == NegativeInfinity {
		return PositiveZero
	}

	f32 := f.ToFloat32()
	result := float32(math.Exp(float64(f32)))
	return ToFloat8(result)
}

// Log returns the natural logarithm of f.
//
// Special cases are:
//
//	Log(+Inf) = +Inf
//	Log(0) = -Inf
//	Log(x < 0) = NaN
//	Log(NaN) = NaN
//
// For finite x > 0, the result is the natural logarithm of x.
// The result is rounded to the nearest representable Float8 value.
func Log(f Float8) Float8 {
	if f == PositiveZero || f == NegativeZero {
		return NegativeInfinity
	}
	if f == PositiveInfinity {
		return PositiveInfinity
	}
	if f.Sign() < 0 {
		// Log of negative number - return zero (NaN equivalent)
		return PositiveZero
	}

	f32 := f.ToFloat32()
	result := float32(math.Log(float64(f32)))
	return ToFloat8(result)
}

// Sin returns the sine of f (in radians).
//
// Special cases are:
//
//	Sin(±0) = ±0
//	Sin(±Inf) = NaN
//	Sin(NaN) = NaN
//
// For finite x, the result is the sine of x in the range [-1, 1].
// The result is rounded to the nearest representable Float8 value.
func Sin(f Float8) Float8 {
	if f == PositiveZero || f == NegativeZero {
		return f // Preserve sign of zero
	}
	if f.IsInf() {
		return PositiveZero // sin(inf) is undefined, return 0
	}

	f32 := f.ToFloat32()
	result := float32(math.Sin(float64(f32)))
	return ToFloat8(result)
}

// Cos returns the cosine of f (in radians).
//
// Special cases are:
//
//	Cos(±0) = 1
//	Cos(±Inf) = NaN
//	Cos(NaN) = NaN
//
// For finite x, the result is the cosine of x in the range [-1, 1].
// The result is rounded to the nearest representable Float8 value.
func Cos(f Float8) Float8 {
	if f == PositiveZero || f == NegativeZero {
		return ToFloat8(1.0)
	}
	if f.IsInf() {
		return PositiveZero // cos(inf) is undefined, return 0
	}

	f32 := f.ToFloat32()
	result := float32(math.Cos(float64(f32)))
	return ToFloat8(result)
}

// Tan returns the tangent of f (in radians).
//
// Special cases are:
//
//	Tan(±0) = ±0
//	Tan(±Inf) = NaN
//	Tan(NaN) = NaN
//
// For finite x, the result is the tangent of x.
// The result is rounded to the nearest representable Float8 value.
// Note that the result may be extremely large or small for inputs near (2n+1)π/2.
func Tan(f Float8) Float8 {
	if f == PositiveZero || f == NegativeZero {
		return f // Preserve sign of zero
	}
	if f.IsInf() {
		return PositiveZero // tan(inf) is undefined, return 0
	}

	f32 := f.ToFloat32()
	result := float32(math.Tan(float64(f32)))
	return ToFloat8(result)
}

// Floor returns the greatest integer value less than or equal to f.
//
// Special cases are:
//
//	Floor(±0) = ±0
//	Floor(±Inf) = ±Inf
//	Floor(NaN) = NaN
//
// For finite x, the result is the greatest integer value ≤ x.
// The result is exact (no rounding occurs).
func Floor(f Float8) Float8 {
	if f.IsZero() || f.IsInf() {
		return f
	}

	f32 := f.ToFloat32()
	result := float32(math.Floor(float64(f32)))
	return ToFloat8(result)
}

// Ceil returns the least integer value greater than or equal to f.
//
// Special cases are:
//
//	Ceil(±0) = ±0
//	Ceil(±Inf) = ±Inf
//	Ceil(NaN) = NaN
//
// For finite x, the result is the least integer value ≥ x.
// The result is exact (no rounding occurs).
func Ceil(f Float8) Float8 {
	if f.IsZero() || f.IsInf() {
		return f
	}

	f32 := f.ToFloat32()
	result := float32(math.Ceil(float64(f32)))
	return ToFloat8(result)
}

// Round returns the nearest integer value to f, rounding ties to even.
//
// Special cases are:
//
//	Round(±0) = ±0
//	Round(±Inf) = ±Inf
//	Round(NaN) = NaN
//
// For finite x, the result is the nearest integer to x.
// Ties are rounded to the nearest even integer.
// The result is exact (no rounding occurs).
func Round(f Float8) Float8 {
	if f.IsZero() || f.IsInf() {
		return f
	}

	f32 := f.ToFloat32()
	result := float32(math.Round(float64(f32)))
	return ToFloat8(result)
}

// Trunc returns the integer value of f with any fractional part removed.
//
// Special cases are:
//
//	Trunc(±0) = ±0
//	Trunc(±Inf) = ±Inf
//	Trunc(NaN) = NaN
//
// For finite x, the result is the integer part of x with the sign of x.
// This is equivalent to rounding toward zero.
// The result is exact (no rounding occurs).
func Trunc(f Float8) Float8 {
	if f.IsZero() || f.IsInf() {
		return f
	}

	f32 := f.ToFloat32()
	result := float32(math.Trunc(float64(f32)))
	return ToFloat8(result)
}

// Fmod returns the floating-point remainder of x/y.
//
// The result has the same sign as x and magnitude less than the magnitude of y.
//
// Special cases are:
//
//	Fmod(±0, y) = ±0 for y != 0
//	Fmod(±Inf, y) = NaN
//	Fmod(x, 0) = NaN
//	Fmod(NaN, y) = NaN
//	Fmod(x, NaN) = NaN
//	Fmod(x, ±Inf) = x for x not infinite
//
// For finite x and y (y ≠ 0), the result is x - n*y where n is the integer
// nearest to x/y. If two integers are equally near, the even one is chosen.
// The result is rounded to the nearest representable Float8 value.
func Fmod(x, y Float8) Float8 {
	if y.IsZero() {
		// fmod(x, ±0) is undefined, but we return 0 for compatibility
		return PositiveZero
	}
	if x.IsZero() {
		// fmod(±0, y) = ±0
		return x
	}
	if y.IsInf() {
		// fmod(x, ±Inf) = x for finite x
		return x
	}
	if x.IsInf() {
		// fmod(±Inf, y) = NaN, but we return 0 for compatibility
		return PositiveZero
	}

	f32 := x.ToFloat32()
	div32 := y.ToFloat32()
	result := float32(math.Mod(float64(f32), float64(div32)))
	return ToFloat8(result)
}

// Constants as Float8 values
var (
	E      = ToFloat8(2.718281828459045)  // Euler's number
	Pi     = ToFloat8(3.141592653589793)  // Pi
	Phi    = ToFloat8(1.618033988749895)  // Golden ratio
	Sqrt2  = ToFloat8(1.4142135623730951) // Square root of 2
	SqrtE  = ToFloat8(1.6487212707001282) // Square root of E
	SqrtPi = ToFloat8(1.7724538509055159) // Square root of Pi
	Ln2    = ToFloat8(0.6931471805599453) // Natural logarithm of 2
	Log2E  = ToFloat8(1.4426950408889634) // Base-2 logarithm of E
	Ln10   = ToFloat8(2.302585092994046)  // Natural logarithm of 10
	Log10E = ToFloat8(0.4342944819032518) // Base-10 logarithm of E
)

// Utility functions

// Clamp restricts f to the range [min, max]
func Clamp(f, min, max Float8) Float8 {
	if Less(f, min) {
		return min
	}
	if Greater(f, max) {
		return max
	}
	return f
}

// Lerp performs linear interpolation between a and b by factor t
func Lerp(a, b, t Float8) Float8 {
	// lerp(a, b, t) = a + t * (b - a)
	diff := Sub(b, a)
	scaled := Mul(t, diff)
	return Add(a, scaled)
}

// Sign returns -1, 0, or 1 depending on the sign of f
func Sign(f Float8) Float8 {
	sign := f.Sign()
	switch sign {
	case -1:
		return ToFloat8(-1.0)
	case 0:
		return PositiveZero
	case 1:
		return ToFloat8(1.0)
	default:
		return PositiveZero
	}
}

// CopySign returns a Float8 with the magnitude of f and the sign of sign
func CopySign(f, sign Float8) Float8 {
	if sign.Sign() < 0 {
		return f.Abs() | SignMask // Set sign bit
	}
	return f.Abs() // Clear sign bit
}
