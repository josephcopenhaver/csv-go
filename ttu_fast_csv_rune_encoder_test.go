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

func Test_runeEncoder_appendText(t *testing.T) {
	is := assert.New(t)

	is.PanicsWithValue(panicInvalidRuneEncoderLen, func() {
		var rw runeEncoder
		rw.appendText(nil)
	})

	is.PanicsWithValue(panicInvalidRuneEncoderLen, func() {
		var rw twoRuneEncoder
		rw.appendText(nil)
	})

	is.PanicsWithValue(panicInvalidRuneEncoderLen, func() {
		var rw twoRuneEncoder
		rw.n = 1
		rw.appendText(nil)
	})

	var bufArr [2 * utf8.UTFMax]byte
	buf := bufArr[:0]

	{
		rw := newRuneEncoder(test4ByteRune)

		resp := rw.appendText(buf)

		is.Same(&resp[0], &bufArr[0])
		is.Equal(4, len(resp))
		is.Equal(test4ByteRuneEnc, string(resp))
	}

	{
		rw := newTwoRuneEncoder(test1ByteRune, test1ByteRune)
		resp := rw.appendText(buf)

		is.Same(&resp[0], &bufArr[0])
		is.Equal(2, len(resp))
		is.Equal(test1ByteRuneEnc+test1ByteRuneEnc, string(resp))
	}

	{
		rw := newTwoRuneEncoder(test1ByteRune, test2ByteRune)
		resp := rw.appendText(buf)

		is.Same(&resp[0], &bufArr[0])
		is.Equal(3, len(resp))
		is.Equal(test1ByteRuneEnc+test2ByteRuneEnc, string(resp))
	}

	{
		rw := newTwoRuneEncoder(test1ByteRune, test3ByteRune)
		resp := rw.appendText(buf)

		is.Same(&resp[0], &bufArr[0])
		is.Equal(4, len(resp))
		is.Equal(test1ByteRuneEnc+test3ByteRuneEnc, string(resp))
	}

	{
		rw := newTwoRuneEncoder(test1ByteRune, test4ByteRune)
		resp := rw.appendText(buf)

		is.Same(&resp[0], &bufArr[0])
		is.Equal(5, len(resp))
		is.Equal(test1ByteRuneEnc+test4ByteRuneEnc, string(resp))
	}

	{
		rw := newTwoRuneEncoder(test4ByteRune, test1ByteRune)
		resp := rw.appendText(buf)

		is.Same(&resp[0], &bufArr[0])
		is.Equal(5, len(resp))
		is.Equal(test4ByteRuneEnc+test1ByteRuneEnc, string(resp))
	}

	{
		rw := newTwoRuneEncoder(test4ByteRune, test2ByteRune)
		resp := rw.appendText(buf)

		is.Same(&resp[0], &bufArr[0])
		is.Equal(6, len(resp))
		is.Equal(test4ByteRuneEnc+test2ByteRuneEnc, string(resp))
	}

	{
		rw := newTwoRuneEncoder(test4ByteRune, test3ByteRune)
		resp := rw.appendText(buf)

		is.Same(&resp[0], &bufArr[0])
		is.Equal(7, len(resp))
		is.Equal(test4ByteRuneEnc+test3ByteRuneEnc, string(resp))
	}

	{
		rw := newTwoRuneEncoder(test4ByteRune, test4ByteRune)
		resp := rw.appendText(buf)

		is.Same(&resp[0], &bufArr[0])
		is.Equal(8, len(resp))
		is.Equal(test4ByteRuneEnc+test4ByteRuneEnc, string(resp))
	}
}
