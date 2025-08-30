package csv_test

import (
	"testing"

	"github.com/josephcopenhaver/csv-go/v2"
)

func TestFunctionalReaderPrepareRowMaxRecordsErrorPaths(t *testing.T) {
	t.Parallel()

	tcs := []functionalReaderTestCase{
		{
			when: "appendRecBufNotAllowed called raises ErrSecOpRecordCountAboveMax",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("\n\n\n\n\n\n\n", "123456\n", ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxRecords(7),
			},
			rows:       [][]string{{``}, {``}, {``}, {``}, {``}, {``}, {``}},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 8, record 8, field 1: " + csv.ErrSecOpRecordCountAboveMax.Error(),
		},
		{
			when: "checkNumFieldsNotAllowed called raises ErrSecOpRecordCountAboveMax",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("\n\n\n\n\n\n\n", ``)),
					csv.ReaderOpts().TerminalRecordSeparatorEmitsRecord(true),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxRecords(7),
			},
			rows:       [][]string{{``}, {``}, {``}, {``}, {``}, {``}, {``}},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 8, record 8, field 1: " + csv.ErrSecOpRecordCountAboveMax.Error(),
		},
	}

	for _, tc := range tcs {
		if tc.then == "" {
			tc.then = "coupled error"
		}
		tc.Run(t)
	}
}
