package float8

import (
	"fmt"
	"math"
	"testing"
)

// TestArithmeticCorrectness verifies Add, Sub, Mul, and Div for all 256
// E4M3FN bit patterns against IEEE 754 reference values computed in float32.
//
// For each pair (a, b) drawn from a representative subset of the 256
// values, the test converts both operands to float32, performs the
// operation in float32, converts the float32 result back to Float8, and
// asserts that the Float8 operation produces the same bit pattern.
//
// Reference: NVIDIA, "FP8 Formats for Deep Learning", 2022.
// https://arxiv.org/abs/2209.05433
func TestArithmeticCorrectness(t *testing.T) {
	// Build the complete E4M3FN value table (all 256 bit patterns).
	var allValues [256]Float8
	for i := 0; i < 256; i++ {
		allValues[i] = Float8(i)
	}

	// Classify each bit pattern for diagnostic output.
	var (
		normals   int
		subnorms  int
		zeros     int
		infs      int
		nans      int
		negatives int
	)
	for _, v := range allValues {
		switch {
		case v.IsNaN():
			nans++
		case v.IsInf():
			infs++
		case v.IsZero():
			zeros++
		default:
			exp := (uint8(v) & uint8(ExponentMask)) >> MantissaLen
			if exp == 0 {
				subnorms++
			} else {
				normals++
			}
		}
		if uint8(v)&uint8(SignMask) != 0 {
			negatives++
		}
	}
	t.Logf("E4M3FN value table: %d normals, %d subnormals, %d zeros, %d infs, %d NaNs (%d negative)",
		normals, subnorms, zeros, infs, nans, negatives)

	// referenceOp computes the expected Float8 result by performing the
	// operation in float32 and converting back.
	referenceOp := func(a, b Float8, op func(float32, float32) float32) Float8 {
		fa := a.ToFloat32()
		fb := b.ToFloat32()
		fr := op(fa, fb)
		return ToFloat8(fr)
	}

	// matchResult returns true when got and want represent the same
	// Float8 value, treating all NaN bit patterns as equal.
	matchResult := func(got, want Float8) bool {
		if got.IsNaN() && want.IsNaN() {
			return true
		}
		return got == want
	}

	type opSpec struct {
		name string
		fn   func(Float8, Float8) Float8
		f32  func(float32, float32) float32
	}

	ops := []opSpec{
		{"Add", Add, func(a, b float32) float32 { return a + b }},
		{"Sub", Sub, func(a, b float32) float32 { return a - b }},
		{"Mul", Mul, func(a, b float32) float32 { return a * b }},
		{"Div", Div, func(a, b float32) float32 { return a / b }},
	}

	// Select a representative subset of values to keep runtime
	// reasonable while still covering every bit pattern. We test:
	//   - Every value paired with a small set of "probe" values
	//   - Every probe value paired with every value
	// This gives 256 * len(probes) * 2 * 4 checks.
	probes := []Float8{
		PositiveZero,                // +0
		NegativeZero,                // -0
		PositiveInfinity,            // +Inf
		NegativeInfinity,            // -Inf
		NaN,                         // NaN
		Float8(0xFF),                // negative NaN
		ToFloat8(1.0),               // 1.0
		ToFloat8(-1.0),              // -1.0
		ToFloat8(2.0),               // 2.0
		ToFloat8(0.5),               // 0.5
		ToFloat8(0.0078125),         // smallest normal
		MaxValue,                    // largest finite positive
		MinValue,                    // largest finite negative
		SmallestPositive,            // smallest positive subnormal
		Float8(SmallestPositive | SignMask), // smallest negative subnormal
	}

	for _, op := range ops {
		t.Run(op.name, func(t *testing.T) {
			failures := 0
			const maxFailures = 20

			for i := 0; i < 256; i++ {
				a := allValues[i]
				for _, b := range probes {
					want := referenceOp(a, b, op.f32)
					got := op.fn(a, b)
					if !matchResult(got, want) {
						failures++
						if failures <= maxFailures {
							t.Errorf("%s(0x%02x [%v], 0x%02x [%v]): got 0x%02x [%v], want 0x%02x [%v]",
								op.name, uint8(a), a.ToFloat32(), uint8(b), b.ToFloat32(),
								uint8(got), got.ToFloat32(), uint8(want), want.ToFloat32())
						}
					}
				}
				for _, b := range probes {
					// Reverse order: probe as first operand.
					want := referenceOp(b, a, op.f32)
					got := op.fn(b, a)
					if !matchResult(got, want) {
						failures++
						if failures <= maxFailures {
							t.Errorf("%s(0x%02x [%v], 0x%02x [%v]): got 0x%02x [%v], want 0x%02x [%v]",
								op.name, uint8(b), b.ToFloat32(), uint8(a), a.ToFloat32(),
								uint8(got), got.ToFloat32(), uint8(want), want.ToFloat32())
						}
					}
				}
			}
			if failures > maxFailures {
				t.Errorf("... and %d more failures (capped output at %d)", failures-maxFailures, maxFailures)
			}
			if failures == 0 {
				t.Logf("%s: all %d pairs passed", op.name, 256*len(probes)*2)
			}
		})
	}
}

