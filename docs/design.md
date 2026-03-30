# float8 Design Document

This document describes the design of `github.com/zerfoo/float8`, a pure-Go implementation of the IEEE 754 FP8 E4M3FN format for quantized ML inference.

## 1. FP8 E4M3FN Bit Layout

The E4M3FN format packs a floating-point number into a single byte:

```
  Bit 7     Bits 6-3      Bits 2-0
 ┌──────┬──────────────┬────────────┐
 │ Sign │   Exponent   │  Mantissa  │
 │ (1)  │     (4)      │    (3)     │
 └──────┴──────────────┴────────────┘
```

| Field | Width | Mask | Description |
|-------|-------|------|-------------|
| Sign | 1 bit | `0x80` | 0 = positive, 1 = negative |
| Exponent | 4 bits | `0x78` | Biased unsigned integer (bias = 7) |
| Mantissa | 3 bits | `0x07` | Explicit significand bits; normal numbers have an implicit leading 1 |

The exponent bias is 7 (`2^(4-1) - 1`), giving an unbiased exponent range of [-6, +8] for stored values 1-15. Stored exponent 0 indicates a subnormal (no implicit leading 1).

**Representable range:** The largest finite value (bit pattern `0x7E` = `0.1111.110`) is 448.0. The smallest positive normal is `0x08` (1.0 x 2^-6 = 0.015625). The smallest positive subnormal is `0x01` (0.001 x 2^-6 = 0.001953125).

**Precision:** With 3 explicit mantissa bits (4 effective bits for normals), relative precision is roughly 2^-3 = 12.5%. This is adequate for storing quantized weights and activations but not for accumulation; ML frameworks accumulate in float16 or float32.

## 2. Lookup Table Strategy

Because the format has only 256 possible bit patterns, exhaustive precomputation is practical:

### Conversion Table

A single 256-entry `[]float32` table maps every `Float8` bit pattern to its exact float32 equivalent. Indexed by `uint8(f)`, a lookup replaces the branch-heavy algorithmic decode path with a single array access. Memory cost: 256 x 4 = **1 KiB**.

### Arithmetic Tables

Each binary operation (add, subtract, multiply, divide) uses a 65,536-entry `[]Float8` table indexed by `uint16(a)<<8 | uint16(b)`. Every (a, b) pair is precomputed once from the algorithmic implementation. Memory cost per table: 65,536 x 1 = **64 KiB** (256 KiB total for all four operations).

### Lazy Initialization

Tables are not allocated at package init. Callers opt in via `EnableFastConversion()` and `EnableFastArithmetic()`, which populate the tables on first call. This keeps the default memory footprint at zero for programs that only need occasional FP8 conversions. Tables can be released with the corresponding `Disable` functions.

### Mode Selection

Three arithmetic modes control dispatch:

| Mode | Behavior |
|------|----------|
| `ArithmeticAuto` (default) | Use table if loaded, otherwise algorithmic |
| `ArithmeticLookup` | Force table path (panics if tables not loaded) |
| `ArithmeticAlgorithmic` | Force algorithmic path regardless of table state |

## 3. Arithmetic Operations

All arithmetic follows a **convert-up, compute, convert-down** pattern:

1. Convert both `Float8` operands to `float32` (exact, since FP8 is a subset of float32).
2. Perform the operation in float32 precision.
3. Convert the float32 result back to `Float8` with round-to-nearest-even.

This strategy inherits float32 IEEE 754 semantics and avoids implementing carry propagation, alignment shifting, or normalization in 8-bit arithmetic.

**Operations provided:** `Add`, `Sub`, `Mul`, `Div`, `Sqrt`, `Pow`, `Exp`, `Log`, `Sin`, `Cos`, `Tan`, `Floor`, `Ceil`, `Round`, `Trunc`, `Fmod`, `Abs`, `Neg`, `Min`, `Max`, `Clamp`, `Lerp`, `CopySign`.

**Comparison operations:** `Equal`, `Less`, `Greater`, `LessEqual`, `GreaterEqual` handle NaN (unordered), signed zeros (+0 == -0), and infinities per IEEE 754 rules.

