package csv

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"unicode/utf8"
	"unsafe"
)

// TODO: add an option to strip a starting utf8/utf16 byte order marker
// TODO: support utf16?

const (
	asciiCarriageReturn = 0x0D
	asciiLineFeed       = 0x0A
	asciiVerticalTab    = 0x0B
	asciiFormFeed       = 0x0C
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
	rStateUnusedUpperBound
)

type posErrType uint8

const (
	posErrTypeIO posErrType = iota + 1
	posErrTypeParsing
)

func (et posErrType) String() string {
	return []string{"io", "parsing"}[et-1]
}

var (
	ErrUnexpectedHeaderRowContents   = errors.New("header row values do not match expectations")
	ErrBadRecordSeparator            = errors.New("record separator can only be one rune long or \"\r\n\"")
	ErrIncompleteQuotedField         = fmt.Errorf("incomplete quoted field: %w", UnexpectedEOFError{})
	ErrUnexpectedEndOfEscapeSequence = fmt.Errorf("expecting an escaped character in a quoted field: %w", UnexpectedEOFError{})
	ErrBadEscapeSequence             = errors.New("found escape character not followed by a quote or another escape character")
	ErrBadEscapeAtStartOfRecord      = errors.New("escape character found outside the context of a quoted field when expecting the start of a record")
	ErrBadEscapeAtStartOfField       = errors.New("escape character found outside the context of a quoted field when expecting the start of a field")
	ErrBadEscapeInUnquotedField      = errors.New("escape character found outside the context of a quoted field when processing an unquoted field")
	ErrInvalidQuotedFieldEnding      = errors.New("unexpected character found after end of quoted field") // expecting field separator, record separator, quote char, or end of file if field count matches expectations
	ErrNoHeaderRow                   = fmt.Errorf("no header row: %w", UnexpectedEOFError{})
	ErrNoRows                        = fmt.Errorf("no rows: %w", UnexpectedEOFError{})
	// config errors
	ErrNilReader = errors.New("nil reader")
)

type UnexpectedEOFError struct{}

func (e UnexpectedEOFError) Error() string {
	return io.ErrUnexpectedEOF.Error()
}

type posTracedErr struct {
	err                                error
	byteIndex, recordIndex, fieldIndex uint
	errType                            posErrType
}

func (e posTracedErr) Error() string {
	return fmt.Sprintf("%s error at byte %d, record %d, field %d: %s", e.errType, e.byteIndex+1, e.recordIndex+1, e.fieldIndex+1, e.err.Error())
}

func (e posTracedErr) Unwrap() error {
	return e.err
}

type IOError struct {
	posTracedErr
}

func newIOError(byteIndex, recordIndex, fieldIndex uint, err error) IOError {
	return IOError{posTracedErr{
		errType:     posErrTypeIO,
		err:         err,
		byteIndex:   byteIndex,
		recordIndex: recordIndex,
		fieldIndex:  fieldIndex,
	}}
}

type ParsingError struct {
	posTracedErr
}

func newParsingError(byteIndex, recordIndex, fieldIndex uint, err error) ParsingError {
	return ParsingError{posTracedErr{
		errType:     posErrTypeParsing,
		err:         err,
		byteIndex:   byteIndex,
		recordIndex: recordIndex,
		fieldIndex:  fieldIndex,
	}}
}

type ConfigurationError struct {
	err error
}

func (e ConfigurationError) Error() string {
	return e.err.Error()
}

func (e ConfigurationError) Unwrap() error {
	return e.err
}

type ErrFieldCountMismatch struct {
	exp, act int
}

func (e ErrFieldCountMismatch) Error() string {
	return fmt.Sprintf("expected %d fields but found %d instead", e.exp, e.act)
}

func fieldCountMismatchErr(exp, act int) ErrFieldCountMismatch {
	return ErrFieldCountMismatch{exp, act}
}

type ErrTooManyFields struct {
	exp int
}

func (e ErrTooManyFields) Error() string {
	return fmt.Sprintf("more than %d fields found in record", e.exp)
}

func tooManyFieldsErr(exp int) ErrTooManyFields {
	return ErrTooManyFields{exp}
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

// TrimHeaders causes the first row to be recognized as a header row and all values are returned with whitespace trimmed.
func (readerOpts) TrimHeaders(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.trimHeaders = b
	}
}

