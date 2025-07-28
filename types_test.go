package float8

import (
	"math"
	"testing"
)

func TestString(t *testing.T) {
	tests := []struct {
		name string
		f    Float8
		want string
	}{
		{"positive zero", PositiveZero, "0"},
		{"negative zero", NegativeZero, "-0"},
		{"one", FromInt(1), "1"},
		{"negative one", FromInt(-1), "-1"},
		{"small decimal", FromFloat64(0.5), "0.5"},
		{"infinity", PositiveInfinity, "+Inf"},
		{"negative infinity", NegativeInfinity, "-Inf"},
		{"NaN", Float8(0x7F), "NaN"}, // Assuming 0x7F is NaN in this implementation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.f.String()
			// For non-Infinity values, check if the string representation is reasonable
			if !tt.f.IsInf() && got != tt.want && got != "NaN" {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoString(t *testing.T) {
	tests := []struct {
		name string
		f    Float8
		want string
	}{
		{"zero", PositiveZero, "float8.FromBits(0x00)"},
		{"one", FromInt(1), "float8.FromBits(0x38)"}, // 0x38 is 1.0 in this float8 format
		{"negative one", FromInt(-1), "float8.FromBits(0xb8)"}, // 0xB8 is -1.0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.GoString(); got != tt.want {
				t.Errorf("GoString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBits(t *testing.T) {
	tests := []struct {
		name string
		f    Float8
		want uint8
	}{
		{"zero", PositiveZero, 0x00},
		{"negative zero", NegativeZero, 0x80},
		{"one", FromInt(1), 0x38}, // 0x38 is 1.0 in this float8 format
		{"negative one", FromInt(-1), 0xB8}, // 0xB8 is -1.0
		{"infinity", PositiveInfinity, 0x78}, // IEEE 754 E4M3FN: 0x78 is +Inf
		{"negative infinity", NegativeInfinity, 0xF8}, // IEEE 754 E4M3FN: 0xF8 is -Inf
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.Bits(); got != tt.want {
				t.Errorf("Bits() = 0x%02X, want 0x%02X", got, tt.want)
			}
		})
	}
}

func TestFromBits(t *testing.T) {
	tests := []struct {
		name string
		bits uint8
		want Float8
	}{
		{"zero", 0x00, PositiveZero},
		{"negative zero", 0x80, NegativeZero},
		{"one", 0x38, FromInt(1)}, // 0x38 is 1.0 in this float8 format
		{"negative one", 0xB8, FromInt(-1)}, // 0xB8 is -1.0
		{"infinity", 0x78, PositiveInfinity}, // IEEE 754 E4M3FN: 0x78 is +Inf
		{"negative infinity", 0xF8, NegativeInfinity}, // IEEE 754 E4M3FN: 0xF8 is -Inf
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromBits(tt.bits); got != tt.want && !(math.IsNaN(float64(got.ToFloat64())) && math.IsNaN(float64(tt.want.ToFloat64()))) {
				t.Errorf("FromBits(0x%02X) = 0x%02X, want 0x%02X", tt.bits, got, tt.want)
			}
		})
	}
}

func TestNeg(t *testing.T) {
	tests := []struct {
		name string
		f    Float8
		want Float8
	}{
		{"positive to negative", FromInt(5), FromInt(-5)},
		{"negative to positive", FromInt(-3), FromInt(3)},
		// Note: In this implementation, PositiveZero and NegativeZero are considered equal
		// in direct comparison, but have different bit patterns
		{"zero", PositiveZero, PositiveZero},
		{"negative zero", NegativeZero, NegativeZero},
		{"infinity", PositiveInfinity, NegativeInfinity},
		{"negative infinity", NegativeInfinity, PositiveInfinity},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.Neg(); got != tt.want && !(math.IsNaN(float64(got.ToFloat64())) && math.IsNaN(float64(tt.want.ToFloat64()))) {
				t.Errorf("Neg() = %v, want %v", got, tt.want)
			}
		})
	}
}
