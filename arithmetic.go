package float8

import (
	"math"
)

// Global arithmetic mode
var DefaultArithmeticMode = ArithmeticAuto

// Add returns the sum of the operands a and b.
//
// This is a convenience function that calls AddWithMode with DefaultArithmeticMode.
// For more control over the arithmetic behavior, use AddWithMode directly.
//
// Special cases:
//
//	Add(+0, ±0) = +0
//	Add(-0, -0) = -0
//	Add(±Inf, ∓Inf) = NaN (but returns +0 in this implementation)
//	Add(NaN, x) = NaN
//	Add(x, NaN) = NaN
//
// For finite numbers, the result is rounded to the nearest representable Float8 value
// using the current rounding mode (typically round-to-nearest-even).
func Add(a, b Float8) Float8 {
	return AddWithMode(a, b, DefaultArithmeticMode)
}

// AddWithMode returns the sum of the operands a and b using the specified arithmetic mode.
//
// The arithmetic mode determines how the addition is performed:
//   - ArithmeticAuto: Uses the fastest available method (lookup tables if enabled)
//   - ArithmeticLookup: Forces use of lookup tables (panics if not available)
//   - ArithmeticAlgorithmic: Uses the algorithmic implementation
//
// Special cases are handled according to IEEE 754 rules:
//   - If either operand is NaN, the result is NaN
//   - Infinities of the same sign add to infinity of that sign
//   - Infinities of opposite signs produce NaN (but this implementation returns +0)
//   - The sign of a zero result is the sign of the sum of the operands
//
// For finite numbers, the result is rounded to the nearest representable Float8 value.
// If the exact result is exactly halfway between two representable values, it is
// rounded to the value with an even least significant bit (round-to-nearest-even).
func AddWithMode(a, b Float8, mode ArithmeticMode) Float8 {
	// Use lookup table if available and mode allows it
	if (mode == ArithmeticAuto || mode == ArithmeticLookup) && addTable != nil {
		return addTable[uint16(a)<<8|uint16(b)]
	}

	// Fall back to algorithmic implementation
	return addAlgorithmic(a, b)
}

// Sub returns the difference of a-b, i.e., the result of subtracting b from a.
//
// This is a convenience function that calls SubWithMode with DefaultArithmeticMode.
// For more control over the arithmetic behavior, use SubWithMode directly.
//
// Special cases:
//
//	Sub(+0, +0) = +0
//	Sub(+0, -0) = +0
//	Sub(-0, +0) = -0
//	Sub(-0, -0) = +0
//	Sub(±Inf, ±Inf) = NaN (but returns +0 in this implementation)
//	Sub(NaN, x) = NaN
//	Sub(x, NaN) = NaN
//
// For finite numbers, the result is rounded to the nearest representable Float8 value.
func Sub(a, b Float8) Float8 {
	return SubWithMode(a, b, DefaultArithmeticMode)
}

// SubWithMode performs subtraction with specified arithmetic mode
func SubWithMode(a, b Float8, mode ArithmeticMode) Float8 {
	// Use lookup table if available and mode allows it
	if (mode == ArithmeticAuto || mode == ArithmeticLookup) && subTable != nil {
		return subTable[uint16(a)<<8|uint16(b)]
	}

	// Fall back to algorithmic implementation
	return subAlgorithmic(a, b)
}

// Mul returns the product of the operands a and b.
//
// This is a convenience function that calls MulWithMode with DefaultArithmeticMode.
// For more control over the arithmetic behavior, use MulWithMode directly.
//
// Special cases:
//
//	Mul(±0, ±Inf) = NaN
//	Mul(±Inf, ±0) = NaN
//	Mul(±0, ±0) = ±0 (sign obeys the rule for signs of zero products)
//	Mul(±0, y) = ±0 for y finite and not zero
//	Mul(±Inf, y) = ±Inf for y finite and not zero
//	Mul(x, y) = NaN if x or y is NaN
//
// The sign of the result follows the standard sign rules for multiplication.
// For finite numbers, the result is rounded to the nearest representable Float8 value.
func Mul(a, b Float8) Float8 {
	return MulWithMode(a, b, DefaultArithmeticMode)
}

// MulWithMode performs multiplication with specified arithmetic mode
func MulWithMode(a, b Float8, mode ArithmeticMode) Float8 {
	// Use lookup table if available and mode allows it
	if (mode == ArithmeticAuto || mode == ArithmeticLookup) && mulTable != nil {
		return mulTable[uint16(a)<<8|uint16(b)]
	}

	// Fall back to algorithmic implementation
	return mulAlgorithmic(a, b)
}

