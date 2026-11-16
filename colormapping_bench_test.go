package colormapping_test

import (
	"strconv"
	"testing"

	"github.com/luanti-golang/colormapping"
)

func newDefaultMapping(b *testing.B) *colormapping.ColorMapping {
	m := colormapping.NewColorMapping()
	if err := m.LoadDefaults(); err != nil {
		b.Fatal(err)
	}
	return m
}

func BenchmarkGetColorHit(b *testing.B) {
	m := newDefaultMapping(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.GetColor("default:stone", 0)
	}
}

func BenchmarkGetColorHitParallel(b *testing.B) {
	m := newDefaultMapping(b)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.GetColor("default:stone", 0)
		}
	})
}

func BenchmarkGetColorFallback(b *testing.B) {
	m := newDefaultMapping(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.GetColor("mymod:my_red_node", 0)
	}
}

func BenchmarkGetColorMiss(b *testing.B) {
	m := newDefaultMapping(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.GetColor("somemod:some_unknown_long_node_name", 0)
	}
}

func BenchmarkGetColorMissParallel(b *testing.B) {
	m := newDefaultMapping(b)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.GetColor("somemod:some_unknown_long_node_name", 0)
		}
	})
}

// first lookup of distinct unknown names: the uncached worst case
func BenchmarkGetColorMissUncached(b *testing.B) {
	m := newDefaultMapping(b)
	names := make([]string, b.N)
	for i := range names {
		names[i] = "somemod:unknown_node_" + strconv.Itoa(i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.GetColor(names[i], 0)
	}
}

func BenchmarkGetColorPalette(b *testing.B) {
	m := newDefaultMapping(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.GetColor("unifiedbricks:brickblock_multicolor_dark", i&255)
	}
}

func BenchmarkNewColorMapping(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		colormapping.NewColorMapping()
	}
}

func BenchmarkLoadDefaults(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		newDefaultMapping(b)
	}
}
