package csv

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"iter"
	"strings"
	"unicode/utf8"
	"unsafe"
)

type BufferedReader interface {
	io.Reader
	ReadRune() (r rune, size int, err error)
	UnreadRune() error
	ReadByte() (byte, error)
}

const (
	asciiCarriageReturn = 0x0D
	asciiLineFeed       = 0x0A
	asciiVerticalTab    = 0x0B
	asciiFormFeed       = 0x0C
	utf8NextLine        = 0x85
	utf8LineSeparator   = 0x2028

	utf8ByteOrderMarker    = 0xEFBBBF
	utf16ByteOrderMarkerBE = 0xFEFF
	utf16ByteOrderMarkerLE = 0xFFFE
)

type rState uint8

const (
	rStateStartOfDoc rState = iota
	rStateStartOfRecord
	rStateInQuotedField
	rStateInQuotedFieldAfterEscape
	rStateEndOfQuotedField
	rStateStartOfField
	rStateInField
	rStateInLineComment
)

const (
	panicRecordSepLen                = "invalid record separator length"
	panicUnknownReaderStateDuringEOF = "reader in unknown state when EOF encountered"
)

var (
	// classifications
	ErrIO              = errors.New("io error")
	ErrParsing         = errors.New("parsing error")
	ErrFieldCount      = errors.New("field count error")
	ErrBadConfig       = errors.New("bad config")
	ErrBadReadRuneImpl = errors.New("bad ReadRune implementation")
	// instances
	ErrTooManyFields               = errors.New("too many fields")
	ErrNotEnoughFields             = errors.New("not enough fields")
	ErrReaderClosed                = errors.New("reader closed")
	ErrUnexpectedHeaderRowContents = errors.New("header row values do not match expectations")
	ErrBadRecordSeparator          = errors.New("record separator can only be one valid utf8 rune long or \"\\r\\n\"")
	ErrIncompleteQuotedField       = fmt.Errorf("incomplete quoted field: %w", io.ErrUnexpectedEOF)
	ErrQuoteInUnquotedField        = errors.New("quote found in unquoted field")
	ErrInvalidQuotedFieldEnding    = errors.New("unexpected character found after end of quoted field") // expecting field separator, record separator, quote char, or end of file if field count matches expectations
	ErrNoHeaderRow                 = fmt.Errorf("no header row: %w", io.ErrUnexpectedEOF)
	ErrNoRows                      = fmt.Errorf("no rows: %w", io.ErrUnexpectedEOF)
	ErrNoByteOrderMarker           = errors.New("no byte order marker")
	ErrNilReader                   = errors.New("nil reader")
	ErrInvalidEscapeInQuotedField  = errors.New("invalid escape sequence in quoted field")
	ErrNewlineInUnquotedField      = errors.New("newline rune found in unquoted field")
	ErrUnexpectedQuoteAfterField   = errors.New("unexpected quote after quoted+escaped field")
	ErrBadUnreadRuneImpl           = errors.New("UnreadRune failed")
	ErrUnsafeCRFileEnd             = fmt.Errorf("ended in a carriage return which must be quoted when record separator is CRLF: %w", io.ErrUnexpectedEOF)
	// ReadByte should never fail because we're always preceding this call with UnreadRune
	//
	// it could happen if someone is trying to read concurrently or made their own bad buffered reader implementation
	ErrBadReadByteImpl = errors.New("ReadByte failed")

	errInvalidEscapeInQuotedFieldUnexpectedByte = fmt.Errorf("%w: unexpected non-UTF8 byte following escape", ErrInvalidEscapeInQuotedField)
	errInvalidEscapeInQuotedFieldUnexpectedRune = fmt.Errorf("%w: unexpected rune following escape", ErrInvalidEscapeInQuotedField)

	errNewlineInUnquotedFieldCarriageReturn = fmt.Errorf("%w: carriage return", ErrNewlineInUnquotedField)
	errNewlineInUnquotedFieldLineFeed       = fmt.Errorf("%w: line feed", ErrNewlineInUnquotedField)
)

type posTracedErr struct {
	errType                error
	err                    error
	byteIndex, recordIndex uint64
	fieldIndex             uint
}

func (e posTracedErr) Error() string {
	return fmt.Sprintf("%s at byte %d, record %d, field %d: %s", e.errType.Error(), e.byteIndex, e.recordIndex, e.fieldIndex, e.err.Error())
}

func (e posTracedErr) Unwrap() []error {
	return []error{e.errType, e.err}
}

func newIOError(byteIndex, recordIndex uint64, fieldIndex uint, err error) posTracedErr {
	return posTracedErr{
		errType:     ErrIO,
		err:         err,
		byteIndex:   byteIndex,
		recordIndex: recordIndex,
		fieldIndex:  fieldIndex,
	}
}

func newParsingError(byteIndex, recordIndex uint64, fieldIndex uint, err error) posTracedErr {
	return posTracedErr{
		errType:     ErrParsing,
		err:         err,
		byteIndex:   byteIndex,
		recordIndex: recordIndex,
		fieldIndex:  fieldIndex,
	}
}

