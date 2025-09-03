//go:build !check_bounds

package csv

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestUnitReaderHelperPaths_decodeMBControlRune(t *testing.T) {
	t.Parallel()

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
