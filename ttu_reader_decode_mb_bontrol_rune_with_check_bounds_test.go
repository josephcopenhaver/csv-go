//go:build check_bounds

package csv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitReaderHelperPaths_decodeMBControlRuneWithCheckBounds(t *testing.T) {

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
