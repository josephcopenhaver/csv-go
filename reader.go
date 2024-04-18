package csv

import (
	"bufio"
	"errors"
	"io"
	"slices"
	"strings"
)

// TODO: add an option to strip a starting utf8 byte order marker

const (
	_asciiCarriageReturn    = 0x0D
	_asciiLineFeed          = 0x0A
	_asciiVerticalTab       = 0x0B
	_asciiFormFeed          = 0x0C
	_unicodeReplacementChar = 0xFFFD
	_unicodeNextLine        = 0x85
	_unicodeLineSeparator   = 0x2028
)

type rState uint8

const (
	rStateStartOfRecord rState = iota
	rStateExpectLineFeed
	rStateInQuotedField
	rStateEndOfQuotedField
	rStateStartOfField
	rStateInField
	rStateInLineComment
)

// TODO: errors should report line numbers and character index

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

// StripHeaders does nothing at the moment except cause a panic
func (readerOpts) MaxNumFields(n int) ReaderOption {
	return func(cfg *rCfg) {
		cfg.maxNumFields = n
	}
}

// StripHeaders does nothing at the moment except cause a panic
func (readerOpts) MaxNumBytes(n int) ReaderOption {
	return func(cfg *rCfg) {
		cfg.maxNumBytes = n
	}
}

// MaxNumRecords does nothing at the moment except cause a panic
func (readerOpts) MaxNumRecords(n int) ReaderOption {
	return func(cfg *rCfg) {
		cfg.maxNumRecords = n
	}
}

// MaxNumRecordBytes does nothing at the moment except cause a panic
func (readerOpts) MaxNumRecordBytes(n int) ReaderOption {
	return func(cfg *rCfg) {
		cfg.maxNumRecordBytes = n
	}
}

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

