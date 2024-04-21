package csv

import (
	"bufio"
	"errors"
	"io"
	"slices"
	"strings"
	"unicode/utf8"
)

// TODO: add an option to strip a starting utf8/utf16 byte order marker
// TODO: support utf16?

const (
	asciiCarriageReturn = 0x0D
	asciiLineFeed       = 0x0A
	asciiVerticalTab    = 0x0B
	asciiFormFeed       = 0x0C
	utf8ReplacementChar = 0xFFFD
	utf8NextLine        = 0x85
	utf8LineSeparator   = 0x2028
)

type rState uint8

const (
	rStateStartOfRecord rState = iota
	rStateInQuotedField
	rStateEndOfQuotedField
	rStateStartOfField
	rStateInField
	rStateInLineComment
)

// TODO: errors should report line numbers and character index

var ErrBadRecordSeparator = errors.New("record separator can only be one rune long or \"\r\n\"")

type ErrNoHeaderRow struct{}

func (e ErrNoHeaderRow) Error() string {
	return "no header row"
}

func newErrNoHeaderRow() ErrNoHeaderRow {
	return ErrNoHeaderRow{}
}

type ErrNoRows struct{}

func (e ErrNoRows) Error() string {
	return "no rows"
}

func newErrNoRows() ErrNoRows {
	return ErrNoRows{}
}

type ErrFieldCountMismatch struct{}

func (e ErrFieldCountMismatch) Error() string {
	return "field counts do not match between rows/config"
}

func newErrFieldCountMismatch() error {
	return ErrFieldCountMismatch{}
}

type ErrTooManyFields struct{}

func (e ErrTooManyFields) Error() string {
	return "too many fields"
}

func newErrTooManyFields() error {
	return ErrTooManyFields{}
}

type ErrInvalidQuotedFieldEnding struct{}

func (e ErrInvalidQuotedFieldEnding) Error() string {
	return "invalid char when expecting field separator, record separator, quote char, or end of file"
}

func newErrInvalidQuotedFieldEnding() error {
	return ErrInvalidQuotedFieldEnding{}
}

type Reader struct {
	scan func() bool
	row  func() []string
	err  error
}

type ReaderOption func(*rCfg)

type readerOpts struct{}

func (readerOpts) Reader(r io.Reader) ReaderOption {
	return func(cfg *rCfg) {
		cfg.reader = r
	}
}

func (readerOpts) ErrorOnNoRows(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.errorOnNoRows = b
	}
}

func (readerOpts) TrimHeaders(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.trimHeaders = b
	}
}

func (readerOpts) ExpectHeaders(h []string) ReaderOption {
	return func(cfg *rCfg) {
		cfg.headers = h
	}
}

func (readerOpts) RemoveHeaderRow(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.removeHeaderRow = b
	}
}

// // MaxNumFields does nothing at the moment except cause a panic
// func (readerOpts) MaxNumFields(n int) ReaderOption {
// 	return func(cfg *rCfg) {
// 		cfg.maxNumFields = n
// 	}
// }

// // MaxNumBytes does nothing at the moment except cause a panic
// func (readerOpts) MaxNumBytes(n int) ReaderOption {
// 	return func(cfg *rCfg) {
// 		cfg.maxNumBytes = n
// 	}
// }

// // MaxNumRecords does nothing at the moment except cause a panic
// func (readerOpts) MaxNumRecords(n int) ReaderOption {
// 	return func(cfg *rCfg) {
// 		cfg.maxNumRecords = n
// 	}
// }

// // MaxNumRecordBytes does nothing at the moment except cause a panic
// func (readerOpts) MaxNumRecordBytes(n int) ReaderOption {
// 	return func(cfg *rCfg) {
// 		cfg.maxNumRecordBytes = n
// 	}
// }

func (readerOpts) FieldSeparator(r rune) ReaderOption {
	return func(cfg *rCfg) {
		cfg.fieldSeparator = r
	}
}

