package csv

import (
	"bufio"
	"errors"
	"io"
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

type rState uint8

const (
	rStateStartOfRecord rState = iota + 1
	rStateInQuotedField
	rStateEndOfQuotedField
	rStateStartOfField
	rStateInField
	rStateInLineComment
)

type Reader struct {
	Scan func() bool
	Row  func() []string
	err  error
}

type rCfg struct {
	recordSep     string
	numCols       int
	delimiter     rune
	quote         rune
	comment       rune
	expectHeaders bool
	numColsSet    bool
	commentSet    bool
}

func NewReader(r io.Reader) *Reader {

	cfg := rCfg{
		recordSep:     "",
		numCols:       0,
		delimiter:     ',',
		quote:         '"',
		comment:       0,
		expectHeaders: false,
		numColsSet:    false,
		commentSet:    false,
	}

	cr := &Reader{}
	cr.Scan, cr.Row = readStrat(cfg, r, &cr.err)

	return cr
}

func (r *Reader) Err() error {
	return r.err
}

func readStrat(cfg rCfg, r io.Reader, errPtr *error) (scan func() bool, read func() []string) {

	quoteBytes := runeBytes(cfg.quote)

	var done bool
	var row []string
	read = func() []string {
		return row
	}

	in := bufio.NewReader(r)

	state := rStateStartOfRecord
	var field []byte

	numFields := -1
	if cfg.numColsSet {
		numFields = cfg.numCols
	}

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

	// TODO: ignore comments after first record line encountered?
	// TODO: max record length option? ( defensive programming )
	// TODO: reuse row option? ( borrow checking )
	// TODO: support custom recordSep
	// TODO: must handle zero columns case in some fashion

	if cfg.recordSep == "" && !cfg.expectHeaders {
		// letting any valid utf8 end of line act as the record separator
		scan = func() bool {
			if done {
				return false
			}

			row = row[:0]

			var lastCWasCR bool

			for !done {
				c, size, rErr := in.ReadRune()
				if size == 1 && c == 0xFFFD {
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
					// We should panic on initialization
					switch state {
					case rStateStartOfRecord, rStateStartOfField:
						if len(field) == 0 && len(row) == numFields {
							done = true
							*errPtr = newErrTooManyFields() // TODO: cleanup
							return false
						}
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
					default:
						panic("csv reader not initialized properly")
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
							return false
						case rStateEndOfQuotedField, rStateStartOfField, rStateInField:
							row = append(row, string(field))
							field = nil
							return checkNumFields()
						default:
							panic("csv reader not initialized properly")
						}
					}
					// right here in the code is the only place where the runtime could loop back around where done = true and the last character
					// has been processed
				}

				switch state {
				case rStateStartOfRecord:
					switch c {
					case cfg.delimiter:
						row = append(row, "")
						if len(row) == numFields {
							done = true
							*errPtr = newErrTooManyFields() // TODO: cleanup
							return false
						}
						// field = nil
						state = rStateStartOfField
					case cfg.quote:
						state = rStateInQuotedField
					case '\r':
						row = append(row, "")
						if len(row) == numFields {
							done = true
							*errPtr = newErrTooManyFields() // TODO: cleanup
							return false
						}
						// field = nil
						// state = rStateStartOfRecord
						lastCWasCR = true
						return checkNumFields()
					case 0x85, 0x2028:
						row = append(row, "")
						if len(row) == numFields {
							done = true
							*errPtr = newErrTooManyFields() // TODO: cleanup
							return false
						}
						// field = nil
						// state = rStateStartOfRecord
						lastCWasCR = false
						return checkNumFields()
					case '\n':
						if !lastCWasCR {
							row = append(row, "")
							if len(row) == numFields {
								done = true
								*errPtr = newErrTooManyFields() // TODO: cleanup
								return false
							}
							// field = nil
							// state = rStateStartOfRecord
						}
						lastCWasCR = false
						return checkNumFields()
					default:
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
					case cfg.delimiter:
						row = append(row, string(field))
						if len(row) == numFields {
							done = true
							*errPtr = newErrTooManyFields() // TODO: cleanup
							return false
						}
						field = nil
						state = rStateStartOfField
					case cfg.quote:
						field = append(field, quoteBytes...)
						state = rStateInQuotedField
					case '\r':
						row = append(row, string(field))
						field = nil
						state = rStateStartOfRecord
						lastCWasCR = true
						return checkNumFields()
					case '\n', 0x85, 0x2028:
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
					case cfg.delimiter:
						row = append(row, string(field))
						// field = nil
						// state = rStateStartOfField
					case cfg.quote:
						state = rStateInQuotedField
					case '\r':
						row = append(row, string(field))
						// field = nil
						state = rStateStartOfRecord
						lastCWasCR = true
						return checkNumFields()
					case '\n', 0x85, 0x2028:
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
					case cfg.delimiter:
						row = append(row, string(field))
						if len(row) == numFields {
							done = true
							*errPtr = newErrTooManyFields() // TODO: cleanup
							return false
						}
						field = nil
						state = rStateStartOfField
						// case cfg.quote:
						// 	state = rStateInField
					case '\r':
						row = append(row, string(field))
						field = nil
						state = rStateStartOfRecord
						lastCWasCR = true
						return checkNumFields()
					case '\n', 0x85, 0x2028:
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
					case '\r', '\n', 0x85, 0x2028:
						state = rStateStartOfRecord
						return checkNumFields()
					default:
					}
				default:
					panic("csv reader not initialized properly")
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
