package csv

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit CSV Reader RecordSeparator error paths", func() {

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
		Context("given an input of record separator ", func() {
			It("should return an option that sets record sep length to invalid (-1)", func() {
				resp := ReaderOpts().RecordSeparator(tc.arg)
				Expect(resp).ToNot(Equal(ReaderOption(badRecordSeparatorRConfig)))
				Expect(fmt.Sprintf("%v", resp)).To(Equal(fmt.Sprintf("%v", ReaderOption(badRecordSeparatorRConfig))))

				var cfg rCfg
				resp(&cfg)
				Expect(cfg.recordSepLen).To(Equal(int8(-1)))
			})
		})
	}
})
