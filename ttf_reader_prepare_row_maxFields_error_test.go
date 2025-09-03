package csv_test

import (
	"testing"

	"github.com/josephcopenhaver/csv-go/v3"
)

func TestFunctionalReaderPrepareRowMaxFieldsErrorPaths(t *testing.T) {
	t.Parallel()

	tcs := []functionalReaderTestCase{
		{
			when: "MaxFields=2 and NumFields=2 and first row contains 3 fields",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(",,34567", ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().NumFields(2),
				csv.ReaderOpts().MaxFields(2),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrFieldCount, csv.ErrTooManyFields},
			iterErrStr: csv.ErrParsing.Error() + " at byte 2, record 1, field 2: " + csv.ErrTooManyFields.Error() + ": field count exceeds 2",
		},
		{
			when: "MaxFields=3 and NumFields=2 and first row contains 3 fields",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(",,34567", ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().NumFields(2),
				csv.ReaderOpts().MaxFields(3),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrFieldCount, csv.ErrTooManyFields},
			iterErrStr: csv.ErrParsing.Error() + " at byte 2, record 1, field 2: " + csv.ErrTooManyFields.Error() + ": field count exceeds 2",
		},
		{
			when: "MaxFields=2 and NumFields=unspecified and first row contains 3 fields",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(",,34567", ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxFields(2),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrFieldCount, csv.ErrTooManyFields, csv.ErrSecOpFieldCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 2, record 1, field 2: " + csv.ErrSecOpFieldCountAboveMax.Error(),
		},
		{
			when: "MaxFields=2 and NumFields=unspecified and first row contains 2 fields and second row contains 3 fields",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(",\n,,567", ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxFields(2),
			},
			rows:       [][]string{{"", ""}},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrFieldCount, csv.ErrTooManyFields},
			iterErrStr: csv.ErrParsing.Error() + " at byte 4, record 2, field 2: " + csv.ErrTooManyFields.Error() + ": field count exceeds 2",
		},
	}

	for _, tc := range tcs {
		if tc.then == "" {
			tc.then = "coupled error"
		}
		tc.Run(t)
	}
}