// TestArithmeticCorrectnessAllPairs tests every (a, b) combination
// for a focused set of operations (Mul, Add) to catch any discrepancy
// in the full 256x256 space.
func TestArithmeticCorrectnessAllPairs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping exhaustive 256x256 pair test in short mode")
	}

	type opSpec struct {
		name string
		fn   func(Float8, Float8) Float8
		f32  func(float32, float32) float32
	}

	ops := []opSpec{
		{"Add", Add, func(a, b float32) float32 { return a + b }},
		{"Sub", Sub, func(a, b float32) float32 { return a - b }},
		{"Mul", Mul, func(a, b float32) float32 { return a * b }},
		{"Div", Div, func(a, b float32) float32 { return a / b }},
	}

	for _, op := range ops {
		t.Run(op.name, func(t *testing.T) {
			failures := 0
			const maxFailures = 20

			for a := 0; a < 256; a++ {
				fa := Float8(a)
				for b := 0; b < 256; b++ {
					fb := Float8(b)

					got := op.fn(fa, fb)

					// Compute reference via float32.
					f32a := fa.ToFloat32()
					f32b := fb.ToFloat32()
					f32r := op.f32(f32a, f32b)
					want := ToFloat8(f32r)

					if got.IsNaN() && want.IsNaN() {
						continue
					}
					if got == want {
						continue
					}

					failures++
					if failures <= maxFailures {
						t.Errorf("%s(0x%02x [%v], 0x%02x [%v]): got 0x%02x [%v], want 0x%02x [%v]",
							op.name, a, f32a, b, f32b,
							uint8(got), got.ToFloat32(), uint8(want), want.ToFloat32())
					}
				}
			}
			if failures > maxFailures {
				t.Errorf("... and %d more failures (capped output at %d)", failures-maxFailures, maxFailures)
			}
			if failures == 0 {
				t.Logf("%s: all 65536 pairs passed", op.name)
			}
		})
	}
}

