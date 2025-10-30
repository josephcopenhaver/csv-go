package csv

import (
	"testing"
)

func Benchmark_runeScape4_containsWideRune(b *testing.B) {
	const (
		// cent sign
		test2ByteRune = '\u00A2'
		// euro sign
		test3ByteRune = '\u20AC'
		// grinning face
		test4ByteRune = '\U0001F600'
	)

	b.ReportAllocs()

	var rs runeScape4
	rs.addRune(test2ByteRune)
	rs.addRune(test3ByteRune)
	rs.addRune(test4ByteRune)
	rs.addRune(test4ByteRune + 1)

	searchRune := test4ByteRune + 2

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rs.containsWideRune(searchRune)
	}
}

func Benchmark_runeScape6_containsWideRune(b *testing.B) {
	const (
		// cent sign
		test2ByteRune = '\u00A2'
		// euro sign
		test3ByteRune = '\u20AC'
		// grinning face
		test4ByteRune = '\U0001F600'
	)

	b.ReportAllocs()

	var rs runeScape6
	rs.addRune(test2ByteRune)
	rs.addRune(test3ByteRune)
	rs.addRune(test4ByteRune)
	rs.addRune(test4ByteRune + 1)
	rs.addRune(test4ByteRune + 2)
	rs.addRune(test4ByteRune + 3)

	searchRune := test4ByteRune + 4

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rs.containsWideRune(searchRune)
	}
}
