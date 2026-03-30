# ADR-001: API Stability Contract for float8 v1.0.0

- **Status:** Accepted
- **Date:** 2026-03-29
- **Authors:** Daniel Ndungu

## Context

The `github.com/zerfoo/float8` package provides IEEE 754 FP8 E4M3FN arithmetic for the Zerfoo ML ecosystem. It is imported by `ztensor` for quantized tensor storage and compute. The package has reached a stable API surface and needs a clear stability contract so downstream consumers can depend on it without fear of breakage.

## Decision

### Stable (v1 guarantee)

The following API surface is covered by Go module compatibility and will not have breaking changes within the v1.x line:

**Core type:**
- `Float8` (defined as `uint8`)

**Constructors and conversions:**
- `ToFloat8(float32) Float8`
- `ToFloat8WithMode(float32, ConversionMode) (Float8, error)`
- `FromFloat64(float64) Float8`
- `FromInt(int) Float8`
- `FromBits(uint8) Float8`
- `Parse(string) (Float8, error)`
- `Zero() Float8`, `One() Float8`

**Methods on Float8:**
- `ToFloat32() float32`, `ToFloat64() float64`, `ToInt() int`
- `Bits() uint8`
- `Abs() Float8`, `Neg() Float8`
- `Sign() int`
- `IsZero() bool`, `IsNaN() bool`, `IsInf() bool`, `IsFinite() bool`, `IsNormal() bool`, `IsValid() bool`
- `String() string`, `GoString() string`

**Arithmetic functions:**
- `Add`, `Sub`, `Mul`, `Div` (and `*WithMode` variants)
- `AddSlice`, `MulSlice`, `ScaleSlice`, `SumSlice`

**Math functions:**
- `Sqrt`, `Pow`, `Exp`, `Log`
- `Sin`, `Cos`, `Tan`
- `Floor`, `Ceil`, `Round`, `Trunc`, `Fmod`
- `Min`, `Max`, `Clamp`, `Lerp`, `Sign`, `CopySign`

**Comparison functions:**
- `Equal`, `Less`, `LessEqual`, `Greater`, `GreaterEqual`

**Batch conversions:**
- `ToSlice8([]float32) []Float8`
- `ToSlice32([]Float8) []float32`

**Configuration:**
- `Config`, `DefaultConfig`, `Configure`
- `ConversionMode` (constants: `ModeDefault`, `ModeStrict`, `ModeFast`)
- `ArithmeticMode` (constants: `ArithmeticAuto`, `ArithmeticAlgorithmic`, `ArithmeticLookup`)
- `EnableFastConversion`, `DisableFastConversion`
- `EnableFastArithmetic`, `DisableFastArithmetic`
- `DefaultConversionMode`, `DefaultArithmeticMode` (package-level variables)

**Constants:**
- Bit masks: `SignMask`, `ExponentMask`, `MantissaMask`, `MantissaLen`
- Exponent: `ExponentBias`, `ExponentMax`, `ExponentMin`, `Float32Bias`
- Special values: `PositiveZero`, `NegativeZero`, `PositiveInfinity`, `NegativeInfinity`, `NaN`, `MaxValue`, `MinValue`, `SmallestPositive`
- Math constants: `E`, `Pi`, `Phi`, `Sqrt2`, `SqrtE`, `SqrtPi`, `Ln2`, `Log2E`, `Ln10`, `Log10E`

**Error types:**
- `Float8Error` (struct with `Op`, `Value`, `Msg` fields)
- Sentinel errors: `ErrOverflow`, `ErrUnderflow`, `ErrNaN`

**Utilities:**
- `Initialize()`, `GetVersion()`, `GetMemoryUsage()`, `DebugInfo()`

### Explicitly deferred

The following are **not** part of v1 and are candidates for v1.1+:

- **FP8 E5M2 format** — A second 8-bit format with 5 exponent bits and 2 mantissa bits, used in some gradient representations. Will be added as a separate type (e.g., `Float8E5M2`) without altering the existing `Float8` (E4M3FN) type.
- **SIMD-accelerated batch operations** — Platform-specific vectorized paths for slice operations.
- **Stochastic rounding mode** — A `ConversionMode` variant that uses probabilistic rounding for training workloads.

### Versioning policy

- Patch releases (v1.0.x): bug fixes, performance improvements, documentation.
- Minor releases (v1.x.0): new functions, types, or constants that do not break existing callers.
- The `Version` constant tracks the current release and is updated by release-please.

## Consequences

- Downstream packages (`ztensor`, `zerfoo`) can pin `float8 v1.x` and upgrade freely within the major version.
- New FP8 formats (E5M2) will be additive and will not modify the `Float8` type or its semantics.
- Any behavioral change to existing functions (e.g., rounding rules, special-value handling) requires a new major version.
