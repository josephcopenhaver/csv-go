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

type rState uint8

const (
	rStateStartOfRecord rState = iota
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

	if cfg.recordSep == "" && !cfg.expectHeaders {
		// letting any valid utf8 end of line act as the record separator
		scan = func() bool {
			if done {
				return false
			}

			row = row[:0]

			var lastCWasCR bool

			for !done {
				c, size, err := in.ReadRune()
				if size == 1 && c == 0xFFFD {
					if err := in.UnreadRune(); err != nil {
						panic(err)
					}
					b, err := in.ReadByte()
					if err != nil {
						panic(err)
					}
					field = append(field, b)
					if len(row) == numFields {
						done = true
						*errPtr = newErrTooManyFields() // TODO: cleanup
						return false
					}
					if err == nil {
						continue
					}
				}
				if err != nil {
					done = true
					if !errors.Is(err, io.EOF) {
						*errPtr = err
						return false
					}
					if size == 0 {
						// check if we're in a terminal state otherwise error
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
						}
						return false
					}
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
						*errPtr = errors.New("close of field not followed by delimiter, quote, or newline") // TODO: extract into var or struct
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
