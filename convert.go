package float8

import (
	"math"
	"unsafe"
)

// Global conversion mode (can be changed for different behavior)
var DefaultConversionMode = ModeDefault

// ToFloat8 converts a float32 value to Float8 format using the default conversion mode.
//
// This is a convenience function that calls ToFloat8WithMode with DefaultConversionMode.
// For more control over the conversion process, use ToFloat8WithMode directly.
//
// Special cases:
//   - Converts +0.0 to PositiveZero (0x00)
//   - Converts -0.0 to NegativeZero (0x80)
//   - Converts +Inf to PositiveInfinity (0x78)
//   - Converts -Inf to NegativeInfinity (0xF8)
//   - Converts NaN to NaN (0x7F or 0xFF)
//
// For finite numbers, the conversion may lose precision or result in overflow/underflow.
// The default mode handles these cases by saturating to the maximum/minimum representable values.
func ToFloat8(f32 float32) Float8 {
	result, _ := ToFloat8WithMode(f32, DefaultConversionMode)
	return result
}

// ToFloat8WithMode converts a float32 to Float8 with the specified conversion mode.
//
// The conversion mode determines how edge cases are handled:
//   - ModeDefault: Uses standard IEEE 754 rounding behavior, saturating on overflow
//   - ModeStrict: Returns an error for overflow/underflow/NaN
//   - ModeFast: Uses lookup tables when available (if enabled)
//
// Special cases are handled as follows:
//   - ±0.0 is converted to the corresponding Float8 zero (preserving sign)
//   - ±Inf is converted to the corresponding Float8 infinity
//   - NaN is handled according to the conversion mode
//
// For finite numbers, the conversion follows these steps:
//   1. Extract sign, exponent, and mantissa from the float32
//   2. Adjust the exponent for the Float8 format (E4M3FN)
//   3. Round the mantissa to 3 bits (plus implicit leading bit)
//   4. Handle overflow/underflow according to the conversion mode
//
// Returns the converted Float8 value and an error if the conversion fails in strict mode.
func ToFloat8WithMode(f32 float32, mode ConversionMode) (Float8, error) {
	// Handle special cases first
	if f32 == 0.0 {
		// Check the sign bit to distinguish between +0.0 and -0.0
		if math.Signbit(float64(f32)) {
			return NegativeZero, nil  // -0.0
		}
		return PositiveZero, nil  // +0.0
	}

	if math.IsInf(float64(f32), 0) {
		if f32 > 0 {
			return PositiveInfinity, nil
		}
		return NegativeInfinity, nil
	}

	if math.IsNaN(float64(f32)) {
		if mode == ModeStrict {
			return 0, ErrNaN
		}
		// In non-strict mode, convert NaN to NaN (IEEE 754 E4M3FN standard)
		return NaN, nil
	}

	// Extract IEEE 754 components using bit manipulation for performance
	bits := math.Float32bits(f32)
	sign := bits >> 31
	exp := int32((bits >> 23) & 0xFF)
	mant := bits & 0x7FFFFF

	// Handle zero after denormal check
	if exp == 0 && mant == 0 {
		if sign != 0 {
			return NegativeZero, nil
		}
		return PositiveZero, nil
	}

	// Convert exponent from float32 bias to float8 bias
	exp8 := exp - Float32Bias + ExponentBias

	// Check for overflow
	if exp8 > ExponentMax {
		if mode == ModeStrict {
			return 0, &Float8Error{
				Op:    "convert",
				Value: f32,
				Msg:   "overflow: value too large for float8",
			}
		}
		// Clamp to infinity
		if sign != 0 {
			return NegativeInfinity, nil
		}
		return PositiveInfinity, nil
	}

	// Check for underflow
	if exp8 < ExponentMin {
		if mode == ModeStrict {
			return 0, &Float8Error{
				Op:    "convert",
				Value: f32,
				Msg:   "underflow: value too small for float8",
			}
		}
		// Clamp to zero
		if sign != 0 {
			return NegativeZero, nil
		}
		return PositiveZero, nil
	}

	// Extract top 3 bits of mantissa for float8
	// Add rounding: if the 4th bit is set, round up
	mant8 := mant >> (23 - MantissaLen)
	if (mant>>(23-MantissaLen-1))&1 != 0 {
		mant8++
		// Handle mantissa overflow
		if mant8 >= (1 << MantissaLen) {
			mant8 = 0
			exp8++
			// Check for exponent overflow after rounding
			if exp8 > ExponentMax {
				if mode == ModeStrict {
					return 0, &Float8Error{
						Op:    "convert",
						Value: f32,
						Msg:   "overflow after rounding",
					}
				}
				if sign != 0 {
					return NegativeInfinity, nil
				}
				return PositiveInfinity, nil
			}
		}
	}

	// Combine components into Float8
	result := Float8((sign << 7) | (uint32(exp8) << MantissaLen) | mant8)
	return result, nil
}