func (readerOpts) Quote(r rune) ReaderOption {
	return func(cfg *rCfg) {
		cfg.quote = r
		cfg.quoteSet = true
	}
}

func (readerOpts) Comment(r rune) ReaderOption {
	return func(cfg *rCfg) {
		cfg.comment = r
		cfg.commentSet = true
	}
}

func (readerOpts) NumFields(n int) ReaderOption {
	return func(cfg *rCfg) {
		cfg.numFields = n
		cfg.numFieldsSet = true
	}
}

// TODO: what should be done if the file is empty and numFields == -1 (field count discovery mode)?
//
// how often are we expecting a file to have no headers, no predetermined field count expectation, and no
// record separator yet still expect an empty string to be yielded in a single element row because the
// writer did not end the single column value (which happens to serialize to an empty string rather
// than a quoted empty string) with a record separator/terminator?
//
// Feels like this should probably be an error explicitly returned that the calling layers can handle
// as they choose. Perhaps State() should be exposed as well as the relevant terminal state values?
// But that feels like exposing plumbing rather than offer porcelain.

// TerminalRecordSeparatorEmitsRecord only exists to acknowledge an edge case
// when processing csv documents that contain one column. If the file contents end
// in a record separator it's impossible to determine if that should indicate that
// a new record with an empty field should be emitted unless that record is enclosed
// in quotes or a config option like this exists.
//
// In most cases this should not be an issue, unless the dataset is a single column
// list that allows empty strings for some use case and the writer used to create the
// file chooses to not always write the last record followed by a record separator.
// (treating the record separator like a record terminator)
func (readerOpts) TerminalRecordSeparatorEmitsRecord(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.trsEmitsRecord = b
	}
}

// BorrowRow alters the row function to return the underlying string slice every time it is called rather than a copy.
//
// Only set to true if the returned row and field strings within it are never used after the next call to Scan.
//
// Please consider this to be a micro optimization in most circumstances.
func (readerOpts) BorrowRow(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.borrowRow = b
	}
}

func (readerOpts) RecordSeparator(s string) ReaderOption {
	return func(cfg *rCfg) {
		cfg.recordSepStr = s
		cfg.recordSepStrSet = true
	}
}

func (readerOpts) DiscoverRecordSeparator(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.discoverRecordSeparator = b
	}
}

func ReaderOpts() readerOpts {
	return readerOpts{}
}

type rCfg struct {
	headers                 []string
	reader                  io.Reader
	recordSepStr            string
	recordSep               []rune
	numFields               int
	maxNumFields            int
	maxNumRecords           int
	maxNumRecordBytes       int
	maxNumBytes             int
	fieldSeparator          rune
	quote                   rune
	comment                 rune
	quoteSet                bool
	removeHeaderRow         bool
	discoverRecordSeparator bool
	trimHeaders             bool
	commentSet              bool
	errorOnNoRows           bool
	borrowRow               bool
	recordSepStrSet         bool
	trsEmitsRecord          bool
	numFieldsSet            bool
	//
	// errorOnBadQuotedFieldEndings bool // TODO: support relaxing this check?
}

