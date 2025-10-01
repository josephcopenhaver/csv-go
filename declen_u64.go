package csv

import (
	"math/bits"
)

// sources:
// - https://lemire.me/blog/2021/05/28/computing-the-number-of-digits-of-an-integer-quickly/
// - https://theswissbay.ch/pdf/Gentoomen%20Library/Security/Addison%20Wesley%20-%20Hackers%20Delight%202002.pdf

var pow10u64 = [...]uint64{
	0x0000000000000001, // 1
	0x000000000000000A, // 10
	0x0000000000000064, // 100
	0x00000000000003E8, // 1_000
	0x0000000000002710, // 10_000
	0x00000000000186A0, // 100_000
	0x00000000000F4240, // 1_000_000
	0x0000000000989680, // 10_000_000
	0x0000000005F5E100, // 100_000_000
	0x000000003B9ACA00, // 1_000_000_000
	0x00000002540BE400, // 10_000_000_000
	0x000000174876E800, // 100_000_000_000
	0x000000E8D4A51000, // 1_000_000_000_000
	0x000009184E72A000, // 10_000_000_000_000
	0x00005AF3107A4000, // 100_000_000_000_000
	0x00038D7EA4C68000, // 1_000_000_000_000_000
	0x002386F26FC10000, // 10_000_000_000_000_000
	0x016345785D8A0000, // 100_000_000_000_000_000
	0x0DE0B6B3A7640000, // 1_000_000_000_000_000_000
	0x8AC7230489E80000, // 10_000_000_000_000_000_000
}

func decLenU64(n uint64) int {
	if n == 0 {
		return 1
	}
	// Approximate floor(log10(n)) + 1 using bit-length, then correct.
	// 1233/4096 ≈ log10(2)
	d := ((bits.Len64(n) * 1233) >> 12) + 1 // in [1..20], at most off by 1
	if n < pow10u64[d-1] {
		d--
	}
	return d
}
