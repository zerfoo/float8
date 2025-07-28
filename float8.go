package float8

import (
	"sync"
)

// Package initialization and configuration

var (
	initOnce sync.Once
)

// Initialize performs one-time package initialization
func Initialize() {
	initOnce.Do(func() {
		// Package is ready to use immediately
		// Lookup tables are loaded lazily when needed
	})
}

// Config holds package configuration options
type Config struct {
	EnableFastArithmetic bool
	EnableFastConversion bool
	DefaultMode          ConversionMode
	ArithmeticMode       ArithmeticMode
}

// DefaultConfig returns the default package configuration
func DefaultConfig() *Config {
	return &Config{
		EnableFastArithmetic: false, // Disabled by default to save memory
		EnableFastConversion: false, // Disabled by default to save memory
		DefaultMode:          ModeDefault,
		ArithmeticMode:       ArithmeticAuto,
	}
}

// Configure applies the given configuration to the package
func Configure(config *Config) {
	if config.EnableFastArithmetic {
		EnableFastArithmetic()
	} else {
		DisableFastArithmetic()
	}

	if config.EnableFastConversion {
		EnableFastConversion()
	} else {
		DisableFastConversion()
	}

	DefaultConversionMode = config.DefaultMode
	DefaultArithmeticMode = config.ArithmeticMode
}

// GetMemoryUsage returns the current memory usage of lookup tables in bytes
func GetMemoryUsage() int {
	var usage int

	if conversionTable != nil {
		usage += 256 * 4 // 256 float32 values
	}

	if addTable != nil {
		usage += 65536 // 65536 uint8 values
	}
	if subTable != nil {
		usage += 65536
	}
	if mulTable != nil {
		usage += 65536
	}
	if divTable != nil {
		usage += 65536
	}

	return usage
}

// Version information
const (
	Version      = "2.0.0"
	VersionMajor = 2
	VersionMinor = 0
	VersionPatch = 0
)

// GetVersion returns the package version string
func GetVersion() string {
	return Version
}

// Convenience functions for common operations

// Zero returns a Float8 zero value
func Zero() Float8 {
	return PositiveZero
}

// One returns a Float8 value representing 1.0
func One() Float8 {
	return ToFloat8(1.0)
}

// FromInt converts an integer to Float8
func FromInt(i int) Float8 {
	return ToFloat8(float32(i))
}

// FromFloat64 converts a float64 to Float8 (with potential precision loss)
func FromFloat64(f float64) Float8 {
	return ToFloat8(float32(f))
}

// ToFloat64 converts a Float8 to float64
func (f Float8) ToFloat64() float64 {
	return float64(f.ToFloat32())
}

// ToInt converts a Float8 to int (truncated)
func (f Float8) ToInt() int {
	return int(f.ToFloat32())
}

// Validation functions

// IsValid returns true if the Float8 represents a valid number
func (f Float8) IsValid() bool {
	// All bit patterns are valid in Float8
	return true
}

// IsNormal returns true if the Float8 is a normal (non-zero, non-infinite) number
func (f Float8) IsNormal() bool {
	return !f.IsZero() && f.IsFinite()
}

// Package information for debugging

// DebugInfo returns debugging information about the package state
func DebugInfo() map[string]interface{} {
	return map[string]interface{}{
		"version":            Version,
		"memory_usage_bytes": GetMemoryUsage(),
		"fast_arithmetic":    addTable != nil,
		"fast_conversion":    conversionTable != nil,
		"default_conv_mode":  DefaultConversionMode,
		"default_arith_mode": DefaultArithmeticMode,
	}
}
