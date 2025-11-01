package csv

import (
	"math/bits"
	"unicode/utf8"
)

const (
	utf8ContinuationByteMask             = 0xC0
	utf8StartWideRuneMinByteValue        = 0xC0
	utf8AfterMaskIsContinuationByteValue = 0x80
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
	numWideRunes    uint8
	singleByteBits  [2]uint64
	wideEndByteBits [2]uint32
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
//
// any change to this function likely needs to be replicated to addRuneUniqueUnchecked()
func (rs *runeScape4) addByte(b byte) {
	rs.singleByteBits[(b>>6)&1] |= (uint64(1) << (b & 63))
}

// containsByte will work with invalid unicode byte-length values as well
// if the set of runes was built with pre-validated ascii compatible unicode bytes.
func (rs *runeScape4) containsByte(b byte) bool {
	return (rs.singleByteBits[(b>>6)&1] & (uint64(1) << (b & 63))) != 0
}

// _containsWideEndByte works with either the last byte of a utf8 encoded rune
// or the last 8 bits of a rune/code-point
func (rs *runeScape4) _containsWideEndByte(b byte) bool {
	return (rs.wideEndByteBits[(b>>5)&1] & (uint32(1) << (b & 31))) != 0
}

// _containsWideRune assumes that the rune is a valid unicode code point value that encodes to more than one byte
//
// it is for internal use only
func (rs *runeScape4) _containsWideRune(r rune) bool {

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

func (rs *runeScape4) containsWideRune(r rune) bool {
	return ( /* inlined call to _containsWideEndByte: */ (rs.wideEndByteBits[(uint32(r)>>5)&1] & (uint32(1) << (r & 31))) != 0) && rs._containsWideRune(r)
}

// addWideRune assumes that the rune is a valid unicode code point value that encodes to more than one byte
//
// any change to this function likely needs to be replicated to addRuneUniqueUnchecked()
func (rs *runeScape4) addWideRune(r rune) {
	if rs._containsWideRune(r) {
		return
	}

	rs.wideRunes[rs.numWideRunes] = r
	rs.numWideRunes++

	rs.wideEndByteBits[(uint32(r)>>5)&1] |= (uint32(1) << (r & 31))
}

// containsByte assumes that the rune is a valid unicode code point value
func (rs *runeScape4) _containsRune(r rune) bool {
	if r < utf8.RuneSelf {
		return ( /* inlined call to containsByte: */ (rs.singleByteBits[(uint32(r)>>6)&1] & (uint64(1) << (r & 63))) != 0)
	}

	return rs._containsWideRune(r)
}

// addRuneUniqueUnchecked assumes that the rune is a valid unicode code point value that (if it is wide) has not already been added before
func (rs *runeScape4) addRuneUniqueUnchecked(r rune) {
	if r < utf8.RuneSelf {
		rs.addByte(byte(r))
		return
	}

	rs.wideRunes[rs.numWideRunes] = r
	rs.numWideRunes++

	rs.wideEndByteBits[(uint32(r)>>5)&1] |= (uint32(1) << (r & 31))
}

func (rs *runeScape4) indexAnyInString(s string) int {
	if rs.numWideRunes == 0 {
		for i := range len(s) {
			b := s[i]
			if /* inlined call to containsByte: */ (rs.singleByteBits[(b>>6)&1] & (uint64(1) << (b & 63))) != 0 {
				return i
			}
		}

		return -1
	}

	var i int
	lastWideStartIdx := -utf8.UTFMax
SCAN_LOOP:
	for {
		if i >= len(s) {
			break
		}

		b := s[i]
		switch {
		case b < utf8.RuneSelf:
			if /* inlined call to containsByte: */ (rs.singleByteBits[(b>>6)&1] & (uint64(1) << (b & 63))) != 0 {
				return i
			}
		case b >= utf8StartWideRuneMinByteValue:
			lastWideStartIdx = i
		case (b & utf8ContinuationByteMask) == utf8AfterMaskIsContinuationByteValue:
			if ( /* inlined call to _containsWideEndByte: */ (rs.wideEndByteBits[(b>>5)&1] & (uint32(1) << (b & 31))) != 0) && (i-lastWideStartIdx) < utf8.UTFMax {
				endIdx := lastWideStartIdx + bits.LeadingZeros8(^s[lastWideStartIdx]) - 1
				if endIdx == i {
					// verified already that
					// 1 - the possible end utf8 multi-byte sequence is an ending value in the set
					// 2 - the distance from the last known start of multi-byte sequence is close enough for this byte to be an ending byte
					// 3 - the current index is indeed the end byte of the potentially invalid sequence
					//
					// so now need to verify that
					// 1 - the rune decodes properly out of the bytes
					// 2 - then we need to ensure that the full rune is recognized as in the set
					if r, n := utf8.DecodeRuneInString(s[lastWideStartIdx:]); n > 1 && rs._containsWideRune(r) {
						return lastWideStartIdx
					}

					lastWideStartIdx = -utf8.UTFMax
					break
				}

				// this is not the ending byte we are looking for, so
				//
				// check if
				// - the multi-byte sequence end was passed already
				// - or if the end is unreasonably distant
				if endIdx < i || endIdx > lastWideStartIdx+utf8.UTFMax-1 {
					// then something is strongly off with this sequence, lets continue onwards to the next byte and make it clear there is no previous start of multi-byte position before we do
					lastWideStartIdx = -utf8.UTFMax
					break
				}

				// otherwise the end is coming up, lets scan forward to that end if legal to do so
				// note that legal step bytes between here and there must all be continuation bytes
				//
				// if an illegal byte is found then the sequence is not possible and we can continue scanning at the illegal byte position again
				// after making it clear there is no valid "start of wide" index
				fwdIdx := i + 1
				for {
					if fwdIdx >= len(s) {
						return -1
					}

					b = s[fwdIdx]
					if (b & utf8ContinuationByteMask) != utf8AfterMaskIsContinuationByteValue {
						// not a continuation byte
						lastWideStartIdx = -utf8.UTFMax
						i = fwdIdx
						continue SCAN_LOOP
					}

					// byte is a utf8 continuation value
					if fwdIdx != endIdx {
						// but it is not the end index
						fwdIdx++
						continue
					}

					// at the end index, lets check if we are done fully

					if /* inlined call to _containsWideEndByte: */ (rs.wideEndByteBits[(b>>5)&1] & (uint32(1) << (b & 31))) != 0 {
						if r, n := utf8.DecodeRuneInString(s[lastWideStartIdx:]); n > 1 && rs._containsWideRune(r) {
							return lastWideStartIdx
						}
					}

					// not a match, lets continue from the upper loop on the next byte
					lastWideStartIdx = -utf8.UTFMax
					i = fwdIdx + 1
					continue SCAN_LOOP
				}
			}
		}

		i++
	}

	return -1
}

func (rs *runeScape4) indexAnyRuneLenInString(s string) (rune, uint8, int) {
	if rs.numWideRunes == 0 {
		for i := range len(s) {
			b := s[i]
			if /* inlined call to containsByte: */ (rs.singleByteBits[(b>>6)&1] & (uint64(1) << (b & 63))) != 0 {
				return rune(b), 1, i
			}
		}

		return 0, 0, -1
	}

	var i int
	lastWideStartIdx := -utf8.UTFMax
SCAN_LOOP:
	for {
		if i >= len(s) {
			break
		}

		b := s[i]
		switch {
		case b < utf8.RuneSelf:
			if /* inlined call to containsByte: */ (rs.singleByteBits[(b>>6)&1] & (uint64(1) << (b & 63))) != 0 {
				return rune(b), 1, i
			}
		case b >= utf8StartWideRuneMinByteValue:
			lastWideStartIdx = i
		case (b & utf8ContinuationByteMask) == utf8AfterMaskIsContinuationByteValue:
			if ( /* inlined call to _containsWideEndByte: */ (rs.wideEndByteBits[(b>>5)&1] & (uint32(1) << (b & 31))) != 0) && (i-lastWideStartIdx) < utf8.UTFMax {
				endIdx := lastWideStartIdx + bits.LeadingZeros8(^s[lastWideStartIdx]) - 1
				if endIdx == i {
					// verified already that
					// 1 - the possible end utf8 multi-byte sequence is an ending value in the set
					// 2 - the distance from the last known start of multi-byte sequence is close enough for this byte to be an ending byte
					// 3 - the current index is indeed the end byte of the potentially invalid sequence
					//
					// so now need to verify that
					// 1 - the rune decodes properly out of the bytes
					// 2 - then we need to ensure that the full rune is recognized as in the set
					if r, n := utf8.DecodeRuneInString(s[lastWideStartIdx:]); n > 1 && rs._containsWideRune(r) {
						return r, uint8(n), lastWideStartIdx
					}

					lastWideStartIdx = -utf8.UTFMax
					break
				}

				// this is not the ending byte we are looking for, so
				//
				// check if
				// - the multi-byte sequence end was passed already
				// - or if the end is unreasonably distant
				if endIdx < i || endIdx > lastWideStartIdx+utf8.UTFMax-1 {
					// then something is strongly off with this sequence, lets continue onwards to the next byte and make it clear there is no previous start of multi-byte position before we do
					lastWideStartIdx = -utf8.UTFMax
					break
				}

				// otherwise the end is coming up, lets scan forward to that end if legal to do so
				// note that legal step bytes between here and there must all be continuation bytes
				//
				// if an illegal byte is found then the sequence is not possible and we can continue scanning at the illegal byte position again
				// after making it clear there is no valid "start of wide" index
				fwdIdx := i + 1
				for {
					if fwdIdx >= len(s) {
						return 0, 0, -1
					}

					b = s[fwdIdx]
					if (b & utf8ContinuationByteMask) != utf8AfterMaskIsContinuationByteValue {
						// not a continuation byte
						lastWideStartIdx = -utf8.UTFMax
						i = fwdIdx
						continue SCAN_LOOP
					}

					// byte is a utf8 continuation value
					if fwdIdx != endIdx {
						// but it is not the end index
						fwdIdx++
						continue
					}

					// at the end index, lets check if we are done fully

					if /* inlined call to _containsWideEndByte: */ (rs.wideEndByteBits[(b>>5)&1] & (uint32(1) << (b & 31))) != 0 {
						if r, n := utf8.DecodeRuneInString(s[lastWideStartIdx:]); n > 1 && rs._containsWideRune(r) {
							return r, uint8(n), lastWideStartIdx
						}
					}

					// not a match, lets continue from the upper loop on the next byte
					lastWideStartIdx = -utf8.UTFMax
					i = fwdIdx + 1
					continue SCAN_LOOP
				}
			}
		}

		i++
	}

	return 0, 0, -1
}

func (rs *runeScape4) indexAnyInBytes(p []byte) int {
	if rs.numWideRunes == 0 {
		for i, b := range p {
			if /* inlined call to containsByte: */ (rs.singleByteBits[(b>>6)&1] & (uint64(1) << (b & 63))) != 0 {
				return i
			}
		}

		return -1
	}

	var i int
	lastWideStartIdx := -utf8.UTFMax
SCAN_LOOP:
	for {
		if i >= len(p) {
			break
		}

		b := p[i]
		switch {
		case b < utf8.RuneSelf:
			if /* inlined call to containsByte: */ (rs.singleByteBits[(b>>6)&1] & (uint64(1) << (b & 63))) != 0 {
				return i
			}
		case b >= utf8StartWideRuneMinByteValue:
			lastWideStartIdx = i
		case (b & utf8ContinuationByteMask) == utf8AfterMaskIsContinuationByteValue:
			if ( /* inlined call to _containsWideEndByte: */ (rs.wideEndByteBits[(b>>5)&1] & (uint32(1) << (b & 31))) != 0) && (i-lastWideStartIdx) < utf8.UTFMax {
				endIdx := lastWideStartIdx + bits.LeadingZeros8(^p[lastWideStartIdx]) - 1
				if endIdx == i {
					// verified already that
					// 1 - the possible end utf8 multi-byte sequence is an ending value in the set
					// 2 - the distance from the last known start of multi-byte sequence is close enough for this byte to be an ending byte
					// 3 - the current index is indeed the end byte of the potentially invalid sequence
					//
					// so now need to verify that
					// 1 - the rune decodes properly out of the bytes
					// 2 - then we need to ensure that the full rune is recognized as in the set
					if r, n := utf8.DecodeRune(p[lastWideStartIdx:]); n > 1 && rs._containsWideRune(r) {
						return lastWideStartIdx
					}

					lastWideStartIdx = -utf8.UTFMax
					break
				}

				// this is not the ending byte we are looking for, so
				//
				// check if
				// - the multi-byte sequence end was passed already
				// - or if the end is unreasonably distant
				if endIdx < i || endIdx > lastWideStartIdx+utf8.UTFMax-1 {
					// then something is strongly off with this sequence, lets continue onwards to the next byte and make it clear there is no previous start of multi-byte position before we do
					lastWideStartIdx = -utf8.UTFMax
					break
				}

				// otherwise the end is coming up, lets scan forward to that end if legal to do so
				// note that legal step bytes between here and there must all be continuation bytes
				//
				// if an illegal byte is found then the sequence is not possible and we can continue scanning at the illegal byte position again
				// after making it clear there is no valid "start of wide" index
				fwdIdx := i + 1
				for {
					if fwdIdx >= len(p) {
						return -1
					}

					b = p[fwdIdx]
					if (b & utf8ContinuationByteMask) != utf8AfterMaskIsContinuationByteValue {
						// not a continuation byte
						lastWideStartIdx = -utf8.UTFMax
						i = fwdIdx
						continue SCAN_LOOP
					}

					// byte is a utf8 continuation value
					if fwdIdx != endIdx {
						// but it is not the end index
						fwdIdx++
						continue
					}

					// at the end index, lets check if we are done fully

					if /* inlined call to _containsWideEndByte: */ (rs.wideEndByteBits[(b>>5)&1] & (uint32(1) << (b & 31))) != 0 {
						if r, n := utf8.DecodeRune(p[lastWideStartIdx:]); n > 1 && rs._containsWideRune(r) {
							return lastWideStartIdx
						}
					}

					// not a match, lets continue from the upper loop on the next byte
					lastWideStartIdx = -utf8.UTFMax
					i = fwdIdx + 1
					continue SCAN_LOOP
				}
			}
		}

		i++
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
			if /* inlined call to containsByte: */ (rs.singleByteBits[(b>>6)&1] & (uint64(1) << (b & 63))) != 0 {
				return rune(b), 1, i
			}
		}

		return 0, 0, -1
	}

	var i int
	lastWideStartIdx := -utf8.UTFMax
SCAN_LOOP:
	for {
		if i >= len(p) {
			break
		}

		b := p[i]
		switch {
		case b < utf8.RuneSelf:
			if /* inlined call to containsByte: */ (rs.singleByteBits[(b>>6)&1] & (uint64(1) << (b & 63))) != 0 {
				return rune(b), 1, i
			}
		case b >= utf8StartWideRuneMinByteValue:
			lastWideStartIdx = i
		case (b & utf8ContinuationByteMask) == utf8AfterMaskIsContinuationByteValue:
			if ( /* inlined call to _containsWideEndByte: */ (rs.wideEndByteBits[(b>>5)&1] & (uint32(1) << (b & 31))) != 0) && (i-lastWideStartIdx) < utf8.UTFMax {
				endIdx := lastWideStartIdx + bits.LeadingZeros8(^p[lastWideStartIdx]) - 1
				if endIdx == i {
					// verified already that
					// 1 - the possible end utf8 multi-byte sequence is an ending value in the set
					// 2 - the distance from the last known start of multi-byte sequence is close enough for this byte to be an ending byte
					// 3 - the current index is indeed the end byte of the potentially invalid sequence
					//
					// so now need to verify that
					// 1 - the rune decodes properly out of the bytes
					// 2 - then we need to ensure that the full rune is recognized as in the set
					if r, n := utf8.DecodeRune(p[lastWideStartIdx:]); n > 1 && rs._containsWideRune(r) {
						return r, uint8(n), lastWideStartIdx
					}

					lastWideStartIdx = -utf8.UTFMax
					break
				}

				// this is not the ending byte we are looking for, so
				//
				// check if
				// - the multi-byte sequence end was passed already
				// - or if the end is unreasonably distant
				if endIdx < i || endIdx > lastWideStartIdx+utf8.UTFMax-1 {
					// then something is strongly off with this sequence, lets continue onwards to the next byte and make it clear there is no previous start of multi-byte position before we do
					lastWideStartIdx = -utf8.UTFMax
					break
				}

				// otherwise the end is coming up, lets scan forward to that end if legal to do so
				// note that legal step bytes between here and there must all be continuation bytes
				//
				// if an illegal byte is found then the sequence is not possible and we can continue scanning at the illegal byte position again
				// after making it clear there is no valid "start of wide" index
				fwdIdx := i + 1
				for {
					if fwdIdx >= len(p) {
						return 0, 0, -1
					}

					b = p[fwdIdx]
					if (b & utf8ContinuationByteMask) != utf8AfterMaskIsContinuationByteValue {
						// not a continuation byte
						lastWideStartIdx = -utf8.UTFMax
						i = fwdIdx
						continue SCAN_LOOP
					}

					// byte is a utf8 continuation value
					if fwdIdx != endIdx {
						// but it is not the end index
						fwdIdx++
						continue
					}

					// at the end index, lets check if we are done fully

					if /* inlined call to _containsWideEndByte: */ (rs.wideEndByteBits[(b>>5)&1] & (uint32(1) << (b & 31))) != 0 {
						if r, n := utf8.DecodeRune(p[lastWideStartIdx:]); n > 1 && rs._containsWideRune(r) {
							return r, uint8(n), lastWideStartIdx
						}
					}

					// not a match, lets continue from the upper loop on the next byte
					lastWideStartIdx = -utf8.UTFMax
					i = fwdIdx + 1
					continue SCAN_LOOP
				}
			}
		}

		i++
	}

	return 0, 0, -1
}

type runeScape6 struct {
	numWideRunes    uint8
	singleByteBits  [2]uint64
	wideEndByteBits [2]uint32
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

// the following code is intentionally commented out and not removed because it has been fully inlined
//
// let this act as a reference and nothing more

// // _containsByte will work with invalid unicode byte-length values as well
// // if the set of runes was built with pre-validated ascii compatible unicode bytes.
// func (rs *runeScape6) _containsByte(b byte) bool {
// 	return (rs.bits[(b>>6)&1] & (uint64(1) << (b & 63))) != 0
// }

// addByte assumes that the byte is a valid unicode value less than 128
//
// any change to this function likely needs to be replicated to addRuneUniqueUnchecked()
func (rs *runeScape6) addByte(b byte) {
	rs.singleByteBits[(b>>6)&1] |= (uint64(1) << (b & 63))
}

// _containsWideEndByte works with either the last byte of a utf8 encoded rune
// or the last 8 bits of a rune/code-point
func (rs *runeScape6) _containsWideEndByte(b byte) bool {
	return (rs.wideEndByteBits[(b>>5)&1] & (uint32(1) << (b & 31))) != 0
}

// _containsWideRune assumes that the rune is a valid unicode code point value that encodes to more than one byte
func (rs *runeScape6) _containsWideRune(r rune) bool {

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
//
// any change to this function likely needs to be replicated to addRuneUniqueUnchecked()
func (rs *runeScape6) addWideRune(r rune) {
	if rs._containsWideRune(r) {
		return
	}

	rs.wideRunes[rs.numWideRunes] = r
	rs.numWideRunes++

	rs.wideEndByteBits[(uint32(r)>>5)&1] |= (uint32(1) << (r & 31))
}

// _containsRune assumes that the rune is a valid unicode code point value
func (rs *runeScape6) _containsRune(r rune) bool {
	if r < utf8.RuneSelf {
		return ( /* inlined call to _containsByte: */ (rs.singleByteBits[(uint32(r)>>6)&1] & (uint64(1) << (r & 63))) != 0)
	}

	return rs._containsWideRune(r)
}

// addRuneUniqueUnchecked assumes that the rune is a valid unicode code point value that (if it is wide) has not already been added before
//
// any change to this function likely needs to be replicated to addWideRune() or addByte()
func (rs *runeScape6) addRuneUniqueUnchecked(r rune) {
	if r < utf8.RuneSelf {
		rs.addByte(byte(r))
		return
	}

	rs.wideRunes[rs.numWideRunes] = r
	rs.numWideRunes++

	rs.wideEndByteBits[(uint32(r)>>5)&1] |= (uint32(1) << (r & 31))
}

func (rs *runeScape6) indexAnyRuneLenInBytes(p []byte) (rune, uint8, int) {
	if rs.numWideRunes == 0 {
		for i, b := range p {
			if /* inlined call to _containsByte: */ (rs.singleByteBits[(b>>6)&1] & (uint64(1) << (b & 63))) != 0 {
				return rune(b), 1, i
			}
		}

		return 0, 0, -1
	}

	var i int
	lastWideStartIdx := -utf8.UTFMax
SCAN_LOOP:
	for {
		if i >= len(p) {
			break
		}

		b := p[i]
		switch {
		case b < utf8.RuneSelf:
			if /* inlined call to containsByte: */ (rs.singleByteBits[(b>>6)&1] & (uint64(1) << (b & 63))) != 0 {
				return rune(b), 1, i
			}
		case b >= utf8StartWideRuneMinByteValue:
			lastWideStartIdx = i
		case (b & utf8ContinuationByteMask) == utf8AfterMaskIsContinuationByteValue:
			if ( /* inlined call to _containsWideEndByte: */ (rs.wideEndByteBits[(b>>5)&1] & (uint32(1) << (b & 31))) != 0) && (i-lastWideStartIdx) < utf8.UTFMax {
				endIdx := lastWideStartIdx + bits.LeadingZeros8(^p[lastWideStartIdx]) - 1
				if endIdx == i {
					// verified already that
					// 1 - the possible end utf8 multi-byte sequence is an ending value in the set
					// 2 - the distance from the last known start of multi-byte sequence is close enough for this byte to be an ending byte
					// 3 - the current index is indeed the end byte of the potentially invalid sequence
					//
					// so now need to verify that
					// 1 - the rune decodes properly out of the bytes
					// 2 - then we need to ensure that the full rune is recognized as in the set
					if r, n := utf8.DecodeRune(p[lastWideStartIdx:]); n > 1 && rs._containsWideRune(r) {
						return r, uint8(n), lastWideStartIdx
					}

					lastWideStartIdx = -utf8.UTFMax
					break
				}

				// this is not the ending byte we are looking for, so
				//
				// check if
				// - the multi-byte sequence end was passed already
				// - or if the end is unreasonably distant
				if endIdx < i || endIdx > lastWideStartIdx+utf8.UTFMax-1 {
					// then something is strongly off with this sequence, lets continue onwards to the next byte and make it clear there is no previous start of multi-byte position before we do
					lastWideStartIdx = -utf8.UTFMax
					break
				}

				// otherwise the end is coming up, lets scan forward to that end if legal to do so
				// note that legal step bytes between here and there must all be continuation bytes
				//
				// if an illegal byte is found then the sequence is not possible and we can continue scanning at the illegal byte position again
				// after making it clear there is no valid "start of wide" index
				fwdIdx := i + 1
				for {
					if fwdIdx >= len(p) {
						return 0, 0, -1
					}

					b = p[fwdIdx]
					if (b & utf8ContinuationByteMask) != utf8AfterMaskIsContinuationByteValue {
						// not a continuation byte
						lastWideStartIdx = -utf8.UTFMax
						i = fwdIdx
						continue SCAN_LOOP
					}

					// byte is a utf8 continuation value
					if fwdIdx != endIdx {
						// but it is not the end index
						fwdIdx++
						continue
					}

					// at the end index, lets check if we are done fully

					if /* inlined call to _containsWideEndByte: */ (rs.wideEndByteBits[(b>>5)&1] & (uint32(1) << (b & 31))) != 0 {
						if r, n := utf8.DecodeRune(p[lastWideStartIdx:]); n > 1 && rs._containsWideRune(r) {
							return r, uint8(n), lastWideStartIdx
						}
					}

					// not a match, lets continue from the upper loop on the next byte
					lastWideStartIdx = -utf8.UTFMax
					i = fwdIdx + 1
					continue SCAN_LOOP
				}
			}
		}

		i++
	}

	return 0, 0, -1
}
