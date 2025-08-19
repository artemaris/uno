package utils

import "testing"

func BenchmarkGenerateShortID(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = GenerateShortID()
		}
	})
}
