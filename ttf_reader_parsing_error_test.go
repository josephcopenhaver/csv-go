package csv_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/josephcopenhaver/csv-go/v2"
	"github.com/stretchr/testify/assert"
)

func TestFunctionalReaderParsingErrorPaths(t *testing.T) {
	t.Parallel()

	tcs := []functionalReaderTestCase{
		{
			when: "reader errors on no rows and file is zero length",
			then: "should return a specific error when .Err() is called and contain no rows",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ErrorOnNoRows(true),
			},
			iterErrIs:         []error{csv.ErrParsing, csv.ErrNoRows, io.ErrUnexpectedEOF},
			iterErrStr:        "parsing error at byte 0, record 0, field 0: no rows: unexpected EOF",
			hadRowsMsgAndArgs: []any{"zero length file should cause an error before any rows are constructed"},
		},
		{
			when: "reader errors on no BOM and file is zero length",
			then: "should return a specific error when .Err() is called and contain no rows",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ErrorOnNoByteOrderMarker(true),
			},
			iterErrIs:         []error{csv.ErrParsing, csv.ErrNoByteOrderMarker, io.ErrUnexpectedEOF},
			iterErrStr:        "parsing error at byte 0, record 0, field 0: no byte order marker\nunexpected EOF",
			hadRowsMsgAndArgs: []any{"zero length file should cause an error before any rows are constructed"},
		},
		{
			when: "read file contains quotes in unquoted field",
			then: "should return a specific error when .Err() is called and contain no rows",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("1,2\",3")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().ErrorOnQuotesInUnquotedField(true),
			},
			iterErrIs:         []error{csv.ErrParsing, csv.ErrQuoteInUnquotedField},
			iterErrStr:        "parsing error at byte 4, record 1, field 2: quote found in unquoted field",
			hadRowsMsgAndArgs: []any{"should error before any rows are constructed"},
		},
		{
			when: "reader errors on no rows with zero length file and row-borrow-disabled reader",
			then: "should return a specific error when .Err() is called and contain no rows",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ErrorOnNoRows(true),
				csv.ReaderOpts().BorrowRow(false),
				csv.ReaderOpts().BorrowFields(false),
			},
			iterErrIs:         []error{csv.ErrParsing, csv.ErrNoRows, io.ErrUnexpectedEOF},
			iterErrStr:        "parsing error at byte 0, record 0, field 0: no rows: unexpected EOF",
			hadRowsMsgAndArgs: []any{"zero length file should cause an error before any rows are constructed"},
		},
		{
			when: "reader errors on no rows with zero length file and row-borrow-enabled reader",
			then: "should return a specific error when .Err() is called and contain no rows",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ErrorOnNoRows(true),
				csv.ReaderOpts().BorrowRow(true),
				csv.ReaderOpts().BorrowFields(true),
			},
			iterErrIs:         []error{csv.ErrParsing, csv.ErrNoRows, io.ErrUnexpectedEOF},
			iterErrStr:        "parsing error at byte 0, record 0, field 0: no rows: unexpected EOF",
			hadRowsMsgAndArgs: []any{"zero length file should cause an error before any rows are constructed"},
		},
		{
			when: "reader has row misalignment: 1 then 2 fields",
			then: "should return an error when trying to read the second row",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("1\n2,3")),
				}
			},
			rows: [][]string{
				{"1"},
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrFieldCount, csv.ErrTooManyFields},
			iterErrStr: "parsing error at byte 4, record 2, field 1: too many fields: field count exceeds 1",
		},
		{
			when: "reader has row misalignment: 2 then 1 fields with eof",
			then: "should return an error when trying to read the second row",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("1,2\n3")),
				}
			},
			rows: [][]string{
				{"1", "2"},
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrFieldCount, csv.ErrNotEnoughFields, io.ErrUnexpectedEOF},
			iterErrStr: "parsing error at byte 5, record 2, field 1: not enough fields: expected 2 fields but found 1\nunexpected EOF",
		},
		{
			when: "reader has one row of two cols with NumFields=1",
			then: "should return an error when trying to read the second row",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("1,2")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().NumFields(1),
			},
			iterErrIs:         []error{csv.ErrParsing, csv.ErrFieldCount, csv.ErrTooManyFields},
			iterErrIsNot:      []error{io.ErrUnexpectedEOF},
			iterErrStr:        "parsing error at byte 2, record 1, field 1: too many fields: field count exceeds 1",
			hadRowsMsgAndArgs: []any{"column alignment should cause an error before any rows are constructed"},
		},
		{
			when: "reader has row misalignment: 1 then 2 fields with NumFields=2",
			then: "should return an error when trying to read the second row",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("1\n2,3")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().NumFields(2),
			},
			iterErrIs:         []error{csv.ErrParsing, csv.ErrFieldCount, csv.ErrNotEnoughFields},
			iterErrIsNot:      []error{io.ErrUnexpectedEOF},
			iterErrStr:        "parsing error at byte 2, record 1, field 1: not enough fields: expected 2 fields but found 1",
			hadRowsMsgAndArgs: []any{"column alignment should cause an error before any rows are constructed"},
		},
		{
			when: "reader has comment: then row misalignment: 1 then 2 fields with NumFields=2",
			then: "should return an error when trying to read the second row",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("#\n1\n2,3")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().NumFields(2),
			},
			iterErrIs:         []error{csv.ErrParsing, csv.ErrFieldCount, csv.ErrNotEnoughFields},
			iterErrIsNot:      []error{io.ErrUnexpectedEOF},
			iterErrStr:        "parsing error at byte 4, record 1, field 1: not enough fields: expected 2 fields but found 1",
			hadRowsMsgAndArgs: []any{"column alignment should cause an error before any rows are constructed"},
		},
		{
			when: "when record sep is CRLF",
			then: "an error should be thrown if EOF is thrown before LF part of CRLF",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("a,b,c\r")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r\n"),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrUnsafeCRFileEnd, io.ErrUnexpectedEOF},
			iterErrStr: csv.ErrParsing.Error() + " at byte 6, record 1, field 3: " + csv.ErrUnsafeCRFileEnd.Error(),
		},
		{
			when: "reader is discovering the record sep and first row has two columns and ends in CR+short-multibyte+EOF",
			then: "one row should be returned and error should be raised to the .Err method",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(bytes.NewReader(append([]byte("a,b\r"), 0xC0))),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().DiscoverRecordSeparator(true),
			},
			rows:       [][]string{strings.Split("a,b", ",")},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNotEnoughFields, io.ErrUnexpectedEOF},
			iterErrStr: csv.ErrParsing.Error() + " at byte 5, record 2, field 1: " + csv.ErrNotEnoughFields.Error() + ": expected 2 fields but found 1\n" + io.ErrUnexpectedEOF.Error(),
		},
		{
			when: "erroring on no rows, expecting headers, but document is empty",
			then: "should receive no header row error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ErrorOnNoRows(true),
				csv.ReaderOpts().ExpectHeaders(strings.Split("a,b", ",")),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNoHeaderRow, io.ErrUnexpectedEOF},
			iterErrStr: csv.ErrParsing.Error() + " at byte 0, record 0, field 0: " + csv.ErrNoHeaderRow.Error(),
		},
		{
			when: "erroring on no rows, removing the first header row, but document is empty",
			then: "should receive no header row error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ErrorOnNoRows(true),
				csv.ReaderOpts().RemoveHeaderRow(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNoHeaderRow, io.ErrUnexpectedEOF},
			iterErrStr: csv.ErrParsing.Error() + " at byte 0, record 0, field 0: " + csv.ErrNoHeaderRow.Error(),
		},
		{
			when: "erroring on no rows, whitespace-trimming the header row values, but document is empty",
			then: "should receive no header row error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ErrorOnNoRows(true),
				csv.ReaderOpts().TrimHeaders(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNoHeaderRow, io.ErrUnexpectedEOF},
			iterErrStr: csv.ErrParsing.Error() + " at byte 0, record 0, field 0: " + csv.ErrNoHeaderRow.Error(),
		},
		{
			when: "whitespace-trimming the header row values, but expecting 3 not 2 headers via ExpectHeaders call",
			then: "should error not enough fields",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(" a , b \n1,2,3")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ExpectHeaders(strings.Split("a,b,c", ",")),
				csv.ReaderOpts().TrimHeaders(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNotEnoughFields},
			iterErrStr: csv.ErrParsing.Error() + " at byte 8, record 1, field 2: " + csv.ErrNotEnoughFields.Error() + ": expected 3 fields but found 2",
		},
		{
			when: "whitespace-trimming the header row values, but expecting 2 not 3 headers via ExpectHeaders call",
			then: "should error too many fields",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(" a , b, c \n1,2,3")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ExpectHeaders(strings.Split("a,b", ",")),
				csv.ReaderOpts().TrimHeaders(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrFieldCount, csv.ErrTooManyFields},
			iterErrStr: csv.ErrParsing.Error() + " at byte 7, record 1, field 2: " + csv.ErrTooManyFields.Error() + ": field count exceeds 2",
		},
		{
			when: "whitespace-trimming the header row values, they do not match",
			then: "should error unexpected header row contents",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(" a , b \n1,2")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ExpectHeaders(strings.Split("a,c", ",")),
				csv.ReaderOpts().TrimHeaders(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrUnexpectedHeaderRowContents},
			iterErrStr: csv.ErrParsing.Error() + " at byte 8, record 2, field 1: " + csv.ErrUnexpectedHeaderRowContents.Error(),
		},
		{
			when: "not whitespace-trimming the header row values by default, they do not match because of whitespace",
			then: "should error unexpected header row contents",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(" a , b \n1,2")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ExpectHeaders(strings.Split("a,b", ",")),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrUnexpectedHeaderRowContents},
			iterErrStr: csv.ErrParsing.Error() + " at byte 8, record 2, field 1: " + csv.ErrUnexpectedHeaderRowContents.Error(),
		},
		{
			when: "not whitespace-trimming the header row values, they do not match because of whitespace",
			then: "should error unexpected header row contents",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(" a , b \n1,2")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ExpectHeaders(strings.Split("a,b", ",")),
				csv.ReaderOpts().TrimHeaders(false),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrUnexpectedHeaderRowContents},
			iterErrStr: csv.ErrParsing.Error() + " at byte 8, record 2, field 1: " + csv.ErrUnexpectedHeaderRowContents.Error(),
		},
		{
			when: "erroring on no rows, there are 2 columns expecting a header row, and there is only a header row",
			then: "should error that there are no rows returned",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("a,b")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RemoveHeaderRow(true),
				csv.ReaderOpts().ErrorOnNoRows(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNoRows},
			iterErrStr: csv.ErrParsing.Error() + " at byte 3, record 1, field 2: " + csv.ErrNoRows.Error(),
		},
		{
			when: "erroring on no rows, there is 1 column expecting a header row, and there is only a header row",
			then: "should error that there are no rows returned",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("a")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RemoveHeaderRow(true),
				csv.ReaderOpts().ErrorOnNoRows(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNoRows},
			iterErrStr: csv.ErrParsing.Error() + " at byte 1, record 1, field 1: " + csv.ErrNoRows.Error(),
		},
		{
			when: "CRLF record sep enabled and comment line ends file with CR",
			then: "returns parsing error ErrUnsafeCRFileEnd",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("# neat\r")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().RecordSeparator("\r\n"),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrUnsafeCRFileEnd, io.ErrUnexpectedEOF},
			iterErrStr: csv.ErrParsing.Error() + " at byte 7, record 0, field 0: " + csv.ErrUnsafeCRFileEnd.Error(),
		},
		{
			when: "comments after start of records and support explicitly disabled",
			then: "error processing second row because it is a comment interpreted as one field",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("#neat1\na,b,c\n#neat2\n1,2,3")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().CommentsAllowedAfterStartOfRecords(false),
			},
			rows:       [][]string{strings.Split("a,b,c", ",")},
			iterErrStr: csv.ErrParsing.Error() + " at byte 20, record 2, field 1: " + csv.ErrNotEnoughFields.Error() + ": expected 3 fields but found 1",
		},
		{
			when: "comments after start of records and support explicitly disabled and numFields=3",
			then: "error processing second row because it is a comment interpreted as one field",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("#neat1\na,b,c\n#neat2\n1,2,3")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().CommentsAllowedAfterStartOfRecords(false),
				csv.ReaderOpts().NumFields(3),
			},
			rows:       [][]string{strings.Split("a,b,c", ",")},
			iterErrStr: csv.ErrParsing.Error() + " at byte 20, record 2, field 1: " + csv.ErrNotEnoughFields.Error() + ": expected 3 fields but found 1",
		},
		{
			when: "comments after start of records and support implicitly disabled",
			then: "error processing second row because it is a comment interpreted as one field",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("#neat1\na,b,c\n#neat2\n1,2,3")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
			},
			rows:       [][]string{strings.Split("a,b,c", ",")},
			iterErrStr: csv.ErrParsing.Error() + " at byte 20, record 2, field 1: " + csv.ErrNotEnoughFields.Error() + ": expected 3 fields but found 1",
		},
		{
			when: "comments after start of records and support implicitly disabled and numFields=3",
			then: "error processing second row because it is a comment interpreted as one field",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("#neat1\na,b,c\n#neat2\n1,2,3")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().NumFields(3),
			},
			rows:       [][]string{strings.Split("a,b,c", ",")},
			iterErrStr: csv.ErrParsing.Error() + " at byte 20, record 2, field 1: " + csv.ErrNotEnoughFields.Error() + ": expected 3 fields but found 1",
		},
	}

	for _, tc := range tcs {
		assert.NotEmpty(t, tc.then, "then segment of test is empty")
		if tc.then == "" {
			t.FailNow()
		}
		tc.Run(t)
	}
}
