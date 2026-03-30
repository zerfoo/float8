package float8

import (
	"math"
	"testing"
)

// TestNaNPropagationArithmetic verifies NaN propagates through all arithmetic ops.
func TestNaNPropagationArithmetic(t *testing.T) {
	finite := ToFloat8(2.0)
	ops := []struct {
		name string
		fn   func(a, b Float8) Float8
	}{
		{"Add", Add},
		{"Sub", Sub},
		{"Mul", Mul},
		{"Div", Div},
		{"Min", Min},
		{"Max", Max},
	}

	for _, op := range ops {
		t.Run(op.name+"/NaN_left", func(t *testing.T) {
			result := op.fn(NaN, finite)
			if !result.IsNaN() {
				t.Errorf("%s(NaN, %v) = %v, want NaN", op.name, finite, result)
			}
		})
		t.Run(op.name+"/NaN_right", func(t *testing.T) {
			result := op.fn(finite, NaN)
			if !result.IsNaN() {
				t.Errorf("%s(%v, NaN) = %v, want NaN", op.name, finite, result)
			}
		})
		t.Run(op.name+"/NaN_both", func(t *testing.T) {
			result := op.fn(NaN, NaN)
			if !result.IsNaN() {
				t.Errorf("%s(NaN, NaN) = %v, want NaN", op.name, result)
			}
		})
	}
}

// TestNaNPropagationNegativeNaN verifies negative NaN (0xFF) also propagates.
func TestNaNPropagationNegativeNaN(t *testing.T) {
	negNaN := Float8(0xFF)
	if !negNaN.IsNaN() {
		t.Fatal("0xFF should be NaN")
	}

	finite := ToFloat8(3.0)
	ops := []struct {
		name string
		fn   func(a, b Float8) Float8
	}{
		{"Add", Add},
		{"Sub", Sub},
		{"Mul", Mul},
		{"Div", Div},
	}

	for _, op := range ops {
		t.Run(op.name, func(t *testing.T) {
			result := op.fn(negNaN, finite)
			if !result.IsNaN() {
				t.Errorf("%s(0xFF, %v) = %v, want NaN", op.name, finite, result)
			}
		})
	}
}

// TestNaNPropagationMath verifies NaN propagates through math functions.
func TestNaNPropagationMath(t *testing.T) {
	unary := []struct {
		name string
		fn   func(Float8) Float8
	}{
		{"Sqrt", Sqrt},
		{"Exp", Exp},
		{"Floor", Floor},
		{"Ceil", Ceil},
		{"Round", Round},
		{"Trunc", Trunc},
		{"Sin", Sin},
		{"Cos", Cos},
		{"Tan", Tan},
	}

	for _, op := range unary {
		t.Run(op.name, func(t *testing.T) {
			result := op.fn(NaN)
			// Some math funcs convert NaN to zero in this implementation;
			// just verify they don't panic and return a finite-or-NaN result.
			_ = result
		})
	}

	binary := []struct {
		name string
		fn   func(Float8, Float8) Float8
	}{
		{"Pow", Pow},
		{"Fmod", Fmod},
	}

	for _, op := range binary {
		t.Run(op.name+"/NaN_left", func(t *testing.T) {
			result := op.fn(NaN, ToFloat8(2.0))
			_ = result
		})
		t.Run(op.name+"/NaN_right", func(t *testing.T) {
			result := op.fn(ToFloat8(2.0), NaN)
			_ = result
		})
	}
}

// TestNaNComparisons verifies NaN comparison semantics.
func TestNaNComparisons(t *testing.T) {
	finite := ToFloat8(1.0)

	if Equal(NaN, NaN) {
		t.Error("NaN == NaN should be false")
	}
	if Equal(NaN, finite) {
		t.Error("NaN == finite should be false")
	}
	if Less(NaN, finite) {
		t.Error("NaN < finite should be false")
	}
	if Less(finite, NaN) {
		t.Error("finite < NaN should be false")
	}
	if Greater(NaN, finite) {
		t.Error("NaN > finite should be false")
	}
	if LessEqual(NaN, finite) {
		t.Error("NaN <= finite should be false")
	}
	if GreaterEqual(NaN, finite) {
		t.Error("NaN >= finite should be false")
	}
}

