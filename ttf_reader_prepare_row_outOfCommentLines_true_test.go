package csv_test

import (
	"testing"

	"github.com/josephcopenhaver/csv-go/v2"
)

func TestFunctionalReaderPrepareRowOutOfCommentLinesRespTruePaths(t *testing.T) {
	t.Parallel()

	tcs := []functionalReaderTestCase{
		{
			when: "MaxComments=1 RecordSep=CRLF first line starts with comment followed by 4 chars with CR and a data byte and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`#2345`+"\r7", ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r\n"),
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().MaxComments(1),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpCommentsAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 7, record 0, field 0: " + csv.ErrSecOpCommentsAboveMax.Error(),
		},
		{
			when: "MaxComments=1 with first comment line being 5 chars with LF and comment char and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("#2345\n#", ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().MaxComments(1),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpCommentsAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 7, record 1, field 1: " + csv.ErrSecOpCommentsAboveMax.Error(), // TODO: record and field indicator are wrong here
		},
	}

	for _, tc := range tcs {
		if tc.then == "" {
			tc.then = "coupled error"
		}
		tc.Run(t)
	}
}
