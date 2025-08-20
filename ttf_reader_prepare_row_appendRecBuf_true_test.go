package csv_test

import (
	"strings"
	"testing"

	"github.com/josephcopenhaver/csv-go/v2"
)

func TestFunctionalReaderPrepareRowAppendRecBufRespTruePaths(t *testing.T) {
	t.Parallel()

	tcs := []functionalReaderTestCase{
		{
			when: "MaxRecordBytes=1 first record column exceeds max record bytes",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("12")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxRecordBytes(1),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 2, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=1 second record first column exceeds max record bytes",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("1\n23")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxRecordBytes(1),
			},
			rows:       [][]string{{"1"}},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 4, record 2, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=1 first record first quoted column exceeds max record bytes and has no end quote",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`"12`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxRecordBytes(1),
				csv.ReaderOpts().Quote('"'),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 3, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=7 first record first column exceeds max record bytes after read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`1234567`, `8`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxRecordBytes(7),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 8, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=1 first record first field exceeds max record bytes before record separator",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`12,`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxRecordBytes(1),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 2, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=1 first record first quoted field exceeds max record bytes with field separator",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`"1,`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxRecordBytes(1),
				csv.ReaderOpts().Quote('"'),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 3, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=1 first record second field exceeds max record bytes before field separator",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`1,2,`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxRecordBytes(1),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 3, record 1, field 2: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=7 first record first field exceeds max record bytes after read-partition before field separator",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`1234567`, `8,`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxRecordBytes(7),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 8, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "record starts with an escape character preceded by 6 numerals with maxRecordBytes set to 6 followed by a read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`123456\`, ``)),
					csv.ReaderOpts().Quote('"'),
					csv.ReaderOpts().Escape('\\'),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxRecordBytes(6),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 7, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "second field starts with an escape character preceded by 6 numerals with maxRecordBytes set to 6 followed by a read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`,123456\`, ``)),
					csv.ReaderOpts().Quote('"'),
					csv.ReaderOpts().Escape('\\'),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxRecordBytes(6),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 8, record 1, field 2: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "with maxRecordBytes set to 5 with quote char before a read-partition and 6 characters followed by an escape character followed by a read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`"`, `123456\`, ``)),
					csv.ReaderOpts().Quote('"'),
					csv.ReaderOpts().Escape('\\'),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxRecordBytes(5),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 7, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "with maxRecordBytes set to 5 with quote char before a read-partition and 6 characters followed by an escape character followed by a read-partition and another escape character and another read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`"`, `123456\`, `\`, ``)),
					csv.ReaderOpts().Quote('"'),
					csv.ReaderOpts().Escape('\\'),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxRecordBytes(6),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 9, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "with maxRecordBytes set to 7 with first field has seven data bytes a read-partition one more data byte and escape and a read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`1234567`, `8\`, ``)),
					csv.ReaderOpts().Quote('"'),
					csv.ReaderOpts().Escape('\\'),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxRecordBytes(7),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 8, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "ErrorOnQuotesInUnquotedField=false, in quote with data byte, quote, and read-partition with maxRecordBytes set to 1",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`1"`, ``)),
					csv.ReaderOpts().Quote('"'),
					csv.ReaderOpts().ErrorOnQuotesInUnquotedField(false),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxRecordBytes(1),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 2, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "maxRecordBytes set to 4 and first field has 5 bytes quoted followed by a read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`"12345"`, ``)),
					csv.ReaderOpts().Quote('"'),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxRecordBytes(4),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 6, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "escape enabled with MaxRecordBytes=5 with first field quoted with 5 data bytes followed by an escaped quote character",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`"12345\`, `"`, `"`)),
					csv.ReaderOpts().Quote('"'),
					csv.ReaderOpts().Escape('\\'),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxRecordBytes(5),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 8, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=5 quoted field of 5 characters followed by a read-partition quote read-partition quote cycle",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`"12345"`, `"`, `"`)),
					csv.ReaderOpts().Quote('"'),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxRecordBytes(5),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 8, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=4 ErrorOnQuotesInUnquotedField=false with empty first field and a second field with 5 characters followed by a quote and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`,12345"`, ``)),
					csv.ReaderOpts().Quote('"'),
					csv.ReaderOpts().ErrorOnQuotesInUnquotedField(false),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxRecordBytes(4),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 6, record 1, field 2: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=7 ErrorOnQuotesInUnquotedField=false first field with 5 characters followed by a read-partition and quote and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`1234567`, `"`, ``)),
					csv.ReaderOpts().Quote('"'),
					csv.ReaderOpts().ErrorOnQuotesInUnquotedField(false),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().MaxRecordBytes(7),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 8, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=5 RecordSeparator=CRLF first field with 5 characters followed by a CR and a byte and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("12345\r7", ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r\n"),
				csv.ReaderOpts().MaxRecordBytes(5),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 6, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=4 RecordSeparator=CRLF second field with 4 characters followed by a CR and a byte and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(",2345\r7", ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r\n"),
				csv.ReaderOpts().MaxRecordBytes(4),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 6, record 1, field 2: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=5 RecordSeparator=CRLF first field with quote 5 characters followed by a CR and a byte and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`"23456`+"\r"+`8`, ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().RecordSeparator("\r\n"),
				csv.ReaderOpts().MaxRecordBytes(5),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 7, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=7 RecordSeparator=CRLF first field with 7 characters followed by a CR and a byte and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`1234567`, "\r"+`8`, ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r\n"),
				csv.ReaderOpts().MaxRecordBytes(7),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 8, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=6 RecordSeparator=CRLF first field with 7 characters followed by a CRLF and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`1234567`+"\r\n", ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r\n"),
				csv.ReaderOpts().MaxRecordBytes(6),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 7, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=6 RecordSeparator=CRLF first field with quote 6 characters followed by read-partition one char CRLF and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`"234567`, "8\r\n", ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r\n"),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().MaxRecordBytes(6),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 8, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=3 RecordSeparator=CRLF first field with fieldsep followed by 4 characters followed by a CRLF and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`,2345`+"\r\n", ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r\n"),
				csv.ReaderOpts().MaxRecordBytes(3),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 5, record 1, field 2: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=7 RecordSeparator=CRLF first field with 7 characters followed by read-partition and a character and CRLF and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`1234567`, `8`+"\r\n", ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r\n"),
				csv.ReaderOpts().MaxRecordBytes(7),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 8, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=6 Comment=# first field with 6 characters followed by comment and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`123456#`, ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().MaxRecordBytes(6),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 7, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=5 Comment=# first field with quote and 5 characters followed by comment and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`"23456#`, ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().MaxRecordBytes(5),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 7, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=6 Comment=# 5 chars and comma and 1 char and read-partition and comment and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`12345,1`, `#`, ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().MaxRecordBytes(6),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 8, record 1, field 2: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=5 first field with quote and 5 characters followed by a CR and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`"23456`+"\r", ``)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().MaxRecordBytes(5),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 7, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
		{
			when: "MaxRecordBytes=5 DiscoverRecordSeparator=true first field with quote and 5 characters followed by a LF and read-partition",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`"23456`+"\n", ``)),
					csv.ReaderOpts().DiscoverRecordSeparator(true),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().MaxRecordBytes(5),
			},
			iterErrIs:  []error{csv.ErrSecOp, csv.ErrSecOpRecordByteCountAboveMax},
			iterErrStr: csv.ErrSecOp.Error() + " at byte 7, record 1, field 1: " + csv.ErrSecOpRecordByteCountAboveMax.Error(),
		},
	}

	for _, tc := range tcs {
		if tc.then == "" {
			tc.then = "coupled error"
		}
		tc.Run(t)
	}
}
