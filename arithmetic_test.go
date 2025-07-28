package float8

import (
	"testing"
)

func TestMulSlice(t *testing.T) {
	tests := []struct {
		a        []Float8
		b        []Float8
		expected []Float8
		hasPanic bool
	}{
		{
			a:        []Float8{One(), FromInt(2), FromInt(3)},
			b:        []Float8{FromInt(2), FromInt(3), FromInt(4)},
			expected: []Float8{FromInt(2), FromInt(6), FromInt(12)},
			hasPanic: false,
		},
		{
			a:        []Float8{},
			b:        []Float8{},
			expected: []Float8{},
			hasPanic: false,
		},
		{
			a:        []Float8{One(), FromInt(2)},
			b:        []Float8{One()}, // Different length - should panic
			hasPanic: true,
		},
	}

	for _, tt := range tests {
		func() {
			if tt.hasPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("Expected panic for slice length mismatch, but got none")
					}
				}()
			}

			result := MulSlice(tt.a, tt.b)

			if tt.hasPanic {
				t.Error("Expected panic but function completed successfully")
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Result length %d, expected %d", len(result), len(tt.expected))
				return
			}

			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("At index %d: got %v, want %v", i, result[i], tt.expected[i])
				}
			}
		}()
	}
}

func TestLess(t *testing.T) {
	tests := []struct {
		a        Float8
		b        Float8
		expected bool
		name     string
	}{
		// Test cases with NaN
		{NaN, One(), false, "NaN < number"},
		{One(), NaN, false, "number < NaN"},
		{NaN, NaN, false, "NaN < NaN"},

		// Test cases with zeros
		{PositiveZero, PositiveZero, false, "+0 < +0"},
		{NegativeZero, NegativeZero, false, "-0 < -0"},
		{NegativeZero, PositiveZero, false, "-0 < +0"},
		{PositiveZero, NegativeZero, false, "+0 < -0"},

		// Test cases with infinities
		{NegativeInfinity, PositiveInfinity, true, "-Inf < +Inf"},
		{PositiveInfinity, NegativeInfinity, false, "+Inf < -Inf"},
		{PositiveInfinity, PositiveInfinity, false, "+Inf < +Inf"},
		{NegativeInfinity, NegativeInfinity, false, "-Inf < -Inf"},
		{FromInt(1), PositiveInfinity, true, "1 < +Inf"},
		{FromInt(-1), NegativeInfinity, false, "-1 < -Inf"},
		{PositiveInfinity, FromInt(1), false, "+Inf < 1"},
		{NegativeInfinity, FromInt(-1), true, "-Inf < -1"},

		// Test cases with regular numbers
		{FromInt(1), FromInt(2), true, "1 < 2"},
		{FromInt(2), FromInt(1), false, "2 < 1"},
		{FromInt(-1), FromInt(1), true, "-1 < 1"},
		{FromInt(1), FromInt(1), false, "1 < 1"},
		{FromInt(1), FromInt(2), true, "1 < 2"},
		{FromInt(-2), FromInt(-1), true, "-2 < -1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Less(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Less(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}

			// Test commutativity for non-NaN cases
			if !tt.a.IsNaN() && !tt.b.IsNaN() && tt.a != tt.b {
				// If a < b is true, then b < a should be false and vice versa
				result2 := Less(tt.b, tt.a)
				if result && result2 {
					t.Errorf("Both Less(%v, %v) and Less(%v, %v) returned true", 
						tt.a, tt.b, tt.b, tt.a)
				}
			}
		})
	}
}

