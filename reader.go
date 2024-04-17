package csv

import (
	"bufio"
	"errors"
	"io"
)

// TODO: add an option to strip a starting utf8 byte order marker

const (
	_asciiCarriageReturn    = 0x0D
	_asciiLineFeed          = 0x0A
	_unicodeReplacementChar = 0xFFFD
	_unicodeNextLine        = 0x85
	_unicodeLineSeparator   = 0x2028
	// _asciiVerticalTab       = 0x0B
	// _asciiFormFeed          = 0x0C
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

// ExpectHeaders does nothing at the moment except cause a panic
func (readerOpts) ExpectHeaders(h []string) ReaderOption {
	return func(cfg *rCfg) {
		cfg.headers = h
	}
}

// StripHeaders does nothing at the moment except cause a panic
func (readerOpts) StripHeaders(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.stripHeaders = b
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
	}
}

func (readerOpts) Comment(r rune) ReaderOption {
	return func(cfg *rCfg) {
		cfg.comment = r
		cfg.commentSet = true
	}
}

// RecordSeparator does nothing at the moment except cause a panic
func (readerOpts) RecordSeparator(s string) ReaderOption {
	return func(cfg *rCfg) {
		n := len(s)
		if n > 2 || (n == 2 && (s[0] != _asciiCarriageReturn || s[1] != _asciiLineFeed)) {
			cfg.err = errors.New("record separator can only be one byte long or \"\r\n\"")
		}
		cfg.recordSep = s
	}
}

func ReaderOpts() readerOpts {
	return readerOpts{}
}

type rCfg struct {
	err               error
	headers           []string
	reader            io.Reader
	recordSep         string
	numFields         int
	maxNumFields      int
	maxNumRecords     int
	maxNumRecordBytes int
	maxNumBytes       int
	fieldSeparator    rune
	quote             rune
	comment           rune
	stripHeaders      bool
	// trimHeaders       bool
	commentSet bool
	//
	// errorOnBadQuotedFieldEndings bool // TODO: support relaxing this check
}

func NewReader(options ...ReaderOption) (*Reader, error) {

	cfg := rCfg{
		numFields:      -1,
		fieldSeparator: ',',
		quote:          '"',
		// recordSep:      string([]rune{_asciiLineFeed}),
	}

	for _, f := range options {
		f(&cfg)
	}

	if err := cfg.err; err != nil {
		return nil, err
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

	var lastCWasCR bool

	// TODO: ignore comments after first record line encountered?
	// TODO: reuse row option? ( borrow checking )
	// TODO: support custom recordSep
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

	if cfg.recordSep == "" && cfg.headers == nil && !cfg.stripHeaders && cfg.maxNumFields == 0 && cfg.maxNumRecordBytes == 0 && cfg.maxNumRecords == 0 && cfg.maxNumBytes == 0 {
		// letting any valid utf8 end of line act as the record separator
		scan = func() bool {
			if done {
				return false
			}

			row = row[:0]

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
					case rStateStartOfRecord:
						field = append(field, b)
						lastCWasCR = false
						state = rStateInField
					case rStateStartOfField:
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
							fallthrough
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
						lastCWasCR = false
						state = rStateStartOfField
					case cfg.quote:
						lastCWasCR = false
						state = rStateInQuotedField
					case _asciiCarriageReturn:
						row = append(row, "")
						// field = nil
						// state = rStateStartOfRecord
						lastCWasCR = true
						return checkNumFields()
					case _unicodeNextLine, _unicodeLineSeparator:
						row = append(row, "")
						// field = nil
						lastCWasCR = false
						// state = rStateStartOfRecord
						return checkNumFields()
					case _asciiLineFeed:
						if !lastCWasCR {
							row = append(row, "")
							// field = nil
							// state = rStateStartOfRecord
						}
						lastCWasCR = false
						return checkNumFields()
					default:
						lastCWasCR = false
						if cfg.commentSet && c == cfg.comment {
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
					case _asciiCarriageReturn:
						row = append(row, string(field))
						field = nil
						state = rStateStartOfRecord
						lastCWasCR = true
						return checkNumFields()
					case _asciiLineFeed, _unicodeNextLine, _unicodeLineSeparator:
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
					case cfg.quote:
						state = rStateInQuotedField
					case _asciiCarriageReturn:
						row = append(row, string(field))
						// field = nil
						state = rStateStartOfRecord
						lastCWasCR = true
						return checkNumFields()
					case _asciiLineFeed, _unicodeNextLine, _unicodeLineSeparator:
						row = append(row, string(field))
						// field = nil
						state = rStateStartOfRecord
						return checkNumFields()
					default:
						field = append(field, runeBytes(c)...)
						state = rStateInField
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
						// case cfg.quote:
						// 	state = rStateInField
					case _asciiCarriageReturn:
						row = append(row, string(field))
						field = nil
						state = rStateStartOfRecord
						lastCWasCR = true
						return checkNumFields()
					case _asciiLineFeed, _unicodeNextLine, _unicodeLineSeparator:
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
						state = rStateStartOfRecord
						return checkNumFields()
					default:
					}
				}
			}

			return checkNumFields()
		}
		return
	}

	panic("unimplemented config selections")
}

func runeBytes(r rune) []byte {
	return []byte(string([]rune{r}))
}