// ExpectHeaders causes the first row to be recognized as a header row.
//
// If the slice of header values does not match then the reader will error.
func (readerOpts) ExpectHeaders(h []string) ReaderOption {
	return func(cfg *rCfg) {
		cfg.headers = h
	}
}

// RemoveHeaderRow causes the first row to be recognized as a header row.
//
// The row will be skipped over by Scan() and will not be returned by Row().
func (readerOpts) RemoveHeaderRow(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.removeHeaderRow = b
	}
}

// func (readerOpts) RemoveByteOrderMarker(b bool) ReaderOption {
// 	return func(cfg *rCfg) {
// 		cfg.removeByteOrderMarker = b
// 	}
// }

// func (readerOpts) ErrOnNoByteOrderMarker(b bool) ReaderOption {
// 	return func(cfg *rCfg) {
// 		cfg.errOnNoByteOrderMarker = b
// 	}
// }

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

func (readerOpts) Escape(r rune) ReaderOption {
	return func(cfg *rCfg) {
		cfg.escape = r
		cfg.escapeSet = true
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
// But that feels like exposing plumbing rather than offering porcelain.

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
	escape                  rune
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
	escapeSet               bool
	// removeByteOrderMarker   bool
	// errOnNoByteOrderMarker  bool
	//
	// errorOnBadQuotedFieldEndings bool // TODO: support relaxing this check?
}

func (cfg *rCfg) validate() error {

	if cfg.reader == nil {
		return ErrNilReader
	}

	if cfg.headers != nil {
		if len(cfg.headers) == 0 {
			return errors.New("empty set of headers expected")
		}
		if !cfg.numFieldsSet {
			cfg.numFields = len(cfg.headers)
		} else if cfg.numFields != len(cfg.headers) {
			return errors.New("explicitly specified NumFields does not match length of ExpectHeaders list")
		}
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
			if !validUtf8Rune(cfg.recordSep[0]) {
				return errors.New("invalid record separator value")
			}
			if cfg.quoteSet && cfg.recordSep[0] == cfg.quote {
				return errors.New("invalid record separator and quote combination")
			}
			if cfg.recordSep[0] == cfg.fieldSeparator {
				return errors.New("invalid record separator and field separator combination")
			}
			if cfg.commentSet && cfg.recordSep[0] == cfg.comment {
				return errors.New("invalid record separator and quote combination")
			}
		}
	}

	if !validUtf8Rune(cfg.fieldSeparator) {
		return errors.New("invalid field separator value")
	}

	if cfg.quoteSet {
		if !validUtf8Rune(cfg.quote) {
			return errors.New("invalid quote value")
		}
		// if escape would behave just like quote alone
		// then just have quote set
		if cfg.escapeSet && cfg.escape == cfg.quote {
			cfg.escapeSet = false
		}

		if cfg.commentSet && cfg.quote == cfg.comment {
			return errors.New("invalid comment and quote combination")
		}

		if cfg.fieldSeparator == cfg.quote {
			return errors.New("invalid field separator and quote combination")
		}
	}

	if cfg.escapeSet {
		if !validUtf8Rune(cfg.escape) {
			return errors.New("invalid escape value")
		}

		if cfg.commentSet && cfg.escape == cfg.comment {
			return errors.New("invalid comment and escape combination")
		}

		if cfg.fieldSeparator == cfg.escape {
			return errors.New("invalid field separator and escape combination")
		}

		if !cfg.quoteSet {
			return errors.New("escape can only be specified when quote is also specified")
		}
	}

	if cfg.commentSet {
		if !validUtf8Rune(cfg.comment) {
			return errors.New("invalid escape value")
		}

		if cfg.fieldSeparator == cfg.escape {
			return errors.New("invalid field separator and escape combination")
		}
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
		return nil, ConfigurationError{err}
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

	var recordIndex, fieldIndex, byteIndex uint

	parsingErr := func(err error) ParsingError {
		return newParsingError(byteIndex, recordIndex, fieldIndex, err)
	}

	ioErr := func(err error) IOError {
		return newIOError(byteIndex, recordIndex, fieldIndex, err)
	}

	numFields := cfg.numFields

	var fieldStart int
	var recordBuf bytes.Buffer

	var done bool
	var fieldLengths []int
	if cfg.borrowRow {
		r.row = func() []string {
			if fieldLengths == nil || len(fieldLengths) != numFields {
				return nil
			}

			row := make([]string, len(fieldLengths))
			r.row = func() []string {
				buf := recordBuf.Bytes()
				p := 0
				for i, s := range fieldLengths {
					if s == 0 {
						row[i] = ""
						continue
					}
					row[i] = unsafe.String(&buf[p], s)
					p += s
				}
				return row
			}
			return r.row()
		}
	} else {
		r.row = func() []string {
			buf := recordBuf.Bytes()
			p := 0
			row := make([]string, len(fieldLengths))
			for i, s := range fieldLengths {
				if s == 0 {
					row[i] = ""
					continue
				}
				row[i] = strings.Clone(unsafe.String(&buf[p], s))
				p += s
			}
			return row
		}
	}

	var prepareRow func() bool

	in := bufio.NewReader(cfg.reader)

	state := rStateStartOfRecord

	checkNumFields := func() bool {
		if len(fieldLengths) == numFields {
			return true
		}

		done = true
		r.err = parsingErr(fieldCountMismatchErr(numFields, len(fieldLengths)))
		return false
	}
	if numFields == -1 {
		next := checkNumFields
		checkNumFields = func() bool {
			numFields = len(fieldLengths)
			checkNumFields = next
			return true
		}
	}

	// TODO: turn off allowing comments after first record/header line encountered?
	// TODO: must handle zero columns case in some fashion
	// TODO: how about ignoring empty newlines encountered before header or data rows?
	// TODO: how about ignoring multiple empty newlines at the end of the document? (probably
	// okay to do if expected field count is greater than 1, field content overlapping with a record separator should be quoted)

	fieldNumOverflow := func() bool {
		if len(fieldLengths) == numFields {
			done = true
			r.err = parsingErr(tooManyFieldsErr(numFields))
			return true
		}
		return false
	}

	nextRuneIsLF := func() bool {
		c, size, err := in.ReadRune()
		if size <= 0 {
			done = true
			if !errors.Is(err, io.EOF) {
				r.err = ioErr(err)
			}
			return false
		}

		if size == 1 && (c == asciiLineFeed) {
			// advance the position indicator
			byteIndex += uint(size)

			if err != nil {
				done = true
				if !errors.Is(err, io.EOF) {
					r.err = ioErr(err)
				}
			}
			return true
		}

		if err := in.UnreadRune(); err != nil {
			panic(err)
		}

		return false
	}

	nextRuneIsEscapeOrQuote := func() int {
		c, size, err := in.ReadRune()
		if size <= 0 {
			done = true
			if !errors.Is(err, io.EOF) {
				r.err = ioErr(err)
			}
			return -1
		}

		if c == cfg.escape {
			// advance the position indicator
			byteIndex += uint(size)

			if err != nil {
				done = true
				if !errors.Is(err, io.EOF) {
					r.err = ioErr(err)
				}
			}
			return 0
		}

		if c == cfg.quote {
			// advance the position indicator
			byteIndex += uint(size)

			if err != nil {
				done = true
				if !errors.Is(err, io.EOF) {
					r.err = ioErr(err)
				}
			}
			return 1
		}

		if err := in.UnreadRune(); err != nil {
			panic(err)
		}

		return -1
	}

	isRecordSeparatorImplForRunes := func(sep []rune) func(rune) bool {
		if len(sep) == 1 {
			v := sep[0]
			return func(c rune) bool {
				return c == v
			}
		}

		return func(c rune) bool {
			return (c == asciiCarriageReturn && nextRuneIsLF())
		}
	}

	var isRecordSeparator func(rune) bool
	if cfg.discoverRecordSeparator {
		isRecordSeparator = func(c rune) bool {
			isCarriageReturn, ok := isNewlineRune(c)
			if !ok {
				return false
			}

			if isCarriageReturn && nextRuneIsLF() {
				isRecordSeparator = isRecordSeparatorImplForRunes([]rune{asciiCarriageReturn, asciiLineFeed})
			} else {
				isRecordSeparator = isRecordSeparatorImplForRunes([]rune{c})
			}

			return true
		}
	} else {
		isRecordSeparator = isRecordSeparatorImplForRunes(cfg.recordSep)
	}

	prepareRow = func() bool {
		for !done {
			c, size, rErr := in.ReadRune()

			// advance the position indicator
			byteIndex += uint(size)

			if size == 1 && c == utf8.RuneError {
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
					recordBuf.WriteByte(b)
					state = rStateInField
				case rStateInField, rStateInQuotedField:
					recordBuf.WriteByte(b)
					// state = rStateInField
				// case rStateInQuotedField:
				// 	recordBuf.WriteByte(b)
				// 	// state = rStateInQuotedField
				case rStateEndOfQuotedField:
					if rErr == nil {
						done = true
						r.err = parsingErr(ErrInvalidQuotedFieldEnding)
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
					r.err = ioErr(rErr)
					return false
				}
				if size == 0 {
					// check if we're in a terminal state otherwise error
					// there is no new character to process
					switch state {
					case rStateInQuotedField:
						r.err = parsingErr(ErrIncompleteQuotedField)
						return false
					case rStateStartOfRecord:
						if cfg.trsEmitsRecord && numFields == 1 {
							fieldLengths = append(fieldLengths, 0)
							// fieldStart = recordBuf.Len()
							if !checkNumFields() {
								r.err = errors.Join(r.err, UnexpectedEOFError{})
								return false
							}
							return true
						}
						return false
					case rStateInLineComment:
						return false
					case rStateStartOfField:
						fieldLengths = append(fieldLengths, 0)
						// fieldStart = recordBuf.Len()
						if !checkNumFields() {
							r.err = errors.Join(r.err, UnexpectedEOFError{})
							return false
						}
						return true
					case rStateEndOfQuotedField, rStateInField:
						{
							rbl := recordBuf.Len()
							fieldLengths = append(fieldLengths, rbl-fieldStart)
							fieldStart = rbl
						}
						if !checkNumFields() {
							r.err = errors.Join(r.err, UnexpectedEOFError{})
							return false
						}
						return true
					}
				}
				// right here in the code is the only place where the runtime could loop back around where done = true and the last character
				// has been processed
			}

			switch state {
			case rStateStartOfRecord:
				switch c {
				case cfg.fieldSeparator:
					fieldLengths = append(fieldLengths, 0)
					// fielStart = recordBuf.Len()
					state = rStateStartOfField
					fieldIndex++
				default:
					if isRecordSeparator(c) {
						fieldLengths = append(fieldLengths, 0)
						// fieldStart = recordBuf.Len()
						// state = rStateStartOfRecord
						if checkNumFields() {
							fieldIndex = 0
							recordIndex++
							return true
						}
						return false
					}
					if cfg.escapeSet && c == cfg.escape {
						done = true
						r.err = parsingErr(ErrBadEscapeAtStartOfRecord)
						return false
					} else if cfg.quoteSet && c == cfg.quote {
						state = rStateInQuotedField
					} else if cfg.commentSet && c == cfg.comment {
						state = rStateInLineComment
					} else {
						recordBuf.Write(runeBytes(c))
						state = rStateInField
					}
				}
			case rStateInQuotedField:
				switch c {
				case cfg.quote:
					state = rStateEndOfQuotedField
				default:
					if cfg.escapeSet && c == cfg.escape {
						if iresp := nextRuneIsEscapeOrQuote(); iresp != -1 {
							if iresp == 1 {
								recordBuf.Write(quoteBytes)
								continue
							}
							// otherwise append escape char
						} else if done {
							if r.err == nil {
								r.err = parsingErr(ErrUnexpectedEndOfEscapeSequence)
							}
							return false
						} else {
							done = true
							r.err = parsingErr(ErrBadEscapeSequence)
							return false
						}
					}
					recordBuf.Write(runeBytes(c))
					// state = rStateInQuotedField
				}
			case rStateEndOfQuotedField:
				switch c {
				case cfg.fieldSeparator:
					{
						rbl := recordBuf.Len()
						fieldLengths = append(fieldLengths, rbl-fieldStart)
						fieldStart = rbl
					}
					if fieldNumOverflow() {
						return false
					}
					state = rStateStartOfField
					fieldIndex++
				case cfg.quote:
					if cfg.escapeSet {
						done = true
						r.err = parsingErr(ErrInvalidQuotedFieldEnding)
						return false
					}
					recordBuf.Write(quoteBytes)
					state = rStateInQuotedField
				default:
					if isRecordSeparator(c) {
						{
							rbl := recordBuf.Len()
							fieldLengths = append(fieldLengths, rbl-fieldStart)
							fieldStart = rbl
						}
						state = rStateStartOfRecord
						if checkNumFields() {
							fieldIndex = 0
							recordIndex++
							return true
						}
						return false
					}
					done = true
					r.err = parsingErr(ErrInvalidQuotedFieldEnding)
					return false
				}
			case rStateStartOfField:
				switch c {
				case cfg.fieldSeparator:
					fieldLengths = append(fieldLengths, 0)
					// fieldStart = recordBuf.Len()
					if fieldNumOverflow() {
						return false
					}
					// state = rStateStartOfField
					fieldIndex++
				default:
					if isRecordSeparator(c) {
						fieldLengths = append(fieldLengths, 0)
						// fieldStart = recordBuf.Len()
						state = rStateStartOfRecord
						if checkNumFields() {
							fieldIndex = 0
							recordIndex++
							return true
						}
						return false
					}
					if cfg.escapeSet && c == cfg.escape {
						done = true
						r.err = parsingErr(ErrBadEscapeAtStartOfField)
						return false
					} else if cfg.quoteSet && c == cfg.quote {
						state = rStateInQuotedField
					} else {
						recordBuf.Write(runeBytes(c))
						state = rStateInField
					}
				}
			case rStateInField:
				switch c {
				case cfg.fieldSeparator:
					{
						rbl := recordBuf.Len()
						fieldLengths = append(fieldLengths, rbl-fieldStart)
						fieldStart = rbl
					}
					if fieldNumOverflow() {
						return false
					}
					state = rStateStartOfField
					fieldIndex++
				default:
					if isRecordSeparator(c) {
						{
							rbl := recordBuf.Len()
							fieldLengths = append(fieldLengths, rbl-fieldStart)
							fieldStart = rbl
						}
						state = rStateStartOfRecord
						if checkNumFields() {
							fieldIndex = 0
							recordIndex++
							return true
						}
						return false
					}
					if cfg.escapeSet && cfg.escape == c {
						done = true
						r.err = parsingErr(ErrBadEscapeInUnquotedField)
						return false
					}
					recordBuf.Write(runeBytes(c))
					// state = rStateInField
				}
			case rStateInLineComment:
				if isRecordSeparator(c) {
					state = rStateStartOfRecord
					// recordIndex++ // not valid in this case because the previous state was not a record
					return prepareRow()
				}
			}
		}

		return checkNumFields()
	}

	resetRecordBuffer := func() {
		fieldLengths = fieldLengths[:0]
		fieldStart = 0
		recordBuf.Reset()
	}

	// letting any valid utf8 end of line act as the record separator
	r.scan = func() bool {
		if done {
			return false
		}

		resetRecordBuffer()

		return prepareRow()
	}

	headersHandled := true
	checkHeadersMatch := (cfg.headers != nil)
	if checkHeadersMatch || cfg.removeHeaderRow || cfg.trimHeaders {
		headersHandled = false
		next := r.scan
		r.scan = func() bool {
			r.scan = next

			if !r.scan() {
				return false
			}

			if cfg.trimHeaders {
				headerBytes := slices.Clone(recordBuf.Bytes())
				recordBuf.Reset()

				var p int
				matching := checkHeadersMatch
				for i, s := range fieldLengths {
					var field string
					if s > 0 {
						field = unsafe.String(&headerBytes[p], s)
					}
					p += s

					field = strings.TrimSpace(field)
					fieldLengths[i] = len(field)

					recordBuf.WriteString(field)

					if matching && field != cfg.headers[i] {
						matching = false
					}
				}

				if checkHeadersMatch && !matching {
					done = true
					r.err = parsingErr(ErrUnexpectedHeaderRowContents)
					return false
				}
			} else if checkHeadersMatch {
				headerBytes := recordBuf.Bytes()

				var p int
				for i, s := range fieldLengths {
					var field string
					if s > 0 {
						field = unsafe.String(&headerBytes[p], s)
					}
					p += s

					if field != cfg.headers[i] {
						done = true
						r.err = parsingErr(ErrUnexpectedHeaderRowContents)
						return false
					}
				}
			}

			headersHandled = true

			if !cfg.removeHeaderRow {
				return true
			}

			resetRecordBuffer()

			return prepareRow()
		}
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
				if !headersHandled {
					r.err = parsingErr(ErrNoHeaderRow)
				} else if cfg.errorOnNoRows {
					r.err = parsingErr(ErrNoRows)
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
	var buf [utf8.UTFMax]byte
	b := buf[:]
	n := utf8.EncodeRune(b, r)
	return b[:n]
}

func validUtf8Rune(r rune) bool {
	return r != utf8.RuneError && utf8.ValidRune(r)
}
