package csv_test

import (
	"io"
	"strings"
	"testing"

	"github.com/josephcopenhaver/csv-go/v3"
)

func TestFunctionalReaderScanErrorFieldIndex(t *testing.T) {
	t.Parallel()

	tcs := []functionalReaderTestCase{
		{
			when: "4 column header and no rows with RemoveHeaderRow=true ErrorOnNoRows=true",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`a,b,c,d` + "\n")),
					csv.ReaderOpts().RemoveHeaderRow(true),
					csv.ReaderOpts().ErrorOnNoRows(true),
				}
			},
			iterErrIs: []error{
				csv.ErrParsing,
				csv.ErrNoRows,
				io.ErrUnexpectedEOF,
			},
			iterErrStr: csv.ErrParsing.Error() + " at byte 8, record 2, field 1: " + csv.ErrNoRows.Error(),
		},
		{
			when: "no header row or data row with ExpectHeaders=4-cols ErrorOnNoRows=true",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(``)),
					csv.ReaderOpts().ExpectHeaders(strings.Split(`a,b,c,d`, ",")...),
					csv.ReaderOpts().ErrorOnNoRows(true),
				}
			},
			iterErrIs: []error{
				csv.ErrParsing,
				csv.ErrNoHeaderRow,
				io.ErrUnexpectedEOF,
			},
			iterErrStr: csv.ErrParsing.Error() + " at byte 0, record 0, field 0: " + csv.ErrNoHeaderRow.Error(),
		},
		{
			when: "4 column header row differs from expected",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`a,b,c,d`)),
					csv.ReaderOpts().ExpectHeaders(strings.Split(`a,b,c,e`, ",")...),
				}
			},
			iterErrIs: []error{
				csv.ErrParsing,
				csv.ErrUnexpectedHeaderRowContents,
			},
			iterErrStr: csv.ErrParsing.Error() + " at byte 7, record 1, field 4: " + csv.ErrUnexpectedHeaderRowContents.Error(),
		},
	}

	for _, tc := range tcs {
		if tc.then == "" {
			tc.then = "error returned has the correct record and field values"
		}
		tc.Run(t)
	}
}
