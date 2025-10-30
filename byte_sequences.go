package csv

import (
	"unicode/utf8"
)

type byteSequenceShort struct {
	n uint8
	b [1 * utf8.UTFMax]byte
}

func (bs *byteSequenceShort) appendText(p []byte) []byte {
	// the switch helps remove bounds checks and adds an
	// implicit single grow instruction
	//
	// so speed increases at the cost of readability
	//
	// it is end-state equivalent to:
	//
	// return append(p, bs.b[:bs.n]...)

	switch bs.n {
	case 1:
		return append(p, bs.b[0])
	case 2:
		return append(p, bs.b[0], bs.b[1])
	case 3:
		return append(p, bs.b[0], bs.b[1], bs.b[2])
	case 4:
		return append(p, bs.b[0], bs.b[1], bs.b[2], bs.b[3])
	default:
		panic(panicInvalidByteSequenceLength)
	}
}

func newSeq(r rune) byteSequenceShort {
	var buf [1 * utf8.UTFMax]byte

	n := uint8(utf8.EncodeRune(buf[:], r))

	return byteSequenceShort{n, buf}
}

type byteSequenceLong struct {
	n uint8
	b [2 * utf8.UTFMax]byte
}

func (bs *byteSequenceLong) appendText(p []byte) []byte {
	// the switch helps remove bounds checks and adds an
	// implicit single grow instruction
	//
	// so speed increases at the cost of readability
	//
	// it is end-state equivalent to:
	//
	// return append(p, bs.b[:bs.n]...)

	switch bs.n {
	case 2:
		return append(p, bs.b[0], bs.b[1])
	case 3:
		return append(p, bs.b[0], bs.b[1], bs.b[2])
	case 4:
		return append(p, bs.b[0], bs.b[1], bs.b[2], bs.b[3])
	case 5:
		return append(p, bs.b[0], bs.b[1], bs.b[2], bs.b[3], bs.b[4])
	case 6:
		return append(p, bs.b[0], bs.b[1], bs.b[2], bs.b[3], bs.b[4], bs.b[5])
	case 7:
		return append(p, bs.b[0], bs.b[1], bs.b[2], bs.b[3], bs.b[4], bs.b[5], bs.b[6])
	case 8:
		return append(p, bs.b[0], bs.b[1], bs.b[2], bs.b[3], bs.b[4], bs.b[5], bs.b[6], bs.b[7])
	default:
		panic(panicInvalidByteSequenceLength)
	}
}

func newSeq2(r1, r2 rune) byteSequenceLong {
	var buf [2 * utf8.UTFMax]byte

	n := uint8(utf8.EncodeRune(buf[:], r1))
	n += uint8(utf8.EncodeRune(buf[n:], r2))

	return byteSequenceLong{n, buf}
}

type runeScape4 struct {
	numWideRunes uint8
	bits         [8]uint32
	// csv writing will only have up to 4 wide runes ever
	// - quote, escape, field sep, record sep
	wideRunes [4]rune
}

// addRune assumes that the rune is a valid unicode code point value
func (rs *runeScape4) addRune(r rune) {
	if r < utf8.RuneSelf {
		rs.addByte(byte(r))
		return
	}

	rs.addWideRune(r)
}

// addByte assumes that the byte is a valid unicode value less than 128
func (rs *runeScape4) addByte(b byte) {
	rs.bits[b>>5] |= (uint32(1) << (b & 31))
}

// containsByte will work with invalid unicode byte-length values as well
// if the set of runes was built with pre-validated ascii compatible unicode bytes.
//
// TODO: may not be getting inlined enough
func (rs *runeScape4) containsByte(b byte) bool {
	return (rs.bits[b>>5] & (uint32(1) << (b & 31))) != 0
}

