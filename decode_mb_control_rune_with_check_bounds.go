//go:build check_bounds

package csv

import (
	"unicode/utf8"
)

func decodeMBControlRune(p []byte) (rune, uint8) {
	r, s := utf8.DecodeRune(p)
	if s <= 1 {
		panic("decode rune failed") // so must have a bad control rune loaded via config options or memory was corrupted somehow
	}

	return r, uint8(s)
}
