// fast_csv_rune_encoder.go
//
// This file defines small, allocation-free helpers for efficiently appending
// known CSV structural runes (quote, delimiter, etc.) into an output buffer.
//
// A runeEncoder pre-encodes a single rune's UTF-8 bytes.
// A twoRuneEncoder pre-encodes two runes' UTF-8 bytes back-to-back.
//
// appendText(dst) uses an unrolled switch on the byte length (1..4 or 2..8)
// instead of append(dst, buf[:n]...) so the compiler can:
//   - elide bounds checks
//   - perform at most one capacity grow
//   - inline straight-line byte copies
//
// These helpers are on the Writer hot path.
// They intentionally trade readability for branchless speed.

package csv

import (
	"unicode/utf8"
)

type runeEncoder struct {
	n uint8
	b [1 * utf8.UTFMax]byte
}

func (re *runeEncoder) appendText(p []byte) []byte {
	// the switch helps remove bounds checks and adds an
	// implicit single grow instruction
	//
	// so speed increases at the cost of readability
	//
	// it is end-state equivalent to:
	//
	// return append(p, re.b[:re.n]...)

	switch re.n {
	case 1:
		return append(p, re.b[0])
	case 2:
		return append(p, re.b[0], re.b[1])
	case 3:
		return append(p, re.b[0], re.b[1], re.b[2])
	case 4:
		return append(p, re.b[0], re.b[1], re.b[2], re.b[3])
	default:
		panic(panicInvalidRuneEncoderLen)
	}
}

func newRuneEncoder(r rune) runeEncoder {
	var buf [1 * utf8.UTFMax]byte

	n := uint8(utf8.EncodeRune(buf[:], r))

	return runeEncoder{n, buf}
}

type twoRuneEncoder struct {
	n uint8
	b [2 * utf8.UTFMax]byte
}

func (re *twoRuneEncoder) appendText(p []byte) []byte {
	// the switch helps remove bounds checks and adds an
	// implicit single grow instruction
	//
	// so speed increases at the cost of readability
	//
	// it is end-state equivalent to:
	//
	// return append(p, re.b[:re.n]...)

	switch re.n {
	case 2:
		return append(p, re.b[0], re.b[1])
	case 3:
		return append(p, re.b[0], re.b[1], re.b[2])
	case 4:
		return append(p, re.b[0], re.b[1], re.b[2], re.b[3])
	case 5:
		return append(p, re.b[0], re.b[1], re.b[2], re.b[3], re.b[4])
	case 6:
		return append(p, re.b[0], re.b[1], re.b[2], re.b[3], re.b[4], re.b[5])
	case 7:
		return append(p, re.b[0], re.b[1], re.b[2], re.b[3], re.b[4], re.b[5], re.b[6])
	case 8:
		return append(p, re.b[0], re.b[1], re.b[2], re.b[3], re.b[4], re.b[5], re.b[6], re.b[7])
	default:
		panic(panicInvalidRuneEncoderLen)
	}
}

func newTwoRuneEncoder(r1, r2 rune) twoRuneEncoder {
	var buf [2 * utf8.UTFMax]byte

	n := uint8(utf8.EncodeRune(buf[:], r1))
	n += uint8(utf8.EncodeRune(buf[n:], r2))

	return twoRuneEncoder{n, buf}
}
