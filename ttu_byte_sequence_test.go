package csv

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

const (
	// tilde
	test1ByteRune = '~'
	// cent sign
	test2ByteRune = '\u00A2'
	// euro sign
	test3ByteRune = '\u20AC'
	// grinning face
	test4ByteRune = '\U0001F600'

	test1ByteRuneEnc = "\x7E"
	test2ByteRuneEnc = "\xC2\xA2"
	test3ByteRuneEnc = "\xE2\x82\xAC"
	test4ByteRuneEnc = "\xF0\x9F\x98\x80"
)

func Test_byteSequenceShort_appendText(t *testing.T) {
	is := assert.New(t)

	is.PanicsWithValue(panicInvalidByteSequenceLength, func() {
		var bs byteSequenceShort
		bs.appendText(nil)
	})

	is.PanicsWithValue(panicInvalidByteSequenceLength, func() {
		var bs byteSequenceLong
		bs.appendText(nil)
	})

	is.PanicsWithValue(panicInvalidByteSequenceLength, func() {
		var bs byteSequenceLong
		bs.n = 1
		bs.appendText(nil)
	})

	var bufArr [2 * utf8.UTFMax]byte
	buf := bufArr[:0]

	{
		bs := newSeq(test4ByteRune)

		resp := bs.appendText(buf)

		is.Same(&resp[0], &bufArr[0])
		is.Equal(4, len(resp))
		is.Equal(test4ByteRuneEnc, string(resp))
	}

	{
		bs := newSeq2(test1ByteRune, test1ByteRune)
		resp := bs.appendText(buf)

		is.Same(&resp[0], &bufArr[0])
		is.Equal(2, len(resp))
		is.Equal(test1ByteRuneEnc+test1ByteRuneEnc, string(resp))
	}

	{
		bs := newSeq2(test1ByteRune, test2ByteRune)
		resp := bs.appendText(buf)

		is.Same(&resp[0], &bufArr[0])
		is.Equal(3, len(resp))
		is.Equal(test1ByteRuneEnc+test2ByteRuneEnc, string(resp))
	}

	{
		bs := newSeq2(test1ByteRune, test3ByteRune)
		resp := bs.appendText(buf)

		is.Same(&resp[0], &bufArr[0])
		is.Equal(4, len(resp))
		is.Equal(test1ByteRuneEnc+test3ByteRuneEnc, string(resp))
	}

	{
		bs := newSeq2(test1ByteRune, test4ByteRune)
		resp := bs.appendText(buf)

		is.Same(&resp[0], &bufArr[0])
		is.Equal(5, len(resp))
		is.Equal(test1ByteRuneEnc+test4ByteRuneEnc, string(resp))
	}

	{
		bs := newSeq2(test4ByteRune, test1ByteRune)
		resp := bs.appendText(buf)

		is.Same(&resp[0], &bufArr[0])
		is.Equal(5, len(resp))
		is.Equal(test4ByteRuneEnc+test1ByteRuneEnc, string(resp))
	}

	{
		bs := newSeq2(test4ByteRune, test2ByteRune)
		resp := bs.appendText(buf)

		is.Same(&resp[0], &bufArr[0])
		is.Equal(6, len(resp))
		is.Equal(test4ByteRuneEnc+test2ByteRuneEnc, string(resp))
	}

	{
		bs := newSeq2(test4ByteRune, test3ByteRune)
		resp := bs.appendText(buf)

		is.Same(&resp[0], &bufArr[0])
		is.Equal(7, len(resp))
		is.Equal(test4ByteRuneEnc+test3ByteRuneEnc, string(resp))
	}

	{
		bs := newSeq2(test4ByteRune, test4ByteRune)
		resp := bs.appendText(buf)

		is.Same(&resp[0], &bufArr[0])
		is.Equal(8, len(resp))
		is.Equal(test4ByteRuneEnc+test4ByteRuneEnc, string(resp))
	}
}

