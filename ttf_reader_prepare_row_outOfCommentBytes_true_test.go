package csv_test

import (
	"testing"

	"github.com/josephcopenhaver/csv-go/v2"
)

func TestFunctionalReaderPrepareRowOutOfCommentBytesRespTruePaths(t *testing.T) {
	t.Parallel()

	tcs := []functionalReaderTestCase{
		{
			when: "MaxCommentBytes=5 first line has a comment followed by 6 chars and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`#234567`, ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().MaxCommentBytes(5),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpCommentBytesAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 7, record 0, field 0: " + csv.ErrSecOpCommentBytesAboveMax.Error(),
		},
		{
			when: "MaxCommentBytes=5 first line has a comment followed by 5 chars and field sep and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`#23456,`, ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().MaxCommentBytes(5),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpCommentBytesAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 7, record 0, field 0: " + csv.ErrSecOpCommentBytesAboveMax.Error(),
		},
		{
			when: "MaxCommentBytes=5 first line has a comment followed by 5 chars and escape char and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`#23456\`, ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
				csv.ReaderOpts().MaxCommentBytes(5),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpCommentBytesAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 7, record 0, field 0: " + csv.ErrSecOpCommentBytesAboveMax.Error(),
		},
		{
			when: "MaxCommentBytes=5 first line has a comment followed by 5 chars and quote char and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`#23456"`, ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().MaxCommentBytes(5),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpCommentBytesAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 7, record 0, field 0: " + csv.ErrSecOpCommentBytesAboveMax.Error(),
		},
		{
			when: "MaxCommentBytes=3 first line has a comment followed by 4 chars and CRLF record separator and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`#2345`+"\r\n", ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().MaxCommentBytes(3),
				csv.ReaderOpts().RecordSeparator("\r\n"),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpCommentBytesAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 5, record 0, field 0: " + csv.ErrSecOpCommentBytesAboveMax.Error(),
		},
		{ // rawr
			when: "MaxCommentBytes=5 first line has a comment followed by 5 chars and comment and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`#23456#`, ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().MaxCommentBytes(5),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpCommentBytesAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 7, record 0, field 0: " + csv.ErrSecOpCommentBytesAboveMax.Error(),
		},
		{
			when: "MaxCommentBytes=5 first line has a comment followed by 5 chars and CR and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`#23456`+"\r", ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().MaxCommentBytes(5),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpCommentBytesAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 7, record 0, field 0: " + csv.ErrSecOpCommentBytesAboveMax.Error(),
		},
	}

	for _, tc := range tcs {
		if tc.then == "" {
			tc.then = "coupled error"
		}

		tc.Run(t)
	}
}
