package csv

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"iter"
	"strconv"
	"strings"
	"unicode/utf8"
	"unsafe"
)

const (
	defaultReaderBufferSize = (1 << 12) // 4096

	asciiCarriageReturn = 0x0D
	asciiLineFeed       = 0x0A
	asciiVerticalTab    = 0x0B
	asciiFormFeed       = 0x0C
	utf8NextLine        = 0x85
	utf8LineSeparator   = 0x2028
	invalidControlRune  = 0x80

	utf8ByteOrderMarker    = 0xEFBBBF
	utf16ByteOrderMarkerBE = 0xFEFF
	utf16ByteOrderMarkerLE = 0xFFFE

	rMaxOverflowNumBytes = utf8.UTFMax - 1
	// ReaderMinBufferSize is the minimum value a ReaderBufferSize
	// option will allow. It is also the minimum length for any
	// ReaderBuffer slice argument. This is exported so
	// configuration which may not be hardcoded by the utilizing
	// author can more easily define validation logic and cite
	// the reason for the limit.
	//
	// Algorithms used in this lib cannot work with a smaller buffer
	// size than this - however in general ReaderBufferSize and
	// ReaderBuffer options should be used to tune and balance mem
	// constraints with performance gained via using larger amounts
	// of buffer space.
	ReaderMinBufferSize = utf8.UTFMax + rMaxOverflowNumBytes
)

type rFlag uint16

const (
	stDone rFlag = 1 << iota
	stAfterSOR
	stEOF

	rFlagDropBOM
	rFlagErrOnNoBOM

	// rFlagStartOfDoc = either of rFlagDropBOM, rFlagErrOnNoBOM

	rFlagErrOnNLInUF
	rFlagErrOnQInUF
	rFlagComment
	rFlagQuote
	rFlagEscape
	rFlagCommentAfterSOR
	rFlagTRSEmitsRecord
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
	ErrIO         = errors.New("io error")
	ErrParsing    = errors.New("parsing error")
	ErrFieldCount = errors.New("field count error")
	ErrBadConfig  = errors.New("bad config")
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
	ErrInvalidEscSeqInQuotedField  = errors.New("invalid escape sequence in quoted field")
	ErrNewlineInUnquotedField      = errors.New("newline rune found in unquoted field")
	ErrUnexpectedQuoteAfterField   = errors.New("unexpected quote after quoted+escaped field")
	ErrUnsafeCRFileEnd             = fmt.Errorf("ended in a carriage return which must be quoted when record separator is CRLF: %w", io.ErrUnexpectedEOF)

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

// ErrorOnNoRows causes cr.Err() to return ErrNoRows should the reader
// stream terminate before any data records are parsed.
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

// TODO: in V3 alter ExpectHeaders to be vararg rather than a single slice

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

// BorrowRow alters the Row function to return the same slice instance each time with the strings inside set to different values.
//
// Only set to true if the returned row slice is never used or modified after the next call to Scan or Close. You must clone the slice if doing otherwise.
//
// See BorrowFields() if you wish to also remove allocations related to cloning strings into the slice.
//
// Please consider this to be a micro optimization in most circumstances just because is tightens the usage
// contract of the returned row in ways most would not normally consider.
func (ReaderOptions) BorrowRow(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.borrowRow = b
	}
}