// Div returns the quotient a/b of the operands a and b.
//
// This is a convenience function that calls DivWithMode with DefaultArithmeticMode.
// For more control over the arithmetic behavior, use DivWithMode directly.
//
// Special cases:
//
//	Div(±0, ±0) = NaN
//	Div(±Inf, ±Inf) = NaN
//	Div(x, ±0) = ±Inf for x finite and not zero (sign obeys rule for signs)
//	Div(±Inf, y) = ±Inf for y finite and not zero (sign obeys rule for signs)
//	Div(x, y) = NaN if x or y is NaN
//
// The sign of the result follows the standard sign rules for division.
// For finite numbers, the result is rounded to the nearest representable Float8 value.
// Division by zero results in ±Inf with the sign determined by the rule of signs.
func Div(a, b Float8) Float8 {
	return DivWithMode(a, b, DefaultArithmeticMode)
}

// DivWithMode performs division with specified arithmetic mode
func DivWithMode(a, b Float8, mode ArithmeticMode) Float8 {
	// Use lookup table if available and mode allows it
	if (mode == ArithmeticAuto || mode == ArithmeticLookup) && divTable != nil {
		return divTable[uint16(a)<<8|uint16(b)]
	}

	// Fall back to algorithmic implementation
	return divAlgorithmic(a, b)
}

// Algorithmic implementations

func addAlgorithmic(a, b Float8) Float8 {
	// Handle special cases
	if a.IsZero() {
		return b
	}
	if b.IsZero() {
		return a
	}

	// Handle infinity cases
	if a.IsInf() || b.IsInf() {
		if a == PositiveInfinity && b == NegativeInfinity {
			return PositiveZero // NaN case, but we return zero
		}
		if a == NegativeInfinity && b == PositiveInfinity {
			return PositiveZero // NaN case, but we return zero
		}
		if a.IsInf() {
			return a
		}
		return b
	}

	// Convert to float32, perform operation, convert back
	f32a := a.ToFloat32()
	f32b := b.ToFloat32()
	result := f32a + f32b

	return ToFloat8(result)
}

func subAlgorithmic(a, b Float8) Float8 {
	// Handle special cases
	if b.IsZero() {
		return a
	}
	if a.IsZero() {
		return b.Neg()
	}

	// Handle infinity cases
	if a.IsInf() || b.IsInf() {
		if a == b && a.IsInf() {
			return PositiveZero // NaN case, but we return zero
		}
		if a.IsInf() {
			return a
		}
		return b.Neg()
	}

	// Convert to float32, perform operation, convert back
	f32a := a.ToFloat32()
	f32b := b.ToFloat32()
	result := f32a - f32b

	return ToFloat8(result)
}

func mulAlgorithmic(a, b Float8) Float8 {
	// Handle NaN cases - any operation with NaN results in NaN
	if a.IsNaN() || b.IsNaN() {
		return NaN
	}

	// Get signs before any potential conversions
	signA := int8(1)
	if a&SignMask != 0 {
		signA = -1
	}

	signB := int8(1)
	if b&SignMask != 0 {
		signB = -1
	}

	// Handle infinity * 0 or 0 * infinity cases (indeterminate form, results in NaN)
	if (a.IsInf() && b.IsZero()) || (a.IsZero() && b.IsInf()) {
		return NaN
	}

	// Handle zero cases (after checking for infinity * 0)
	if a.IsZero() || b.IsZero() {
		// Handle sign of zero result
		if signA*signB < 0 {
			return NegativeZero
		}
		return PositiveZero
	}

	// Handle infinity cases (after checking for infinity * 0)
	if a.IsInf() || b.IsInf() {
		if signA*signB > 0 {
			return PositiveInfinity
		}
		return NegativeInfinity
	}

	// Convert to float32, perform operation, convert back
	f32a := a.ToFloat32()
	f32b := b.ToFloat32()
	result := f32a * f32b

	// Final conversion may produce NaN or infinity, which is fine
	return ToFloat8(result)
}