func TestAddSlice(t *testing.T) {
	tests := []struct {
		a        []Float8
		b        []Float8
		expected []Float8
		panic    bool
		name     string
	}{
		// Test case with regular numbers
		{
			a:        []Float8{One(), FromInt(2), FromInt(3)},
			b:        []Float8{One(), FromInt(3), FromInt(5)},
			expected: []Float8{FromInt(2), FromInt(5), FromInt(8)},
			panic:    false,
			name:     "regular numbers",
		},
		// Test case with empty slices
		{
			a:        []Float8{},
			b:        []Float8{},
			expected: []Float8{},
			panic:    false,
			name:     "empty slices",
		},
		// Test case with special values (infinity, zero, etc.)
		{
			a:        []Float8{PositiveInfinity, NegativeInfinity, PositiveZero, NegativeZero, NaN},
			b:        []Float8{One(), One(), One(), One(), One()},
			expected: []Float8{PositiveInfinity, NegativeInfinity, One(), One(), NaN},
			panic:    false,
			name:     "special values",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("Expected panic but got none")
					}
				}()
			}

			result := AddSlice(tt.a, tt.b)

			if tt.panic {
				t.Error("Expected panic but function completed successfully")
				return
			}

			if len(result) != len(tt.expected) {
				t.Fatalf("Expected result length %d, got %d", len(tt.expected), len(result))
			}

			for i := range result {
				if !(result[i] == tt.expected[i] || (result[i].IsNaN() && tt.expected[i].IsNaN())) {
					t.Errorf("At index %d: expected %v, got %v", i, tt.expected[i], result[i])
				}
			}
		})
	}

	// Test panic case separately
	t.Run("panic on length mismatch", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for slice length mismatch, but got none")
			}
		}()

		a := []Float8{One(), FromInt(2)}
		b := []Float8{One()} // Different length - should panic
		_ = AddSlice(a, b)
		t.Error("Expected panic but function completed successfully")
	})
}

func TestDisableFastArithmetic(t *testing.T) {
	// First enable fast arithmetic to set up the tables
	EnableFastArithmetic()

	// Verify tables are initialized
	if addTable == nil || subTable == nil || mulTable == nil || divTable == nil {
		t.Error("Expected tables to be initialized after EnableFastArithmetic")
	}

	// Now disable fast arithmetic
	DisableFastArithmetic()

	// Verify tables are nil after disabling
	if addTable != nil || subTable != nil || mulTable != nil || divTable != nil {
		t.Error("Expected tables to be nil after DisableFastArithmetic")
	}

	// Test that operations still work after disabling (should fall back to algorithmic)
	a := FromInt(5)
	b := FromInt(3)
	sum := Add(a, b)
	if sum != FromInt(8) {
		t.Errorf("Expected 5 + 3 = 8 after DisableFastArithmetic, got %v", sum)
	}
}

// TestDivisionEdgeCases tests edge cases in division to achieve 100% coverage
func TestDivisionEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		a, b     Float8
		expected Float8
	}{
		// Test infinity / infinity = NaN
		{"positive inf / positive inf", PositiveInfinity, PositiveInfinity, NaN},
		{"negative inf / negative inf", NegativeInfinity, NegativeInfinity, NaN},
		{"positive inf / negative inf", PositiveInfinity, NegativeInfinity, NaN},
		{"negative inf / positive inf", NegativeInfinity, PositiveInfinity, NaN},
		
		// Test finite / infinity = 0 with proper sign
		{"positive / positive inf", FromInt(5), PositiveInfinity, PositiveZero},
		{"negative / positive inf", FromInt(-5), PositiveInfinity, NegativeZero},
		{"positive / negative inf", FromInt(5), NegativeInfinity, NegativeZero},
		{"negative / negative inf", FromInt(-5), NegativeInfinity, PositiveZero},
		
		// Test overflow cases that result in infinity (using max values)
		{"max positive / small positive", ToFloat8(448.0), ToFloat8(0.0625), PositiveInfinity},
		{"max negative / small positive", ToFloat8(-448.0), ToFloat8(0.0625), NegativeInfinity},
		
		// Test cases that trigger overflow detection in float32 division
		{"large positive / tiny positive", ToFloat8(240.0), ToFloat8(0.0078125), PositiveInfinity},
		{"large negative / tiny positive", ToFloat8(-240.0), ToFloat8(0.0078125), NegativeInfinity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Div(tt.a, tt.b)
			if tt.expected.IsNaN() {
				if !result.IsNaN() {
					t.Errorf("Div(%v, %v) = %v, want NaN", tt.a, tt.b, result)
				}
			} else if result != tt.expected {
				t.Errorf("Div(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
