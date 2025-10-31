package csv

import (
	"fmt"
	"math/rand"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

const (
	utf16SurrogateCodePointStart   = 0xD800
	utf16SurrogateCodePointEnd     = 0xDFFF
	utf16SurrogateCodePointGapSize = utf16SurrogateCodePointEnd - utf16SurrogateCodePointStart + 1
)

func Test_runeSet4_containsSingleByteRune(t *testing.T) {
	t.Parallel()

	is := assert.New(t)

	var rs runeSet4

	for i := range utf8.RuneSelf {
		is.False(rs.containsSingleByteRune(byte(i)))
		r, n, x := rs.indexAnyRuneLenInBytes([]byte{byte(i)})
		is.Equal(rune(0), r)
		is.Equal(uint8(0), n)
		is.Equal(int(-1), x)
		r, n, x = rs.indexAnyRuneLenInString(string([]byte{byte(i)}))
		is.Equal(rune(0), r)
		is.Equal(uint8(0), n)
		is.Equal(int(-1), x)
		x = rs.indexAnyInBytes([]byte{byte(i)})
		is.Equal(int(-1), x)
		x = rs.indexAnyInString(string([]byte{byte(i)}))
		is.Equal(int(-1), x)

		rs.addByte(byte(i))

		is.True(rs.containsSingleByteRune(byte(i)))
		r, n, x = rs.indexAnyRuneLenInBytes([]byte{byte(i)})
		is.Equal(rune(i), r)
		is.Equal(uint8(1), n)
		is.Equal(int(0), x)
		r, n, x = rs.indexAnyRuneLenInString(string([]byte{byte(i)}))
		is.Equal(rune(i), r)
		is.Equal(uint8(1), n)
		is.Equal(int(0), x)
		x = rs.indexAnyInBytes([]byte{byte(i)})
		is.Equal(int(0), x)
		x = rs.indexAnyInString(string([]byte{byte(i)}))
		is.Equal(int(0), x)
	}

	for i := utf8.RuneSelf; i < 256; i++ {
		is.False(rs.containsSingleByteRune(byte(i)))
		r, n, x := rs.indexAnyRuneLenInBytes([]byte{byte(i)})
		is.Equal(rune(0), r)
		is.Equal(uint8(0), n)
		is.Equal(int(-1), x)
		r, n, x = rs.indexAnyRuneLenInString(string([]byte{byte(i)}))
		is.Equal(rune(0), r)
		is.Equal(uint8(0), n)
		is.Equal(int(-1), x)
		x = rs.indexAnyInBytes([]byte{byte(i)})
		is.Equal(int(-1), x)
		x = rs.indexAnyInString(string([]byte{byte(i)}))
		is.Equal(int(-1), x)
	}
}

func Test_runeSet4_containsWideRune(t *testing.T) {
	t.Parallel()

	is := assert.New(t)

	{
		var rs runeSet4

		rs.addRune(test4ByteRune)
		rs.addRune(test4ByteRune + 1)
		rs.addRune(test4ByteRune + 2)
		rs.addRune(test4ByteRune + 3)

		is.False(rs.internalContainsRune(test4ByteRune + 4))
		is.True(rs.internalContainsRune(test4ByteRune + 3))
		is.True(rs.internalContainsRune(test4ByteRune + 2))
		is.True(rs.internalContainsRune(test4ByteRune + 1))
		is.True(rs.internalContainsRune(test4ByteRune))
	}

	// adding the same wide rune twice should not change rune count
	{
		var rs runeSet4

		rs.addRune(test4ByteRune)
		is.Equal(uint8(1), rs.mbRuneCount)

		rs.addRune(test4ByteRune)
		is.Equal(uint8(1), rs.mbRuneCount)
		is.False(rs.internalContainsRune(test1ByteRune))
	}

	// when there is a mix of wide rune in the set and in the search string and hit is wide rune
	// then (0, 0, -1) should not be returned from indexAnyRuneLenInString
	{
		var rs runeSet4

		rs.addRune(test4ByteRune)
		rs.addRune(test1ByteRune)

		r, n, i := rs.indexAnyRuneLenInString(test2ByteRuneEnc + "0" + test3ByteRuneEnc + "1" + test1ByteRuneEnc)
		is.Equal(test1ByteRune, r)
		is.Equal(uint8(1), n)
		is.Equal(7, i)

		i = rs.indexAnyInString(test2ByteRuneEnc + "0" + test3ByteRuneEnc + "1" + test1ByteRuneEnc)
		is.Equal(7, i)
	}

	// when there is a mix of wide rune in the set and in the search string and hit is wide rune
	// then (0, 0, -1) should not be returned from indexAnyRuneLenInBytes
	{

		var rs runeSet4

		rs.addRune(test4ByteRune)
		rs.addRune(test1ByteRune)

		r, n, i := rs.indexAnyRuneLenInBytes([]byte(test2ByteRuneEnc + "0" + test3ByteRuneEnc + "1" + test1ByteRuneEnc))
		is.Equal(test1ByteRune, r)
		is.Equal(uint8(1), n)
		is.Equal(7, i)

		i = rs.indexAnyInBytes([]byte(test2ByteRuneEnc + "0" + test3ByteRuneEnc + "1" + test1ByteRuneEnc))
		is.Equal(7, i)
	}

	// when there is a mix of wide rune in the set and in the search string and hit is wide rune
	// then (0, 0, -1) should not be returned from indexAnyRuneLenInString
	{
		var rs runeSet4

		rs.addRune(test4ByteRune)
		rs.addRune(test1ByteRune)

		r, n, i := rs.indexAnyRuneLenInString(test2ByteRuneEnc + "0" + test3ByteRuneEnc + "1" + test4ByteRuneEnc)
		is.Equal(test4ByteRune, r)
		is.Equal(uint8(4), n)
		is.Equal(7, i)

		i = rs.indexAnyInString(test2ByteRuneEnc + "0" + test3ByteRuneEnc + "1" + test4ByteRuneEnc)
		is.Equal(7, i)
	}

	// when there is a mix of wide rune in the set and in the search string and hit is wide rune
	// then (0, 0, -1) should not be returned from indexAnyRuneLenInBytes
	{

		var rs runeSet4

		rs.addRune(test4ByteRune)
		rs.addRune(test1ByteRune)

		r, n, i := rs.indexAnyRuneLenInBytes([]byte(test2ByteRuneEnc + "0" + test3ByteRuneEnc + "1" + test4ByteRuneEnc))
		is.Equal(test4ByteRune, r)
		is.Equal(uint8(4), n)
		is.Equal(7, i)

		i = rs.indexAnyInBytes([]byte(test2ByteRuneEnc + "0" + test3ByteRuneEnc + "1" + test4ByteRuneEnc))
		is.Equal(7, i)
	}

	// when there is a wide rune in the set and not in the search string
	// then indexAnyRuneLenInString should return (0, 0, -1)
	{

		var rs runeSet4

		rs.addRune(test4ByteRune)

		r, n, i := rs.indexAnyRuneLenInString(test1ByteRuneEnc)
		is.Equal(rune(0), r)
		is.Equal(uint8(0), n)
		is.Equal(int(-1), i)

		i = rs.indexAnyInString(test1ByteRuneEnc)
		is.Equal(int(-1), i)
	}

	// when there is a mix of wide rune in the set and not in the search string
	// then indexAnyRuneLenInString should not return (0, 0, -1)
	{

		var rs runeSet4

		rs.addRune(test1ByteRune)
		rs.addRune(test4ByteRune)

		r, n, i := rs.indexAnyRuneLenInString(test3ByteRuneEnc + test1ByteRuneEnc)
		is.Equal(test1ByteRune, r)
		is.Equal(uint8(1), n)
		is.Equal(int(3), i)

		i = rs.indexAnyInString(test3ByteRuneEnc + test1ByteRuneEnc)
		is.Equal(int(3), i)

		r, n, i = rs.indexAnyRuneLenInString(test3ByteRuneEnc + test4ByteRuneEnc)
		is.Equal(test4ByteRune, r)
		is.Equal(uint8(4), n)
		is.Equal(int(3), i)

		i = rs.indexAnyInString(test3ByteRuneEnc + test4ByteRuneEnc)
		is.Equal(int(3), i)
	}

	// when there is a mix of wide rune in the set and no overlap in a mixed string
	// then indexAnyRuneLenInBytes should return (0, 0, -1)
	{

		var rs runeSet4

		rs.addRune(test1ByteRune)
		rs.addRune(test4ByteRune)

		r, n, i := rs.indexAnyRuneLenInBytes([]byte("0" + test3ByteRuneEnc))
		is.Equal(rune(0), r)
		is.Equal(uint8(0), n)
		is.Equal(int(-1), i)

		i = rs.indexAnyInBytes([]byte("0" + test3ByteRuneEnc))
		is.Equal(int(-1), i)
	}

	// when there is a mix of wide rune in the set and overlap in a mixed string
	// then indexAnyRuneLenInBytes should not return (0, 0, -1)
	{

		var rs runeSet4

		rs.addRune(test1ByteRune)
		rs.addRune(test4ByteRune)

		r, n, i := rs.indexAnyRuneLenInBytes([]byte("0" + test3ByteRuneEnc + test4ByteRuneEnc))
		is.Equal(test4ByteRune, r)
		is.Equal(uint8(4), n)
		is.Equal(int(4), i)

		i = rs.indexAnyInBytes([]byte("0" + test3ByteRuneEnc + test4ByteRuneEnc))
		is.Equal(int(4), i)
	}

	// when a random wide rune is added
	//
	// containsWideEndByte should be able to find it by last encoded byte form
	// as well as full rune form.
	{
		var buf [utf8.UTFMax]byte
		r := rune(rand.Intn(utf8.MaxRune+1-utf8.RuneSelf-utf16SurrogateCodePointGapSize) + utf8.RuneSelf)
		if r >= utf16SurrogateCodePointStart {
			r += utf16SurrogateCodePointGapSize
		}

		// added via addRuneUniqueUnchecked
		{
			var rs runeSet4
			rs.addRuneUniqueUnchecked(r)

			is.True(rs.internalContainsMBEndByte(byte(r)), "r=%d", uint32(r))

			// last 6 bits of the last byte in an encoded rune are always the
			// last 6 bits of the code point rune value
			is.True(rs.internalContainsMBEndByte(byte(r)&0x3F), "r=%d", uint32(r))

			// but for good measure, test the long way too
			n := utf8.EncodeRune(buf[:], r)
			is.NotEqual(1, n)
			is.True(rs.internalContainsMBEndByte(buf[n-1]), "r=%d", uint32(r))
		}

		// added via addMBRune
		{
			var rs runeSet4
			rs.addMBRune(r)

			is.True(rs.internalContainsMBEndByte(byte(r)), "r=%d", uint32(r))

			// last 6 bits of the last byte in an encoded rune are always the
			// last 6 bits of the code point rune value
			is.True(rs.internalContainsMBEndByte(byte(r)&0x3F), "r=%d", uint32(r))

			// but for good measure, test the long way too
			n := utf8.EncodeRune(buf[:], r)
			is.NotEqual(1, n)
			is.True(rs.internalContainsMBEndByte(buf[n-1]), "r=%d", uint32(r))
		}
	}
}

func Test_runeSet6_containsWideRune(t *testing.T) {
	t.Parallel()

	is := assert.New(t)

	{
		var rs runeSet6

		rs.addRune(test4ByteRune)
		rs.addRune(test4ByteRune + 1)
		rs.addRune(test4ByteRune + 2)
		rs.addRune(test4ByteRune + 3)
		rs.addRune(test4ByteRune + 4)
		rs.addRune(test4ByteRune + 5)

		is.False(rs.internalContainsRune(test4ByteRune + 6))
		is.True(rs.internalContainsRune(test4ByteRune + 5))
		is.True(rs.internalContainsRune(test4ByteRune + 4))
		is.True(rs.internalContainsRune(test4ByteRune + 3))
		is.True(rs.internalContainsRune(test4ByteRune + 2))
		is.True(rs.internalContainsRune(test4ByteRune + 1))
		is.True(rs.internalContainsRune(test4ByteRune))
	}

	// adding the same wide rune twice should not change rune count
	{
		var rs runeSet6

		rs.addRune(test4ByteRune)
		is.Equal(uint8(1), rs.mbRuneCount)

		rs.addRune(test4ByteRune)
		is.Equal(uint8(1), rs.mbRuneCount)
		is.False(rs.internalContainsRune(test1ByteRune))
	}

	// when there is a mix of wide rune in the set and in the search string and hit is wide rune
	// then (0, 0, -1) should not be returned from indexAnyRuneLenInString
	{
		var rs runeSet6

		rs.addRune(test4ByteRune)
		rs.addRune(test1ByteRune)

		r, n, i := rs.indexAnyRuneLenInString(test2ByteRuneEnc + "0" + test3ByteRuneEnc + "1" + test1ByteRuneEnc)
		is.Equal(test1ByteRune, r)
		is.Equal(uint8(1), n)
		is.Equal(7, i)

		i = rs.indexAnyInString(test2ByteRuneEnc + "0" + test3ByteRuneEnc + "1" + test1ByteRuneEnc)
		is.Equal(7, i)
	}

	// when there is a mix of wide rune in the set and in the search string and hit is wide rune
	// then (0, 0, -1) should not be returned from indexAnyRuneLenInBytes
	{

		var rs runeSet6

		rs.addRune(test4ByteRune)
		rs.addRune(test1ByteRune)

		r, n, i := rs.indexAnyRuneLenInBytes([]byte(test2ByteRuneEnc + "0" + test3ByteRuneEnc + "1" + test1ByteRuneEnc))
		is.Equal(test1ByteRune, r)
		is.Equal(uint8(1), n)
		is.Equal(7, i)

		i = rs.indexAnyInBytes([]byte(test2ByteRuneEnc + "0" + test3ByteRuneEnc + "1" + test1ByteRuneEnc))
		is.Equal(7, i)
	}

	// when there is a mix of wide rune in the set and in the search string and hit is wide rune
	// then (0, 0, -1) should not be returned from indexAnyRuneLenInString
	{
		var rs runeSet6

		rs.addRune(test4ByteRune)
		rs.addRune(test1ByteRune)

		r, n, i := rs.indexAnyRuneLenInString(test2ByteRuneEnc + "0" + test3ByteRuneEnc + "1" + test4ByteRuneEnc)
		is.Equal(test4ByteRune, r)
		is.Equal(uint8(4), n)
		is.Equal(7, i)

		i = rs.indexAnyInString(test2ByteRuneEnc + "0" + test3ByteRuneEnc + "1" + test4ByteRuneEnc)
		is.Equal(7, i)
	}

	// when there is a mix of wide rune in the set and in the search string and hit is wide rune
	// then (0, 0, -1) should not be returned from indexAnyRuneLenInBytes
	{

		var rs runeSet6

		rs.addRune(test4ByteRune)
		rs.addRune(test1ByteRune)

		r, n, i := rs.indexAnyRuneLenInBytes([]byte(test2ByteRuneEnc + "0" + test3ByteRuneEnc + "1" + test4ByteRuneEnc))
		is.Equal(test4ByteRune, r)
		is.Equal(uint8(4), n)
		is.Equal(7, i)

		i = rs.indexAnyInBytes([]byte(test2ByteRuneEnc + "0" + test3ByteRuneEnc + "1" + test4ByteRuneEnc))
		is.Equal(7, i)
	}

	// when there is a wide rune in the set and not in the search string
	// then indexAnyRuneLenInString should return (0, 0, -1)
	{

		var rs runeSet6

		rs.addRune(test4ByteRune)

		r, n, i := rs.indexAnyRuneLenInString(test1ByteRuneEnc)
		is.Equal(rune(0), r)
		is.Equal(uint8(0), n)
		is.Equal(int(-1), i)

		i = rs.indexAnyInString(test1ByteRuneEnc)
		is.Equal(int(-1), i)
	}

	// when there is a mix of wide rune in the set and not in the search string
	// then indexAnyRuneLenInString should not return (0, 0, -1)
	{

		var rs runeSet6

		rs.addRune(test1ByteRune)
		rs.addRune(test4ByteRune)

		r, n, i := rs.indexAnyRuneLenInString(test3ByteRuneEnc + test1ByteRuneEnc)
		is.Equal(test1ByteRune, r)
		is.Equal(uint8(1), n)
		is.Equal(int(3), i)

		i = rs.indexAnyInString(test3ByteRuneEnc + test1ByteRuneEnc)
		is.Equal(int(3), i)

		r, n, i = rs.indexAnyRuneLenInString(test3ByteRuneEnc + test4ByteRuneEnc)
		is.Equal(test4ByteRune, r)
		is.Equal(uint8(4), n)
		is.Equal(int(3), i)

		i = rs.indexAnyInString(test3ByteRuneEnc + test4ByteRuneEnc)
		is.Equal(int(3), i)
	}

	// when there is a mix of wide rune in the set and no overlap in a mixed string
	// then indexAnyRuneLenInBytes should return (0, 0, -1)
	{

		var rs runeSet6

		rs.addRune(test1ByteRune)
		rs.addRune(test4ByteRune)

		r, n, i := rs.indexAnyRuneLenInBytes([]byte("0" + test3ByteRuneEnc))
		is.Equal(rune(0), r)
		is.Equal(uint8(0), n)
		is.Equal(int(-1), i)

		i = rs.indexAnyInBytes([]byte("0" + test3ByteRuneEnc))
		is.Equal(int(-1), i)
	}

	// when there is a mix of wide rune in the set and overlap in a mixed string
	// then indexAnyRuneLenInBytes should not return (0, 0, -1)
	{

		var rs runeSet6

		rs.addRune(test1ByteRune)
		rs.addRune(test4ByteRune)

		r, n, i := rs.indexAnyRuneLenInBytes([]byte("0" + test3ByteRuneEnc + test4ByteRuneEnc))
		is.Equal(test4ByteRune, r)
		is.Equal(uint8(4), n)
		is.Equal(int(4), i)

		i = rs.indexAnyInBytes([]byte("0" + test3ByteRuneEnc + test4ByteRuneEnc))
		is.Equal(int(4), i)
	}

	// when a random wide rune is added
	//
	// containsWideEndByte should be able to find it by last encoded byte form
	// as well as full rune form.
	{
		var buf [utf8.UTFMax]byte
		r := rune(rand.Intn(utf8.MaxRune+1-utf8.RuneSelf-utf16SurrogateCodePointGapSize) + utf8.RuneSelf)
		if r >= utf16SurrogateCodePointStart {
			r += utf16SurrogateCodePointGapSize
		}

		// added via addRuneUniqueUnchecked
		{
			var rs runeSet6
			rs.addRuneUniqueUnchecked(r)
			is.True(rs.containsMBRune(r))

			is.True(rs.internalContainsMBEndByte(byte(r)), "r=%d", uint32(r))

			// last 6 bits of the last byte in an encoded rune are always the
			// last 6 bits of the code point rune value
			is.True(rs.internalContainsMBEndByte(byte(r)&0x3F), "r=%d", uint32(r))

			// but for good measure, test the long way too
			n := utf8.EncodeRune(buf[:], r)
			is.NotEqual(1, n)
			is.True(rs.internalContainsMBEndByte(buf[n-1]), "r=%d", uint32(r))
		}

		// added via addMBRune
		{
			var rs runeSet6
			rs.addMBRune(r)
			is.True(rs.containsMBRune(r))

			is.True(rs.internalContainsMBEndByte(byte(r)), "r=%d", uint32(r))

			// last 6 bits of the last byte in an encoded rune are always the
			// last 6 bits of the code point rune value
			is.True(rs.internalContainsMBEndByte(byte(r)&0x3F), "r=%d", uint32(r))

			// but for good measure, test the long way too
			n := utf8.EncodeRune(buf[:], r)
			is.NotEqual(1, n)
			is.True(rs.internalContainsMBEndByte(buf[n-1]), "r=%d", uint32(r))
		}
	}
}

func Test_runeSet6_containsSingleByteRune(t *testing.T) {
	t.Parallel()

	is := assert.New(t)

	var rs runeSet6

	for i := range utf8.RuneSelf {
		is.False(rs.containsSingleByteRune(byte(i)))
		r, n, x := rs.indexAnyRuneLenInBytes([]byte{byte(i)})
		is.Equal(rune(0), r)
		is.Equal(uint8(0), n)
		is.Equal(int(-1), x)
		r, n, x = rs.indexAnyRuneLenInString(string([]byte{byte(i)}))
		is.Equal(rune(0), r)
		is.Equal(uint8(0), n)
		is.Equal(int(-1), x)
		x = rs.indexAnyInBytes([]byte{byte(i)})
		is.Equal(int(-1), x)
		x = rs.indexAnyInString(string([]byte{byte(i)}))
		is.Equal(int(-1), x)

		rs.addByte(byte(i))

		is.True(rs.containsSingleByteRune(byte(i)))
		r, n, x = rs.indexAnyRuneLenInBytes([]byte{byte(i)})
		is.Equal(rune(i), r)
		is.Equal(uint8(1), n)
		is.Equal(int(0), x)
		r, n, x = rs.indexAnyRuneLenInString(string([]byte{byte(i)}))
		is.Equal(rune(i), r)
		is.Equal(uint8(1), n)
		is.Equal(int(0), x)
		x = rs.indexAnyInBytes([]byte{byte(i)})
		is.Equal(int(0), x)
		x = rs.indexAnyInString(string([]byte{byte(i)}))
		is.Equal(int(0), x)
	}

	for i := utf8.RuneSelf; i < 256; i++ {
		is.False(rs.containsSingleByteRune(byte(i)))
		r, n, x := rs.indexAnyRuneLenInBytes([]byte{byte(i)})
		is.Equal(rune(0), r)
		is.Equal(uint8(0), n)
		is.Equal(int(-1), x)
		r, n, x = rs.indexAnyRuneLenInString(string([]byte{byte(i)}))
		is.Equal(rune(0), r)
		is.Equal(uint8(0), n)
		is.Equal(int(-1), x)
		x = rs.indexAnyInBytes([]byte{byte(i)})
		is.Equal(int(-1), x)
		x = rs.indexAnyInString(string([]byte{byte(i)}))
		is.Equal(int(-1), x)
	}
}

type rsOpKind uint8

const (
	rsk4 rsOpKind = iota
	rsk6
)

type runeSetTestCase struct {
	when string
	then string
	k    rsOpKind
	// s is the set of runes to initialize with
	s string
	// a is the string search argument to use with the target operation type
	a string
	// r is a response rune
	r rune
	// n is a response number of bytes the rune occupies in the search bytes
	n uint8
	// i is the response index of where in the search bytes the rune can be found
	i int
}

func (tc *runeSetTestCase) Run(t *testing.T) {
	t.Helper()

	f := func(t *testing.T, tc *runeSetTestCase) {
		t.Helper()

		f := func(t *testing.T) {
			t.Helper()

			is := assert.New(t)

			var rs4 runeSet4
			var rs6 runeSet6

			// initialize the set
			{
				var addRune func(rune)

				switch tc.k {
				case rsk4:
					addRune = rs4.addRune
				case rsk6:
					addRune = rs6.addRune
				default:
					is.FailNow("unknown operation type", "kind=%d", uint8(tc.k))
				}

				for _, r := range tc.s {
					addRune(r)
				}
			}

			switch tc.k {
			case rsk4:
				r, n, i := rs4.indexAnyRuneLenInString(tc.a)
				is.Equal(tc.r, r)
				is.Equal(tc.n, n)
				is.Equal(tc.i, i)

				i = rs4.indexAnyInString(tc.a)
				is.Equal(tc.i, i)

				r, n, i = rs4.indexAnyRuneLenInBytes([]byte(tc.a))
				is.Equal(tc.r, r)
				is.Equal(tc.n, n)
				is.Equal(tc.i, i)

				i = rs4.indexAnyInBytes([]byte(tc.a))
				is.Equal(tc.i, i)
			case rsk6:
				r, n, i := rs6.indexAnyRuneLenInString(tc.a)
				is.Equal(tc.r, r)
				is.Equal(tc.n, n)
				is.Equal(tc.i, i)

				i = rs6.indexAnyInString(tc.a)
				is.Equal(tc.i, i)

				r, n, i = rs6.indexAnyRuneLenInBytes([]byte(tc.a))
				is.Equal(tc.r, r)
				is.Equal(tc.n, n)
				is.Equal(tc.i, i)

				i = rs6.indexAnyInBytes([]byte(tc.a))
				is.Equal(tc.i, i)
			default:
				is.FailNow("unknown operation type", "kind=%d", uint8(tc.k))
			}
		}

		name := tc.then
		if name == "" {
			name = "then results should match"
		} else {
			name = "then " + name
		}

		if tc.when != "" {
			name = "when " + tc.when + "/" + name
		}

		t.Run(name, f)
	}

	tcClone := func() *runeSetTestCase {
		tc_copy := *tc
		return &tc_copy
	}

	switch tc.k {
	case rsk4:
		ctc := tcClone()
		ctc.when = "rS4 " + ctc.when
		f(t, ctc)

		ctc = tcClone()
		ctc.k = rsk6
		ctc.when = "rS6 " + ctc.when
		f(t, ctc)
	case rsk6:
		ctc := tcClone()
		ctc.when = "rS6 " + ctc.when
		f(t, ctc)

		var mbRuneCount int
		for _, r := range tc.s {
			if r >= utf8.RuneSelf {
				mbRuneCount++
			}
		}
		if mbRuneCount < 5 {
			ctc := tcClone()
			ctc.k = rsk4
			ctc.when = "rS4 " + ctc.when
			f(t, ctc)
		}
	default:
		panic(fmt.Sprintf("kind=%d", uint8(tc.k)))
	}
}

func TestRuneSetIndexNonePaths(t *testing.T) {
	t.Parallel()

	tcs := []runeSetTestCase{
		{
			when: "no runes in the set and empty search arg",
			s:    "",
			a:    "",
			i:    -1,
		},
		{
			when: "one single-byte rune in the set and empty search arg",
			s:    ",",
			a:    "",
			i:    -1,
		},
		{
			when: "one multi-byte rune in the set and empty search arg",
			s:    "\U0001F600",
			a:    "",
			i:    -1,
		},
		{
			when: "single+multi byte rune in the set and empty search arg",
			s:    ",\U0001F600",
			a:    "",
			i:    -1,
		},
		{
			when: "4-byte rune in the set ends in zero byte and arg contains 0xF08F8080",
			s:    "\U0001F600",
			a:    "\xF0\x8F\x80\x80",
			i:    -1,
		},
		{
			when: "4-byte rune in the set ends in zero byte and arg contains 0xF4908080",
			s:    "\U0001F600",
			a:    "\xF4\x90\x80\x80",
			i:    -1,
		},
		{
			when: "3-byte rune in the set ends in zero byte and arg contains 0xE09F80",
			s:    "\u0800",
			a:    "\xE0\x9F\x80",
			i:    -1,
		},
		{
			when: "3-byte rune in the set ends in zero byte and arg contains 0xEDA080",
			s:    "\u0800",
			a:    "\xED\xA0\x80",
			i:    -1,
		},
		{
			when: "2-byte rune in the set ends in zero byte and arg contains 0xC080",
			s:    "\xC2\x80",
			a:    "\xC0\x80",
			i:    -1,
		},
		{
			when: "2-byte rune in the set ends in zero byte and arg contains 0x8080808080",
			s:    "\U0001F600",
			a:    "\x80\x80\x80\x80\x80",
			i:    -1,
		},
		{
			when: "2-byte rune in the set 0xC280 ends in zero byte and arg contains 0xC081",
			s:    "\xC2\x80",
			a:    "\xC0\x81",
			i:    -1,
		},
		{
			when: "2-byte rune in the set 0xC380 ends in zero byte and arg contains 0xC280",
			s:    "\xC3\x80",
			a:    "\xC2\x80",
			i:    -1,
		},
		{
			when: "3-byte rune in the set 0xE18080 ends in zero byte and arg contains 0xE18180",
			s:    "\xE1\x80\x80",
			a:    "\xE1\x81\x80",
			i:    -1,
		},
		{
			when: "4-byte rune in the set 0xF1808080 ends in zero byte and arg contains 0xF1818080",
			s:    "\xF1\x80\x80\x80",
			a:    "\xF1\x81\x80\x80",
			i:    -1,
		},
		{
			when: "4-byte rune in the set 0xF1808080 ends in zero byte and arg contains 0xF8808080",
			s:    "\xF1\x80\x80\x80",
			a:    "\xF8\x80\x80\x80",
			i:    -1,
		},
		{
			when: "4-byte rune in the set and arg contains a single ASCII rune",
			s:    "\U0001F600",
			a:    ",",
			i:    -1,
		},
	}

	for _, tc := range tcs {
		if tc.then == "" {
			tc.then = "indexAny returns -1"
		}
		tc.Run(t)
	}
}