// containsWideRune assumes that the rune is a valid unicode code point value that encodes to more than one byte
func (rs *runeScape4) containsWideRune(r rune) bool {

	// unwound the loop search to avoid loop overhead
	//
	// becomes a fast series of opcodes when compiled
	//
	// it is end-state equivalent to the following except
	// that the wide-rune check order is reversed:
	//
	//
	// for i := range rs.numWideRunes {
	// 	if rs.wideRunes[i] == r {
	// 		return true
	// 	}
	// }
	//
	// return false

	switch rs.numWideRunes {
	case 4:
		if rs.wideRunes[3] == r {
			return true
		}
		fallthrough
	case 3:
		if rs.wideRunes[2] == r {
			return true
		}
		fallthrough
	case 2:
		if rs.wideRunes[1] == r {
			return true
		}
		fallthrough
	case 1:
		return (rs.wideRunes[0] == r)
	}

	return false
}

// addWideRune assumes that the rune is a valid unicode code point value that encodes to more than one byte
func (rs *runeScape4) addWideRune(r rune) {
	if rs.containsWideRune(r) {
		return
	}

	rs.wideRunes[rs.numWideRunes] = r
	rs.numWideRunes++
}

// addWideRune assumes that the rune is a valid unicode code point value
//
// TODO: may not be getting inlined enough still
func (rs *runeScape4) containsRune(r rune) bool {
	if r < utf8.RuneSelf {
		return (rs.bits[byte(r)>>5] & (uint32(1) << (r & 31))) != 0
	}

	return rs.containsWideRune(r)
}

// addRuneUniqueUnchecked assumes that the rune is a valid unicode code point value that (if it is wide) has not already been added before
func (rs *runeScape4) addRuneUniqueUnchecked(r rune) {
	if r < utf8.RuneSelf {
		rs.addByte(byte(r))
		return
	}

	rs.wideRunes[rs.numWideRunes] = r
	rs.numWideRunes++
}

func (rs *runeScape4) indexAnyInString(s string) int {
	if rs.numWideRunes == 0 {
		for i := range len(s) {
			b := s[i]
			if (rs.bits[b>>5] & (uint32(1) << (b & 31))) != 0 {
				return i
			}
		}

		return -1
	}

	var i int
	for i < len(s) {
		b := s[i]
		if b < utf8.RuneSelf {
			if (rs.bits[b>>5] & (uint32(1) << (b & 31))) != 0 {
				return i
			}

			i++
			continue
		}

		r, n := utf8.DecodeRuneInString(s[i:])
		if n != 1 && rs.containsWideRune(r) {
			return i
		}

		i += n
	}

	return -1
}

func (rs *runeScape4) indexAnyRuneLenInString(s string) (rune, uint8, int) {
	if rs.numWideRunes == 0 {
		for i := range len(s) {
			b := s[i]
			if (rs.bits[b>>5] & (uint32(1) << (b & 31))) != 0 {
				return rune(b), 1, i
			}
		}

		return 0, 0, -1
	}

	var i int
	for i < len(s) {
		b := s[i]
		if b < utf8.RuneSelf {
			if (rs.bits[b>>5] & (uint32(1) << (b & 31))) != 0 {
				return rune(b), 1, i
			}

			i++
			continue
		}

		r, n := utf8.DecodeRuneInString(s[i:])
		if n != 1 && rs.containsWideRune(r) {
			return r, uint8(n), i
		}

		i += n
	}

	return 0, 0, -1
}

func (rs *runeScape4) indexAnyInBytes(p []byte) int {
	if rs.numWideRunes == 0 {
		for i, b := range p {
			if (rs.bits[b>>5] & (uint32(1) << (b & 31))) != 0 {
				return i
			}
		}

		return -1
	}

	var i int
	for i < len(p) {
		b := p[i]
		if b < utf8.RuneSelf {
			if (rs.bits[b>>5] & (uint32(1) << (b & 31))) != 0 {
				return i
			}

			i++
			continue
		}

		r, n := utf8.DecodeRune(p[i:])
		if n != 1 && rs.containsWideRune(r) {
			return i
		}

		i += n
	}

	return -1
}