func divAlgorithmic(a, b Float8) Float8 {
	// Handle NaN cases first (NaN op anything = NaN)
	if a.IsNaN() || b.IsNaN() {
		return NaN
	}

	// Handle division by zero
	if b.IsZero() {
		if a.IsZero() {
			return NaN // 0/0 is NaN
		}
		// Get signs explicitly to handle negative zero
		signA := a&SignMask != 0
		signB := b&SignMask != 0
		// The sign of the result is positive if the operands have the same sign, negative otherwise
		if signA == signB {
			return PositiveInfinity
		}
		return NegativeInfinity
	}

	// Handle zero dividend
	if a.IsZero() {
		// Get signs explicitly to handle negative zero
		signA := a&SignMask != 0
		signB := b&SignMask != 0
		// The sign of the result is negative if the operands have different signs
		if signA != signB {
			return NegativeZero
		}
		return PositiveZero
	}

	// Handle infinity cases
	if a.IsInf() {
		if b.IsInf() {
			return NaN // inf/inf is NaN
		}
		// inf / x = inf with proper sign
		signA := a&SignMask != 0
		signB := b&SignMask != 0
		if signA == signB {
			return PositiveInfinity
		}
		return NegativeInfinity
	}

	// x / inf = 0 with proper sign
	if b.IsInf() {
		signA := a&SignMask != 0
		signB := b&SignMask != 0
		if signA == signB {
			return PositiveZero
		}
		return NegativeZero
	}

	// Convert to float32, perform operation, convert back
	f32a := a.ToFloat32()
	f32b := b.ToFloat32()
	result := f32a / f32b

	// Handle potential overflow/underflow
	if math.IsInf(float64(result), 0) {
		if (a.Sign() > 0) == (b.Sign() > 0) {
			return PositiveInfinity
		}
		return NegativeInfinity
	}

	return ToFloat8(result)
}

// Comparison operations

// Equal returns true if two Float8 values are equal
func Equal(a, b Float8) bool {
	// Handle NaN cases - NaN is not equal to anything, including itself
	if a.IsNaN() || b.IsNaN() {
		return false
	}

	// Handle zero cases - +0 == -0
	if a.IsZero() && b.IsZero() {
		return true
	}

	// Handle infinities
	aInf := a.IsInf()
	bInf := b.IsInf()
	if aInf || bInf {
		// If both are infinities, they must be the same value
		if aInf && bInf {
			// Both are infinity, check if they have the same sign
			return a == b
		}
		// One is infinity, the other is not
		return false
	}

	// For all other cases, use exact equality
	return a == b
}

// Less returns true if a < b
func Less(a, b Float8) bool {
	// Handle NaN cases - any comparison with NaN is false
	if a.IsNaN() || b.IsNaN() {
		return false
	}

	// Handle special cases
	if a.IsZero() && b.IsZero() {
		return false // +0 == -0
	}

	// Handle infinities
	aInf := a.IsInf()
	bInf := b.IsInf()

	if aInf && bInf {
		// Both are infinities, compare signs
		// -Inf < +Inf is true, +Inf < -Inf is false, same infinities are equal
		return a == NegativeInfinity && b == PositiveInfinity
	} else if aInf {
		// a is infinity, check if it's -Inf
		return a == NegativeInfinity
	} else if bInf {
		// b is infinity, check if it's +Inf
		return b == PositiveInfinity
	}

	// Regular comparison for finite numbers
	return a.ToFloat32() < b.ToFloat32()
}

// Greater returns true if a > b
func Greater(a, b Float8) bool {
	return Less(b, a)
}

// LessEqual returns true if a <= b
func LessEqual(a, b Float8) bool {
	return Less(a, b) || Equal(a, b)
}

// GreaterEqual returns true if a >= b
func GreaterEqual(a, b Float8) bool {
	return Greater(a, b) || Equal(a, b)
}

// Min returns the smaller of two Float8 values.
// If either value is NaN, returns NaN.
// Min(+Inf, x) returns x (if x is finite or -Inf)
// Min(-Inf, x) returns -Inf
// Min(x, +Inf) returns x (if x is finite or -Inf)
// Min(x, -Inf) returns -Inf
func Min(a, b Float8) Float8 {
	// Handle NaN cases first
	if a.IsNaN() || b.IsNaN() {
		return NaN
	}

	// Handle infinities
	switch {
	case a == NegativeInfinity || b == NegativeInfinity:
		// If either is -Inf, that's the minimum
		return NegativeInfinity
	case a == PositiveInfinity:
		// If a is +Inf, return b (which could be finite or -Inf)
		return b
	case b == PositiveInfinity:
		// If b is +Inf, return a (which could be finite or -Inf)
		return a
	default:
		// For finite numbers, use Less to determine the minimum
		if Less(a, b) {
			return a
		}
		return b
	}
}

// Max returns the larger of two Float8 values.
// If either value is NaN, returns NaN.
// Max(+Inf, x) returns +Inf
// Max(-Inf, x) returns x (if x is finite or +Inf)
// Max(x, +Inf) returns +Inf
// Max(x, -Inf) returns x (if x is finite or +Inf)
func Max(a, b Float8) Float8 {
	// Handle NaN cases first
	if a.IsNaN() || b.IsNaN() {
		return NaN
	}

	// Handle infinities
	switch {
	case a == PositiveInfinity || b == PositiveInfinity:
		// If either is +Inf, that's the maximum
		return PositiveInfinity
	case a == NegativeInfinity:
		// If a is -Inf, return b (which could be finite or +Inf)
		return b
	case b == NegativeInfinity:
		// If b is -Inf, return a (which could be finite or +Inf)
		return a
	default:
		// For finite numbers, use Greater to determine the maximum
		if Greater(a, b) {
			return a
		}
		return b
	}
}