**Batch operations:** `AddSlice`, `MulSlice`, `ScaleSlice`, `SumSlice` operate element-wise on `[]Float8` slices. `ToSlice8` and `ToSlice32` handle bulk conversion between `[]float32` and `[]Float8`.

## 4. Conversion To/From float32

### float32 to Float8 (`ToFloat8`)

1. Handle special cases first: signed zeros, infinities, NaN.
2. Extract sign, exponent, and mantissa from the float32 IEEE 754 bits.
3. Re-bias the exponent: `exp8 = exp32 - 127 + 7`.
4. Check for overflow (exp8 > 15 -> clamp to infinity) and underflow (exp8 < -7 -> clamp to zero).
5. Truncate the 23-bit mantissa to 3 bits, applying round-to-nearest-even: if the 4th bit is set, round up. Handle mantissa carry into the exponent.
6. Pack sign (1 bit), exponent (4 bits), and mantissa (3 bits) into a `uint8`.

Three conversion modes control edge-case behavior:

| Mode | Overflow | Underflow | NaN |
|------|----------|-----------|-----|
| `ModeDefault` | Saturate to infinity | Saturate to zero | Convert to `0x7F` |
| `ModeStrict` | Return error | Return error | Return error |
| `ModeFast` | Use lookup table | Use lookup table | Use lookup table |

### Float8 to float32 (`ToFloat32`)

The conversion is always exact (no rounding) because every FP8 value is representable in float32. The algorithmic path extracts sign, exponent, and mantissa, re-biases the exponent (`exp32 = exp8 - 7 + 127`), shifts the 3-bit mantissa to float32 position (left-shift by 20), and assembles the 32-bit IEEE 754 pattern. With the conversion table enabled, this reduces to a single array lookup.

## 5. No-Infinities Design Rationale

The E4M3FN format (the "FN" stands for "Finite, NaN") intentionally eliminates infinity encodings to maximize the finite representable range:

- In standard IEEE 754, the all-ones exponent (`1111`) with a zero mantissa encodes infinity. E4M3FN repurposes this encoding as a normal finite value, extending the maximum magnitude from 240 to **448**.
- Only the all-ones exponent with all-ones mantissa (`0x7F`, `0xFF`) is reserved for NaN. This gives exactly two NaN encodings (positive and negative) instead of the usual 14 quiet/signaling NaN patterns.
- ML inference rarely produces or consumes infinities. Overflows during quantized GEMM/GEMV saturate to the maximum representable value rather than propagating infinity, which is more numerically stable for downstream operations like softmax and layer normalization.

**Note:** The current implementation defines `PositiveInfinity` and `NegativeInfinity` constants for API compatibility with IEEE 754 conventions (e.g., overflow from float32 conversion maps to these bit patterns), but in E4M3FN semantics these are finite values equal to +/-448.

## 6. Use in ML Inference

FP8 E4M3FN is the standard quantization format for weights and activations in transformer inference:

### Quantized Storage

Model weights stored in GGUF files use FP8 to reduce memory bandwidth by 4x compared to float32. The `ToSlice8`/`ToSlice32` batch conversion functions support bulk quantization and dequantization with negative-zero preservation.

### GEMM/GEMV Kernels

In the Zerfoo ecosystem, `ztensor` imports `float8` to implement quantized matrix-multiply kernels. Weights are stored as `[]Float8` and dequantized to float16 or float32 in register before the fused multiply-accumulate. Accumulation always occurs in higher precision to avoid catastrophic rounding error.

### Where FP8 Fits in the Precision Hierarchy

| Type | Bits | Use Case |
|------|------|----------|
| float32 | 32 | Accumulation, loss computation, optimizer state |
| float16 / bfloat16 | 16 | Activations, KV cache, intermediate results |
| **float8 (E4M3FN)** | **8** | **Weight storage, activation quantization** |
| int4 (Q4_K_M) | 4 | Aggressive weight quantization (GGUF) |

### Scope

This library covers E4M3FN only. The E5M2 variant (5 exponent bits, 2 mantissa bits, with infinities) is planned for a future release (see T46.4.8) and targets gradient storage in mixed-precision training, where the wider dynamic range matters more than precision.