type errNotEnoughFields struct {
	Expected, Actual int
}

func (e errNotEnoughFields) Unwrap() []error {
	return []error{ErrFieldCount, ErrNotEnoughFields}
}

func (e errNotEnoughFields) Error() string {
	return fmt.Sprintf("%s: expected %d fields but found %d", ErrNotEnoughFields.Error(), e.Expected, e.Actual)
}

func notEnoughFieldsErr(exp, act int) errNotEnoughFields {
	return errNotEnoughFields{exp, act}
}

type errTooManyFields struct {
	expected int
}

func (e errTooManyFields) Unwrap() []error {
	return []error{ErrFieldCount, ErrTooManyFields}
}

func (e errTooManyFields) Error() string {
	return fmt.Sprintf("%s: field count exceeds %d", ErrTooManyFields.Error(), e.expected)
}

func tooManyFieldsErr(exp int) errTooManyFields {
	return errTooManyFields{exp}
}

type ReaderOption func(*rCfg)

// ReaderOptions should never be instantiated manually
//
// Instead call ReaderOpts()
//
// This is only exported to allow godocs to discover the exported methods.
//
// ReaderOptions will never have exported members and the zero value is not
// part of the semver guarantee. Instantiate it incorrectly at your own peril.
//
// Calling the function is a nop that is compiled away anyways, you will not
// optimize anything at all. Use ReaderOpts()!
type ReaderOptions struct{}

func (ReaderOptions) Reader(r io.Reader) ReaderOption {
	return func(cfg *rCfg) {
		cfg.reader = r
	}
}

// ClearFreedDataMemory ensures that whenever a shared memory buffer
// that contains data goes out of scope that zero values are written
// to every byte within the buffer.
//
// This may significantly degrade performance and is recommended only
// for sensitive data or long-lived processes.
func (ReaderOptions) ClearFreedDataMemory(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.clearMemoryAfterFree = b
	}
}

func (ReaderOptions) ErrorOnNoRows(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.errOnNoRows = b
	}
}

// TrimHeaders causes the first row to be recognized as a header row and all values are returned with whitespace trimmed.
func (ReaderOptions) TrimHeaders(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.trimHeaders = b
	}
}

// ExpectHeaders causes the first row to be recognized as a header row.
//
// If the slice of header values does not match then the reader will error.
func (ReaderOptions) ExpectHeaders(h []string) ReaderOption {
	return func(cfg *rCfg) {
		cfg.headers = h
	}
}

// RemoveHeaderRow causes the first row to be recognized as a header row.
//
// The row will be skipped over by Scan() and will not be returned by Row().
func (ReaderOptions) RemoveHeaderRow(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.removeHeaderRow = b
	}
}

func (ReaderOptions) RemoveByteOrderMarker(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.removeByteOrderMarker = b
	}
}

func (ReaderOptions) ErrorOnNoByteOrderMarker(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.errOnNoByteOrderMarker = b
	}
}

func (ReaderOptions) ErrorOnQuotesInUnquotedField(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.errOnQuotesInUnquotedField = b
	}
}

func (ReaderOptions) ErrorOnNewlineInUnquotedField(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.errOnNewlineInUnquotedField = b
	}
}

// // MaxNumFields does nothing at the moment except cause a panic
// func (ReaderOptions) MaxNumFields(n int) ReaderOption {
// 	return func(cfg *rCfg) {
// 		cfg.maxNumFields = n
// 	}
// }

// // MaxNumBytes does nothing at the moment except cause a panic
// func (ReaderOptions) MaxNumBytes(n int) ReaderOption {
// 	return func(cfg *rCfg) {
// 		cfg.maxNumBytes = n
// 	}
// }

// // MaxNumRecords does nothing at the moment except cause a panic
// func (ReaderOptions) MaxNumRecords(n int) ReaderOption {
// 	return func(cfg *rCfg) {
// 		cfg.maxNumRecords = n
// 	}
// }

// // MaxNumRecordBytes does nothing at the moment except cause a panic
// func (ReaderOptions) MaxNumRecordBytes(n int) ReaderOption {
// 	return func(cfg *rCfg) {
// 		cfg.maxNumRecordBytes = n
// 	}
// }

func (ReaderOptions) FieldSeparator(r rune) ReaderOption {
	return func(cfg *rCfg) {
		cfg.fieldSeparator = r
	}
}

func (ReaderOptions) Quote(r rune) ReaderOption {
	return func(cfg *rCfg) {
		cfg.quote = r
		cfg.quoteSet = true
	}
}

// Escape is useful for specifying what character
// is used to escape a quote in a field and the literal
// escape character itself.
//
// Without specifying this option a quote character is
// expected to be escaped by it just being doubled while
// the overall field is wrapped in quote characters.
//
// This is mainly useful when processing a spark csv
// file as it does not follow strict rfc4180.
//
// So set to '\\' if you have this need.
//
// It is not valid to use this option without specifically
// setting a quote. Doing so will result in an error being
// returned on Reader creation.
func (ReaderOptions) Escape(r rune) ReaderOption {
	return func(cfg *rCfg) {
		cfg.escape = r
		cfg.escapeSet = true
	}
}