func Test_runeScape4_containsWideRune(t *testing.T) {
	is := assert.New(t)

	{
		var rs runeScape4

		rs.addRune(test4ByteRune)
		rs.addRune(test4ByteRune + 1)
		rs.addRune(test4ByteRune + 2)
		rs.addRune(test4ByteRune + 3)

		is.False(rs.containsRune(test4ByteRune + 4))
		is.True(rs.containsRune(test4ByteRune + 3))
		is.True(rs.containsRune(test4ByteRune + 2))
		is.True(rs.containsRune(test4ByteRune + 1))
		is.True(rs.containsRune(test4ByteRune))
	}

	// adding the same wide rune twice should not change rune count
	{
		var rs runeScape4

		rs.addRune(test4ByteRune)
		is.Equal(uint8(1), rs.numWideRunes)

		rs.addRune(test4ByteRune)
		is.Equal(uint8(1), rs.numWideRunes)
		is.False(rs.containsRune(test1ByteRune))
	}

	// when there is a wide rune in the set but not in the search string
	// then -1 should be returned from indexAnyInString
	{
		var rs runeScape4

		rs.addRune(test4ByteRune)

		is.Equal(-1, rs.indexAnyInString(test1ByteRuneEnc))
	}

	// when there is a wide rune in the set and in the search string
	// then -1 should not be returned from indexAnyInString
	{
		var rs runeScape4

		rs.addRune(test4ByteRune)

		is.Equal(0, rs.indexAnyInString(test4ByteRuneEnc))
	}

	// when there is a mix of wide rune in the set and in the search string and hit is an ascii rune
	// then -1 should not be returned from indexAnyInString
	{
		var rs runeScape4

		rs.addRune(test4ByteRune)
		rs.addRune(test1ByteRune)

		is.Equal(7, rs.indexAnyInString(test2ByteRuneEnc+"0"+test3ByteRuneEnc+"1"+test1ByteRuneEnc))
	}

	// when there is a mix of wide rune in the set and in the search string and hit is an ascii rune
	// then -1 should not be returned from indexAnyInBytes
	{
		var rs runeScape4

		rs.addRune(test4ByteRune)
		rs.addRune(test1ByteRune)

		is.Equal(7, rs.indexAnyInBytes([]byte(test2ByteRuneEnc+"0"+test3ByteRuneEnc+"1"+test1ByteRuneEnc)))
	}

	// when there is a mix of wide rune in the set and in the search string and hit is wide rune
	// then -1 should not be returned from indexAnyInString
	{
		var rs runeScape4

		rs.addRune(test4ByteRune)
		rs.addRune(test1ByteRune)

		is.Equal(7, rs.indexAnyInString(test2ByteRuneEnc+"0"+test3ByteRuneEnc+"1"+test4ByteRuneEnc))
	}

	// when there is a mix of wide rune in the set and in the search string and hit is wide rune
	// then -1 should not be returned from indexAnyInBytes
	{
		var rs runeScape4

		rs.addRune(test4ByteRune)
		rs.addRune(test1ByteRune)

		is.Equal(7, rs.indexAnyInBytes([]byte(test2ByteRuneEnc+"0"+test3ByteRuneEnc+"1"+test4ByteRuneEnc)))
	}

	// when there is a mix of wide rune in the set and in the search string and hit is wide rune
	// then (0, 0, -1) should not be returned from indexAnyRuneLenInString
	{
		var rs runeScape4

		rs.addRune(test4ByteRune)
		rs.addRune(test1ByteRune)

		r, n, i := rs.indexAnyRuneLenInString(test2ByteRuneEnc + "0" + test3ByteRuneEnc + "1" + test1ByteRuneEnc)
		is.Equal(test1ByteRune, r)
		is.Equal(uint8(1), n)
		is.Equal(7, i)
	}

	// when there is a mix of wide rune in the set and in the search string and hit is wide rune
	// then (0, 0, -1) should not be returned from indexAnyRuneLenInBytes
	{

		var rs runeScape4

		rs.addRune(test4ByteRune)
		rs.addRune(test1ByteRune)

		r, n, i := rs.indexAnyRuneLenInBytes([]byte(test2ByteRuneEnc + "0" + test3ByteRuneEnc + "1" + test1ByteRuneEnc))
		is.Equal(test1ByteRune, r)
		is.Equal(uint8(1), n)
		is.Equal(7, i)
	}

	// when there is a mix of wide rune in the set and in the search string and hit is wide rune
	// then (0, 0, -1) should not be returned from indexAnyRuneLenInString
	{
		var rs runeScape4

		rs.addRune(test4ByteRune)
		rs.addRune(test1ByteRune)

		r, n, i := rs.indexAnyRuneLenInString(test2ByteRuneEnc + "0" + test3ByteRuneEnc + "1" + test4ByteRuneEnc)
		is.Equal(test4ByteRune, r)
		is.Equal(uint8(4), n)
		is.Equal(7, i)
	}

	// when there is a mix of wide rune in the set and in the search string and hit is wide rune
	// then (0, 0, -1) should not be returned from indexAnyRuneLenInBytes
	{

		var rs runeScape4

		rs.addRune(test4ByteRune)
		rs.addRune(test1ByteRune)

		r, n, i := rs.indexAnyRuneLenInBytes([]byte(test2ByteRuneEnc + "0" + test3ByteRuneEnc + "1" + test4ByteRuneEnc))
		is.Equal(test4ByteRune, r)
		is.Equal(uint8(4), n)
		is.Equal(7, i)
	}

	// when there is a wide rune in the set and not in the search string
	// then indexAnyRuneLenInString should return (0, 0, -1)
	{

		var rs runeScape4

		rs.addRune(test4ByteRune)

		r, n, i := rs.indexAnyRuneLenInString(test1ByteRuneEnc)
		is.Equal(rune(0), r)
		is.Equal(uint8(0), n)
		is.Equal(int(-1), i)
	}

	// when there is a mix of wide rune in the set and not in the search string
	// then indexAnyRuneLenInString should not return (0, 0, -1)
	{

		var rs runeScape4

		rs.addRune(test1ByteRune)
		rs.addRune(test4ByteRune)

		r, n, i := rs.indexAnyRuneLenInString(test3ByteRuneEnc + test1ByteRuneEnc)
		is.Equal(test1ByteRune, r)
		is.Equal(uint8(1), n)
		is.Equal(int(3), i)

		r, n, i = rs.indexAnyRuneLenInString(test3ByteRuneEnc + test4ByteRuneEnc)
		is.Equal(test4ByteRune, r)
		is.Equal(uint8(4), n)
		is.Equal(int(3), i)
	}

	// when there is a mix of wide rune in the set and no overlap in a mixed string
	// then indexAnyInBytes should return -1
	{

		var rs runeScape4

		rs.addRune(test1ByteRune)
		rs.addRune(test4ByteRune)

		i := rs.indexAnyInBytes([]byte("0" + test3ByteRuneEnc))
		is.Equal(int(-1), i)
	}

	// when there is a mix of wide rune in the set and overlap in a mixed string
	// then indexAnyInBytes should not return -1
	{

		var rs runeScape4

		rs.addRune(test1ByteRune)
		rs.addRune(test4ByteRune)

		i := rs.indexAnyInBytes([]byte("0" + test3ByteRuneEnc + test4ByteRuneEnc))
		is.Equal(int(4), i)
	}

	// when there is a mix of wide rune in the set and no overlap in a mixed string
	// then indexAnyRuneLenInBytes should return (0, 0, -1)
	{

		var rs runeScape4

		rs.addRune(test1ByteRune)
		rs.addRune(test4ByteRune)

		r, n, i := rs.indexAnyRuneLenInBytes([]byte("0" + test3ByteRuneEnc))
		is.Equal(rune(0), r)
		is.Equal(uint8(0), n)
		is.Equal(int(-1), i)
	}

	// when there is a mix of wide rune in the set and overlap in a mixed string
	// then indexAnyRuneLenInBytes should not return (0, 0, -1)
	{

		var rs runeScape4

		rs.addRune(test1ByteRune)
		rs.addRune(test4ByteRune)

		r, n, i := rs.indexAnyRuneLenInBytes([]byte("0" + test3ByteRuneEnc + test4ByteRuneEnc))
		is.Equal(test4ByteRune, r)
		is.Equal(uint8(4), n)
		is.Equal(int(4), i)
	}
}