// BorrowFields alters the Row function to return strings that directly reference the internal buffer
// without copying. This is UNSAFE and can lead to memory corruption if not handled properly.
//
// WARNING: Specifying this option as true while BorrowRow is false will result in an error.
//
// DANGER: Only set to true if you guarantee that field strings are NEVER used after the next call to
// Scan or Close. Otherwise, you MUST clone both the slice AND the strings within it via strings.Clone().
// Failure to do so can lead to memory corruption as the underlying buffer will be reused.
//
// Example of safe usage:
//
//	for reader.Scan() {
//	  row := reader.Row()
//	  // Process row immediately without storing references
//	  processRow(row[0], row[1])
//	}
//	if reader.Err() != nil { ... }
//
// Example of UNSAFE usage that will lead to bugs:
//
//	var savedStrings []string
//	for reader.Scan() {
//	  row := reader.Row()
//	  savedStrings = append(savedStrings, row[0]) // WRONG! Will be corrupted
//	}
//	if reader.Err() != nil { ... }
//
// This should be considered a micro-optimization only for performance-critical code paths
// where profiling has identified string copying as a bottleneck.
func (ReaderOptions) BorrowFields(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.borrowFields = b
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

// ReaderBufferSize will only accept a value greater than or equal to ReaderMinBufferSize otherwise
// an error will be thrown when creating the reader instance.
func (ReaderOptions) ReaderBufferSize(v int) ReaderOption {
	return func(cfg *rCfg) {
		cfg.rawBufSize = v
		cfg.rawBufSizeSet = true
	}
}

// ReaderBuffer will only accept a slice with a length greater than or equal to ReaderMinBufferSize
// otherwise an error will be thrown when creating the reader instance. Only up to the length of the
// slice is utilized during buffering operations. Capacity of the provided slice is not utilized in
// any way.
func (ReaderOptions) ReaderBuffer(v []byte) ReaderOption {
	return func(cfg *rCfg) {
		cfg.rawBuf = v
		cfg.rawBufSet = true
	}
}

func ReaderOpts() ReaderOptions {
	return ReaderOptions{}
}

type rCfg struct {
	headers    []string
	rawBuf     []byte
	recordBuf  []byte
	reader     io.Reader
	recordSep  [2]rune
	rawBufSize int
	numFields  int
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
	borrowFields                       bool
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
	rawBufSet                          bool
	rawBufSizeSet                      bool
}

type Reader struct {
	controlRunes string
	rawBuf       []byte
	rawIndex     int
	//
	readErr           error
	prepareRow        func() bool
	scan              func() bool
	err               error
	row               func() []string
	checkNumFields    func(errTrailer error) bool
	close             func() error
	reader            io.Reader
	recordSep         [2]rune
	recordBuf         []byte
	fieldLengths      []int
	rowBuf            []string
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
	rawNumHiddenBytes uint8
	// afterStartOfRecords is a flag set to communicate when the first records have been observed and potentially emitted
	//
	// this is useful for supporting comments after start of records explicitly with disabled being the default
	//
	// the recordIndex could also have been used for this purpose but it may have overflow issues for some input types
	// and keeping its purpose singular and disconnected from parsing management is likely ideal
	recordSepLen int8
	bitFlags     rFlag
}

func (cfg *rCfg) validate() error {

	if cfg.reader == nil {
		return ErrNilReader
	}

	if cfg.rawBufSet && cfg.rawBufSizeSet {
		return errors.New("cannot specify both ReaderBuffer and ReaderBufferSize")
	}

	if cfg.rawBufSet && len(cfg.rawBuf) < ReaderMinBufferSize {
		return errors.New("ReaderBuffer must have a length greater than or equal to " + strconv.Itoa(ReaderMinBufferSize))
	}

	if cfg.rawBufSizeSet && cfg.rawBufSize < ReaderMinBufferSize {
		return errors.New("ReaderBufferSize must be greater than or equal to " + strconv.Itoa(ReaderMinBufferSize))
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
		// TODO: instead of setting a default value that is not reachable
		// under normal circumstances, lets make sure the value cannot be used
		// until after the separator is fully resolved by switching handlers.
		//
		// note that there is a problem
		//
		// the original loop design is not re-entrant whenever the record sep is discovered
		// so loops and older strategies continue
		//
		// current generated code also allows the loops to continue by keeping the runtime discovery checks rather
		// than the compile time optimized strategies
		cfg.recordSep = [2]rune{invalidControlRune, 0}
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
			if cfg.quoteSet && (cfg.quote == asciiCarriageReturn || cfg.quote == asciiLineFeed) {
				return errors.New("invalid record separator and quote combination")
			}
			if cfg.fieldSeparator == asciiCarriageReturn || cfg.fieldSeparator == asciiLineFeed {
				return errors.New("invalid record separator and field separator combination")
			}
			if cfg.commentSet && (cfg.comment == asciiCarriageReturn || cfg.comment == asciiLineFeed) {
				return errors.New("invalid record separator and comment combination")
			}
			if cfg.escapeSet && (cfg.escape == asciiCarriageReturn || cfg.escape == asciiLineFeed) {
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

	if cfg.borrowFields && !cfg.borrowRow {
		return errors.New("field borrowing cannot be enabled without enabling row borrowing")
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

	// a mode affects what runes are relevant
	// and the reachable set of control-rune handlers
	// because of that change to relevancy
	//
	// it may not be ideal to have the switch statements
	// change dynamically as that could lead to cache
	// misses
	//
	// however at least dynamically adjusting the set
	// of control runes should be effective and likely
	// not cause cache misses if allocated in one block
	//
	// TODO: implement it and see if it is more efficient
	//
	// ---
	//
	// modes:
	// - in-comment
	// - in-quoted-field
	// - in-escape

	var controlRunes []rune
	if cfg.recordSepLen == 0 {
		var buf [11]rune
		controlRunes = append(buf[:0], cfg.fieldSeparator)
	} else {
		var buf [7]rune
		controlRunes = append(buf[:0], cfg.fieldSeparator)
	}

	var bitFlags rFlag
	if cfg.trsEmitsRecord {
		bitFlags |= rFlagTRSEmitsRecord
	}
	if cfg.quoteSet {
		controlRunes = append(controlRunes, cfg.quote)
		bitFlags |= rFlagQuote
	} else {
		cfg.quote = invalidControlRune
	}
	if cfg.escapeSet {
		controlRunes = append(controlRunes, cfg.escape)
		bitFlags |= rFlagEscape
	} else {
		cfg.escape = invalidControlRune
	}
	if cfg.commentSet {
		controlRunes = append(controlRunes, cfg.comment)
		bitFlags |= rFlagComment
	} else {
		cfg.comment = invalidControlRune
	}
	if cfg.removeByteOrderMarker {
		bitFlags |= rFlagDropBOM
	}
	if cfg.errOnNoByteOrderMarker {
		bitFlags |= rFlagErrOnNoBOM
	}
	if cfg.commentsAllowedAfterStartOfRecords {
		bitFlags |= rFlagCommentAfterSOR
	}
	if cfg.errOnQuotesInUnquotedField {
		bitFlags |= rFlagErrOnQInUF
	}

	if cfg.recordSepLen != 0 {
		controlRunes = append(controlRunes, cfg.recordSep[0])
	}

	if cfg.errOnNewlineInUnquotedField {
		bitFlags |= rFlagErrOnNLInUF

		crs := []byte(string(controlRunes))

		if !bytes.Contains(crs, []byte{asciiCarriageReturn}) {
			controlRunes = append(controlRunes, asciiCarriageReturn)
		}

		if !bytes.Contains(crs, []byte{asciiLineFeed}) {
			controlRunes = append(controlRunes, asciiLineFeed)
		}
	}

	if cfg.recordSepLen == 0 {
		allPossibleNLRunes := []rune{asciiCarriageReturn, asciiLineFeed, asciiVerticalTab, asciiFormFeed, utf8NextLine, utf8LineSeparator}
		crs := []byte(string(controlRunes))

		for _, r := range allPossibleNLRunes {
			if bytes.Contains(crs, []byte(string(r))) {
				continue
			}

			controlRunes = append(controlRunes, r)
			crs = []byte(string(controlRunes))
		}
	}

	var rowBuf []string
	if cfg.numFields > 0 {
		rowBuf = make([]string, cfg.numFields)
	}

	cr := Reader{
		controlRunes:   string(controlRunes),
		rawBuf:         cfg.rawBuf[0:0:len(cfg.rawBuf)],
		quote:          cfg.quote,
		escape:         cfg.escape,
		numFields:      cfg.numFields,
		fieldSeparator: cfg.fieldSeparator,
		comment:        cfg.comment,
		headers:        headers,
		recordBuf:      cfg.recordBuf[0:0:len(cfg.recordBuf)],
		rowBuf:         rowBuf,
		recordSep:      cfg.recordSep,
		recordSepLen:   cfg.recordSepLen,
		bitFlags:       bitFlags,
	}

	cr.initPipeline(cfg)

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

// streamParsingErr performs a side effect of increasing the byteIndex of state
//
// therefore it should never be called more than once in an invocation to Scan
// and it should always be followed with logic that short circuits the scan
// and returns false for the scan operation
func (r *Reader) streamParsingErr(err error) {
	r.setDone()

	// r.byteIndex++ moves from control byte idx
	// to next byte idx
	//
	// this matches v1 flow
	r.byteIndex++

	r.parsingErr(err)
}

func (r *Reader) ioErr(err error) {
	if r.err == nil {
		recordIndex, fieldIndex := r.humanIndexes()
		r.err = newIOError(r.byteIndex, recordIndex, fieldIndex, err)
	}
}

func (r *Reader) setRowBorrowedAndFieldsCloned() {
	if r.rowBuf == nil {
		r.row = func() []string {
			if r.fieldLengths == nil || len(r.fieldLengths) != r.numFields || r.err != nil {
				return nil
			}

			r.rowBuf = make([]string, len(r.fieldLengths))

			r.row = func() []string {
				if len(r.fieldLengths) != r.numFields || r.err != nil {
					return nil
				}

				strBuf := string(r.recordBuf)

				var p int
				for i, s := range r.fieldLengths {
					r.rowBuf[i] = strBuf[p : p+s]
					p += s
				}

				return r.rowBuf
			}

			return r.row()
		}

		return
	}

	r.row = func() []string {
		if len(r.fieldLengths) != r.numFields || r.err != nil {
			return nil
		}

		strBuf := string(r.recordBuf)

		var p int
		for i, s := range r.fieldLengths {
			r.rowBuf[i] = strBuf[p : p+s]
			p += s
		}

		return r.rowBuf
	}
}

func (r *Reader) setRowBorrowedAndFieldsBorrowed() {
	if r.rowBuf == nil {
		r.row = func() []string {
			if r.fieldLengths == nil || len(r.fieldLengths) != r.numFields || r.err != nil {
				return nil
			}

			r.rowBuf = make([]string, len(r.fieldLengths))

			r.row = func() []string {
				if len(r.fieldLengths) != r.numFields || r.err != nil {
					return nil
				}

				var p int
				for i, s := range r.fieldLengths {
					if s == 0 {
						r.rowBuf[i] = ""
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
					r.rowBuf[i] = unsafe.String(&r.recordBuf[p], s)

					p += s
				}

				return r.rowBuf
			}

			return r.row()
		}

		return
	}

	r.row = func() []string {
		if len(r.fieldLengths) != r.numFields || r.err != nil {
			return nil
		}

		var p int
		for i, s := range r.fieldLengths {
			if s == 0 {
				r.rowBuf[i] = ""
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
			r.rowBuf[i] = unsafe.String(&r.recordBuf[p], s)

			p += s
		}

		return r.rowBuf
	}
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
		r.bitFlags |= stAfterSOR
		r.checkNumFields = r.defaultCheckNumFields
		return true
	}

	return false
}

// errTrailer is unused since we're discovering the row length
func (r *Reader) checkNumFieldsWithDiscovery(errTrailer error) bool {
	r.numFields = len(r.fieldLengths)
	r.bitFlags |= stAfterSOR
	r.checkNumFields = r.defaultCheckNumFields
	return true
}

func (r *Reader) resetRecordBuffers() {
	r.fieldLengths = r.fieldLengths[:0]
	r.fieldStart = 0
	r.recordBuf = r.recordBuf[:0]
}

func (r *Reader) zeroRecordBuffers() {
	{
		v := r.rawBuf[:cap(r.rawBuf)]
		for i := range v {
			v[i] = 0
		}
	}

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

	if r.rowBuf != nil {
		v := r.rowBuf[:cap(r.rowBuf)]
		for i := range v {
			v[i] = ""
		}
	}

	r.resetRecordBuffers()
}

func (r *Reader) defaultScan() bool {

	r.resetRecordBuffers()

	return r.prepareRow()
}

func (r *Reader) initPipeline(cfg rCfg) {

	if !cfg.clearMemoryAfterFree {
		r.close = r.defaultClose
		r.prepareRow = r.prepareRow_memclearOff
	} else {
		r.close = r.closeWithMemClear
		r.prepareRow = r.prepareRow_memclearOn
	}

	if cfg.rawBufSizeSet {
		r.rawBuf = make([]byte, 0, cfg.rawBufSize)
	} else if !cfg.rawBufSet {
		r.rawBuf = make([]byte, 0, defaultReaderBufferSize)
	}

	if cfg.initialRecordBufferSize > 0 {
		r.recordBuf = make([]byte, 0, cfg.initialRecordBufferSize)
	}

	if (r.bitFlags & (rFlagDropBOM | rFlagErrOnNoBOM)) == 0 {
		r.state = rStateStartOfRecord
	}

	r.reader = cfg.reader

	if r.numFields == -1 {
		r.checkNumFields = r.checkNumFieldsWithDiscovery
	} else {
		if r.numFields > 0 {
			r.fieldLengths = make([]int, 0, r.numFields)
		}
		if (r.bitFlags & rFlagComment) != 0 {
			r.checkNumFields = r.checkNumFieldsWithStartOfRecordTracking
		} else {
			r.checkNumFields = r.defaultCheckNumFields
		}
	}

	if !cfg.borrowRow {
		r.row = r.defaultRow
	} else if cfg.borrowFields {
		r.setRowBorrowedAndFieldsBorrowed()
	} else {
		r.setRowBorrowedAndFieldsCloned()
	}

	// letting any valid utf8 end of line act as the record separator
	// also building preflight and post-flight pipeline dynamically
	r.scan = r.defaultScan

	headersHandled := true
	checkHeadersMatch := (r.headers != nil)
	if checkHeadersMatch || cfg.removeHeaderRow || cfg.trimHeaders {
		headersHandled = false
		trimHeaders := cfg.trimHeaders
		removeHeaderRow := cfg.removeHeaderRow
		next := r.scan
		r.scan = func() bool {
			r.scan = next

			if !r.scan() {
				return false
			}

			if trimHeaders {
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

			if !removeHeaderRow {
				return true
			}

			r.resetRecordBuffers()

			return r.scan()
		}
	}

	if r.headers == nil && !cfg.errOnNoRows {
		return
	}

	// verify that true is returned at least once
	// using a slip closure
	{
		errOnNoRows := cfg.errOnNoRows
		next := r.scan
		r.scan = func() bool {
			r.scan = next
			v := r.scan()
			if !v {
				if !headersHandled {
					r.parsingErr(ErrNoHeaderRow)
				} else if errOnNoRows {
					r.parsingErr(ErrNoRows)
				}
			}
			return v
		}
	}
}

// fieldNumOverflow invokes streamParsingErr if
// the current field length currently matches configured
// max limit
//
// all caveats of calling streamParsingErr apply here including
// side effects caused by streamParsingErr when return value is
// true
//
// only call when the state machine would increment the number
// of fields when processing the stream and short circuit
// that processes - returning the error path - when the result
// of this function is true
func (r *Reader) fieldNumOverflow() bool {
	if len(r.fieldLengths) == r.numFields {
		r.streamParsingErr(tooManyFieldsErr(r.numFields))
		return true
	}

	return false
}

func (r *Reader) setDone() {
	if (r.bitFlags & stDone) != 0 {
		return
	}
	r.bitFlags |= stDone

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

func (r *Reader) handleEOF() bool {
	// r.done is always true when this function is called

	// check if we're in a terminal state otherwise error
	// there is no new character to process
	switch r.state {
	case rStateStartOfDoc:
		if (r.bitFlags & rFlagErrOnNoBOM) != 0 {
			r.parsingErr(errors.Join(ErrNoByteOrderMarker, io.ErrUnexpectedEOF))
			return false
		}

		r.state = rStateStartOfRecord // might be removable, but leaving in because it's not a hot path and it's good practice to ensure the state machine is fully deterministic
		fallthrough
	case rStateStartOfRecord:
		if (r.bitFlags&rFlagTRSEmitsRecord) != 0 && r.numFields == 1 {
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
		return r.checkNumFields(io.ErrUnexpectedEOF)
	}

	panic(panicUnknownReaderStateDuringEOF)
}

// endsInValidUTF8 should get inlined by the compiler
//
// keep it small and only one function call
func endsInValidUTF8(p []byte) bool {
	r, s := utf8.DecodeLastRune(p)
	return (r != utf8.RuneError || s > 1)
}