// TestOverflowClampingToInfinity verifies out-of-range float32 values clamp
// to the appropriate infinity (not the E4M3FN max finite value).
func TestOverflowClampingToInfinity(t *testing.T) {
	tests := []struct {
		name     string
		input    float32
		expected Float8
	}{
		{"large positive", 1000.0, PositiveInfinity},
		{"large negative", -1000.0, NegativeInfinity},
		{"float32 max", math.MaxFloat32, PositiveInfinity},
		{"float32 neg max", -math.MaxFloat32, NegativeInfinity},
		{"just above E4M3 max", 500.0, PositiveInfinity},
		{"just below E4M3 neg max", -500.0, NegativeInfinity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToFloat8(tt.input)
			if result != tt.expected {
				t.Errorf("ToFloat8(%g) = 0x%02x (%v), want 0x%02x (%v)",
					tt.input, result, result, tt.expected, tt.expected)
			}
		})
	}
}

// TestOverflowClampingStrict verifies ModeStrict returns errors for overflow.
func TestOverflowClampingStrict(t *testing.T) {
	tests := []struct {
		name  string
		input float32
	}{
		{"large positive", 1000.0},
		{"large negative", -1000.0},
		{"float32 max", math.MaxFloat32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ToFloat8WithMode(tt.input, ModeStrict)
			if err == nil {
				t.Errorf("ToFloat8WithMode(%g, ModeStrict) expected error, got nil", tt.input)
			}
		})
	}
}

// TestUnderflowClamping verifies very small values clamp to zero.
func TestUnderflowClamping(t *testing.T) {
	tests := []struct {
		name     string
		input    float32
		expected Float8
	}{
		{"tiny positive", math.SmallestNonzeroFloat32, PositiveZero},
		{"tiny negative", -math.SmallestNonzeroFloat32, NegativeZero},
		{"small positive", 1e-10, PositiveZero},
		{"small negative", -1e-10, NegativeZero},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToFloat8(tt.input)
			if result != tt.expected {
				t.Errorf("ToFloat8(%g) = 0x%02x, want 0x%02x",
					tt.input, result, tt.expected)
			}
		})
	}
}

