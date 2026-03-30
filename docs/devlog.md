# Float8 Development Log

## 2026-03-29 -- Baseline Benchmarks

Recorded on Apple M4 (darwin/arm64), Go 1.25, `-benchmem -count=3`.

### Conversion

| Benchmark | ns/op | B/op | allocs/op |
|-----------|------:|-----:|----------:|
| FromFloat32/Normal | 2.50 | 0 | 0 |
| FromFloat32/Subnormal | 2.53 | 0 | 0 |
| FromFloat32/Zero | 0.98 | 0 | 0 |
| FromFloat32/Large | 2.54 | 0 | 0 |
| ToFloat32/Algorithmic | 1.36 | 0 | 0 |
| ToFloat32/Lookup | 0.38 | 0 | 0 |

### Arithmetic (Algorithmic vs Lookup)

| Benchmark | Algorithmic ns/op | Lookup ns/op | Speedup |
|-----------|------------------:|-------------:|--------:|
| Add | 7.08 | 0.99 | 7.2x |
| Sub | 6.91 | 0.99 | 7.0x |
| Mul | 7.27 | 1.00 | 7.3x |
| Div | 7.85 | 1.00 | 7.9x |

### Batch Operations (1000 elements)

| Benchmark | ns/op | B/op | allocs/op |
|-----------|------:|-----:|----------:|
| ToSlice8 | 3351 | 1024 | 1 |
| ToSlice32 | 1592 | 4096 | 1 |

All operations are zero-allocation for scalar paths. Lookup tables provide
a consistent ~7x speedup over algorithmic arithmetic at the cost of 256 KB
of memory (4 tables x 64K entries x 1 byte).
