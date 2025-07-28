package float8

import (
	"fmt"
)

// Float8 represents an 8-bit floating-point number using the IEEE 754 FP8 E4M3FN format.
// This format is commonly used in machine learning for reduced-precision arithmetic.
//
// Bit layout:
//   - 1 bit  : Sign (0 = positive, 1 = negative)
//   - 4 bits : Exponent (biased by 7, range [-6, 7])
//   - 3 bits : Mantissa (3 explicit bits, 1 implicit leading bit for normal numbers)
//
// Special values:
//   - PositiveZero/NegativeZero: Exponent=0000, Mantissa=000
//   - PositiveInfinity/NegativeInfinity: Exponent=1111, Mantissa=000
//   - NaN: Exponent=1111, Mantissa=111
//
// This implementation follows the E4M3FN variant which has no infinities and two NaNs.
type Float8 uint8

// Bit masks and constants for Float8 format
const (
	SignMask     = 0b10000000 // 0x80 - Sign bit mask
	ExponentMask = 0b01111000 // 0x78 - Exponent bits mask
	MantissaMask = 0b00000111 // 0x07 - Mantissa bits mask
	MantissaLen  = 3          // Number of mantissa bits

	// Exponent bias and limits
	// See https://en.wikipedia.org/wiki/Exponent_bias
	// bias = 2^(|exponent|-1) - 1
	ExponentBias = 7  // Bias for 4-bit exponent
	ExponentMax  = 15 // Maximum exponent value
	ExponentMin  = -7 // Minimum exponent value

	// Float32 constants for conversion
	Float32Bias = 127 // IEEE 754 single precision bias

	// Special values
	PositiveZero     Float8 = 0x00
	NegativeZero     Float8 = 0x80
	PositiveInfinity Float8 = 0x78 // IEEE 754 E4M3FN: S.1111.000 = 0.1111.000₂
	NegativeInfinity Float8 = 0xF8 // IEEE 754 E4M3FN: S.1111.000 = 1.1111.000₂
	NaN              Float8 = 0x7F // IEEE 754 E4M3FN: S.1111.111 (0x7F or 0xFF)
	MaxValue         Float8 = 0x7E // Largest finite positive value
	MinValue         Float8 = 0xFE // Largest finite negative value
	SmallestPositive Float8 = 0x01 // Smallest positive normalized value
)

// ConversionMode defines how conversions handle edge cases
type ConversionMode int

const (
	// ModeDefault uses standard IEEE 754 rounding behavior
	ModeDefault ConversionMode = iota
	// ModeStrict returns errors for overflow/underflow
	ModeStrict
	// ModeFast uses lookup tables when available (default for arithmetic)
	ModeFast
)

// ArithmeticMode defines which implementation to use for arithmetic operations
type ArithmeticMode int

const (
	// ArithmeticAuto chooses the best implementation automatically
	ArithmeticAuto ArithmeticMode = iota
	// ArithmeticAlgorithmic forces algorithmic implementation
	ArithmeticAlgorithmic
	// ArithmeticLookup forces lookup table implementation (if available)
	ArithmeticLookup
)

// Float8Error represents errors that can occur during Float8 operations
type Float8Error struct {
	Op    string  // Operation that caused the error
	Value float32 // Input value that caused the error (if applicable)
	Msg   string  // Error message
}

func (e *Float8Error) Error() string {
	if e.Value != 0 {
		return fmt.Sprintf("float8.%s: %s (value: %g)", e.Op, e.Msg, e.Value)
	}
	return fmt.Sprintf("float8.%s: %s", e.Op, e.Msg)
}

// Common error instances
var (
	ErrOverflow  = &Float8Error{Op: "convert", Msg: "value too large for float8"}
	ErrUnderflow = &Float8Error{Op: "convert", Msg: "value too small for float8"}
	ErrNaN       = &Float8Error{Op: "convert", Msg: "NaN not representable in float8"}
)