func (ReaderOptions) Comment(r rune) ReaderOption {
	return func(cfg *rCfg) {
		cfg.comment = r
		cfg.commentSet = true
	}
}

func (ReaderOptions) CommentsAllowedAfterStartOfRecords(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.commentsAllowedAfterStartOfRecords = b
	}
}

func (ReaderOptions) NumFields(n int) ReaderOption {
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
func (ReaderOptions) TerminalRecordSeparatorEmitsRecord(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.trsEmitsRecord = b
	}
}

// BorrowRow alters the row function to return the underlying string slice every time it is called rather than a copy.
//
// Only set to true if the returned row slice and field strings within it are never used after the next call to Scan or Close. You must copy the slice and copy the strings within it via strings.Copy() if doing otherwise.
//
// Please consider this to be a micro optimization in most circumstances just because is tightens the usage
// contract of the returned row in ways most would not normally consider.
func (ReaderOptions) BorrowRow(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.borrowRow = b
	}
}

func (ReaderOptions) RecordSeparator(s string) ReaderOption {
	if len(s) == 0 {
		return badRecordSeparatorRConfig
	}

	// usage of unsafe here is actually safe because v is
	// never modified and no parts of its contents exist
	// without cloning values to other parts of memory
	// past the lifecycle of this function
	//
	// It will also never be called if the len is zero,
	// just as an extra precaution.
	v := unsafe.Slice(unsafe.StringData(s), len(s))

	r1, n1 := utf8.DecodeRune(v)
	if r1 == utf8.RuneError {
		// note that even when explicitly setting to utf8.RuneError
		// we're not allowing it
		//
		// it's just not a good practice as this character has special meaning
		//
		// I'm open to a PR to enable it though should there be strong evidence to
		// need it supported
		return badRecordSeparatorRConfig
	}
	if n1 == len(v) {
		return func(cfg *rCfg) {
			cfg.recordSep[0] = r1
			cfg.recordSepLen = 1
			cfg.recordSepSet = true
		}
	}

	r2, n2 := utf8.DecodeRune(v[n1:])
	if r2 == utf8.RuneError {
		// note that even when explicitly setting to utf8.RuneError
		// we're not allowing it
		//
		// it's just not a good practice as this character has special meaning
		//
		// I'm open to a PR to enable it though should there be strong evidence to
		// need it supported
		return badRecordSeparatorRConfig
	}
	if n1+n2 == len(v) && r1 == asciiCarriageReturn && r2 == asciiLineFeed {
		return func(cfg *rCfg) {
			cfg.recordSep[0] = r1
			cfg.recordSep[1] = r2
			cfg.recordSepLen = 2
			cfg.recordSepSet = true
		}
	}

	return badRecordSeparatorRConfig
}

func (ReaderOptions) DiscoverRecordSeparator(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.discoverRecordSeparator = b
	}
}

// InitialRecordBufferSize is a hint to pre-allocate record buffer space once
// and reduce the number of re-allocations when processing a reader.
//
// Please consider this to be a micro optimization in most circumstances just
// because it's not likely that most users will know the maximum total record
// size they wish to target / be under and it's generally a better practice
// to leave these details to the go runtime to coordinate via standard
// garbage collection.
func (ReaderOptions) InitialRecordBufferSize(v int) ReaderOption {
	return func(cfg *rCfg) {
		cfg.initialRecordBufferSize = v
		cfg.initialRecordBufferSizeSet = true
	}
}

// InitialRecordBuffer is a hint to pre-allocate record buffer space once
// externally and pipe it in to reduce the number of re-allocations when
// processing a reader and reuse it at a later time after the reader is closed.
//
// This option should generally not be used. It only exists to assist with
// processing large numbers of CSV files should memory be a clear constraint.
// There is no guarantee this buffer will always be used till the end of the
// csv Reader's lifecycle.
//
// Please consider this to be a micro optimization in most circumstances just because is tightens the usage
// contract of the csv Reader in ways most would not normally consider.
func (ReaderOptions) InitialRecordBuffer(v []byte) ReaderOption {
	return func(cfg *rCfg) {
		cfg.recordBuf = v
		cfg.recordBufSet = true
	}
}

func ReaderOpts() ReaderOptions {
	return ReaderOptions{}
}

type rCfg struct {
	headers   []string
	recordBuf []byte
	reader    io.Reader
	recordSep [2]rune
	numFields int
	// maxNumFields                       int
	// maxNumRecords                      int
	// maxNumRecordBytes                  int
	// maxNumBytes                        int
	initialRecordBufferSize            int
	fieldSeparator                     rune
	quote                              rune
	escape                             rune
	comment                            rune
	recordSepLen                       int8
	quoteSet                           bool
	escapeSet                          bool
	removeHeaderRow                    bool
	discoverRecordSeparator            bool
	trimHeaders                        bool
	commentSet                         bool
	errOnNoRows                        bool
	borrowRow                          bool
	trsEmitsRecord                     bool
	numFieldsSet                       bool
	removeByteOrderMarker              bool
	errOnNoByteOrderMarker             bool
	commentsAllowedAfterStartOfRecords bool
	errOnQuotesInUnquotedField         bool
	errOnNewlineInUnquotedField        bool
	recordSepSet                       bool
	clearMemoryAfterFree               bool
	initialRecordBufferSizeSet         bool
	recordBufSet                       bool
}

