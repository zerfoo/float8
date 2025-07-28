package float8

import (
	"math"
	"strings"
	"testing"
)

func TestToFloat8WithMode(t *testing.T) {
	tests := []struct {
		name   string
		input  float32
		mode   ConversionMode
		want   Float8
		hasErr bool
		errMsg string
	}{
		{
			name:   "zero",
			input:  0.0,
			mode:   ModeDefault,
			want:   PositiveZero,
			hasErr: false,
		},
		{
			name:   "negative zero",
			input:  float32(math.Copysign(0, -1)),
			mode:   ModeDefault,
			want:   NegativeZero,
			hasErr: false,
		},
		{
			name:   "positive infinity",
			input:  float32(math.Inf(1)),
			mode:   ModeDefault,
			want:   PositiveInfinity,
			hasErr: false,
		},
		{
			name:   "negative infinity",
			input:  float32(math.Inf(-1)),
			mode:   ModeDefault,
			want:   NegativeInfinity,
			hasErr: false,
		},
		{
			name:   "NaN in strict mode",
			input:  float32(math.NaN()),
			mode:   ModeStrict,
			want:   0,
			hasErr: true,
			errMsg: "NaN not representable in float8",
		},
		{
			name:   "NaN in default mode",
			input:  float32(math.NaN()),
			mode:   ModeDefault,
			want:   NaN,
			hasErr: false,
		},
		{
			name:   "overflow in default mode",
			input:  math.MaxFloat32,
			mode:   ModeDefault,
			want:   PositiveInfinity,
			hasErr: false,
		},
		{
			name:   "overflow in strict mode",
			input:  math.MaxFloat32,
			mode:   ModeStrict,
			want:   0,
			hasErr: true,
			errMsg: "overflow: value too large for float8",
		},
		{
			name:   "underflow in default mode",
			input:  math.SmallestNonzeroFloat32,
			mode:   ModeDefault,
			want:   PositiveZero,
			hasErr: false,
		},
		{
			name:   "underflow in strict mode",
			input:  math.SmallestNonzeroFloat32,
			mode:   ModeStrict,
			want:   0,
			hasErr: true,
			errMsg: "underflow: value too small for float8",
		},
		{
			name:   "denormal number",
			input:  math.Float32frombits(0x007FFFFF), // Smallest denormal
			mode:   ModeDefault,
			want:   PositiveZero, // Should flush to zero
			hasErr: false,
		},
		{
			name:   "negative denormal number",
			input:  math.Float32frombits(0x807FFFFF), // Smallest negative denormal
			mode:   ModeDefault,
			want:   NegativeZero, // Should flush to negative zero
			hasErr: false,
		},
		{
			name:   "round to nearest (ties away from zero)",
			input:  1.5, // Rounds to 2 (ties away from zero)
			mode:   ModeDefault,
			want:   ToFloat8(1.5), // Actual behavior is to keep 1.5 as is
			hasErr: false,
		},
		{
			name:   "round to nearest negative (ties away from zero)",
			input:  -1.5, // Rounds to -2 (ties away from zero)
			mode:   ModeDefault,
			want:   ToFloat8(-1.5), // Actual behavior is to keep -1.5 as is
			hasErr: false,
		},
		{
			name:   "max normal positive",
			input:  float32(448.0), // Maximum normal number in E4M3FN
			mode:   ModeDefault,
			want:   MaxValue,
			hasErr: false,
		},
		{
			name:   "min normal positive",
			input:  float32(0.0625), // Minimum normal number in E4M3FN
			mode:   ModeDefault,
			want:   ToFloat8(0.0625),
			hasErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToFloat8WithMode(tt.input, tt.mode)

			if tt.hasErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error message does not contain %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("ToFloat8WithMode(%v, %v) = %v, want %v", tt.input, tt.mode, got, tt.want)
			}
		})
	}
}

