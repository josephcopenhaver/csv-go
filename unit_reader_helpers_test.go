package csv

import (
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestUnitReaderHelperPaths_isByteOrderMarker(t *testing.T) {

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

func TestUnitReaderHelperPaths_decodeMBControlRune(t *testing.T) {

	info, ok := debug.ReadBuildInfo()
	if !ok {
		t.Error("failed to determine build info")
		t.FailNow()
	}

	var checkBounds bool
	for _, setting := range info.Settings {
		if setting.Key == "-tags" {
			for _, tag := range strings.Split(setting.Value, ",") {
				if v := strings.ToLower(strings.TrimSpace(tag)); v == "check_bounds" {
					checkBounds = true
					break
				}
			}
			break
		}
	}

	if !checkBounds {
		testUnitReaderHelperPaths_decodeMBControlRune(t)
		return
	}

	testUnitReaderHelperPaths_decodeMBControlRuneWithCheckBounds(t)
}

func testUnitReaderHelperPaths_decodeMBControlRune(t *testing.T) {

	type Resp struct {
		r rune
		s uint8
	}

	type TC struct {
		n    string
		in   []byte
		resp Resp
	}
	tcs := []TC{
		{"0x0", []byte{0x0}, Resp{0, 1}},
		{"a", []byte("a"), Resp{'a', 1}},
		{"0x7F", []byte{0x7F}, Resp{0x7F, 1}},
		{"0xC2A2 - Cent Sign", []byte{0xC2, 0xA2}, Resp{([]rune(string([]byte{0xC2, 0xA2})))[0], 2}},
		{"0xE282AC - Euro sign", []byte{0xE2, 0x82, 0xAC}, Resp{([]rune(string([]byte{0xE2, 0x82, 0xAC})))[0], 3}},
		{"0xF09FA4AF - Mind Blown emoji", []byte{0xF0, 0x9F, 0xA4, 0xAF}, Resp{([]rune(string([]byte{0xF0, 0x9F, 0xA4, 0xAF})))[0], 4}},
		// more than one rune worth of bytes
		{"0x0 - followed by a", []byte{0x0, 0x10}, Resp{0, 1}},
		{"a - followed by a", []byte("aa"), Resp{'a', 1}},
		{"0x7F - followed by a", []byte{0x7F, 0x10}, Resp{0x7F, 1}},
		{"0xC2A210 - Cent Sign followed by a", []byte{0xC2, 0xA2, 0x10}, Resp{([]rune(string([]byte{0xC2, 0xA2})))[0], 2}},
		{"0xE282AC10 - Euro sign followed by a", []byte{0xE2, 0x82, 0xAC, 0x10}, Resp{([]rune(string([]byte{0xE2, 0x82, 0xAC})))[0], 3}},
		{"0xF09FA4AF10 - Mind Blown emoji followed by a", []byte{0xF0, 0x9F, 0xA4, 0xAF, 0x10}, Resp{([]rune(string([]byte{0xF0, 0x9F, 0xA4, 0xAF})))[0], 4}},
		// incomplete number of bytes
		{"0xC2 - Cent Sign missing last byte", []byte{0xC2}, Resp{utf8.RuneError, 1}},
		{"0xE282 - Euro sign missing last byte", []byte{0xE2, 0x82}, Resp{utf8.RuneError, 1}},
		{"0xF09FA4 - Mind Blown emoji missing last byte", []byte{0xF0, 0x9F, 0xA4}, Resp{utf8.RuneError, 1}},
		// empty
		{"nil", nil, Resp{utf8.RuneError, 0}},
		{"empty slice", []byte{}, Resp{utf8.RuneError, 0}},
	}

	for _, tc := range tcs {
		t.Run("given the input pattern "+tc.n, func(t *testing.T) {
			is := assert.New(t)

			r, s := decodeMBControlRune(tc.in)
			is.Equal(tc.resp.r, r)
			is.Equal(tc.resp.s, s)
		})
	}
}

func testUnitReaderHelperPaths_decodeMBControlRuneWithCheckBounds(t *testing.T) {

	type Resp struct {
		r rune
		s uint8
		p bool
	}

	type TC struct {
		n    string
		in   []byte
		resp Resp
	}
	tcs := []TC{
		{"0x0", []byte{0x0}, Resp{p: true}},
		{"a", []byte("a"), Resp{p: true}},
		{"0x7F", []byte{0x7F}, Resp{p: true}},
		{"0xC2A2 - Cent Sign", []byte{0xC2, 0xA2}, Resp{r: ([]rune(string([]byte{0xC2, 0xA2})))[0], s: 2}},
		{"0xE282AC - Euro sign", []byte{0xE2, 0x82, 0xAC}, Resp{r: ([]rune(string([]byte{0xE2, 0x82, 0xAC})))[0], s: 3}},
		{"0xF09FA4AF - Mind Blown emoji", []byte{0xF0, 0x9F, 0xA4, 0xAF}, Resp{r: ([]rune(string([]byte{0xF0, 0x9F, 0xA4, 0xAF})))[0], s: 4}},
		// more than one rune worth of bytes
		{"0x0 - followed by a", []byte{0x0, 0x10}, Resp{p: true}},
		{"a - followed by a", []byte("aa"), Resp{p: true}},
		{"0x7F - followed by a", []byte{0x7F, 0x10}, Resp{p: true}},
		{"0xC2A210 - Cent Sign followed by a", []byte{0xC2, 0xA2, 0x10}, Resp{r: ([]rune(string([]byte{0xC2, 0xA2})))[0], s: 2}},
		{"0xE282AC10 - Euro sign followed by a", []byte{0xE2, 0x82, 0xAC, 0x10}, Resp{r: ([]rune(string([]byte{0xE2, 0x82, 0xAC})))[0], s: 3}},
		{"0xF09FA4AF10 - Mind Blown emoji followed by a", []byte{0xF0, 0x9F, 0xA4, 0xAF, 0x10}, Resp{r: ([]rune(string([]byte{0xF0, 0x9F, 0xA4, 0xAF})))[0], s: 4}},
		// incomplete number of bytes
		{"0xC2 - Cent Sign missing last byte", []byte{0xC2}, Resp{p: true}},
		{"0xE282 - Euro sign missing last byte", []byte{0xE2, 0x82}, Resp{p: true}},
		{"0xF09FA4 - Mind Blown emoji missing last byte", []byte{0xF0, 0x9F, 0xA4}, Resp{p: true}},
		// empty
		{"nil", nil, Resp{p: true}},
		{"empty slice", []byte{}, Resp{p: true}},
	}

	run := func(p []byte) (_ rune, _ uint8, _r any) {
		defer func() {
			if r := recover(); r != nil {
				_r = r
			}
		}()

		r, s := decodeMBControlRune(p)
		return r, s, nil
	}

	for _, tc := range tcs {
		t.Run("given the input pattern "+tc.n, func(t *testing.T) {
			is := assert.New(t)

			r, s, p := run(tc.in)
			if tc.resp.p {
				//
				// should have panicked
				//
				is.NotNil(p)
				is.Equal("decode rune failed", p)
				is.Equal(rune(0), r)
				is.Equal(rune(0), tc.resp.r)
				is.Equal(uint8(0), s)
				is.Equal(uint8(0), tc.resp.s)
				return
			}

			is.Nil(p)
			is.Equal(tc.resp.r, r)
			is.Equal(tc.resp.s, s)
		})
	}
}
