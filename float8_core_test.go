package float8

import (
	"math"
	"testing"
)

func TestInitialize(t *testing.T) {
	// Test default initialization
	Initialize() // No error expected

	// Test reinitialization with custom config
	config := DefaultConfig()
	config.EnableFastArithmetic = true

	// No error expected from Configure
	Configure(config)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config == nil {
		t.Error("DefaultConfig() returned nil")
	}

	// Verify default values
	if config != nil {
		if config.EnableFastArithmetic {
			t.Error("Expected EnableFastArithmetic to be false by default")
		}
		if config.EnableFastConversion {
			t.Error("Expected EnableFastConversion to be false by default")
		}
		if config.DefaultMode != ModeDefault {
			t.Errorf("Expected DefaultMode to be ModeDefault, got %v", config.DefaultMode)
		}
		if config.ArithmeticMode != ArithmeticAuto {
			t.Errorf("Expected ArithmeticMode to be ArithmeticAuto, got %v", config.ArithmeticMode)
		}
	} else {
		t.Error("DefaultConfig() returned nil")
	}
}

func TestGetMemoryUsage(t *testing.T) {
	// Save the current configuration to restore it later
	origConfig := DefaultConfig()

	// Restore the original configuration after the test
	defer Configure(origConfig)

	tests := []struct {
		name           string
		config         *Config
		expectedMemory int
	}{
		{
			name: "no tables enabled",
			config: &Config{
				EnableFastArithmetic: false,
				EnableFastConversion: false,
			},
			expectedMemory: 0,
		},
		{
			name: "only conversion table enabled",
			config: &Config{
				EnableFastArithmetic: false,
				EnableFastConversion: true,
			},
			expectedMemory: 256 * 4, // 256 float32 values (4 bytes each)
		},
		{
			name: "only arithmetic tables enabled",
			config: &Config{
				EnableFastArithmetic: true,
				EnableFastConversion: false,
			},
			expectedMemory: 65536 * 4, // 4 tables (add, sub, mul, div) with 65536 uint8 values each (1 byte each)
		},
		{
			name: "all tables enabled",
			config: &Config{
				EnableFastArithmetic: true,
				EnableFastConversion: true,
			},
			expectedMemory: (256 * 4) + (65536 * 4), // conversion table + 4 arithmetic tables
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply the test configuration
			Configure(tt.config)

			// Get the memory usage
			memUsage := GetMemoryUsage()

			// Verify the memory usage matches expectations
			if memUsage != tt.expectedMemory {
				t.Errorf("GetMemoryUsage() = %d, want %d", memUsage, tt.expectedMemory)
			}
		})
	}
}

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	if version == "" {
		t.Error("Empty version string")
	}

	// Version should be in format X.Y.Z
	if version != "2.0.0" {
		t.Errorf("Unexpected version: %s", version)
	}
}

func TestZero(t *testing.T) {
	zero := Zero()
	if zero != PositiveZero {
		t.Errorf("Zero() = 0x%02X, expected 0x00", zero)
	}
}

func TestFromInt(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected Float8
	}{
		{"Zero", 0, 0x00},
		{"One", 1, 0x38},
		{"NegativeOne", -1, 0xB8},
		{"Two", 2, 0x40},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FromInt(tt.input)
			if result != tt.expected {
				t.Errorf("FromInt(%d) = 0x%02X, want 0x%02X", tt.input, result, tt.expected)
			}
		})
	}
}

func TestToInt(t *testing.T) {
	tests := []struct {
		name     string
		input    Float8
		expected int
	}{
		{"Zero", 0x00, 0},
		{"One", 0x38, 1},
		{"NegativeOne", 0xB8, -1},
		{"Two", 0x40, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.ToInt()
			if result != tt.expected {
				t.Errorf("%v.ToInt() = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFloat64Conversions(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected Float8
	}{
		{"Zero", 0.0, 0x00},
		{"One", 1.0, 0x38},
		{"NegativeOne", -1.0, 0xB8},
		{"Half", 0.5, 0x30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test FromFloat64
			result := FromFloat64(tt.value)
			if result != tt.expected {
				t.Errorf("FromFloat64(%v) = 0x%02X, want 0x%02X", tt.value, result, tt.expected)
			}

			// Test ToFloat64
			f64 := result.ToFloat64()
			// Allow small floating point imprecision
			if tt.name != "Zero" && math.Abs(f64-tt.value) > 0.02 {
				t.Errorf("ToFloat64() = %v, want %v", f64, tt.value)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name     string
		value    Float8
		expected bool
	}{
		{"Zero", 0x00, true},
		{"One", 0x38, true},
		{"NegativeOne", 0xB8, true},
		{"Infinity", 0x7C, true},
		{"NegativeInfinity", 0xFC, true},
		{"NaN", 0x7F, true}, // In this implementation, 0x7F is a valid number
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.value.IsValid()
			if result != tt.expected {
				t.Errorf("IsValid(0x%02X) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestIsNormal(t *testing.T) {
	tests := []struct {
		name     string
		value    Float8
		expected bool
	}{
		{"Zero", 0x00, false},
		{"One", 0x38, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.value.IsNormal()
			if result != tt.expected {
				t.Errorf("IsNormal(0x%02X) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestDebugInfo(t *testing.T) {
	// This is a basic test since we can't predict the exact output
	info := DebugInfo()
	if info == nil {
		t.Error("DebugInfo() returned nil")
	}

	// Check for version key which should always be present
	if _, exists := info["version"]; !exists {
		t.Error("DebugInfo() missing version key")
	}
}
