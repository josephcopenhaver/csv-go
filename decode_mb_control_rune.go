//go:build !check_bounds

package csv

import (
	"unicode/utf8"
)

func decodeMBControlRune(p []byte) (rune, uint8) {
	r, s := utf8.DecodeRune(p)

	return r, uint8(s)
}