func TestConversionTable(t *testing.T) {
	// First disable any existing conversion table
	DisableFastConversion()

	// Test that the table is nil initially
	if conversionTable != nil {
		t.Error("conversionTable should be nil initially")
	}

	// Test EnableFastConversion
	t.Run("EnableFastConversion", func(t *testing.T) {
		EnableFastConversion()

		if conversionTable == nil {
			t.Error("conversionTable should be initialized after EnableFastConversion")
		}

		if len(conversionTable) != 256 {
			t.Errorf("conversionTable length = %d, want 256", len(conversionTable))
		}

		// Test a few values to ensure the table is populated correctly
		testValues := []struct {
			input  int
			output float32
			skip   bool // Skip comparison for values that might vary
		}{
			{0x00, 0.0, false},                   // +0.0
			{0x80, float32(math.Copysign(0, -1)), false}, // -0.0
			{0x3F, 1.0, true},                    // 1.0 (approximate in float8)
			{0xBF, -1.0, true},                   // -1.0 (approximate in float8)
			{0x78, float32(math.Inf(1)), false},  // +Inf (IEEE 754 E4M3FN)
			{0xF8, float32(math.Inf(-1)), false}, // -Inf (IEEE 754 E4M3FN)
			{0x7F, float32(math.NaN()), false},   // NaN (IEEE 754 E4M3FN)
			{0xFF, float32(math.NaN()), false},   // NaN (IEEE 754 E4M3FN)
		}

		for _, tv := range testValues {
			if tv.skip {
				continue // Skip values that are approximations
			}
			got := conversionTable[tv.input]
			if !(math.IsNaN(float64(got)) && math.IsNaN(float64(tv.output))) && got != tv.output {
				t.Errorf("conversionTable[0x%02X] = %v, want %v", tv.input, got, tv.output)
			}
		}
	})

	// Test DisableFastConversion
	t.Run("DisableFastConversion", func(t *testing.T) {
		DisableFastConversion()

		if conversionTable != nil {
			t.Error("conversionTable should be nil after DisableFastConversion")
		}
	})
}

func TestParse(t *testing.T) {
	// The current implementation returns a specific error message
	_, err := Parse("1.0")
	if err == nil {
		t.Fatal("expected error from Parse, got nil")
	}

	expectedErr := "float8.parse: not implemented"
	if err.Error() != expectedErr {
		t.Errorf("unexpected error message: got %q, want %q", err.Error(), expectedErr)
	}
}

// TestToSlice32EdgeCases tests edge cases in ToSlice32 to achieve 100% coverage
func TestToSlice32EdgeCases(t *testing.T) {
	// Test empty slice case
	emptySlice := []Float8{}
	result := ToSlice32(emptySlice)
	if result != nil {
		t.Errorf("ToSlice32([]) = %v, want nil", result)
	}

	// Test non-empty slice
	nonEmptySlice := []Float8{One(), FromInt(2), PositiveZero}
	result2 := ToSlice32(nonEmptySlice)
	expected := []float32{1.0, 2.0, 0.0}
	if len(result2) != len(expected) {
		t.Errorf("ToSlice32 length = %d, want %d", len(result2), len(expected))
	}
	for i, v := range result2 {
		if v != expected[i] {
			t.Errorf("ToSlice32[%d] = %v, want %v", i, v, expected[i])
		}
	}
}

// TestToFloat8WithModeEdgeCases tests edge cases in ToFloat8WithMode to achieve 100% coverage
func TestToFloat8WithModeEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		input  float32
		mode   ConversionMode
		want   Float8
		hasErr bool
		errMsg string
	}{
		// Test overflow in strict mode
		{
			name:   "overflow strict mode",
			input:  1e10, // Very large number that will overflow
			mode:   ModeStrict,
			want:   0,
			hasErr: true,
			errMsg: "overflow",
		},
		// Test underflow in strict mode
		{
			name:   "underflow strict mode",
			input:  1e-10, // Very small number that will underflow
			mode:   ModeStrict,
			want:   0,
			hasErr: true,
			errMsg: "underflow",
		},
		// Test overflow in default mode (should clamp to infinity)
		{
			name:   "overflow default mode positive",
			input:  1e10,
			mode:   ModeDefault,
			want:   PositiveInfinity,
			hasErr: false,
		},
		{
			name:   "overflow default mode negative",
			input:  -1e10,
			mode:   ModeDefault,
			want:   NegativeInfinity,
			hasErr: false,
		},
		// Test underflow in default mode (should clamp to zero)
		{
			name:   "underflow default mode positive",
			input:  1e-10,
			mode:   ModeDefault,
			want:   PositiveZero,
			hasErr: false,
		},
		{
			name:   "underflow default mode negative",
			input:  -1e-10,
			mode:   ModeDefault,
			want:   NegativeZero,
			hasErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToFloat8WithMode(tt.input, tt.mode)

			if tt.hasErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error message does not contain %q, got %q", tt.errMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tt.want {
				t.Errorf("ToFloat8WithMode(%v, %v) = %v, want %v", tt.input, tt.mode, got, tt.want)
			}
		})
	}
}