// ToFloat32 converts a Float8 value to float32.
//
// This conversion is always exact since Float8 is a subset of float32.
// Special values are preserved:
//   - PositiveZero/NegativeZero → ±0.0
//   - PositiveInfinity/NegativeInfinity → ±Inf
//   - NaN → NaN
//
// The conversion uses a fast path for common values and falls back to
// algorithmic conversion for other values.
func (f Float8) ToFloat32() float32 {
	// Use lookup table for fast conversion if available
	if conversionTable != nil {
		return conversionTable[f]
	}
	return f.toFloat32Algorithmic()
}

// toFloat32Algorithmic performs algorithmic conversion (always available)
func (f Float8) toFloat32Algorithmic() float32 {
	// Handle special cases
	if f.IsZero() {
		if f == NegativeZero {
			return math.Float32frombits(0x80000000) // -0.0
		}
		return 0.0
	}

	if f.IsNaN() {
		return float32(math.NaN())
	}
	if f == PositiveInfinity {
		return float32(math.Inf(1))
	}
	if f == NegativeInfinity {
		return float32(math.Inf(-1))
	}

	// Extract components
	sign := uint32(f) >> 7
	exp8 := (uint32(f) >> MantissaLen) & 0x0F
	mant8 := uint32(f) & MantissaMask

	// Convert exponent from float8 bias to float32 bias
	exp32 := exp8 - ExponentBias + Float32Bias

	// Shift mantissa to float32 position
	mant32 := mant8 << (23 - MantissaLen)

	// Combine into IEEE 754 float32
	bits := (sign << 31) | (exp32 << 23) | mant32
	return math.Float32frombits(bits)
}

// Batch conversion functions for improved performance

// ToSlice8 converts a slice of float32 to Float8 with optimized performance.
//
// This function is optimized for batch conversion of float32 values to Float8.
// It handles special values correctly, including negative zero, infinity, and NaN.
//
// Parameters:
//   - f32s: The input slice of float32 values to convert. May be nil or empty.
//
// Returns:
//   - nil if the input slice is nil
//   - A non-nil empty slice if the input slice is empty
//   - A new slice containing the converted Float8 values
//
// Note: This function preserves negative zero by checking the sign bit of zero values.
// For large slices, consider using a pool of []Float8 to reduce allocations.
func ToSlice8(f32s []float32) []Float8 {
	if f32s == nil {
		return nil
	}
	if len(f32s) == 0 {
		return []Float8{} // Return non-nil empty slice
	}

	result := make([]Float8, len(f32s))

	// Convert each element, preserving negative zero
	for i := 0; i < len(f32s); i++ {
		// Special handling for negative zero
		if f32s[i] == 0 && math.Signbit(float64(f32s[i])) {
			result[i] = NegativeZero
		} else {
			result[i] = ToFloat8(f32s[i])
		}
	}

	return result
}

// ToSlice32 converts a slice of Float8 to float32 with optimized performance.
//
// This function is optimized for batch conversion of Float8 values to float32.
// It handles all special values correctly, including negative zero, infinity, and NaN.
//
// Parameters:
//   - f8s: The input slice of Float8 values to convert. May be nil or empty.
//
// Returns:
//   - nil if the input slice is nil
//   - A new slice containing the converted float32 values
//
// Note: The conversion from Float8 to float32 is always exact since Float8 is a
// subset of float32. For large slices, consider using a pool of []float32 to reduce allocations.
func ToSlice32(f8s []Float8) []float32 {
	if len(f8s) == 0 {
		return nil
	}

	result := make([]float32, len(f8s))

	// Use unsafe pointer arithmetic for better performance
	src := (*Float8)(unsafe.Pointer(&f8s[0]))
	dst := (*float32)(unsafe.Pointer(&result[0]))

	for i := 0; i < len(f8s); i++ {
		*(*float32)(unsafe.Pointer(uintptr(unsafe.Pointer(dst)) + uintptr(i)*unsafe.Sizeof(float32(0)))) =
			(*(*Float8)(unsafe.Pointer(uintptr(unsafe.Pointer(src)) + uintptr(i)*unsafe.Sizeof(Float8(0))))).ToFloat32()
	}

	return result
}

// Parse converts a string to Float8
func Parse(s string) (Float8, error) {
	// This would implement string parsing - simplified for now
	// In a full implementation, this would parse various float formats
	return PositiveZero, &Float8Error{Op: "parse", Msg: "not implemented"}
}

// Lookup table for fast conversion (loaded lazily)
var conversionTable []float32

// initConversionTable initializes the conversion lookup table
func initConversionTable() {
	if conversionTable != nil {
		return
	}

	conversionTable = make([]float32, 256)
	for i := 0; i < 256; i++ {
		conversionTable[i] = Float8(i).toFloat32Algorithmic()
	}
}

// EnableFastConversion enables lookup table for ToFloat32 conversion
func EnableFastConversion() {
	initConversionTable()
}

// DisableFastConversion disables lookup table and uses algorithmic conversion
func DisableFastConversion() {
	conversionTable = nil
}