func (cfg *rCfg) validate() error {

	if cfg.reader == nil {
		return errors.New("nil reader")
	}

	if cfg.headers != nil && len(cfg.headers) == 0 {
		return errors.New("empty set of headers expected")
	}

	if cfg.recordSepStrSet {
		s := cfg.recordSepStr
		cfg.recordSepStr = ""

		numBytes := len(s)
		if numBytes == 0 {
			return ErrBadRecordSeparator
		}

		r1, n1 := utf8.DecodeRuneInString(s)
		if n1 == numBytes {
			cfg.recordSep = []rune{r1}
		} else {

			r2, n2 := utf8.DecodeRuneInString(s[n1:])
			if n1+n2 != numBytes || (r1 != asciiCarriageReturn || r2 != asciiLineFeed) {
				return ErrBadRecordSeparator
			}

			cfg.recordSep = []rune{r1, r2}
		}
	}

	{
		n := len(cfg.recordSep)
		if (n == 0) == (!cfg.discoverRecordSeparator) {
			return errors.New("must specify one and only one of automatic record separator discovery or a specific recordSeparator")
		}

		if n < 1 || n > 2 || n == 2 && (cfg.recordSep[0] != asciiCarriageReturn || cfg.recordSep[1] != asciiLineFeed) {
			return errors.New("invalid record separator value")
		}
		if n == 1 {
			if cfg.quoteSet && cfg.recordSep[0] == cfg.quote {
				return errors.New("invalid record separator and quote combination")
			}
			if cfg.recordSep[0] == cfg.fieldSeparator {
				return errors.New("invalid record separator and field separator combination")
			}
			if cfg.recordSep[0] == utf8ReplacementChar {
				return errors.New("invalid record separator value: utf8 replacement character")
			}
		}
	}

	if cfg.quoteSet && cfg.fieldSeparator == cfg.quote {
		return errors.New("invalid field separator and quote combination")
	}

	if !validUtf8Rune(cfg.fieldSeparator) {
		return errors.New("invalid field separator value")
	}

	if cfg.quoteSet && !validUtf8Rune(cfg.quote) {
		return errors.New("invalid quote value")
	}

	if cfg.commentSet && !validUtf8Rune(cfg.comment) {
		return errors.New("invalid quote value")
	}

	if cfg.numFieldsSet && cfg.numFields <= 0 {
		return errors.New("num fields must be greater than zero or not specified")
	}

	if cfg.maxNumFields != 0 || cfg.maxNumRecordBytes != 0 || cfg.maxNumRecords != 0 || cfg.maxNumBytes != 0 {
		panic("unimplemented config selections")
	}

	return nil
}