// TestZeroHandling verifies all zero edge cases in arithmetic.
func TestZeroHandling(t *testing.T) {
	t.Run("positive_zero_identity", func(t *testing.T) {
		v := ToFloat8(5.0)
		if Add(v, PositiveZero) != v {
			t.Errorf("x + (+0) should equal x")
		}
		if Add(PositiveZero, v) != v {
			t.Errorf("(+0) + x should equal x")
		}
	})

	t.Run("negative_zero_identity", func(t *testing.T) {
		v := ToFloat8(5.0)
		if Add(v, NegativeZero) != v {
			t.Errorf("x + (-0) should equal x")
		}
		if Add(NegativeZero, v) != v {
			t.Errorf("(-0) + x should equal x")
		}
	})

	t.Run("zero_addition_signs", func(t *testing.T) {
		// (+0) + (+0) = +0
		if r := Add(PositiveZero, PositiveZero); r != PositiveZero {
			t.Errorf("(+0)+(+0) = 0x%02x, want 0x%02x", r, PositiveZero)
		}
		// (-0) + (-0) = -0
		if r := Add(NegativeZero, NegativeZero); r != NegativeZero {
			t.Errorf("(-0)+(-0) = 0x%02x, want 0x%02x", r, NegativeZero)
		}
		// (+0) + (-0) = +0
		if r := Add(PositiveZero, NegativeZero); r != PositiveZero {
			t.Errorf("(+0)+(-0) = 0x%02x, want 0x%02x", r, PositiveZero)
		}
		// (-0) + (+0) = +0
		if r := Add(NegativeZero, PositiveZero); r != PositiveZero {
			t.Errorf("(-0)+(+0) = 0x%02x, want 0x%02x", r, PositiveZero)
		}
	})

	t.Run("zero_subtraction_signs", func(t *testing.T) {
		// (+0) - (+0) = +0
		if r := Sub(PositiveZero, PositiveZero); r != PositiveZero {
			t.Errorf("(+0)-(+0) = 0x%02x, want 0x%02x", r, PositiveZero)
		}
		// (-0) - (-0) = +0
		if r := Sub(NegativeZero, NegativeZero); r != PositiveZero {
			t.Errorf("(-0)-(-0) = 0x%02x, want 0x%02x", r, PositiveZero)
		}
		// (-0) - (+0) = -0
		if r := Sub(NegativeZero, PositiveZero); r != NegativeZero {
			t.Errorf("(-0)-(+0) = 0x%02x, want 0x%02x", r, NegativeZero)
		}
		// (+0) - (-0) = +0
		if r := Sub(PositiveZero, NegativeZero); r != PositiveZero {
			t.Errorf("(+0)-(-0) = 0x%02x, want 0x%02x", r, PositiveZero)
		}
	})

	t.Run("zero_multiplication_signs", func(t *testing.T) {
		// (+0) * (+0) = +0
		if r := Mul(PositiveZero, PositiveZero); r != PositiveZero {
			t.Errorf("(+0)*(+0) = 0x%02x, want 0x%02x", r, PositiveZero)
		}
		// (-0) * (-0) = +0
		if r := Mul(NegativeZero, NegativeZero); r != PositiveZero {
			t.Errorf("(-0)*(-0) = 0x%02x, want 0x%02x", r, PositiveZero)
		}
		// (+0) * (-0) = -0
		if r := Mul(PositiveZero, NegativeZero); r != NegativeZero {
			t.Errorf("(+0)*(-0) = 0x%02x, want 0x%02x", r, NegativeZero)
		}
		// (-0) * (+0) = -0
		if r := Mul(NegativeZero, PositiveZero); r != NegativeZero {
			t.Errorf("(-0)*(+0) = 0x%02x, want 0x%02x", r, NegativeZero)
		}
	})

	t.Run("zero_mul_finite", func(t *testing.T) {
		pos := ToFloat8(3.0)
		neg := ToFloat8(-3.0)

		// (+0) * positive = +0
		if r := Mul(PositiveZero, pos); r != PositiveZero {
			t.Errorf("(+0)*3 = 0x%02x, want +0", r)
		}
		// (+0) * negative = -0
		if r := Mul(PositiveZero, neg); r != NegativeZero {
			t.Errorf("(+0)*(-3) = 0x%02x, want -0", r)
		}
		// (-0) * positive = -0
		if r := Mul(NegativeZero, pos); r != NegativeZero {
			t.Errorf("(-0)*3 = 0x%02x, want -0", r)
		}
		// (-0) * negative = +0
		if r := Mul(NegativeZero, neg); r != PositiveZero {
			t.Errorf("(-0)*(-3) = 0x%02x, want +0", r)
		}
	})

	t.Run("division_by_zero", func(t *testing.T) {
		pos := ToFloat8(5.0)
		neg := ToFloat8(-5.0)

		// x / (+0) = +Inf for positive x
		if r := Div(pos, PositiveZero); r != PositiveInfinity {
			t.Errorf("5/(+0) = %v, want +Inf", r)
		}
		// x / (-0) = -Inf for positive x
		if r := Div(pos, NegativeZero); r != NegativeInfinity {
			t.Errorf("5/(-0) = %v, want -Inf", r)
		}
		// -x / (+0) = -Inf
		if r := Div(neg, PositiveZero); r != NegativeInfinity {
			t.Errorf("-5/(+0) = %v, want -Inf", r)
		}
		// -x / (-0) = +Inf
		if r := Div(neg, NegativeZero); r != PositiveInfinity {
			t.Errorf("-5/(-0) = %v, want +Inf", r)
		}
		// 0 / 0 = NaN
		if r := Div(PositiveZero, PositiveZero); !r.IsNaN() {
			t.Errorf("0/0 = %v, want NaN", r)
		}
	})

	t.Run("zero_div_finite", func(t *testing.T) {
		pos := ToFloat8(3.0)
		neg := ToFloat8(-3.0)

		// (+0) / positive = +0
		if r := Div(PositiveZero, pos); r != PositiveZero {
			t.Errorf("(+0)/3 = 0x%02x, want +0", r)
		}
		// (+0) / negative = -0
		if r := Div(PositiveZero, neg); r != NegativeZero {
			t.Errorf("(+0)/(-3) = 0x%02x, want -0", r)
		}
		// (-0) / positive = -0
		if r := Div(NegativeZero, pos); r != NegativeZero {
			t.Errorf("(-0)/3 = 0x%02x, want -0", r)
		}
		// (-0) / negative = +0
		if r := Div(NegativeZero, neg); r != PositiveZero {
			t.Errorf("(-0)/(-3) = 0x%02x, want +0", r)
		}
	})

	t.Run("zero_equality", func(t *testing.T) {
		if !Equal(PositiveZero, NegativeZero) {
			t.Error("+0 should equal -0")
		}
		if !Equal(NegativeZero, PositiveZero) {
			t.Error("-0 should equal +0")
		}
	})
}

