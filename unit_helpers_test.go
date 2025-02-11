package csv

import (
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit CSV Reader Helper Paths: isByteOrderMarker", func() {

	type TC struct {
		n      string
		r      uint32
		s      int
		result bool
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
		Context("given a BOM of "+tc.n+" and size "+strconv.Itoa(tc.s), func() {
			It("should be recognized as a Byte Order Marker", func() {
				resp := isByteOrderMarker(tc.r, tc.s)
				Expect(resp).To(Equal(tc.result))
			})
		})
	}
})

var _ = Describe("Unit CSV Reader Helper Paths: isNewlineRune", func() {

	type result struct {
		isCR bool
		ok   bool
	}

	type TC struct {
		n string
		c rune
		result
	}
	tcs := []TC{
		{"ascii carriage return", asciiCarriageReturn, result{true, true}},
		{"ascii line feed", asciiLineFeed, result{false, true}},
		{"ascii vertical tab", asciiVerticalTab, result{false, true}},
		{"ascii form feed", asciiFormFeed, result{false, true}},
		{"utf8 next line", utf8NextLine, result{false, true}},
		{"ut8 line separator", utf8LineSeparator, result{false, true}},
		{"comma", ',', result{false, false}},
	}

	for _, tc := range tcs {
		Context("given the character "+tc.n, func() {
			It("should return ()"+strconv.FormatBool(tc.result.isCR)+","+strconv.FormatBool(tc.result.ok)+")", func() {
				isCR, ok := isNewlineRune(tc.c)
				Expect(result{isCR, ok}).To(Equal(tc.result))
			})
		})
	}
})
