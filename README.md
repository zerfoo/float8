# float8

[![Go Reference](https://pkg.go.dev/badge/github.com/zerfoo/float8.svg)](https://pkg.go.dev/github.com/zerfoo/float8)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

FP8 E4M3FN arithmetic library for Go, commonly used in quantized ML inference.

Part of the [Zerfoo](https://github.com/zerfoo) ML ecosystem.

## Features

- **IEEE 754 FP8 E4M3FN format** — 1 sign, 4 exponent, 3 mantissa bits
- **Fast lookup tables** — optional pre-computed tables for arithmetic and conversion
- **Full arithmetic** — add, subtract, multiply, divide, sqrt, abs, neg
- **No infinities** — the E4M3FN variant uses the infinity encoding for additional finite values
- **Zero dependencies** — pure Go, no CGo

## Installation

```bash
go get github.com/zerfoo/float8
```

Requires Go 1.26+.

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/zerfoo/float8"
)

func main() {
    a := float8.FromFloat32(3.14)
    b := float8.FromFloat32(2.71)

    sum := a.Add(b)
    product := a.Mul(b)

    fmt.Printf("a = %f\n", a.ToFloat32())
    fmt.Printf("a + b = %f\n", sum.ToFloat32())
    fmt.Printf("a * b = %f\n", product.ToFloat32())
}
```

## Format

| Field | Bits | Description |
|-------|------|-------------|
| Sign | 1 | 0 = positive, 1 = negative |
| Exponent | 4 | Biased by 7, range [-6, 7] |
| Mantissa | 3 | 3 explicit + 1 implicit leading bit |

Special values: ±0 (exp=0, mant=0), NaN (exp=1111, mant=111). No infinities.

## Performance Modes

```go
// Enable lookup tables for faster arithmetic (trades memory for speed)
float8.EnableFastArithmetic()
float8.EnableFastConversion()
```

## Used By

- [ztensor](https://github.com/zerfoo/ztensor) — GPU-accelerated tensor library

## License

Apache 2.0
