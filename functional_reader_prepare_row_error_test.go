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
			when: "In unquoted field, encounter CR without LF then EOF with RecSep=CRLF",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("a234567", "\r")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r\n"),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrUnsafeCRFileEnd, io.ErrUnexpectedEOF},
			iterErrStr: csv.ErrParsing.Error() + " at byte 8, record 1, field 1: " + csv.ErrUnsafeCRFileEnd.Error(),
		},
		{
			when: "In quoted field, encounter CR without LF then EOF with RecSep=CRLF",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("\"234567", "\r")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r\n"),
				csv.ReaderOpts().Quote('"'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrUnsafeCRFileEnd, io.ErrUnexpectedEOF},
			iterErrStr: csv.ErrParsing.Error() + " at byte 8, record 1, field 1: " + csv.ErrUnsafeCRFileEnd.Error(),
		},
		{
			when: "CR without LF at start of doc after a BOM while DropBOM=true and RecSep=CRLF",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(string(bomBytes()) + "\r")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RemoveByteOrderMarker(true),
				csv.ReaderOpts().RecordSeparator("\r\n"),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrUnsafeCRFileEnd, io.ErrUnexpectedEOF},
			iterErrStr: csv.ErrParsing.Error() + " at byte 4, record 1, field 1: " + csv.ErrUnsafeCRFileEnd.Error(),
		},
		{
			when: "CR without LF at start of doc while DropBOM=true and RecSep=CRLF",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("\r")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RemoveByteOrderMarker(true),
				csv.ReaderOpts().RecordSeparator("\r\n"),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrUnsafeCRFileEnd, io.ErrUnexpectedEOF},
			iterErrStr: csv.ErrParsing.Error() + " at byte 1, record 1, field 1: " + csv.ErrUnsafeCRFileEnd.Error(),
		},
		{
			when: "CR without LF at start of doc while BOM required and RecSep=CRLF",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("\r")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ErrorOnNoByteOrderMarker(true),
				csv.ReaderOpts().RecordSeparator("\r\n"),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNoByteOrderMarker},
			iterErrStr: csv.ErrParsing.Error() + " at byte 0, record 0, field 0: " + csv.ErrNoByteOrderMarker.Error(),
		},
		{
			when: "quote char found in middle of field and rFlagErrOnQInUF=true explicitly",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`a234567`, `"`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ErrorOnQuotesInUnquotedField(true),
				csv.ReaderOpts().Quote('"'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrQuoteInUnquotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 8, record 1, field 1: " + csv.ErrQuoteInUnquotedField.Error(),
		},
		{
			when: "quote char found in middle of field and rFlagErrOnQInUF=true implicitly",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`a234567`, `"`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrQuoteInUnquotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 8, record 1, field 1: " + csv.ErrQuoteInUnquotedField.Error(),
		},
		{
			when: "quote char found in after quoted field end but not in first char after the escape",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`"23456"`, `a"`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrInvalidQuotedFieldEnding},
			iterErrStr: csv.ErrParsing.Error() + " at byte 8, record 1, field 1: " + csv.ErrInvalidQuotedFieldEnding.Error(),
		},
		{
			when: "quote char found in quoted field after escape but not in first char after the escape",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`"23456\`, `a"`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrInvalidEscSeqInQuotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 8, record 1, field 1: " + csv.ErrInvalidEscSeqInQuotedField.Error(),
		},
		{
			when: "quote char found at start of record but not in first char and rFlagErrOnQInUF=true explicitly",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`a"b"`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().ErrorOnQuotesInUnquotedField(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrQuoteInUnquotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 2, record 1, field 1: " + csv.ErrQuoteInUnquotedField.Error(),
		},
		{
			when: "quote char found at start of record but not in first char and rFlagErrOnQInUF=true implicitly",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`a"b"`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrQuoteInUnquotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 2, record 1, field 1: " + csv.ErrQuoteInUnquotedField.Error(),
		},
		{
			when: "quote char found at start of doc while ErrOnNoBOM=true",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`"`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().ErrorOnNoByteOrderMarker(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNoByteOrderMarker},
			iterErrStr: csv.ErrParsing.Error() + " at byte 0, record 0, field 0: " + csv.ErrNoByteOrderMarker.Error(),
		},
		{
			when: "escape char found after the end of a quoted field",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("a234,\"\"", "\\")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrInvalidQuotedFieldEnding},
			iterErrStr: csv.ErrParsing.Error() + " at byte 8, record 1, field 2: " + csv.ErrInvalidQuotedFieldEnding.Error(),
		},
		{
			when: "escape char found after the first escape in a quoted field, but not right after",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("a234,\"\\", "z\\nice\"")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrInvalidEscSeqInQuotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 8, record 1, field 2: " + csv.ErrInvalidEscSeqInQuotedField.Error(),
		},
		{
			when: "require BOM and starts with escape char",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("\\,b,c")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ErrorOnNoByteOrderMarker(true),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNoByteOrderMarker},
			iterErrStr: csv.ErrParsing.Error() + " at byte 0, record 0, field 0: " + csv.ErrNoByteOrderMarker.Error(),
		},
		{
			when: "field count mismatch after in-field state and a read flush operation",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`1,2-flush-padding`, `,3`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().NumFields(2),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrTooManyFields},
			iterErrStr: csv.ErrParsing.Error() + " at byte 18, record 1, field 2: " + csv.ErrTooManyFields.Error() + ": field count exceeds 2",
		},
		{
			when: "field separator after quoted field with data between end of field and separator",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`"a" ,`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrInvalidQuotedFieldEnding},
			iterErrStr: csv.ErrParsing.Error() + " at byte 4, record 1, field 1: " + csv.ErrInvalidQuotedFieldEnding.Error(),
		},
		{
			when: "field separator in quoted field after escape rune",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`"\,"`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrInvalidEscSeqInQuotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 3, record 1, field 1: " + csv.ErrInvalidEscSeqInQuotedField.Error(),
		},
		{
			when: "Err on no BOM, no BOM in input, input has a row",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("a,b,c")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ErrorOnNoByteOrderMarker(true),
			},
			iterErrIs:  []error{csv.ErrParsing, csv.ErrNoByteOrderMarker},
			iterErrStr: csv.ErrParsing.Error() + " at byte 0, record 0, field 0: " + csv.ErrNoByteOrderMarker.Error(),
		},
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
			iterErrIs:  []error{csv.ErrParsing, csv.ErrInvalidEscSeqInQuotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 13, record 1, field 1: " + csv.ErrInvalidEscSeqInQuotedField.Error(),
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
			iterErrIs:  []error{csv.ErrParsing, csv.ErrInvalidEscSeqInQuotedField},
			iterErrStr: csv.ErrParsing.Error() + " at byte 5, record 1, field 1: " + csv.ErrInvalidEscSeqInQuotedField.Error(),
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
			when: "BOM required but doc starts with another multi-byte rune instead",
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

type partReader struct {
	p     []string
	i, i2 int
}

func (r *partReader) eof() bool {
	return (r.i >= len(r.p)-1 && r.i2 >= len(r.p[len(r.p)-1]))
}

func (r *partReader) Read(p []byte) (int, error) {
	var n int

	if r.eof() {
		return 0, io.EOF
	}

	if len(p) != 0 {
		n = copy(p, ([]byte(r.p[r.i]))[r.i2:])
		if n == 0 {
			r.i++
			r.i2 = 0
			n = copy(p, []byte(r.p[r.i]))
		}
		r.i2 += n
	}

	return n, nil
}

func newPartReader(p ...string) *partReader {
	return &partReader{p: p}
}
