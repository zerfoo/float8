package float8

import (
	"testing"
)

// BenchmarkFromFloat32 benchmarks float32 → Float8 conversion.
func BenchmarkFromFloat32(b *testing.B) {
	b.Run("Normal", func(b *testing.B) {
		f32 := float32(1.5)
		for i := 0; i < b.N; i++ {
			_ = ToFloat8(f32)
		}
	})
	b.Run("Subnormal", func(b *testing.B) {
		f32 := float32(0.001953125) // smallest normal float8 boundary
		for i := 0; i < b.N; i++ {
			_ = ToFloat8(f32)
		}
	})
	b.Run("Zero", func(b *testing.B) {
		f32 := float32(0.0)
		for i := 0; i < b.N; i++ {
			_ = ToFloat8(f32)
		}
	})
	b.Run("Large", func(b *testing.B) {
		f32 := float32(448.0) // max finite float8
		for i := 0; i < b.N; i++ {
			_ = ToFloat8(f32)
		}
	})
}

// BenchmarkToFloat32_Modes benchmarks Float8 → float32 conversion with
// algorithmic and lookup-table paths.
func BenchmarkToFloat32_Modes(b *testing.B) {
	f8 := ToFloat8(1.5)

	b.Run("Algorithmic", func(b *testing.B) {
		DisableFastConversion()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = f8.ToFloat32()
		}
	})
	b.Run("Lookup", func(b *testing.B) {
		EnableFastConversion()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = f8.ToFloat32()
		}
		b.StopTimer()
		DisableFastConversion()
	})
}

// BenchmarkAddModes benchmarks addition with algorithmic and lookup-table paths.
func BenchmarkAddModes(b *testing.B) {
	a := ToFloat8(1.5)
	c := ToFloat8(2.5)

	b.Run("Algorithmic", func(b *testing.B) {
		DisableFastArithmetic()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Add(a, c)
		}
	})
	b.Run("Lookup", func(b *testing.B) {
		EnableFastArithmetic()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Add(a, c)
		}
		b.StopTimer()
		DisableFastArithmetic()
	})
}

// BenchmarkMulModes benchmarks multiplication with algorithmic and lookup-table paths.
func BenchmarkMulModes(b *testing.B) {
	a := ToFloat8(1.5)
	c := ToFloat8(2.5)

	b.Run("Algorithmic", func(b *testing.B) {
		DisableFastArithmetic()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Mul(a, c)
		}
	})
	b.Run("Lookup", func(b *testing.B) {
		EnableFastArithmetic()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Mul(a, c)
		}
		b.StopTimer()
		DisableFastArithmetic()
	})
}

// BenchmarkSub benchmarks subtraction.
func BenchmarkSub(b *testing.B) {
	a := ToFloat8(3.5)
	c := ToFloat8(1.5)

	b.Run("Algorithmic", func(b *testing.B) {
		DisableFastArithmetic()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Sub(a, c)
		}
	})
	b.Run("Lookup", func(b *testing.B) {
		EnableFastArithmetic()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Sub(a, c)
		}
		b.StopTimer()
		DisableFastArithmetic()
	})
}

// BenchmarkDiv benchmarks division.
func BenchmarkDiv(b *testing.B) {
	a := ToFloat8(3.5)
	c := ToFloat8(1.5)

	b.Run("Algorithmic", func(b *testing.B) {
		DisableFastArithmetic()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Div(a, c)
		}
	})
	b.Run("Lookup", func(b *testing.B) {
		EnableFastArithmetic()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Div(a, c)
		}
		b.StopTimer()
		DisableFastArithmetic()
	})
}
