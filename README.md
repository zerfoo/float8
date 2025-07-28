# float8

[![Go Reference](https://pkg.go.dev/badge/github.com/zerfoo/float16.svg)](https://pkg.go.dev/github.com/zerfoo/float8)
[![Go Report Card](https://goreportcard.com/badge/github.com/zerfoo/float16)](https://goreportcard.com/report/github.com/zerfoo/float8)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

A high-performance Go library implementing IEEE 754 FP8 E4M3FN format for 8-bit floating-point arithmetic, commonly used in machine learning applications for reduced-precision computations.

## Features

- **IEEE 754 FP8 E4M3FN Format**: Complete implementation of the 8-bit floating-point format
- **High Performance**: Optimized arithmetic operations with optional fast lookup tables
- **Comprehensive API**: Full support for conversion, arithmetic, and mathematical operations
- **Machine Learning Ready**: Designed for ML workloads requiring reduced precision
- **Zero Dependencies**: Pure Go implementation with no external dependencies

## Format Specification

The Float8 type uses the E4M3FN variant of IEEE 754 FP8:

- **1 bit**: Sign (0 = positive, 1 = negative)
- **4 bits**: Exponent (biased by 7, range [-6, 7])
- **3 bits**: Mantissa (3 explicit bits, 1 implicit leading bit for normal numbers)

### Special Values

- **Zero**: Exponent=0000, Mantissa=000 (both positive and negative)
- **NaN**: Exponent=1111, Mantissa=111
- **No Infinities**: The E4M3FN variant does not support infinity values

## Installation

```bash
go get github.com/zerfoo/float8
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/zerfoo/float8"
)

func main() {
    // Initialize the package (optional, done automatically)
    float8.Initialize()
    
    // Create Float8 values from float32
    a := float8.FromFloat32(3.14)
    b := float8.FromFloat32(2.71)
    
    // Perform arithmetic operations
    sum := a.Add(b)
    product := a.Mul(b)
    
    // Convert back to float32
    fmt.Printf("a = %f\n", a.ToFloat32())
    fmt.Printf("b = %f\n", b.ToFloat32())
    fmt.Printf("a + b = %f\n", sum.ToFloat32())
    fmt.Printf("a * b = %f\n", product.ToFloat32())
}
```

## Configuration

The library supports various configuration options for performance optimization:

```go
// Configure with custom settings
config := &float8.Config{
    EnableFastArithmetic: true,  // Enable lookup tables for faster arithmetic
    EnableFastConversion: true,  // Enable lookup tables for faster conversion
    DefaultMode:          float8.ModeDefault,
    ArithmeticMode:       float8.ArithmeticAuto,
}

float8.Configure(config)
```

## API Reference

### Core Types

- `Float8`: The main 8-bit floating-point type
- `Config`: Configuration options for the package

### Conversion Functions

```go
// From other numeric types
func FromFloat32(f float32) Float8
func FromFloat64(f float64) Float8
func FromInt(i int) Float8

// To other numeric types
func (f Float8) ToFloat32() float32
func (f Float8) ToFloat64() float64
func (f Float8) ToInt() int
```

### Arithmetic Operations

```go
func (f Float8) Add(other Float8) Float8
func (f Float8) Sub(other Float8) Float8
func (f Float8) Mul(other Float8) Float8
func (f Float8) Div(other Float8) Float8
```

### Mathematical Functions

```go
func (f Float8) Abs() Float8
func (f Float8) Neg() Float8
func (f Float8) Sqrt() Float8
// ... and more
```

### Utility Functions

```go
func (f Float8) IsZero() bool
func (f Float8) IsNaN() bool
func (f Float8) IsInf() bool
func (f Float8) String() string
```

## Performance

The library offers two performance modes:

1. **Standard Mode**: Compact implementation with minimal memory usage
2. **Fast Mode**: Uses pre-computed lookup tables for faster operations at the cost of memory

Enable fast mode for performance-critical applications:

```go
float8.EnableFastArithmetic()
float8.EnableFastConversion()
```

## Testing

Run the comprehensive test suite:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Benchmarks

Run performance benchmarks:

```bash
go test -bench=. -benchmem ./...
```

## Use Cases

- **Machine Learning**: Reduced precision training and inference
- **Neural Networks**: Memory-efficient model parameters
- **Scientific Computing**: Applications requiring controlled precision
- **Embedded Systems**: Resource-constrained environments

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- IEEE 754 standard for floating-point arithmetic
- The machine learning community for driving FP8 adoption
- Contributors and maintainers of this project
