package csv

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitReaderRecordSeparatorErrorPaths(t *testing.T) {

	type TC struct {
		n   string
		arg string
	}
	tcs := []TC{
		{"empty string", ""},
		{"non-utf8 rune", string([]byte{0xFE, 0xFF})},
		{"\\r and non-utf8 rune", "\r" + string([]byte{0xFE, 0xFF})},
		{"\\r,", "\r,"},
	}

	for _, tc := range tcs {
		t.Run("with record separator of "+tc.n, func(t *testing.T) {
			// should return an option that sets record sep length to invalid (-1)
			resp := ReaderOpts().RecordSeparator(tc.arg)
			assert.Equal(t, fmt.Sprintf("%v", ReaderOption(badRecordSeparatorRConfig)), fmt.Sprintf("%v", resp))

			var cfg rCfg
			resp(&cfg)
			assert.Equal(t, int8(-1), cfg.recordSepLen)
		})
	}
}