func Test_runeScape6_containsWideRune(t *testing.T) {
	is := assert.New(t)

	{
		var rs runeScape6

		rs.addRune(test4ByteRune)
		rs.addRune(test4ByteRune + 1)
		rs.addRune(test4ByteRune + 2)
		rs.addRune(test4ByteRune + 3)
		rs.addRune(test4ByteRune + 4)
		rs.addRune(test4ByteRune + 5)

		is.False(rs.containsRune(test4ByteRune + 6))
		is.True(rs.containsRune(test4ByteRune + 5))
		is.True(rs.containsRune(test4ByteRune + 4))
		is.True(rs.containsRune(test4ByteRune + 3))
		is.True(rs.containsRune(test4ByteRune + 2))
		is.True(rs.containsRune(test4ByteRune + 1))
		is.True(rs.containsRune(test4ByteRune))
	}

	// adding the same wide rune twice to a runeScape6 should not change rune count
	{
		var rs runeScape6

		rs.addRune(test4ByteRune)
		is.Equal(uint8(1), rs.numWideRunes)

		rs.addRune(test4ByteRune)
		is.Equal(uint8(1), rs.numWideRunes)
		is.False(rs.containsRune(test1ByteRune))
	}

	// when there is a mix of wide rune in the set and in the search string and hit is an ascii rune
	// then -1 should not be returned from indexAnyInBytes
	{
		var rs runeScape6

		rs.addRune(test4ByteRune)
		rs.addRune(test1ByteRune)

		is.Equal(7, rs.indexAnyInBytes([]byte(test2ByteRuneEnc+"0"+test3ByteRuneEnc+"1"+test1ByteRuneEnc)))
	}

	// when there is a mix of wide rune in the set and in the search string and hit is wide rune
	// then -1 should not be returned from indexAnyInBytes
	{
		var rs runeScape6

		rs.addRune(test4ByteRune)
		rs.addRune(test1ByteRune)

		is.Equal(7, rs.indexAnyInBytes([]byte(test2ByteRuneEnc+"0"+test3ByteRuneEnc+"1"+test4ByteRuneEnc)))
	}

	// when there is a mix of wide rune in the set and no overlap in a mixed string
	// then indexAnyInBytes should return -1
	{

		var rs runeScape6

		rs.addRune(test1ByteRune)
		rs.addRune(test4ByteRune)

		i := rs.indexAnyInBytes([]byte("0" + test3ByteRuneEnc))
		is.Equal(int(-1), i)
	}

	// when there is a mix of wide rune in the set and overlap in a mixed string
	// then indexAnyInBytes should not return -1
	{

		var rs runeScape6

		rs.addRune(test1ByteRune)
		rs.addRune(test4ByteRune)

		i := rs.indexAnyInBytes([]byte("0" + test3ByteRuneEnc + test4ByteRuneEnc))
		is.Equal(int(4), i)
	}
}