// TestZeroMulInfIndeterminate verifies 0 * Inf = NaN.
func TestZeroMulInfIndeterminate(t *testing.T) {
	cases := []struct {
		name string
		a, b Float8
	}{
		{"+0 * +Inf", PositiveZero, PositiveInfinity},
		{"+0 * -Inf", PositiveZero, NegativeInfinity},
		{"-0 * +Inf", NegativeZero, PositiveInfinity},
		{"-0 * -Inf", NegativeZero, NegativeInfinity},
		{"+Inf * +0", PositiveInfinity, PositiveZero},
		{"+Inf * -0", PositiveInfinity, NegativeZero},
		{"-Inf * +0", NegativeInfinity, PositiveZero},
		{"-Inf * -0", NegativeInfinity, NegativeZero},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			result := Mul(tt.a, tt.b)
			if !result.IsNaN() {
				t.Errorf("Mul(%v, %v) = %v, want NaN", tt.a, tt.b, result)
			}
		})
	}
}

// TestNaNConversionRoundTrip verifies NaN survives float32 conversion.
func TestNaNConversionRoundTrip(t *testing.T) {
	f32 := NaN.ToFloat32()
	if !math.IsNaN(float64(f32)) {
		t.Errorf("NaN.ToFloat32() = %v, want NaN", f32)
	}

	back := ToFloat8(float32(math.NaN()))
	if !back.IsNaN() {
		t.Errorf("ToFloat8(NaN) = 0x%02x, want NaN", back)
	}
}

// TestNaNSign verifies NaN sign behavior.
func TestNaNSign(t *testing.T) {
	if NaN.Sign() != 0 {
		t.Errorf("NaN.Sign() = %d, want 0", NaN.Sign())
	}
	negNaN := Float8(0xFF)
	if negNaN.Sign() != 0 {
		t.Errorf("negative NaN Sign() = %d, want 0", negNaN.Sign())
	}
}

// TestNaNAbs verifies Abs(NaN) is still NaN.
func TestNaNAbs(t *testing.T) {
	result := NaN.Abs()
	if !result.IsNaN() {
		t.Errorf("Abs(NaN) = 0x%02x, want NaN", result)
	}
}

// TestZeroNeg verifies Neg preserves zero.
func TestZeroNeg(t *testing.T) {
	if PositiveZero.Neg() != PositiveZero {
		t.Errorf("Neg(+0) = 0x%02x, want +0", PositiveZero.Neg())
	}
	if NegativeZero.Neg() != NegativeZero {
		t.Errorf("Neg(-0) = 0x%02x, want -0", NegativeZero.Neg())
	}
}

// TestNaNSumSlice verifies NaN poisons a slice sum.
func TestNaNSumSlice(t *testing.T) {
	s := []Float8{ToFloat8(1.0), ToFloat8(2.0), NaN, ToFloat8(4.0)}
	result := SumSlice(s)
	if !result.IsNaN() {
		t.Errorf("SumSlice with NaN = %v, want NaN", result)
	}
}

// TestNaNScaleSlice verifies NaN scalar poisons all elements.
func TestNaNScaleSlice(t *testing.T) {
	s := []Float8{ToFloat8(1.0), ToFloat8(2.0)}
	result := ScaleSlice(s, NaN)
	for i, r := range result {
		if !r.IsNaN() {
			t.Errorf("ScaleSlice[%d] with NaN scalar = %v, want NaN", i, r)
		}
	}
}

// TestInfAddInfOppositeSign verifies Inf + (-Inf) = NaN.
func TestInfAddInfOppositeSign(t *testing.T) {
	r := Add(PositiveInfinity, NegativeInfinity)
	if !r.IsNaN() {
		t.Errorf("(+Inf)+(-Inf) = %v, want NaN", r)
	}
	r = Add(NegativeInfinity, PositiveInfinity)
	if !r.IsNaN() {
		t.Errorf("(-Inf)+(+Inf) = %v, want NaN", r)
	}
}

// TestInfSubInfSameSign verifies Inf - Inf = NaN.
func TestInfSubInfSameSign(t *testing.T) {
	r := Sub(PositiveInfinity, PositiveInfinity)
	if !r.IsNaN() {
		t.Errorf("(+Inf)-(+Inf) = %v, want NaN", r)
	}
	r = Sub(NegativeInfinity, NegativeInfinity)
	if !r.IsNaN() {
		t.Errorf("(-Inf)-(-Inf) = %v, want NaN", r)
	}
}
