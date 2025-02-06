package csv_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/josephcopenhaver/csv-go"
)

func TestFunctionalReaderPrepareRowErrorPaths(t *testing.T) {

	tcs := []functionalReaderTestCase{
		{
			when: "EOF in quoted field",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader(`"hi th`)),
				csv.ReaderOpts().Quote('"'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrIncompleteQuotedField, io.ErrUnexpectedEOF},
			iterErrStr: csv.ErrParsing.Error() + " at byte 6, record 1, field 1: " + csv.ErrIncompleteQuotedField.Error(),
		},
		{
			when: "EOF in quoted field after escape",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader(`"\"hi there\`)),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrIncompleteQuotedField, io.ErrUnexpectedEOF},
			iterErrStr: csv.ErrParsing.Error() + " at byte 12, record 1, field 1: " + csv.ErrIncompleteQuotedField.Error(),
		},
		{
			when: "in quoted field reader ends in incomplete utf8 rune after enabled escape rune",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(bytes.NewReader(append([]byte(`"\"hi there\`), 0xC0))),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrInvalidEscapeInQuotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 13, record 1, field 1: " + csv.ErrInvalidEscapeInQuotedField.Error() + ": unexpected non-UTF8 byte following escape",
		},
		{
			when: "at end of quoted field reader ends in incomplete utf8 rune ",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(bytes.NewReader(append([]byte(`"hi there"`), 0xC0))),
				csv.ReaderOpts().Quote('"'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrInvalidQuotedFieldEnding},
			iterErrStr: csv.ErrParsing.Error() + " at byte 11, record 1, field 1: " + csv.ErrInvalidQuotedFieldEnding.Error(),
		},
		{
			when: "BOM is required but reader is empty",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().ErrorOnNoByteOrderMarker(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNoByteOrderMarker},
			iterErrStr: csv.ErrParsing.Error() + " at byte 0, record 0, field 0: " + errors.Join(csv.ErrNoByteOrderMarker, io.ErrUnexpectedEOF).Error(),
		},
		{
			when: "numFields=2 and record start is CR",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("\n")),
				csv.ReaderOpts().NumFields(2),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNotEnoughFields},
			iterErrStr: csv.ErrParsing.Error() + " at byte 1, record 1, field 1: " + csv.ErrNotEnoughFields.Error() + ": expected 2 fields but found 1",
		},
		{
			when: "escape set, quote set, and an invalid character follows the escape character",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader(`"\"\x`)),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrInvalidEscapeInQuotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 5, record 1, field 1: " + csv.ErrInvalidEscapeInQuotedField.Error() + ": unexpected rune following escape",
		},
		{
			when: "numFields=1, first column is quoted, second column is unquoted",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader(`"",`)),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().NumFields(1),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrTooManyFields},
			iterErrStr: csv.ErrParsing.Error() + " at byte 3, record 1, field 1: " + csv.ErrTooManyFields.Error() + ": field count exceeds 1",
		},
		{
			when: "three quote chars and escape is set",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader(`"""`)),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrUnexpectedQuoteAfterField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 3, record 1, field 1: " + csv.ErrUnexpectedQuoteAfterField.Error(),
		},
		{
			when: "numFields=2, two quote chars, LF, then EOF",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("\"\"\n")),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().NumFields(2),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNotEnoughFields},
			iterErrStr: csv.ErrParsing.Error() + " at byte 3, record 1, field 1: " + csv.ErrNotEnoughFields.Error() + ": expected 2 fields but found 1",
		},
		{
			when: "numFields=2, two quote chars, then x+EOF",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("\"\"x")),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().NumFields(2),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrInvalidQuotedFieldEnding},
			iterErrStr: csv.ErrParsing.Error() + " at byte 3, record 1, field 1: " + csv.ErrInvalidQuotedFieldEnding.Error(),
		},
		{
			when: "numFields=2 comma comma",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader(",,")),
				csv.ReaderOpts().NumFields(2),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrTooManyFields},
			iterErrStr: csv.ErrParsing.Error() + " at byte 2, record 1, field 2: " + csv.ErrTooManyFields.Error() + ": field count exceeds 2",
		},
		{
			when: "numFields=3 and comma+LF",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader(",\n")),
				csv.ReaderOpts().NumFields(3),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNotEnoughFields},
			iterErrStr: csv.ErrParsing.Error() + " at byte 2, record 1, field 2: " + csv.ErrNotEnoughFields.Error() + ": expected 3 fields but found 2",
		},
		{
			when: "BOM is required but missing and there is an incomplete utf8 rune+EOF",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(bytes.NewReader([]byte{0xC0})),
				csv.ReaderOpts().ErrorOnNoByteOrderMarker(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNoByteOrderMarker},
			iterErrStr: csv.ErrParsing.Error() + " at byte 0, record 0, field 0: " + csv.ErrNoByteOrderMarker.Error(),
		},
		{
			when: "BOM is required but missing on a normal row",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("hi")),
				csv.ReaderOpts().ErrorOnNoByteOrderMarker(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNoByteOrderMarker},
			iterErrStr: csv.ErrParsing.Error() + " at byte 0, record 0, field 0: " + csv.ErrNoByteOrderMarker.Error(),
		},
	}

	for i := range tcs {
		tcs[i].Run(t)
	}
}