// indexAnyRuneLenInBytes finds the index of a rune in a byte sequence
// and returns that with the rune value and byte length
//
// if you do not need the rune value nor the byte length then use indexAnyInBytes instead.
func (rs *runeScape4) indexAnyRuneLenInBytes(p []byte) (rune, uint8, int) {
	if rs.numWideRunes == 0 {
		for i, b := range p {
			if (rs.bits[b>>5] & (uint32(1) << (b & 31))) != 0 {
				return rune(b), 1, i
			}
		}

		return 0, 0, -1
	}

	var i int
	for i < len(p) {
		b := p[i]
		if b < utf8.RuneSelf {
			if (rs.bits[b>>5] & (uint32(1) << (b & 31))) != 0 {
				return rune(b), 1, i
			}

			i++
			continue
		}

		r, n := utf8.DecodeRune(p[i:])
		if n != 1 && rs.containsWideRune(r) {
			return r, uint8(n), i
		}

		i += n
	}

	return 0, 0, -1
}

type runeScape6 struct {
	numWideRunes uint8
	bits         [8]uint32
	// csv parsing will only have up to 6 wide runes ever
	wideRunes [6]rune
}

// addRune assumes that the rune is a valid unicode code point value
func (rs *runeScape6) addRune(r rune) {
	if r < utf8.RuneSelf {
		rs.addByte(byte(r))
		return
	}

	rs.addWideRune(r)
}

// addByte assumes that the byte is a valid unicode value less than 128
func (rs *runeScape6) addByte(b byte) {
	rs.bits[b>>5] |= (uint32(1) << (b & 31))
}

// containsWideRune assumes that the rune is a valid unicode code point value that encodes to more than one byte
func (rs *runeScape6) containsWideRune(r rune) bool {

	// unwound the loop search to avoid loop overhead
	//
	// becomes a fast series of opcodes when compiled
	//
	// it is end-state equivalent to the following except
	// that the wide-rune check order is reversed:
	//
	//
	// for i := range rs.numWideRunes {
	// 	if rs.wideRunes[i] == r {
	// 		return true
	// 	}
	// }
	//
	// return false

	switch rs.numWideRunes {
	case 6:
		if rs.wideRunes[5] == r {
			return true
		}
		fallthrough
	case 5:
		if rs.wideRunes[4] == r {
			return true
		}
		fallthrough
	case 4:
		if rs.wideRunes[3] == r {
			return true
		}
		fallthrough
	case 3:
		if rs.wideRunes[2] == r {
			return true
		}
		fallthrough
	case 2:
		if rs.wideRunes[1] == r {
			return true
		}
		fallthrough
	case 1:
		return (rs.wideRunes[0] == r)
	}

	return false
}

// addWideRune assumes that the rune is a valid unicode code point value that encodes to more than one byte
func (rs *runeScape6) addWideRune(r rune) {
	if rs.containsWideRune(r) {
		return
	}

	rs.wideRunes[rs.numWideRunes] = r
	rs.numWideRunes++
}

// addWideRune assumes that the rune is a valid unicode code point value
//
// TODO: may not be getting inlined enough still
func (rs *runeScape6) containsRune(r rune) bool {
	if r < utf8.RuneSelf {
		return (rs.bits[byte(r)>>5] & (uint32(1) << (r & 31))) != 0
	}

	return rs.containsWideRune(r)
}

// addRuneUniqueUnchecked assumes that the rune is a valid unicode code point value that (if it is wide) has not already been added before
func (rs *runeScape6) addRuneUniqueUnchecked(r rune) {
	if r < utf8.RuneSelf {
		rs.addByte(byte(r))
		return
	}

	rs.wideRunes[rs.numWideRunes] = r
	rs.numWideRunes++
}

func (rs *runeScape6) indexAnyInBytes(p []byte) int {
	if rs.numWideRunes == 0 {
		for i, b := range p {
			if (rs.bits[b>>5] & (uint32(1) << (b & 31))) != 0 {
				return i
			}
		}

		return -1
	}

	var i int
	for i < len(p) {
		b := p[i]
		if b < utf8.RuneSelf {
			if (rs.bits[b>>5] & (uint32(1) << (b & 31))) != 0 {
				return i
			}

			i++
			continue
		}

		r, n := utf8.DecodeRune(p[i:])
		if n != 1 && rs.containsWideRune(r) {
			return i
		}

		i += n
	}

	return -1
}
