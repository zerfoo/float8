# Contributing to float8

Thank you for your interest in contributing to the float8 library! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Contributing Guidelines](#contributing-guidelines)
- [Testing](#testing)
- [Code Style](#code-style)
- [Submitting Changes](#submitting-changes)
- [Issue Reporting](#issue-reporting)
- [Performance Considerations](#performance-considerations)

## Code of Conduct

This project adheres to a code of conduct that promotes a welcoming and inclusive environment. Please be respectful and professional in all interactions.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/yourusername/float8.git
   cd float8
   ```
3. **Add the upstream remote**:
   ```bash
   git remote add upstream https://github.com/zerfoo/float8.git
   ```

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git

### Local Development

1. **Install dependencies** (if any):
   ```bash
   go mod download
   ```

2. **Run tests** to ensure everything works:
   ```bash
   go test ./...
   ```

3. **Run benchmarks**:
   ```bash
   go test -bench=. -benchmem ./...
   ```

## Contributing Guidelines

### Types of Contributions

We welcome several types of contributions:

- **Bug fixes**: Fix issues in existing functionality
- **Performance improvements**: Optimize existing algorithms
- **New features**: Add new mathematical functions or operations
- **Documentation**: Improve README, comments, or examples
- **Tests**: Add test cases for better coverage
- **Benchmarks**: Add performance benchmarks

### Before You Start

1. **Check existing issues** to see if your contribution is already being worked on
2. **Open an issue** to discuss major changes before implementing them
3. **Keep changes focused** - one feature or fix per pull request

### Implementation Guidelines

#### Core Principles

1. **Correctness**: All operations must be mathematically correct according to IEEE 754 FP8 E4M3FN specification
2. **Performance**: Consider both speed and memory usage
3. **Compatibility**: Maintain backward compatibility unless absolutely necessary
4. **Simplicity**: Prefer clear, readable code over clever optimizations

#### Adding New Functions

When adding new mathematical functions:

1. **Follow the pattern** established by existing functions
2. **Handle edge cases** (NaN, zero, overflow/underflow)
3. **Add comprehensive tests** including edge cases
4. **Include benchmarks** for performance-critical functions
5. **Document the function** with clear comments

Example function structure:
```go
// NewFunction performs a specific mathematical operation on Float8 values.
// It handles special cases like NaN and zero appropriately.
func (f Float8) NewFunction() Float8 {
    // Handle special cases first
    if f.IsNaN() {
        return NaN()
    }
    
    // Main implementation
    // ...
    
    return result
}
```

## Testing

### Test Requirements

- **All new code must have tests**
- **Aim for high test coverage** (>90%)
- **Include edge cases** and boundary conditions
- **Test both normal and error paths**

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test files
go test -v ./arithmetic_test.go

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Categories

1. **Unit tests**: Test individual functions
2. **Property tests**: Test mathematical properties (commutativity, associativity, etc.)
3. **Edge case tests**: Test boundary conditions and special values
4. **Performance tests**: Benchmark critical operations

### Writing Tests

Follow this pattern for test functions:

```go
func TestNewFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    Float8
        expected Float8
    }{
        {"normal case", FromFloat32(2.0), FromFloat32(4.0)},
        {"zero", Zero(), Zero()},
        {"NaN", NaN(), NaN()},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := tt.input.NewFunction()
            if !result.Equal(tt.expected) {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## Code Style

### Go Standards

- Follow standard Go formatting (`go fmt`)
- Use `go vet` to check for common issues
- Follow Go naming conventions
- Write clear, descriptive variable names

### Documentation

- **Public functions** must have godoc comments
- **Complex algorithms** should have implementation comments
- **Edge cases** should be documented
- **Performance characteristics** should be noted where relevant

### Example Documentation:

```go
// Add performs floating-point addition of two Float8 values.
// It handles special cases including NaN propagation and zero addition.
// The result follows IEEE 754 rounding rules for the E4M3FN format.
//
// Special cases:
//   - Add(NaN, x) = NaN for any x
//   - Add(x, NaN) = NaN for any x
//   - Add(+0, -0) = +0
//
// Performance: O(1) with fast arithmetic enabled, otherwise involves
// conversion to float32 and back.
func (f Float8) Add(other Float8) Float8 {
    // Implementation...
}
```

## Submitting Changes

### Pull Request Process

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following the guidelines above

3. **Test thoroughly**:
   ```bash
   go test ./...
   go test -race ./...
   ```

4. **Update documentation** if needed

5. **Commit with clear messages**:
   ```bash
   git commit -m "Add NewFunction for mathematical operation X
   
   - Implements IEEE 754 compliant operation
   - Handles NaN and zero edge cases
   - Includes comprehensive tests and benchmarks"
   ```

6. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

7. **Open a pull request** with:
   - Clear description of changes
   - Reference to related issues
   - Test results and benchmark comparisons

### Pull Request Guidelines

- **One feature per PR** - keep changes focused
- **Include tests** for all new functionality
- **Update documentation** as needed
- **Maintain backward compatibility** unless discussed
- **Follow the existing code style**

## Issue Reporting

### Bug Reports

When reporting bugs, please include:

- **Go version** and operating system
- **Minimal reproduction case**
- **Expected vs actual behavior**
- **Stack trace** if applicable

### Feature Requests

For new features, please provide:

- **Use case description**
- **Proposed API design**
- **Performance considerations**
- **Compatibility impact**

## Performance Considerations

### Optimization Guidelines

1. **Measure first** - use benchmarks to identify bottlenecks
2. **Consider memory usage** - Float8 is designed for efficiency
3. **Lookup tables** - balance speed vs memory for critical paths
4. **Avoid allocations** in hot paths
5. **Profile regularly** - use `go tool pprof` for analysis

### Benchmarking

Always include benchmarks for performance-critical changes:

```go
func BenchmarkNewFunction(b *testing.B) {
    f := FromFloat32(3.14)
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _ = f.NewFunction()
    }
}
```

## Questions?

If you have questions about contributing, please:

1. Check existing issues and documentation
2. Open a new issue for discussion
3. Reach out to maintainers

Thank you for contributing to float8!