func (readerOpts) RecordSeparator(s string) ReaderOption {
	return func(cfg *rCfg) {
		runeSlice := []rune(s)
		n := len(runeSlice)
		if n > 2 || (n == 2 && (runeSlice[0] != _asciiCarriageReturn || runeSlice[1] != _asciiLineFeed)) {
			cfg.err = errors.New("record separator can only be one rune long or \"\r\n\"")
		}
		cfg.recordSep = runeSlice
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
	err                     error
	headers                 []string
	reader                  io.Reader
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
	//
	// errorOnBadQuotedFieldEndings bool // TODO: support relaxing this check
}

func NewReader(options ...ReaderOption) (*Reader, error) {

	cfg := rCfg{
		numFields:      -1,
		fieldSeparator: ',',
		recordSep:      []rune{_asciiLineFeed},
	}

	for _, f := range options {
		f(&cfg)
	}

	if err := cfg.err; err != nil {
		return nil, err
	}

	if cfg.reader == nil {
		return nil, errors.New("nil reader")
	}

	if cfg.headers != nil && len(cfg.headers) == 0 {
		return nil, errors.New("empty set of headers expected")
	}

	{
		n := len(cfg.recordSep)
		if (n == 0) == (!cfg.discoverRecordSeparator) {
			return nil, errors.New("must specify one and only one of automatic record separator discovery or a specific recordSeparator")
		}

		if n < 1 || n > 2 || n == 2 && (cfg.recordSep[0] != _asciiCarriageReturn || cfg.recordSep[1] != _asciiLineFeed) {
			return nil, errors.New("invalid record separator value")
		}
		if n == 1 {
			if cfg.quoteSet && cfg.recordSep[0] == cfg.quote {
				return nil, errors.New("invalid record separator and quote combination")
			}
			if cfg.recordSep[0] == cfg.fieldSeparator {
				return nil, errors.New("invalid record separator and field separator combination")
			}
			if cfg.recordSep[0] == _unicodeReplacementChar {
				return nil, errors.New("invalid record separator value: unicode replacement character")
			}
		}
	}

	if cfg.quoteSet && cfg.fieldSeparator == cfg.quote {
		return nil, errors.New("invalid field separator and quote combination")
	}

	cr := &Reader{}
	cr.scan, cr.row = readStrat(cfg, &cr.err)

	return cr, nil
}

func (r *Reader) Err() error {
	return r.err
}

func (r *Reader) Row() []string {
	return r.row()
}

func (r *Reader) Scan() bool {
	return r.scan()
}

func readStrat(cfg rCfg, errPtr *error) (scan func() bool, read func() []string) {

	quoteBytes := runeBytes(cfg.quote)

	var done bool
	var row []string
	read = func() []string {
		return row
	}

	var prepareRow func() bool

	in := bufio.NewReader(cfg.reader)

	state := rStateStartOfRecord
	var field []byte

	numFields := cfg.numFields

	checkNumFields := func() bool {
		// TODO: error if file is empty?
		if len(row) == numFields {
			return true
		}

		done = true
		*errPtr = newErrFieldCountMismatch()
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
				*errPtr = errors.New("header row does not match expectations")
				return false
			}

			checkNumFields = next
			if cfg.removeHeaderRow {
				if !checkNumFields() {
					return false
				}
				return prepareRow()
			}

			return checkNumFields()
		}
	}

	recordSep := []rune(cfg.recordSep)
	var prevState rState

	// TODO: ignore comments after first record line encountered?
	// TODO: reuse row option? ( borrow checking )
	// TODO: must handle zero columns case in some fashion
	// TODO: how about ignoring empty newlines encountered before header or data rows?
	// TODO: how about ignoring empty newlines in general

	rowOverflow := func() bool {
		if len(row) == numFields {
			done = true
			*errPtr = newErrTooManyFields() // TODO: cleanup
			return true
		}
		return false
	}

	//
	// important state transition logic:
	//
	// 1. when moving from start-of-record to a different state: lastCWasCR = false
	// 2. when in start-of-record and not changing then: lastCWasCR must be set accordingly
	// 3. when moving to a field-start state from a non start-of-record start: need to make sure we're not exceeding the expected field count
	// 4. when moving from a start-of-record state to any other state: need to make sure we have the right count of fields in the row
	// 5. when the end of a record/row has been found, need to make sure expected number of fields are in the row

	prepareRowWithKnownRecordSeparator := func() bool {
		for !done {
			c, size, rErr := in.ReadRune()
			if size == 1 && c == _unicodeReplacementChar {
				if err := in.UnreadRune(); err != nil {
					panic(err)
				}
				var b byte
				if v, err := in.ReadByte(); err != nil {
					panic(err)
				} else {
					b = v
				}
				// TODO: Note: all separators, comment indicators, and quote chars are guaranteed not to be invalid rune byte sequences
				// We should panic/error on initialization
				switch state {
				case rStateExpectLineFeed:
					if prevState == rStateInQuotedField {
						done = true
						*errPtr = errors.New("found an awkward byte instead of a newline character after carriage return at the end of a quoted field")
						return false
					}
					// start of record, in field, start of field
					field = append(field, runeBytes(recordSep[0])...)
					fallthrough
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
						*errPtr = newErrInvalidQuotedFieldEnding()
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
					*errPtr = rErr
					return false
				}
				if size == 0 {
					// check if we're in a terminal state otherwise error
					// there is no new character to process
					switch state {
					case rStateInQuotedField:
						*errPtr = errors.New("unexpected end of record") // TODO: extract into var or struct
						return false
					case rStateStartOfRecord, rStateInLineComment:
						// TODO: what about if zero records or headers of any kind have been parsed?
						return false
					case rStateExpectLineFeed:
						if prevState == rStateEndOfQuotedField {
							*errPtr = errors.New("reached EOF after a quoted field followed by a carriage return: expected newline character")
							return false
						}
						field = append(field, runeBytes(recordSep[0])...)
						fallthrough
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
				case recordSep[0]:
					if len(recordSep) == 1 {
						row = append(row, "")
						// field = nil
						// state = rStateStartOfRecord
						return checkNumFields()
					}
					prevState = state // rStateStartOfRecord
					state = rStateExpectLineFeed
				default:
					if cfg.quoteSet && c == cfg.quote {
						state = rStateInQuotedField
					} else if cfg.commentSet && c == cfg.comment {
						state = rStateInLineComment
					} else {
						field = append(field, runeBytes(c)...)
						state = rStateInField
					}
				}
			case rStateExpectLineFeed:
				if c == _asciiLineFeed {
					row = append(row, string(field))
					field = nil
					state = rStateStartOfRecord
					return checkNumFields()
				}

				if prevState == rStateEndOfQuotedField {
					done = true
					*errPtr = errors.New("reached unexpected character after a quoted field followed by a carriage return: expected newline character")
					return false
				}

				if err := in.UnreadRune(); err != nil {
					panic(err)
				}

				field = append(field, runeBytes(recordSep[0])...)
				state = rStateInField
				continue
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
					if rowOverflow() {
						return false
					}
					state = rStateStartOfField
				case cfg.quote:
					field = append(field, quoteBytes...)
					state = rStateInQuotedField
				case recordSep[0]:
					if len(recordSep) == 1 {
						row = append(row, string(field))
						field = nil
						state = rStateStartOfRecord
						return checkNumFields()
					}
					prevState = state // rStateEndOfQuotedField
					state = rStateExpectLineFeed
				default:
					done = true
					*errPtr = newErrInvalidQuotedFieldEnding()
					return false
				}
			case rStateStartOfField:
				switch c {
				case cfg.fieldSeparator:
					row = append(row, string(field))
					// field = nil
					if rowOverflow() {
						return false
					}
					// state = rStateStartOfField
				case recordSep[0]:
					if len(recordSep) == 1 {
						row = append(row, string(field))
						// field = nil
						state = rStateStartOfRecord
						return checkNumFields()
					}
					prevState = state // rStateStartOfField
					state = rStateExpectLineFeed
				default:
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
					if rowOverflow() {
						return false
					}
					state = rStateStartOfField
				case recordSep[0]:
					if len(recordSep) == 1 {
						row = append(row, string(field))
						field = nil
						state = rStateStartOfRecord
						return checkNumFields()
					}
					prevState = state // rStateInField
					state = rStateExpectLineFeed
				default:
					field = append(field, runeBytes(c)...)
					// state = rStateInField
				}
			case rStateInLineComment:
				switch c {
				case _asciiCarriageReturn, _asciiLineFeed, _unicodeNextLine, _unicodeLineSeparator:
					state = rStateStartOfRecord
					return checkNumFields()
				default:
				}
			}
		}

		return checkNumFields()
	}

	if cfg.discoverRecordSeparator {
		nextCharIsLF := func() bool {
			c, size, err := in.ReadRune()
			if size <= 0 {
				if err != nil {
					done = true
					if !errors.Is(err, io.EOF) {
						*errPtr = err
					}
				}
				return false
			}

			if size == 1 && (c == _asciiLineFeed) {
				if err != nil {
					done = true
					if !errors.Is(err, io.EOF) {
						*errPtr = err
					}
				}
				return true
			}

			if err := in.UnreadRune(); err != nil {
				panic(err)
			}

			return false
		}
		prepareRow = func() bool {
			for !done {
				c, size, rErr := in.ReadRune()
				if size == 1 && c == _unicodeReplacementChar {
					if err := in.UnreadRune(); err != nil {
						panic(err)
					}
					var b byte
					if v, err := in.ReadByte(); err != nil {
						panic(err)
					} else {
						b = v
					}
					// TODO: Note: all separators, comment indicators, and quote chars are guaranteed not to be invalid rune byte sequences
					// We should panic/error on initialization
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
							*errPtr = newErrInvalidQuotedFieldEnding()
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
						*errPtr = rErr
						return false
					}
					if size == 0 {
						// check if we're in a terminal state otherwise error
						// there is no new character to process
						switch state {
						case rStateInQuotedField:
							*errPtr = errors.New("unexpected end of record") // TODO: extract into var or struct
							return false
						case rStateStartOfRecord, rStateInLineComment:
							// TODO: what about if zero records or headers of any kind have been parsed?
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
					case _asciiCarriageReturn, _asciiLineFeed, _asciiVerticalTab, _asciiFormFeed, _unicodeNextLine, _unicodeLineSeparator:
						if c == _asciiCarriageReturn && nextCharIsLF() {
							recordSep = []rune{_asciiCarriageReturn, _asciiLineFeed}
						} else {
							recordSep = []rune{c}
						}
						prepareRow = prepareRowWithKnownRecordSeparator
						row = append(row, "")
						// field = nil
						// state = rStateStartOfRecord
						return checkNumFields()
					default:
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
						if rowOverflow() {
							return false
						}
						state = rStateStartOfField
					case cfg.quote:
						field = append(field, quoteBytes...)
						state = rStateInQuotedField
					case _asciiCarriageReturn, _asciiLineFeed, _asciiVerticalTab, _asciiFormFeed, _unicodeNextLine, _unicodeLineSeparator:
						if c == _asciiCarriageReturn && nextCharIsLF() {
							recordSep = []rune{_asciiCarriageReturn, _asciiLineFeed}
						} else {
							recordSep = []rune{c}
						}
						prepareRow = prepareRowWithKnownRecordSeparator
						row = append(row, string(field))
						field = nil
						state = rStateStartOfRecord
						return checkNumFields()
					default:
						done = true
						*errPtr = newErrInvalidQuotedFieldEnding()
						return false
					}
				case rStateStartOfField:
					switch c {
					case cfg.fieldSeparator:
						row = append(row, string(field))
						// field = nil
						if rowOverflow() {
							return false
						}
						// state = rStateStartOfField
					case _asciiCarriageReturn, _asciiLineFeed, _asciiVerticalTab, _asciiFormFeed, _unicodeNextLine, _unicodeLineSeparator:
						if c == _asciiCarriageReturn && nextCharIsLF() {
							recordSep = []rune{_asciiCarriageReturn, _asciiLineFeed}
						} else {
							recordSep = []rune{c}
						}
						prepareRow = prepareRowWithKnownRecordSeparator
						row = append(row, string(field))
						// field = nil
						state = rStateStartOfRecord
						return checkNumFields()
					default:
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
						if rowOverflow() {
							return false
						}
						state = rStateStartOfField
					case _asciiCarriageReturn, _asciiLineFeed, _asciiVerticalTab, _asciiFormFeed, _unicodeNextLine, _unicodeLineSeparator:
						if c == _asciiCarriageReturn && nextCharIsLF() {
							recordSep = []rune{_asciiCarriageReturn, _asciiLineFeed}
						} else {
							recordSep = []rune{c}
						}
						prepareRow = prepareRowWithKnownRecordSeparator
						row = append(row, string(field))
						field = nil
						state = rStateStartOfRecord
						return checkNumFields()
					default:
						field = append(field, runeBytes(c)...)
						// state = rStateInField
					}
				case rStateInLineComment:
					switch c {
					case _asciiCarriageReturn, _asciiLineFeed, _unicodeNextLine, _unicodeLineSeparator:
						if c == _asciiCarriageReturn && nextCharIsLF() {
							recordSep = []rune{_asciiCarriageReturn, _asciiLineFeed}
						} else {
							recordSep = []rune{c}
						}
						prepareRow = prepareRowWithKnownRecordSeparator
						state = rStateStartOfRecord
						return prepareRow()
					default:
					}
				}
			}

			return checkNumFields()
		}
	} else {
		prepareRow = prepareRowWithKnownRecordSeparator
	}

	if cfg.maxNumFields == 0 && cfg.maxNumRecordBytes == 0 && cfg.maxNumRecords == 0 && cfg.maxNumBytes == 0 {
		// letting any valid utf8 end of line act as the record separator
		scan = func() bool {
			if done {
				return false
			}

			row = row[:0]

			return prepareRow()
		}
	} else {
		panic("unimplemented config selections")
	}

	return
}

func runeBytes(r rune) []byte {
	return []byte(string([]rune{r}))
}