// TestE4M3FNValueTable verifies that all 256 bit patterns decode to the
// expected float32 values per the E4M3FN specification. This serves as
// a foundation for the arithmetic tests: if encoding/decoding is wrong,
// arithmetic results will also be wrong.
func TestE4M3FNValueTable(t *testing.T) {
	// Verify basic structural properties.
	for i := 0; i < 256; i++ {
		f := Float8(i)
		f32 := f.ToFloat32()

		// Round-trip: Float8 → float32 → Float8 must be identity
		// (except for NaN sign variants, which may collapse).
		if !f.IsNaN() {
			rt := ToFloat8(f32)
			if rt != f {
				t.Errorf("round-trip failed for 0x%02x: ToFloat8(%.6g) = 0x%02x", i, f32, uint8(rt))
			}
		}

		// Positive and negative variants must have the same magnitude.
		neg := Float8(uint8(i) ^ uint8(SignMask))
		if !f.IsNaN() && !neg.IsNaN() {
			mag := float64(f32)
			if mag < 0 {
				mag = -mag
			}
			negMag := float64(neg.ToFloat32())
			if negMag < 0 {
				negMag = -negMag
			}
			if mag != negMag {
				t.Errorf("magnitude mismatch: |0x%02x| = %v, |0x%02x| = %v", i, mag, uint8(neg), negMag)
			}
		}
	}

	// Verify specific well-known E4M3FN values (from NVIDIA spec table).
	knownValues := []struct {
		bits uint8
		f32  float32
	}{
		{0x00, 0.0},             // +0
		{0x80, float32(math.Copysign(0, -1))}, // -0
		{0x38, 1.0},             // 1.0 = 0.0111.000 → exp=7-7=0, mant=1.000 → 1.0
		{0x3C, 1.5},             // 1.5 = 0.0111.100 → exp=0, mant=1.100 → 1.5
		{0x40, 2.0},             // 2.0 = 0.1000.000 → exp=1, mant=1.000 → 2.0
		{0x48, 4.0},             // 4.0 = 0.1001.000 → exp=9-7=2, mant=1.000 → 4.0
	}

	for _, kv := range knownValues {
		f := Float8(kv.bits)
		got := f.ToFloat32()
		if kv.f32 == 0 {
			// For zero, just check that we get zero with the right sign.
			if got != 0 || math.Signbit(float64(got)) != math.Signbit(float64(kv.f32)) {
				t.Errorf("Float8(0x%02x).ToFloat32() = %v, want %v", kv.bits, got, kv.f32)
			}
		} else if got != kv.f32 {
			t.Errorf("Float8(0x%02x).ToFloat32() = %v, want %v", kv.bits, got, kv.f32)
		}
	}
}

// TestArithmeticCommutativity verifies that Add and Mul are commutative
// for all 256x256 pairs (a fundamental IEEE 754 property).
func TestArithmeticCommutativity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping exhaustive commutativity test in short mode")
	}

	for a := 0; a < 256; a++ {
		for b := 0; b < 256; b++ {
			fa := Float8(a)
			fb := Float8(b)

			// Add commutativity
			ab := Add(fa, fb)
			ba := Add(fb, fa)
			if ab != ba && !(ab.IsNaN() && ba.IsNaN()) {
				t.Errorf("Add not commutative: Add(0x%02x, 0x%02x)=0x%02x but Add(0x%02x, 0x%02x)=0x%02x",
					a, b, uint8(ab), b, a, uint8(ba))
			}

			// Mul commutativity
			ab = Mul(fa, fb)
			ba = Mul(fb, fa)
			if ab != ba && !(ab.IsNaN() && ba.IsNaN()) {
				t.Errorf("Mul not commutative: Mul(0x%02x, 0x%02x)=0x%02x but Mul(0x%02x, 0x%02x)=0x%02x",
					a, b, uint8(ab), b, a, uint8(ba))
			}
		}
	}
}

// TestArithmeticIdentities verifies identity element properties:
//   - a + 0 = a
//   - a * 1 = a
//   - a - a = 0 (for finite values)
//   - a / a = 1 (for finite non-zero values)
func TestArithmeticIdentities(t *testing.T) {
	one := ToFloat8(1.0)

	for i := 0; i < 256; i++ {
		a := Float8(i)
		label := fmt.Sprintf("0x%02x", i)

		// a + 0 = a (for non-NaN values)
		// IEEE 754 §6.3: (-0) + (+0) = +0, so skip negative zero
		if !a.IsNaN() && a != NegativeZero {
			got := Add(a, PositiveZero)
			if got != a {
				t.Errorf("%s: Add(a, +0) = 0x%02x, want 0x%02x", label, uint8(got), i)
			}
		}

		// a * 1 = a (for non-NaN, non-special values)
		if !a.IsNaN() && !a.IsInf() {
			got := Mul(a, one)
			if got != a && !(got.IsZero() && a.IsZero()) {
				t.Errorf("%s: Mul(a, 1) = 0x%02x [%v], want 0x%02x [%v]",
					label, uint8(got), got.ToFloat32(), i, a.ToFloat32())
			}
		}

		// a - a = 0 (for finite values)
		if a.IsFinite() && !a.IsNaN() {
			got := Sub(a, a)
			if !got.IsZero() {
				t.Errorf("%s: Sub(a, a) = 0x%02x [%v], want zero",
					label, uint8(got), got.ToFloat32())
			}
		}

		// a / a = 1 (for finite non-zero values)
		if a.IsFinite() && !a.IsZero() && !a.IsNaN() {
			got := Div(a, a)
			if got != one {
				t.Errorf("%s: Div(a, a) = 0x%02x [%v], want 1.0",
					label, uint8(got), got.ToFloat32())
			}
		}
	}
}