type Reader struct {
	prepareRow func() bool
	scan       func() bool
	err        error
	row        func() []string
	// isRecordSeparator can set the reader state to errored
	//
	// note that it can return true and still set the error state
	// for the next iteration
	isRecordSeparator func(c rune) (bool, bool)
	checkNumFields    func(errTrailer error) bool
	close             func() error
	reader            BufferedReader
	recordSep         [2]rune
	recordBuf         []byte
	fieldLengths      []int
	headers           []string
	fieldStart        int
	numFields         int
	recordIndex       uint64
	byteIndex         uint64
	fieldIndex        uint
	quote             rune
	escape            rune
	fieldSeparator    rune
	comment           rune
	state             rState
	// afterStartOfRecords is a flag set to communicate when the first records have been observed and potentially emitted
	//
	// this is useful for supporting comments after start of records explicitly with disabled being the default
	//
	// the recordIndex could also have been used for this purpose but it may have overflow issues for some input types
	// and keeping its purpose singular and disconnected from parsing management is likely ideal
	recordSepLen                       int8
	afterStartOfRecords                bool
	commentsAllowedAfterStartOfRecords bool
	quoteSet                           bool
	escapeSet                          bool
	commentSet                         bool
	errOnQuotesInUnquotedField         bool
	errOnNewlineInUnquotedField        bool
	trsEmitsRecord                     bool
	trimHeaders                        bool
	removeHeaderRow                    bool
	errOnNoRows                        bool
	removeByteOrderMarker              bool
	errOnNoByteOrderMarker             bool
	done                               bool
}

