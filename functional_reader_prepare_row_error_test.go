package csv_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/josephcopenhaver/csv-go"
)

const (
	utf8LineSeparator = "\u2028"
)

func TestFunctionalReaderPrepareRowErrorPaths(t *testing.T) {

	tcs := []functionalReaderTestCase{
		{
			when: "EOF in quoted field",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`"hi th`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrIncompleteQuotedField, io.ErrUnexpectedEOF},
			iterErrStr: csv.ErrParsing.Error() + " at byte 6, record 1, field 1: " + csv.ErrIncompleteQuotedField.Error(),
		},
		{
			when: "EOF in quoted field after escape",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`"\"hi there\`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrIncompleteQuotedField, io.ErrUnexpectedEOF},
			iterErrStr: csv.ErrParsing.Error() + " at byte 12, record 1, field 1: " + csv.ErrIncompleteQuotedField.Error(),
		},
		{
			when: "in quoted field reader ends in incomplete utf8 rune after enabled escape rune",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(bytes.NewReader(append([]byte(`"\"hi there\`), 0xC0))),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrInvalidEscapeInQuotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 13, record 1, field 1: " + csv.ErrInvalidEscapeInQuotedField.Error() + ": unexpected non-UTF8 byte following escape",
		},
		{
			when: "at end of quoted field reader ends in incomplete utf8 rune ",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(bytes.NewReader(append([]byte(`"hi there"`), 0xC0))),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrInvalidQuotedFieldEnding},
			iterErrStr: csv.ErrParsing.Error() + " at byte 11, record 1, field 1: " + csv.ErrInvalidQuotedFieldEnding.Error(),
		},
		{
			when: "BOM is required but reader is empty",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ErrorOnNoByteOrderMarker(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNoByteOrderMarker},
			iterErrStr: csv.ErrParsing.Error() + " at byte 0, record 0, field 0: " + errors.Join(csv.ErrNoByteOrderMarker, io.ErrUnexpectedEOF).Error(),
		},
		{
			when: "numFields=2 and record start is CR",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().NumFields(2),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNotEnoughFields},
			iterErrStr: csv.ErrParsing.Error() + " at byte 1, record 1, field 1: " + csv.ErrNotEnoughFields.Error() + ": expected 2 fields but found 1",
		},
		{
			when: "escape set, quote set, and an invalid character follows the escape character",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`"\"\x`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrInvalidEscapeInQuotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 5, record 1, field 1: " + csv.ErrInvalidEscapeInQuotedField.Error() + ": unexpected rune following escape",
		},
		{
			when: "numFields=1, first column is quoted, second column is unquoted",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`"",`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().NumFields(1),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrTooManyFields},
			iterErrStr: csv.ErrParsing.Error() + " at byte 3, record 1, field 1: " + csv.ErrTooManyFields.Error() + ": field count exceeds 1",
		},
		{
			when: "three quote chars and escape is set",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`"""`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrUnexpectedQuoteAfterField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 3, record 1, field 1: " + csv.ErrUnexpectedQuoteAfterField.Error(),
		},
		{
			when: "numFields=2, two quote chars, LF, then EOF",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("\"\"\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().NumFields(2),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNotEnoughFields},
			iterErrStr: csv.ErrParsing.Error() + " at byte 3, record 1, field 1: " + csv.ErrNotEnoughFields.Error() + ": expected 2 fields but found 1",
		},
		{
			when: "numFields=2, two quote chars, then x+EOF",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("\"\"x")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().NumFields(2),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrInvalidQuotedFieldEnding},
			iterErrStr: csv.ErrParsing.Error() + " at byte 3, record 1, field 1: " + csv.ErrInvalidQuotedFieldEnding.Error(),
		},
		{
			when: "numFields=2 comma comma",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(",,")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().NumFields(2),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrTooManyFields},
			iterErrStr: csv.ErrParsing.Error() + " at byte 2, record 1, field 2: " + csv.ErrTooManyFields.Error() + ": field count exceeds 2",
		},
		{
			when: "numFields=3 and comma+LF",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(",\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().NumFields(3),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNotEnoughFields},
			iterErrStr: csv.ErrParsing.Error() + " at byte 2, record 1, field 2: " + csv.ErrNotEnoughFields.Error() + ": expected 3 fields but found 2",
		},
		{
			when: "BOM is required but missing and there is an incomplete utf8 rune+EOF",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(bytes.NewReader([]byte{0xC0})),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ErrorOnNoByteOrderMarker(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNoByteOrderMarker},
			iterErrStr: csv.ErrParsing.Error() + " at byte 0, record 0, field 0: " + csv.ErrNoByteOrderMarker.Error(),
		},
		{
			when: "BOM is required but missing on a normal row",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("hi")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ErrorOnNoByteOrderMarker(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNoByteOrderMarker},
			iterErrStr: csv.ErrParsing.Error() + " at byte 0, record 0, field 0: " + csv.ErrNoByteOrderMarker.Error(),
		},
		{
			when: "implicit error on newline in field where CR in middle of field",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("h\ri")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\n"),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNewlineInUnquotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 2, record 1, field 1: " + csv.ErrNewlineInUnquotedField.Error() + ": carriage return",
		},
		{
			when: "explicit error on newline in field where CR in middle of field",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("h\ri")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\n"),
				csv.ReaderOpts().ErrorOnNewlineInUnquotedField(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNewlineInUnquotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 2, record 1, field 1: " + csv.ErrNewlineInUnquotedField.Error() + ": carriage return",
		},
		{
			when: "implicit error on newline in field where LF in middle of field",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("h\ni")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator(utf8LineSeparator),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNewlineInUnquotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 2, record 1, field 1: " + csv.ErrNewlineInUnquotedField.Error() + ": line feed",
		},
		{
			when: "explicit error on newline in field where LF in middle of field",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("h\ni")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator(utf8LineSeparator),
				csv.ReaderOpts().ErrorOnNewlineInUnquotedField(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNewlineInUnquotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 2, record 1, field 1: " + csv.ErrNewlineInUnquotedField.Error() + ": line feed",
		},
		{
			when: "implicit error on newline in field where CR at start of record",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("\r")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\n"),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNewlineInUnquotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 1, record 1, field 1: " + csv.ErrNewlineInUnquotedField.Error() + ": carriage return",
		},
		{
			when: "explicit error on newline in field where CR at start of record",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("\r")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\n"),
				csv.ReaderOpts().ErrorOnNewlineInUnquotedField(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNewlineInUnquotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 1, record 1, field 1: " + csv.ErrNewlineInUnquotedField.Error() + ": carriage return",
		},
		{
			when: "implicit error on newline in field where LF at start of record",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator(utf8LineSeparator),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNewlineInUnquotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 1, record 1, field 1: " + csv.ErrNewlineInUnquotedField.Error() + ": line feed",
		},
		{
			when: "explicit error on newline in field where LF at start of record",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator(utf8LineSeparator),
				csv.ReaderOpts().ErrorOnNewlineInUnquotedField(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNewlineInUnquotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 1, record 1, field 1: " + csv.ErrNewlineInUnquotedField.Error() + ": line feed",
		},
		{
			when: "implicit error on newline in field where CR at start of second field",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(",\r")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\n"),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNewlineInUnquotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 2, record 1, field 2: " + csv.ErrNewlineInUnquotedField.Error() + ": carriage return",
		},
		{
			when: "explicit error on newline in field where CR at start of second field",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(",\r")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\n"),
				csv.ReaderOpts().ErrorOnNewlineInUnquotedField(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNewlineInUnquotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 2, record 1, field 2: " + csv.ErrNewlineInUnquotedField.Error() + ": carriage return",
		},
		{
			when: "implicit error on newline in field where LF at start of second field",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(",\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator(utf8LineSeparator),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNewlineInUnquotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 2, record 1, field 2: " + csv.ErrNewlineInUnquotedField.Error() + ": line feed",
		},
		{
			when: "explicit error on newline in field where LF at start of second field",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(",\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator(utf8LineSeparator),
				csv.ReaderOpts().ErrorOnNewlineInUnquotedField(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNewlineInUnquotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 2, record 1, field 2: " + csv.ErrNewlineInUnquotedField.Error() + ": line feed",
		},
		{
			when: "expecting only one column but record starts with field separator",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(",")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().NumFields(1),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrTooManyFields},
			iterErrStr: csv.ErrParsing.Error() + " at byte 1, record 1, field 1: " + csv.ErrTooManyFields.Error() + ": field count exceeds 1",
		},
		{
			when: "BOM required but doc starts with another multibyte rune instead",
			then: "error at byte 0 - no BOM",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(func() string {
						faceWithColdSweat := 0x1F613
						return string(rune(faceWithColdSweat))
					}())),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RemoveByteOrderMarker(true),
				csv.ReaderOpts().ErrorOnNoByteOrderMarker(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNoByteOrderMarker},
			iterErrStr: csv.ErrParsing.Error() + " at byte 0, record 0, field 0: " + csv.ErrNoByteOrderMarker.Error(),
		},
		{
			when: "escape set and field count overflows after closing quote",
			then: "coupled error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("\"\",")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
				csv.ReaderOpts().NumFields(1),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrTooManyFields},
			iterErrStr: csv.ErrParsing.Error() + " at byte 3, record 1, field 1: " + csv.ErrTooManyFields.Error() + ": field count exceeds 1",
		},
		{
			when: "escape set and io error after closing quote and CR and record sep is CRLF",
			then: "coupled error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(func() *errBufferedReader {
						ebr := newErrBufferedReader(errReader{
							t:        t,
							reader:   strings.NewReader("\"\"\r"),
							numBytes: 4,
							err:      io.ErrClosedPipe,
						})

						return ebr
					}()),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r\n"),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			iterErrIs:  []error{csv.ErrIO, io.ErrClosedPipe},
			iterErrStr: csv.ErrIO.Error() + " at byte 3, record 1, field 1: " + io.ErrClosedPipe.Error(),
		},
		{
			when: "escape set and record sep CRLF after closing quote but field count under-flows",
			then: "coupled error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("\"\"\r\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r\n"),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
				csv.ReaderOpts().NumFields(2),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNotEnoughFields},
			iterErrStr: csv.ErrParsing.Error() + " at byte 4, record 1, field 1: " + csv.ErrNotEnoughFields.Error() + ": expected 2 fields but found 1",
		},
		{
			when: "escape set and unexpected normal ascii char after closing quote",
			then: "coupled error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("\"\"a")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
				csv.ReaderOpts().NumFields(2),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrInvalidQuotedFieldEnding},
			iterErrStr: csv.ErrParsing.Error() + " at byte 3, record 1, field 1: " + csv.ErrInvalidQuotedFieldEnding.Error(),
		},
	}

	for i := range tcs {
		tcs[i].Run(t)
	}
}
