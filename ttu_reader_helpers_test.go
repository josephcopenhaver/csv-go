package csv

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitReaderHelperPaths_isByteOrderMarker(t *testing.T) {
	t.Parallel()

	type TC struct {
		n    string
		r    uint32
		s    int
		resp bool
	}
	tcs := []TC{
		{"utf8", utf8ByteOrderMarker, 3, true},
		{"utf16 LE 2 byte", utf16ByteOrderMarkerLE, 2, true},
		{"utf16 LE 4 byte (utf32 LE)", (utf16ByteOrderMarkerLE << 16), 4, true},
		{"utf16 BE 2 byte", utf16ByteOrderMarkerBE, 2, true},
		{"utf16 BE 3 byte", utf16ByteOrderMarkerBE, 3, true},
		{"utf16 BE 4 byte", utf16ByteOrderMarkerBE, 4, true},
		{"comma", ',', 1, false},
	}

	for _, tc := range tcs {
		t.Run("given a BOM of "+tc.n+" and size "+strconv.Itoa(tc.s), func(t *testing.T) {
			resp := isByteOrderMarker(tc.r, tc.s)
			assert.Equal(t, tc.resp, resp, "should be recognized as a Byte Order Marker")
		})
	}
}

func TestUnitReaderHelperPaths_isNewlineRune(t *testing.T) {
	t.Parallel()

	type Resp struct {
		isCR bool
		ok   bool
	}

	type TC struct {
		n    string
		c    rune
		resp Resp
	}
	tcs := []TC{
		{"ascii carriage return", asciiCarriageReturn, Resp{true, true}},
		{"ascii line feed", asciiLineFeed, Resp{false, true}},
		{"ascii vertical tab", asciiVerticalTab, Resp{false, true}},
		{"ascii form feed", asciiFormFeed, Resp{false, true}},
		{"utf8 next line", utf8NextLine, Resp{false, true}},
		{"ut8 line separator", utf8LineSeparator, Resp{false, true}},
		{"comma", ',', Resp{false, false}},
	}

	for _, tc := range tcs {
		t.Run("given the character "+tc.n, func(t *testing.T) {
			isCR, ok := isNewlineRune(tc.c)
			assert.Equal(t, tc.resp, Resp{isCR, ok}, "should return ("+strconv.FormatBool(tc.resp.isCR)+","+strconv.FormatBool(tc.resp.ok)+")")
		})
	}
}