// IsZero reports whether f represents the floating-point value zero (either positive or negative).
//
// According to IEEE 754, both +0 and -0 are considered zero, though they may have different
// bit patterns and behave differently in certain operations (like 1/+0 = +Inf, 1/-0 = -Inf).
//
// Returns:
//   - true if f is +0 or -0
//   - false otherwise, including for NaN and infinity values
func (f Float8) IsZero() bool {
	return f == PositiveZero || f == NegativeZero
}

// IsInf reports whether f is an infinity, either positive or negative.
//
// In the E4M3FN format, infinity values have all exponent bits set (0x78 for +Inf, 0xF8 for -Inf)
// and a zero mantissa. This is different from the standard IEEE 754 format used in float32/float64.
//
// Returns:
//   - true if f is positive or negative infinity
//   - false otherwise, including for NaN and finite values
func (f Float8) IsInf() bool {
	return f == PositiveInfinity || f == NegativeInfinity
}

// IsFinite reports whether f is a finite value (not infinite and not NaN).
//
// A Float8 value is finite if its exponent is not all 1s (0x0F). This includes
// both normal numbers (with an implicit leading 1 bit) and subnormal numbers
// (with an implicit leading 0 bit).
//
// Returns:
//   - true if f is a finite number (including zero and subnormals)
//   - false if f is infinity or NaN
func (f Float8) IsFinite() bool {
	// Extract exponent bits (bits 3-6)
	exp := (f & ExponentMask) >> MantissaLen
	return exp < 0x0F // Finite if exponent is not all 1s (0x0F = 1111)
}

// IsNaN reports whether f is a "not-a-number" (NaN) value.
//
// In the E4M3FN format, NaN is represented with all exponent bits set (0x0F)
// and all mantissa bits set (0x07). This results in two possible NaN values:
// 0x7F (positive NaN) and 0xFF (negative NaN).
//
// Returns:
//   - true if f is a NaN value
//   - false otherwise, including for infinity and finite values
func (f Float8) IsNaN() bool {
	// IEEE 754 E4M3FN: NaN has exponent=1111 and mantissa=111 (non-zero)
	// This corresponds to bit patterns 0x7F (positive) and 0xFF (negative)
	return (f&0x7F == 0x7F) && (f&0x07 == 0x07)
}

// Sign returns the sign of the Float8 value.
//
// The return values are:
//   - 1  if f > 0
//   - -1 if f < 0
//   - 0  if f is zero (including -0) or NaN
//
// Note that negative zero is treated as zero (returns 0), following the IEEE 754
// standard where +0 and -0 compare as equal. However, they can be distinguished
// using bitwise operations or by examining the sign bit directly.
//
// For NaN values, Sign returns 0, consistent with math/big.Float's behavior.
func (f Float8) Sign() int {
	if f.IsNaN() {
		return 0
	}
	if f.IsZero() {
		return 0
	}
	if f&SignMask != 0 {
		return -1
	}
	return 1
}

// Abs returns the absolute value of f.
//
// Special cases are:
//
//	Abs(±Inf) = +Inf
//	Abs(NaN) = NaN
//	Abs(±0) = +0
//
// For all other values, Abs clears the sign bit to return a positive number.
func (f Float8) Abs() Float8 {
	return f &^ SignMask // Clear sign bit
}

// Neg returns the negation of the Float8
func (f Float8) Neg() Float8 {
	if f.IsZero() {
		return f // Preserve zero sign for IEEE compliance
	}
	return f ^ SignMask // Flip sign bit
}

// String returns a string representation of the Float8 value
func (f Float8) String() string {
	return fmt.Sprintf("%.6g", f.ToFloat32())
}

// GoString returns a Go syntax representation of the Float8 value
func (f Float8) GoString() string {
	return fmt.Sprintf("float8.FromBits(0x%02x)", uint8(f))
}

// Bits returns the underlying uint8 representation
func (f Float8) Bits() uint8 {
	return uint8(f)
}

// FromBits creates a Float8 from its bit representation
func FromBits(bits uint8) Float8 {
	return Float8(bits)
}