func (cfg *rCfg) validate() error {

	if cfg.reader == nil {
		return ErrNilReader
	}

	if cfg.initialRecordBufferSizeSet && cfg.recordBufSet {
		return errors.New("initial record buffer size cannot be specified when also setting the initial record buffer")
	}

	if cfg.initialRecordBufferSizeSet && cfg.initialRecordBufferSize < 0 {
		return errors.New("initial record buffer size must be greater than or equal to zero")
	}

	if cfg.quoteSet && cfg.escapeSet && cfg.quote == cfg.escape {
		cfg.escapeSet = false
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

	if cfg.recordSepLen == -1 {
		return ErrBadRecordSeparator
	}

	if cfg.discoverRecordSeparator {
		cfg.recordSepLen = 0
		cfg.recordSep = [2]rune{}
	}

	{
		// (cfg.recordSepLen == 0 && !cfg.discoverRecordSeparator) here is a defensive check
		// should the defaults ever be changed to no record sep specified by default
		if (cfg.recordSepLen == 0 && !cfg.discoverRecordSeparator) || (cfg.discoverRecordSeparator && cfg.recordSepSet) {
			return errors.New("must specify one and only one of automatic record separator discovery or a specific recordSeparator")
		}

		switch cfg.recordSepLen {
		case 0:
			// must be discovering record separator
			// so ensure there are no overlaps between discovery and other configuration values
			if cfg.quoteSet {
				if _, ok := isNewlineRune(cfg.quote); ok {
					return errors.New("quote cannot be a discoverable newline character when record separator discovery is enabled")
				}
			}
			if _, ok := isNewlineRune(cfg.fieldSeparator); ok {
				return errors.New("field separator cannot be a discoverable newline character when record separator discovery is enabled")
			}
			if cfg.commentSet {
				if _, ok := isNewlineRune(cfg.comment); ok {
					return errors.New("comment cannot be a discoverable newline character when record separator discovery is enabled")
				}
			}
			if cfg.escapeSet {
				if _, ok := isNewlineRune(cfg.escape); ok {
					return errors.New("escape cannot be a discoverable newline character when record separator discovery is enabled")
				}
			}
		case 1:
			if cfg.quoteSet && cfg.recordSep[0] == cfg.quote {
				return errors.New("invalid record separator and quote combination")
			}
			if cfg.recordSep[0] == cfg.fieldSeparator {
				return errors.New("invalid record separator and field separator combination")
			}
			if cfg.commentSet && cfg.recordSep[0] == cfg.comment {
				return errors.New("invalid record separator and comment combination")
			}
			if cfg.escapeSet && cfg.recordSep[0] == cfg.escape {
				return errors.New("invalid record separator and escape combination")
			}
		case 2:
			if cfg.quoteSet && (cfg.quote == '\r' || cfg.quote == '\n') {
				return errors.New("invalid record separator and quote combination")
			}
			if cfg.fieldSeparator == '\r' || cfg.fieldSeparator == '\n' {
				return errors.New("invalid record separator and field separator combination")
			}
			if cfg.commentSet && (cfg.comment == '\r' || cfg.comment == '\n') {
				return errors.New("invalid record separator and comment combination")
			}
			if cfg.escapeSet && (cfg.escape == '\r' || cfg.escape == '\n') {
				return errors.New("invalid record separator and escape combination")
			}
		default:
			panic(panicRecordSepLen)
		}
	}

	if !validUtf8Rune(cfg.fieldSeparator) {
		return errors.New("invalid field separator value")
	}

	if cfg.quoteSet {
		if !validUtf8Rune(cfg.quote) {
			return errors.New("invalid quote value")
		}

		if cfg.commentSet && cfg.quote == cfg.comment {
			return errors.New("invalid comment and quote combination")
		}

		if cfg.fieldSeparator == cfg.quote {
			return errors.New("invalid field separator and quote combination")
		}
	}

	if cfg.commentSet {
		if !validUtf8Rune(cfg.comment) {
			return errors.New("invalid comment value")
		}

		if cfg.escapeSet && cfg.comment == cfg.escape {
			return errors.New("invalid comment and escape combination")
		}
	}

	if cfg.escapeSet {
		if !validUtf8Rune(cfg.escape) {
			return errors.New("invalid escape value")
		}

		if !cfg.quoteSet {
			return errors.New("escape can only be used when quoting is enabled")
		}
	}

	if cfg.numFieldsSet && cfg.numFields <= 0 {
		return errors.New("num fields must be greater than zero or not specified")
	}

	// if cfg.maxNumFields != 0 || cfg.maxNumRecordBytes != 0 || cfg.maxNumRecords != 0 || cfg.maxNumBytes != 0 {
	// 	panic("unimplemented config selections")
	// }

	return nil
}

// NewReader creates a new instance of a CSV reader which is not safe for concurrent reads.
func NewReader(options ...ReaderOption) (*Reader, error) {

	cfg := rCfg{
		numFields:                   -1,
		fieldSeparator:              ',',
		recordSep:                   [2]rune{asciiLineFeed, 0},
		recordSepLen:                1,
		errOnQuotesInUnquotedField:  true,
		errOnNewlineInUnquotedField: true,
	}

	for _, f := range options {
		f(&cfg)
	}

	if err := cfg.validate(); err != nil {
		return nil, errors.Join(ErrBadConfig, err)
	}

	var headers []string
	if len(cfg.headers) > 0 {
		headers = make([]string, len(cfg.headers))
		strLens := make([]int, len(cfg.headers))

		var buf []byte
		if cfg.trimHeaders {
			for i := range cfg.headers {
				v := strings.TrimSpace(cfg.headers[i])
				strLens[i] = len(v)
				buf = append(buf, []byte(v)...)
			}
		} else {
			for i := range cfg.headers {
				v := cfg.headers[i]
				strLens[i] = len(v)
				buf = append(buf, []byte(v)...)
			}
		}

		strBuf := string(buf)

		var p int
		for i, s := range strLens {
			headers[i] = strBuf[p : p+s]
			p += s
		}
	}

	cr := Reader{
		quote:                              cfg.quote,
		escape:                             cfg.escape,
		numFields:                          cfg.numFields,
		fieldSeparator:                     cfg.fieldSeparator,
		comment:                            cfg.comment,
		trsEmitsRecord:                     cfg.trsEmitsRecord,
		quoteSet:                           cfg.quoteSet,
		escapeSet:                          cfg.escapeSet,
		commentSet:                         cfg.commentSet,
		trimHeaders:                        cfg.trimHeaders,
		removeHeaderRow:                    cfg.removeHeaderRow,
		errOnNoRows:                        cfg.errOnNoRows,
		headers:                            headers,
		recordBuf:                          cfg.recordBuf[:0],
		recordSep:                          cfg.recordSep,
		recordSepLen:                       cfg.recordSepLen,
		removeByteOrderMarker:              cfg.removeByteOrderMarker,
		errOnNoByteOrderMarker:             cfg.errOnNoByteOrderMarker,
		commentsAllowedAfterStartOfRecords: cfg.commentsAllowedAfterStartOfRecords,
		errOnQuotesInUnquotedField:         cfg.errOnQuotesInUnquotedField,
		errOnNewlineInUnquotedField:        cfg.errOnNewlineInUnquotedField,
	}

	cr.initPipeline(cfg.reader, cfg.initialRecordBufferSize, cfg.borrowRow, cfg.discoverRecordSeparator, cfg.clearMemoryAfterFree)

	return &cr, nil
}

// Close should be called after reading all rows
// successfully from the underlying reader and checking
// the result of r.Err().
//
// Close currently always returns nil, but in the future
// it may not. It is not a substitute for checking r.Err().
//
// Should any configuration options require post-flight
// checks they will be implemented here.
//
// It will never attempt to close the underlying reader.
func (r *Reader) Close() error {
	return r.close()
}

func (r *Reader) defaultClose() error {
	r.setDone()
	r.err = ErrReaderClosed
	return nil
}

func (r *Reader) closeWithMemClear() error {
	v := r.defaultClose()
	r.zeroRecordBuffers()
	return v
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

func (r *Reader) humanIndexes() (recordIndex uint64, fieldIndex uint) {
	recordIndex = r.recordIndex
	fieldIndex = r.fieldIndex

	switch r.state {
	case rStateStartOfDoc, rStateInLineComment:
		// do nothing
	case rStateStartOfRecord:
		if r.byteIndex != 0 {
			// convert from zero index to 1 index
			recordIndex += 1
			fieldIndex += 1
		}
	default:
		// convert from zero index to 1 index
		recordIndex += 1
		fieldIndex += 1
	}

	return recordIndex, fieldIndex
}

func (r *Reader) parsingErr(err error) {
	if r.err == nil {
		recordIndex, fieldIndex := r.humanIndexes()
		r.err = newParsingError(r.byteIndex, recordIndex, fieldIndex, err)
	}
}

func (r *Reader) ioErr(err error) {
	if r.err == nil {
		recordIndex, fieldIndex := r.humanIndexes()
		r.err = newIOError(r.byteIndex, recordIndex, fieldIndex, err)
	}
}

func (r *Reader) borrowedRow() []string {
	if r.fieldLengths == nil || len(r.fieldLengths) != r.numFields || r.err != nil {
		return nil
	}

	row := make([]string, len(r.fieldLengths))

	r.row = func() []string {
		if len(r.fieldLengths) != r.numFields || r.err != nil {
			return nil
		}

		var p int
		for i, s := range r.fieldLengths {
			if s == 0 {
				row[i] = ""
				continue
			}

			// Usage of unsafe here is what empowers borrowing.
			//
			// This is why this option is NOT enabled by default
			// and never will be. The caller must assure that they
			// will never write to the backing slice or utilize it
			// beyond the next call to Row() or Close()
			//
			// It will also never be called if the len is zero,
			// just as an extra precaution.
			row[i] = unsafe.String(&r.recordBuf[p], s)

			p += s
		}

		return row
	}

	return r.row()
}

func (r *Reader) defaultRow() []string {
	if r.fieldLengths == nil || len(r.fieldLengths) != r.numFields || r.err != nil {
		return nil
	}

	r.row = r.defaultClonedRow
	return r.row()
}

func (r *Reader) defaultClonedRow() []string {
	if len(r.fieldLengths) != r.numFields || r.err != nil {
		return nil
	}

	row := make([]string, len(r.fieldLengths))
	strBuf := string(r.recordBuf)

	var p int
	for i, s := range r.fieldLengths {
		row[i] = strBuf[p : p+s]
		p += s
	}

	return row
}

// nextRuneIsLF can set the reader state to errored
//
// note that it can return true in the first param and still set the error state
// for the next iteration, in this case the second param will be false
//
// if it changes r.err to non-nil then r.done will also be true
//
// this should only be called when searching for pairs of CRLF
//
// if the second param is true then the reader should stop processing as an
// immediate (non-deferred) error has occurred
func (r *Reader) nextRuneIsLF() (bool, bool) {
	c, size, err := r.reader.ReadRune()
	if size <= 0 {
		if errors.Is(err, io.EOF) {
			// should only be true when the record separator is known to be CRLF
			// r.recordSepLen will be 2 in this case
			//
			// in addition if the state is rStateEndOfQuotedField then a less ambiguous error message
			// will be conveyed, so lets not set the error state in that case before the state machine
			// gets that chance
			if r.recordSepLen == 2 && r.state != rStateEndOfQuotedField {
				r.setDone()
				r.parsingErr(ErrUnsafeCRFileEnd)
				return false, true
			}

			// set done on next scan
			// due to the EOF encountered
			r.scan = func() bool {
				r.setDone()
				return false
			}
			return false, false
		} else {
			r.setDone()
			r.ioErr(err)
			return false, true
		}
	} else if err != nil {
		r.setDone()
		r.ioErr(errors.Join(ErrBadReadRuneImpl, err))
		return false, true
	}

	if size == 1 && (c == asciiLineFeed) {
		// advance the position indicator
		r.byteIndex += uint64(size)

		return true, false
	}

	if err := r.reader.UnreadRune(); err != nil {
		r.setDone()
		r.ioErr(errors.Join(ErrBadUnreadRuneImpl, err))
		return false, true
	}

	return false, false
}

func (r *Reader) defaultCheckNumFields(errTrailer error) bool {
	if len(r.fieldLengths) == r.numFields {
		return true
	}

	// note: it's not possible for field lengths size to exceed the number of fields target
	//
	// so if we're here, then we're missing some fields in this record

	r.setDone()
	if r.err == nil {
		r.parsingErr(notEnoughFieldsErr(r.numFields, len(r.fieldLengths)))
		if errTrailer != nil {
			r.err = errors.Join(r.err, errTrailer)
		}
	}

	return false
}

func (r *Reader) checkNumFieldsWithStartOfRecordTracking(errTrailer error) bool {
	if r.defaultCheckNumFields(errTrailer) {
		r.afterStartOfRecords = true
		r.checkNumFields = r.defaultCheckNumFields
		return true
	}

	return false
}

// errTrailer is unused since we're discovering the row length
func (r *Reader) checkNumFieldsWithDiscovery(errTrailer error) bool {
	r.numFields = len(r.fieldLengths)
	r.afterStartOfRecords = true
	r.checkNumFields = r.defaultCheckNumFields
	return true
}

func (r *Reader) isSingleRuneRecordSeparator(c rune) (bool, bool) {
	return (c == r.recordSep[0]), false
}

// isCRLFRecordSeparator can set the reader state to errored
func (r *Reader) isCRLFRecordSeparator(c rune) (bool, bool) {
	if c == asciiCarriageReturn {
		return r.nextRuneIsLF()
	}

	return false, false
}

func (r *Reader) updateIsRecordSeparatorImpl() {
	if r.recordSepLen != 2 {
		r.isRecordSeparator = r.isSingleRuneRecordSeparator
		return
	}

	r.isRecordSeparator = r.isCRLFRecordSeparator
}

// isRecordSeparatorWithDiscovery can set the reader state to errored
func (r *Reader) isRecordSeparatorWithDiscovery(c rune) (bool, bool) {
	isCarriageReturn, ok := isNewlineRune(c)
	if !ok {
		return false, false
	}

	if isCarriageReturn {
		var isSep, immediateErr bool
		if isLF, isErr := r.nextRuneIsLF(); isLF {
			isSep = true
			immediateErr = isErr

			r.recordSep = [2]rune{asciiCarriageReturn, asciiLineFeed}
			r.recordSepLen = 2
		} else {
			if isErr {
				immediateErr = true
			} else {
				isSep = true
			}

			r.recordSep = [2]rune{c, 0}
			r.recordSepLen = 1
		}

		r.updateIsRecordSeparatorImpl()

		return isSep, immediateErr
	}

	r.recordSep = [2]rune{c, 0}
	r.recordSepLen = 1

	r.updateIsRecordSeparatorImpl()

	return true, false
}

func (r *Reader) resetRecordBuffers() {
	r.fieldLengths = r.fieldLengths[:0]
	r.fieldStart = 0
	r.recordBuf = r.recordBuf[:0]
}

func (r *Reader) zeroRecordBuffers() {
	if r.fieldLengths != nil {
		v := r.fieldLengths[:cap(r.fieldLengths)]
		for i := range v {
			v[i] = 0
		}
	}

	if r.recordBuf != nil {
		v := r.recordBuf[:cap(r.recordBuf)]
		for i := range v {
			v[i] = 0
		}
	}

	r.resetRecordBuffers()
}

func (r *Reader) defaultScan() bool {

	r.resetRecordBuffers()

	return r.prepareRow()
}

func (r *Reader) initPipeline(reader io.Reader, initialRecordBufferSize int, borrowRow, discoverRecordSeparator, clearMemoryAfterFree bool) {

	if clearMemoryAfterFree {
		r.prepareRow = r.prepareRow_memclearEnabled
		r.close = r.closeWithMemClear
	} else {
		r.prepareRow = r.prepareRow_memclearDisabled
		r.close = r.defaultClose
	}

	if initialRecordBufferSize > 0 {
		r.recordBuf = make([]byte, 0, initialRecordBufferSize)
	}

	if !(r.removeByteOrderMarker || r.errOnNoByteOrderMarker) {
		r.state = rStateStartOfRecord
	}

	if v, ok := reader.(BufferedReader); ok {
		r.reader = v
	} else {
		r.reader = bufio.NewReader(reader)
	}

	if borrowRow {
		r.row = r.borrowedRow
	} else {
		r.row = r.defaultRow
	}

	if r.numFields == -1 {
		r.checkNumFields = r.checkNumFieldsWithDiscovery
	} else if r.commentSet {
		r.checkNumFields = r.checkNumFieldsWithStartOfRecordTracking
	} else {
		r.checkNumFields = r.defaultCheckNumFields
	}

	if !discoverRecordSeparator {
		r.updateIsRecordSeparatorImpl()
	} else {
		r.isRecordSeparator = r.isRecordSeparatorWithDiscovery
	}

	// letting any valid utf8 end of line act as the record separator
	// also building preflight and post-flight pipeline dynamically
	r.scan = r.defaultScan

	headersHandled := true
	checkHeadersMatch := (r.headers != nil)
	if checkHeadersMatch || r.removeHeaderRow || r.trimHeaders {
		headersHandled = false
		next := r.scan
		r.scan = func() bool {
			r.scan = next

			if !r.scan() {
				return false
			}

			if r.trimHeaders {
				headersStr := string(r.recordBuf)
				r.recordBuf = r.recordBuf[:0]

				var p int
				matching := checkHeadersMatch
				for i, s := range r.fieldLengths {
					field := headersStr[p : p+s]
					p += s

					field = strings.TrimSpace(field)
					r.fieldLengths[i] = len(field)

					r.recordBuf = append(r.recordBuf, []byte(field)...)

					if matching && field != r.headers[i] {
						matching = false
					}
				}

				if checkHeadersMatch && !matching {
					r.setDone()
					r.parsingErr(ErrUnexpectedHeaderRowContents)
					return false
				}
			} else if checkHeadersMatch {

				var p int
				for i, s := range r.fieldLengths {
					var field string
					if s > 0 {
						// usage of unsafe here is actually safe because field is
						// never modified and no parts of its contents exist
						// without cloning values to other parts of memory
						// past the lifecycle of this function
						//
						// It will also never be called if the len is zero,
						// just as an extra precaution.
						field = unsafe.String(&r.recordBuf[p], s)
					}
					p += s

					if field != r.headers[i] {
						r.setDone()
						r.parsingErr(ErrUnexpectedHeaderRowContents)
						return false
					}
				}
			}

			headersHandled = true

			if !r.removeHeaderRow {
				return true
			}

			r.resetRecordBuffers()

			return r.scan()
		}
	}

	if r.headers == nil && !r.errOnNoRows {
		return
	}

	// verify that true is returned at least once
	// using a slip closure
	{
		next := r.scan
		r.scan = func() bool {
			r.scan = next
			v := r.scan()
			if !v {
				if !headersHandled {
					r.parsingErr(ErrNoHeaderRow)
				} else if r.errOnNoRows {
					r.parsingErr(ErrNoRows)
				}
			}
			return v
		}
	}
}

func (r *Reader) fieldNumOverflow() bool {
	if len(r.fieldLengths) == r.numFields {
		r.setDone()
		r.parsingErr(tooManyFieldsErr(r.numFields))
		return true
	}

	return false
}

func (r *Reader) handleEOF() bool {

	// r.done is always true when this function is called

	// check if we're in a terminal state otherwise error
	// there is no new character to process
	switch r.state {
	case rStateStartOfDoc:
		if r.errOnNoByteOrderMarker {
			r.parsingErr(errors.Join(ErrNoByteOrderMarker, io.ErrUnexpectedEOF))
			return false
		}

		r.state = rStateStartOfRecord
		fallthrough
	case rStateStartOfRecord:
		if r.trsEmitsRecord && r.numFields == 1 {
			r.fieldLengths = append(r.fieldLengths, 0)
			// field start is unchanged because the last one was zero length
			// r.fieldStart = len(r.recordBuf)
			return r.checkNumFields(io.ErrUnexpectedEOF)
		}
		return false
	case rStateInQuotedField, rStateInQuotedFieldAfterEscape:
		r.parsingErr(ErrIncompleteQuotedField)
		return false
	case rStateInLineComment:
		return false
	case rStateStartOfField:
		r.fieldLengths = append(r.fieldLengths, 0)
		// field start is unchanged because the last one was zero length
		// r.fieldStart = len(r.recordBuf)
		return r.checkNumFields(io.ErrUnexpectedEOF)
	case rStateEndOfQuotedField, rStateInField:
		r.fieldLengths = append(r.fieldLengths, len(r.recordBuf)-r.fieldStart)
		r.fieldStart = len(r.recordBuf)
		return r.checkNumFields(io.ErrUnexpectedEOF)
	}

	panic(panicUnknownReaderStateDuringEOF)
}

func (r *Reader) setDone() {
	if r.done {
		return
	}
	r.done = true

	r.scan = func() bool {
		return false
	}
}

func (r *Reader) appendRecBuf(b ...byte) {
	oldRef := r.recordBuf

	r.recordBuf = append(r.recordBuf, b...)

	if cap(oldRef) == 0 {
		return
	}

	oldRef = oldRef[:cap(oldRef)]

	if &oldRef[0] == &(r.recordBuf[:1])[0] {
		return
	}

	for i := range oldRef {
		oldRef[i] = 0
	}
}

// IntoIter converts the reader state into an iterator.
// Calling this method more than once returns the same iterator instance.
//
// If the reader is configured with BorrowRow(true) then the resulting
// slice and field strings are only valid to use up until the next
// iteration and should not be saved to persistent memory.
//
// It is best practice to check if Err() returns a non-nil
// error after fully traversing this iterator.
//
// This is just a syntactic sugar method to work with range statements
// in go1.23 and later.
func (r *Reader) IntoIter() iter.Seq[[]string] {
	return r.iter
}

func (r *Reader) iter(yield func([]string) bool) {
	for r.scan() {
		if !yield(r.row()) {
			return
		}
	}
}

//
// helpers
//

func badRecordSeparatorRConfig(cfg *rCfg) {
	cfg.recordSepLen = -1
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

func validUtf8Rune(r rune) bool {
	return r != utf8.RuneError && utf8.ValidRune(r)
}

// TODO: write tests that cover inputs that span the isByteOrderMarker cases
//
// it is likely not all code paths are possible in the current implementation

// https://en.wikipedia.org/wiki/Byte_order_mark
func isByteOrderMarker(r uint32, size int) bool {
	switch uint32(r) {
	case utf8ByteOrderMarker:
		return size == 3
	case utf16ByteOrderMarkerLE:
		return size == 2
	case (utf16ByteOrderMarkerLE << 16):
		return size == 4
	case utf16ByteOrderMarkerBE:
		// size of 3 would typically indicate variable width charset for utf32
		// size of 4 would typically indicate all characters are 4 characters wide
		return 2 <= size && size <= 4
	default:
		return false
	}
}