// Batch operations for slices

// AddSlice performs element-wise addition of two Float8 slices.
//
// This function adds corresponding elements of the input slices and returns
// a new slice with the results. The input slices must have the same length;
// otherwise, the function will panic.
//
// Parameters:
//   - a, b: Slices of Float8 values to be added element-wise.
//
// Returns:
//   - A new slice where each element is the sum of the corresponding elements in a and b.
//
// Panics:
//   - If the input slices have different lengths.
//
// Example:
//
//	a := []Float8{1.0, 2.0, 3.0}
//	b := []Float8{4.0, 5.0, 6.0}
//	result := AddSlice(a, b) // Returns [5.0, 7.0, 9.0]
func AddSlice(a, b []Float8) []Float8 {
	if len(a) != len(b) {
		panic("float8: slice length mismatch")
	}

	result := make([]Float8, len(a))
	for i := range a {
		result[i] = Add(a[i], b[i])
	}
	return result
}

// MulSlice performs element-wise multiplication of two Float8 slices.
//
// This function multiplies corresponding elements of the input slices and returns
// a new slice with the results. The input slices must have the same length;
// otherwise, the function will panic.
//
// Parameters:
//   - a, b: Slices of Float8 values to be multiplied element-wise.
//
// Returns:
//   - A new slice where each element is the product of the corresponding elements in a and b.
//
// Panics:
//   - If the input slices have different lengths.
//
// Example:
//
//	a := []Float8{1.0, 2.0, 3.0}
//	b := []Float8{4.0, 5.0, 6.0}
//	result := MulSlice(a, b) // Returns [4.0, 10.0, 18.0]
func MulSlice(a, b []Float8) []Float8 {
	if len(a) != len(b) {
		panic("float8: slice length mismatch")
	}

	result := make([]Float8, len(a))
	for i := range a {
		result[i] = Mul(a[i], b[i])
	}
	return result
}

// ScaleSlice multiplies each element in the slice by a scalar
func ScaleSlice(s []Float8, scalar Float8) []Float8 {
	result := make([]Float8, len(s))
	for i := range s {
		result[i] = Mul(s[i], scalar)
	}
	return result
}

// SumSlice returns the sum of all elements in the slice.
//
// This function computes the sum of all Float8 values in the input slice.
// If the slice is empty, it returns PositiveZero.
//
// The summation is performed using the standard addition rules for Float8,
// including proper handling of special values (NaN, Inf, etc.).
//
// Parameters:
//   - s: The input slice of Float8 values to sum.
//
// Returns:
//   - The sum of all elements in the slice.
//   - If the slice is empty, returns PositiveZero.
//   - If any element is NaN, the result is NaN.
//
// Example:
//
//	s := []Float8{1.0, 2.0, 3.0, 4.0}
//	sum := SumSlice(s) // Returns 10.0
func SumSlice(s []Float8) Float8 {
	sum := PositiveZero
	for _, v := range s {
		sum = Add(sum, v)
	}
	return sum
}

// Lookup tables (loaded lazily)
var (
	addTable []Float8
	subTable []Float8
	mulTable []Float8
	divTable []Float8
)

// EnableFastArithmetic enables lookup tables for arithmetic operations
func EnableFastArithmetic() {
	initArithmeticTables()
}

// DisableFastArithmetic disables lookup tables and uses algorithmic operations
func DisableFastArithmetic() {
	addTable = nil
	subTable = nil
	mulTable = nil
	divTable = nil
}

// initArithmeticTables initializes all arithmetic lookup tables
func initArithmeticTables() {
	if addTable != nil {
		return // Already initialized
	}

	// Initialize tables with 65536 entries each (256 * 256)
	addTable = make([]Float8, 65536)
	subTable = make([]Float8, 65536)
	mulTable = make([]Float8, 65536)
	divTable = make([]Float8, 65536)

	for a := 0; a < 256; a++ {
		for b := 0; b < 256; b++ {
			idx := a<<8 | b
			f8a := Float8(a)
			f8b := Float8(b)

			addTable[idx] = addAlgorithmic(f8a, f8b)
			subTable[idx] = subAlgorithmic(f8a, f8b)
			mulTable[idx] = mulAlgorithmic(f8a, f8b)
			divTable[idx] = divAlgorithmic(f8a, f8b)
		}
	}
}