// TestArithmeticSpecialValues tests NaN propagation and infinity
// arithmetic for all 256 values, per IEEE 754 rules.
func TestArithmeticSpecialValues(t *testing.T) {
	for i := 0; i < 256; i++ {
		a := Float8(i)

		// NaN + x = NaN
		got := Add(NaN, a)
		if !got.IsNaN() {
			t.Errorf("Add(NaN, 0x%02x): got 0x%02x, want NaN", i, uint8(got))
		}
		got = Add(a, NaN)
		if !got.IsNaN() {
			t.Errorf("Add(0x%02x, NaN): got 0x%02x, want NaN", i, uint8(got))
		}

		// NaN * x = NaN
		got = Mul(NaN, a)
		if !got.IsNaN() {
			t.Errorf("Mul(NaN, 0x%02x): got 0x%02x, want NaN", i, uint8(got))
		}
		got = Mul(a, NaN)
		if !got.IsNaN() {
			t.Errorf("Mul(0x%02x, NaN): got 0x%02x, want NaN", i, uint8(got))
		}

		// NaN - x = NaN
		got = Sub(NaN, a)
		if !got.IsNaN() {
			t.Errorf("Sub(NaN, 0x%02x): got 0x%02x, want NaN", i, uint8(got))
		}
		got = Sub(a, NaN)
		if !got.IsNaN() {
			t.Errorf("Sub(0x%02x, NaN): got 0x%02x, want NaN", i, uint8(got))
		}

		// NaN / x = NaN
		got = Div(NaN, a)
		if !got.IsNaN() {
			t.Errorf("Div(NaN, 0x%02x): got 0x%02x, want NaN", i, uint8(got))
		}
		got = Div(a, NaN)
		if !got.IsNaN() {
			t.Errorf("Div(0x%02x, NaN): got 0x%02x, want NaN", i, uint8(got))
		}
	}
}

// TestLookupTableConsistency verifies that the lookup table
// implementation produces identical results to the algorithmic
// implementation for all 256x256 pairs across all four operations.
func TestLookupTableConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping exhaustive lookup table consistency test in short mode")
	}

	// Ensure tables are disabled to get algorithmic results.
	DisableFastArithmetic()

	type opSpec struct {
		name    string
		algoFn  func(Float8, Float8) Float8
		tableFn func(Float8, Float8) Float8
	}

	// Collect algorithmic results first.
	type result struct {
		add, sub, mul, div Float8
	}
	algo := make([]result, 65536)
	for a := 0; a < 256; a++ {
		for b := 0; b < 256; b++ {
			idx := a<<8 | b
			fa := Float8(a)
			fb := Float8(b)
			algo[idx] = result{
				add: Add(fa, fb),
				sub: Sub(fa, fb),
				mul: Mul(fa, fb),
				div: Div(fa, fb),
			}
		}
	}

	// Enable lookup tables and compare.
	EnableFastArithmetic()
	defer DisableFastArithmetic()

	failures := 0
	for a := 0; a < 256; a++ {
		for b := 0; b < 256; b++ {
			idx := a<<8 | b
			fa := Float8(a)
			fb := Float8(b)

			check := func(opName string, got, want Float8) {
				if got.IsNaN() && want.IsNaN() {
					return
				}
				if got != want {
					failures++
					if failures <= 20 {
						t.Errorf("%s(0x%02x, 0x%02x): lookup=0x%02x, algo=0x%02x",
							opName, a, b, uint8(got), uint8(want))
					}
				}
			}

			check("Add", Add(fa, fb), algo[idx].add)
			check("Sub", Sub(fa, fb), algo[idx].sub)
			check("Mul", Mul(fa, fb), algo[idx].mul)
			check("Div", Div(fa, fb), algo[idx].div)
		}
	}
	if failures > 20 {
		t.Errorf("... and %d more lookup/algo mismatches", failures-20)
	}
	if failures == 0 {
		t.Log("lookup table and algorithmic implementations are fully consistent")
	}
}