func NewReader(options ...ReaderOption) (*Reader, error) {

	cfg := rCfg{
		numFields:      -1,
		fieldSeparator: ',',
		recordSep:      []rune{asciiLineFeed},
	}

	for _, f := range options {
		f(&cfg)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	cr := &Reader{}
	cr.init(cfg)

	return cr, nil
}

func (r *Reader) Err() error {
	return r.err
}

// Row returns a slice of strings that represents a row of a dataset.
//
// Row only returns valid results after a call to Scan() return true.
// For efficiency reasons this method should not be called more than once between
// calls to Scan().
//
// If the reader is configured with BorrowRow(true) then the resulting slice and
// field strings are only valid to use up until the next call to Scan and
// should not be saved to persistent memory.
func (r *Reader) Row() []string {
	return r.row()
}

func (r *Reader) Scan() bool {
	return r.scan()
}

func (r *Reader) init(cfg rCfg) {

	quoteBytes := runeBytes(cfg.quote)

	var done bool
	var row []string
	if cfg.borrowRow {
		r.row = func() []string {
			n := len(row)
			return row[:n:n]
		}
	} else {
		r.row = func() []string {
			n := len(row)
			c := make([]string, n)
			copy(c, row)
			return c
		}
	}

	var prepareRow func() bool

	in := bufio.NewReader(cfg.reader)

	state := rStateStartOfRecord
	var field []byte

	numFields := cfg.numFields

	checkNumFields := func() bool {
		if len(row) == numFields {
			return true
		}

		done = true
		r.err = newErrFieldCountMismatch()
		return false
	}
	if numFields == -1 {
		next := checkNumFields
		checkNumFields = func() bool {
			numFields = len(row)
			checkNumFields = next
			return true
		}
	}

	var headersVerified bool
	if !(cfg.headers == nil && !cfg.removeHeaderRow && !cfg.trimHeaders) {
		next := checkNumFields
		checkNumFields = func() bool {
			if cfg.trimHeaders {
				for i := range row {
					row[i] = strings.TrimSpace(row[i])
				}
			}

			if cfg.headers != nil && !slices.Equal(cfg.headers, row) {
				done = true
				r.err = errors.New("header row does not match expectations")
				return false
			}

			checkNumFields = next
			if cfg.removeHeaderRow {
				if !checkNumFields() {
					return false
				}
				headersVerified = true
				return prepareRow()
			}

			return checkNumFields()
		}
	}

	recordSep := cfg.recordSep

	// TODO: turn off allowing comments after first record/header line encountered?
	// TODO: must handle zero columns case in some fashion
	// TODO: how about ignoring empty newlines encountered before header or data rows?
	// TODO: how about ignoring multiple empty newlines at the end of the document? (probably
	// okay to do if expected field count is greater than 1, field content overlapping with a record separator should be quoted)

	fieldNumOverflow := func() bool {
		if len(row) == numFields {
			done = true
			r.err = newErrTooManyFields() // TODO: cleanup
			return true
		}
		return false
	}

	nextCharIsLF := func() bool {
		c, size, err := in.ReadRune()
		if size <= 0 {
			if err != nil {
				done = true
				if !errors.Is(err, io.EOF) {
					r.err = err
				}
			}
			return false
		}

		if size == 1 && (c == asciiLineFeed) {
			if err != nil {
				done = true
				if !errors.Is(err, io.EOF) {
					r.err = err
				}
			}
			return true
		}

		if err := in.UnreadRune(); err != nil {
			panic(err)
		}

		return false
	}

	isRecordSeparatorImplForRunes := func(sep []rune) func(rune) bool {
		if len(sep) == 1 {
			v := sep[0]
			return func(c rune) bool {
				return c == v
			}
		}

		return func(c rune) bool {
			return (c == asciiCarriageReturn && nextCharIsLF())
		}
	}

	var isRecordSeparator func(rune) bool
	if cfg.discoverRecordSeparator {
		isRecordSeparator = func(c rune) bool {
			isCarriageReturn, ok := isNewlineRune(c)
			if !ok {
				return false
			}

			if isCarriageReturn && nextCharIsLF() {
				recordSep = []rune{asciiCarriageReturn, asciiLineFeed}
			} else {
				recordSep = []rune{c}
			}

			isRecordSeparator = isRecordSeparatorImplForRunes(recordSep)

			return true
		}
	} else {
		isRecordSeparator = isRecordSeparatorImplForRunes(recordSep)
	}

	prepareRow = func() bool {
		for !done {
			c, size, rErr := in.ReadRune()
			if size == 1 && c == utf8ReplacementChar {
				if err := in.UnreadRune(); err != nil {
					panic(err)
				}
				var b byte
				if v, err := in.ReadByte(); err != nil {
					panic(err)
				} else {
					b = v
				}
				switch state {
				case rStateStartOfRecord, rStateStartOfField:
					field = append(field, b)
					state = rStateInField
				case rStateInField, rStateInQuotedField:
					field = append(field, b)
					// state = rStateInField
				// case rStateInQuotedField:
				// 	field = append(field, b)
				// 	// state = rStateInQuotedField
				case rStateEndOfQuotedField:
					if rErr == nil {
						done = true
						r.err = newErrInvalidQuotedFieldEnding()
						return false
					}
				case rStateInLineComment:
					// state = rStateInLineComment
				}
				if rErr == nil {
					continue
				}
			}
			if rErr != nil {
				done = true
				if !errors.Is(rErr, io.EOF) {
					r.err = rErr
					return false
				}
				if size == 0 {
					// check if we're in a terminal state otherwise error
					// there is no new character to process
					switch state {
					case rStateInQuotedField:
						r.err = errors.New("unexpected end of record") // TODO: extract into var or struct
						return false
					case rStateStartOfRecord:
						if cfg.trsEmitsRecord && numFields == 1 {
							row = append(row, "")
							return checkNumFields()
						}
						return false
					case rStateInLineComment:
						return false
					case rStateEndOfQuotedField, rStateStartOfField, rStateInField:
						row = append(row, string(field))
						field = nil
						return checkNumFields()
					}
				}
				// right here in the code is the only place where the runtime could loop back around where done = true and the last character
				// has been processed
			}

			switch state {
			case rStateStartOfRecord:
				switch c {
				case cfg.fieldSeparator:
					row = append(row, "")
					// field = nil
					state = rStateStartOfField
				default:
					if isRecordSeparator(c) {
						row = append(row, "")
						// field = nil
						// state = rStateStartOfRecord
						return checkNumFields()
					}
					if cfg.quoteSet && c == cfg.quote {
						state = rStateInQuotedField
					} else if cfg.commentSet && c == cfg.comment {
						state = rStateInLineComment
					} else {
						field = append(field, runeBytes(c)...)
						state = rStateInField
					}
				}
			case rStateInQuotedField:
				switch c {
				case cfg.quote:
					state = rStateEndOfQuotedField
				default:
					field = append(field, runeBytes(c)...)
					// state = rStateInQuotedField
				}
			case rStateEndOfQuotedField:
				switch c {
				case cfg.fieldSeparator:
					row = append(row, string(field))
					field = nil
					if fieldNumOverflow() {
						return false
					}
					state = rStateStartOfField
				case cfg.quote:
					field = append(field, quoteBytes...)
					state = rStateInQuotedField
				default:
					if isRecordSeparator(c) {
						row = append(row, string(field))
						field = nil
						state = rStateStartOfRecord
						return checkNumFields()
					}
					done = true
					r.err = newErrInvalidQuotedFieldEnding()
					return false
				}
			case rStateStartOfField:
				switch c {
				case cfg.fieldSeparator:
					row = append(row, string(field))
					// field = nil
					if fieldNumOverflow() {
						return false
					}
					// state = rStateStartOfField
				default:
					if isRecordSeparator(c) {
						row = append(row, string(field))
						// field = nil
						state = rStateStartOfRecord
						return checkNumFields()
					}
					if cfg.quoteSet && c == cfg.quote {
						state = rStateInQuotedField
					} else {
						field = append(field, runeBytes(c)...)
						state = rStateInField
					}
				}
			case rStateInField:
				switch c {
				case cfg.fieldSeparator:
					row = append(row, string(field))
					field = nil
					if fieldNumOverflow() {
						return false
					}
					state = rStateStartOfField
				default:
					if isRecordSeparator(c) {
						row = append(row, string(field))
						field = nil
						state = rStateStartOfRecord
						return checkNumFields()
					}
					field = append(field, runeBytes(c)...)
					// state = rStateInField
				}
			case rStateInLineComment:
				if isRecordSeparator(c) {
					state = rStateStartOfRecord
					return prepareRow()
				}
			}
		}

		return checkNumFields()
	}

	// letting any valid utf8 end of line act as the record separator
	r.scan = func() bool {
		if done {
			return false
		}

		row = row[:0]

		return prepareRow()
	}

	if cfg.headers == nil && !cfg.errorOnNoRows {
		return
	}

	// verify that true is returned at least once
	// using a slip closure
	{
		next := r.scan
		r.scan = func() bool {
			r.scan = next
			v := r.scan()
			if !v && r.err == nil {
				if cfg.headers != nil && !headersVerified {
					r.err = newErrNoHeaderRow()
				} else if cfg.errorOnNoRows {
					r.err = newErrNoRows()
				}
			}
			return v
		}
	}
}

func isNewlineRune(c rune) (isCarriageReturn bool, ok bool) {
	switch c {
	case asciiCarriageReturn:
		return true, true
	case asciiLineFeed, asciiVerticalTab, asciiFormFeed, utf8NextLine, utf8LineSeparator:
		return false, true
	}
	return false, false
}

func runeBytes(r rune) []byte {
	return []byte(string([]rune{r}))
}

func validUtf8Rune(r rune) bool {
	v, n := utf8.DecodeRuneInString(string([]rune{r}))
	return n != 0 && r == v
}
