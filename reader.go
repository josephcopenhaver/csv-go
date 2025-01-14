package csv

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"slices"
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
	utf8ByteOrderMarker = 0xEFBBBF
)

type rState uint8

const (
	rStateStartOfDoc rState = iota
	rStateStartOfRecord
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
	ErrBadConfig                   = errors.New("bad config")
	ErrUnexpectedHeaderRowContents = errors.New("header row values do not match expectations")
	ErrBadRecordSeparator          = errors.New("record separator can only be one rune long or \"\\r\\n\"")
	ErrIncompleteQuotedField       = fmt.Errorf("incomplete quoted field: %w", io.ErrUnexpectedEOF)
	ErrInvalidQuotedFieldEnding    = errors.New("unexpected character found after end of quoted field") // expecting field separator, record separator, quote char, or end of file if field count matches expectations
	ErrNoHeaderRow                 = fmt.Errorf("no header row: %w", io.ErrUnexpectedEOF)
	ErrNoRows                      = fmt.Errorf("no rows: %w", io.ErrUnexpectedEOF)
	ErrNoByteOrderMarker           = errors.New("no byte order marker")
	ErrNilReader                   = errors.New("nil reader")
)

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

func (readerOpts) RemoveByteOrderMarker(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.removeByteOrderMarker = b
	}
}

func (readerOpts) ErrOnNoByteOrderMarker(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.errOnNoByteOrderMarker = b
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

func (readerOpts) CommentsAllowedAfterStartOfRecords(b bool) ReaderOption {
	return func(cfg *rCfg) {
		cfg.commentsAllowedAfterStartOfRecords = b
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
	if len(s) == 0 {
		return badRecordSeparatorRConfig
	}
	// usage of unsafe here is actually safe because v is
	// never modified and no parts of its contents exist
	// without cloning values to other parts of memory
	// past the lifecycle of this function
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
		}
	}

	return badRecordSeparatorRConfig
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
	headers                            []string
	reader                             io.Reader
	recordSep                          [2]rune
	numFields                          int
	maxNumFields                       int
	maxNumRecords                      int
	maxNumRecordBytes                  int
	maxNumBytes                        int
	fieldSeparator                     rune
	quote                              rune
	comment                            rune
	recordSepLen                       int8
	quoteSet                           bool
	removeHeaderRow                    bool
	discoverRecordSeparator            bool
	trimHeaders                        bool
	commentSet                         bool
	errorOnNoRows                      bool
	borrowRow                          bool
	trsEmitsRecord                     bool
	numFieldsSet                       bool
	removeByteOrderMarker              bool
	errOnNoByteOrderMarker             bool
	commentsAllowedAfterStartOfRecords bool
}

type Reader struct {
	scan              func() bool
	err               error
	row               func() []string
	isRecordSeparator func(c rune) bool
	checkNumFields    func() bool
	reader            BufferedReader
	recordSep         [2]rune
	recordBuf         []byte
	fieldLengths      []int
	headers           []string
	fieldStart        int
	numFields         int
	recordIndex       uint
	fieldIndex        uint
	byteIndex         uint
	quote             rune
	fieldSeparator    rune
	comment           rune
	done              bool
	state             rState
	// afterStartOfRecords is a flag set to communicate when the first records have been observed and potentially emitted
	//
	// this is useful for supporting comments after start of records explicitly with disabled being the default
	//
	// the recordIndex could also have been used for this purpose but it may have overflow issues for some input types
	// and keeping its purpose singular and disconnected from parsing management is likely ideal
	afterStartOfRecords                bool
	recordSepLen                       int8
	commentsAllowedAfterStartOfRecords bool
	quoteSet                           bool
	commentSet                         bool
	trsEmitsRecord                     bool
	trimHeaders                        bool
	removeHeaderRow                    bool
	errorOnNoRows                      bool
	removeByteOrderMarker              bool
	errOnNoByteOrderMarker             bool
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

	if cfg.recordSepLen == -1 {
		return ErrBadRecordSeparator
	}

	{
		n := cfg.recordSepLen
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

		if cfg.commentSet && cfg.quote == cfg.comment {
			return errors.New("invalid comment and quote combination")
		}

		if cfg.fieldSeparator == cfg.quote {
			return errors.New("invalid field separator and quote combination")
		}
	}

	if cfg.commentSet && !validUtf8Rune(cfg.comment) {
		return errors.New("invalid comment value")
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
		recordSep:      [2]rune{asciiLineFeed, 0},
		recordSepLen:   1,
	}

	for _, f := range options {
		f(&cfg)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrBadConfig, err)
	}

	cr := Reader{
		quote:                              cfg.quote,
		numFields:                          cfg.numFields,
		fieldSeparator:                     cfg.fieldSeparator,
		comment:                            cfg.comment,
		trsEmitsRecord:                     cfg.trsEmitsRecord,
		quoteSet:                           cfg.quoteSet,
		commentSet:                         cfg.commentSet,
		trimHeaders:                        cfg.trimHeaders,
		removeHeaderRow:                    cfg.removeHeaderRow,
		errorOnNoRows:                      cfg.errorOnNoRows,
		headers:                            cfg.headers,
		recordSep:                          cfg.recordSep,
		recordSepLen:                       cfg.recordSepLen,
		removeByteOrderMarker:              cfg.removeByteOrderMarker,
		errOnNoByteOrderMarker:             cfg.errOnNoByteOrderMarker,
		commentsAllowedAfterStartOfRecords: cfg.commentsAllowedAfterStartOfRecords,
	}

	cr.initPipeline(cfg.reader, cfg.borrowRow, cfg.discoverRecordSeparator)

	return &cr, nil
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

func (r *Reader) parsingErr(err error) {
	r.err = newParsingError(r.byteIndex, r.recordIndex, r.fieldIndex, err)
}

func (r *Reader) ioErr(err error) {
	r.err = newIOError(r.byteIndex, r.recordIndex, r.fieldIndex, err)
}

func (r *Reader) borrowedRow() []string {
	if r.fieldLengths == nil || len(r.fieldLengths) != r.numFields {
		return nil
	}

	row := make([]string, len(r.fieldLengths))

	r.row = func() []string {
		p := 0
		for i, s := range r.fieldLengths {
			if s == 0 {
				row[i] = ""
				continue
			}
			row[i] = unsafe.String(&r.recordBuf[p], s)
			p += s
		}
		return row
	}

	return r.row()
}

func (r *Reader) defaultRow() []string {
	if r.fieldLengths == nil || len(r.fieldLengths) != r.numFields {
		return nil
	}

	r.row = r.defaultClonedRow
	return r.row()
}

func (r *Reader) defaultClonedRow() []string {
	if r.fieldLengths == nil || len(r.fieldLengths) != r.numFields {
		return nil
	}

	p := 0
	row := make([]string, len(r.fieldLengths))
	for i, s := range r.fieldLengths {
		if s == 0 {
			row[i] = ""
			continue
		}
		row[i] = strings.Clone(unsafe.String(&r.recordBuf[p], s))
		p += s
	}
	return row
}

func (r *Reader) nextRuneIsLF() bool {
	c, size, err := r.reader.ReadRune()
	if size <= 0 {
		r.done = true
		if !errors.Is(err, io.EOF) {
			r.ioErr(err)
		}
		return false
	}

	if size == 1 && (c == asciiLineFeed) {
		// advance the position indicator
		r.byteIndex += uint(size)

		if err != nil {
			r.done = true
			if !errors.Is(err, io.EOF) {
				r.ioErr(err)
			}
		}
		return true
	}

	if err := r.reader.UnreadRune(); err != nil {
		panic(err)
	}

	return false
}

func (r *Reader) defaultCheckNumFields() bool {
	if len(r.fieldLengths) == r.numFields {
		r.afterStartOfRecords = true
		return true
	}

	r.done = true
	r.parsingErr(fieldCountMismatchErr(r.numFields, len(r.fieldLengths)))
	return false
}

func (r *Reader) checkNumFieldsWithDiscovery() bool {
	r.numFields = len(r.fieldLengths)
	r.checkNumFields = r.defaultCheckNumFields
	return true
}

func (r *Reader) isSingleRuneRecordSeparator(c rune) bool {
	return c == r.recordSep[0]
}

func (r *Reader) isCRLFRecordSeparator(c rune) bool {
	return (c == asciiCarriageReturn && r.nextRuneIsLF())
}

func (r *Reader) updateIsRecordSeparatorImpl() {
	if r.recordSepLen != 2 {
		r.isRecordSeparator = r.isSingleRuneRecordSeparator
		return
	}

	r.isRecordSeparator = r.isCRLFRecordSeparator
}

func (r *Reader) isRecordSeparatorWithDiscovery(c rune) bool {
	isCarriageReturn, ok := isNewlineRune(c)
	if !ok {
		return false
	}

	if isCarriageReturn && r.nextRuneIsLF() {
		r.recordSep = [2]rune{asciiCarriageReturn, asciiLineFeed}
		r.recordSepLen = 2
	} else {
		r.recordSep = [2]rune{c, 0}
		r.recordSepLen = 1
	}

	r.updateIsRecordSeparatorImpl()

	return true
}

func (r *Reader) resetRecordBuffers() {
	r.fieldLengths = r.fieldLengths[:0]
	r.fieldStart = 0
	r.recordBuf = r.recordBuf[:0]
}

func (r *Reader) defaultScan() bool {
	if r.done {
		return false
	}

	r.resetRecordBuffers()

	return r.prepareRow()
}

func (r *Reader) initPipeline(reader io.Reader, borrowRow, discoverRecordSeparator bool) {

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

	if r.numFields != -1 {
		r.checkNumFields = r.defaultCheckNumFields
	} else {
		r.checkNumFields = r.checkNumFieldsWithDiscovery
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
				headerBytes := slices.Clone(r.recordBuf)
				r.recordBuf = r.recordBuf[:0]

				var p int
				matching := checkHeadersMatch
				for i, s := range r.fieldLengths {
					var field string
					if s > 0 {
						field = unsafe.String(&headerBytes[p], s)
					}
					p += s

					field = strings.TrimSpace(field)
					r.fieldLengths[i] = len(field)

					r.recordBuf = append(r.recordBuf, []byte(field)...)

					if matching && field != r.headers[i] {
						matching = false
					}
				}

				if checkHeadersMatch && !matching {
					r.done = true
					r.parsingErr(ErrUnexpectedHeaderRowContents)
					return false
				}
			} else if checkHeadersMatch {

				var p int
				for i, s := range r.fieldLengths {
					var field string
					if s > 0 {
						field = unsafe.String(&r.recordBuf[p], s)
					}
					p += s

					if field != r.headers[i] {
						r.done = true
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

			return r.prepareRow()
		}
	}

	if r.headers == nil && !r.errorOnNoRows {
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
					r.parsingErr(ErrNoHeaderRow)
				} else if r.errorOnNoRows {
					r.parsingErr(ErrNoRows)
				}
			}
			return v
		}
	}
}

func (r *Reader) fieldNumOverflow() bool {
	if len(r.fieldLengths) == r.numFields {
		r.done = true
		r.parsingErr(tooManyFieldsErr(r.numFields))
		return true
	}

	return false
}

func (r *Reader) handleEOF() bool {

	// check if we're in a terminal state otherwise error
	// there is no new character to process
	switch r.state {
	case rStateStartOfDoc:
		if r.errOnNoByteOrderMarker {
			r.done = true
			r.ioErr(fmt.Errorf("%w: %w", ErrNoByteOrderMarker, io.EOF))
			return false
		}

		r.state = rStateStartOfRecord
		fallthrough
	case rStateStartOfRecord:
		if r.trsEmitsRecord && r.numFields == 1 {
			r.fieldLengths = append(r.fieldLengths, 0)
			// r.fieldStart = len(r.recordBuf)
			if !r.checkNumFields() {
				r.err = errors.Join(r.err, io.ErrUnexpectedEOF)
				return false
			}
			return true
		}
		return false
	case rStateInQuotedField:
		r.parsingErr(ErrIncompleteQuotedField)
		return false
	case rStateInLineComment:
		return false
	case rStateStartOfField:
		r.fieldLengths = append(r.fieldLengths, 0)
		// r.fieldStart = len(r.recordBuf)
		if !r.checkNumFields() {
			r.err = errors.Join(r.err, io.ErrUnexpectedEOF)
			return false
		}
		return true
	case rStateEndOfQuotedField, rStateInField:
		r.fieldLengths = append(r.fieldLengths, len(r.recordBuf)-r.fieldStart)
		r.fieldStart = len(r.recordBuf)
		if !r.checkNumFields() {
			r.err = errors.Join(r.err, io.ErrUnexpectedEOF)
			return false
		}
		return true
	}

	panic("reader in unknown state when EOF encountered")
}

func (r *Reader) prepareRow() bool {

	for {
		c, size, rErr := r.reader.ReadRune()

		// advance the position indicator
		r.byteIndex += uint(size)

		if size == 1 && c == utf8.RuneError {

			//
			// handle a non UTF8 byte
			//

			if rStateStartOfDoc == r.state {
				if r.errOnNoByteOrderMarker {
					r.done = true
					r.parsingErr(ErrNoByteOrderMarker)
					return false
				}

				r.state = rStateStartOfRecord
			}

			if err := r.reader.UnreadRune(); err != nil {
				panic(err)
			}
			var b byte
			if v, err := r.reader.ReadByte(); err != nil {
				panic(err)
			} else {
				b = v
			}

			switch r.state {
			case rStateStartOfRecord, rStateStartOfField:
				r.recordBuf = append(r.recordBuf, b)
				r.state = rStateInField
			case rStateInField, rStateInQuotedField:
				r.recordBuf = append(r.recordBuf, b)
				// r.state = rStateInField
			// case rStateInQuotedField:
			// 	r.recordBuf = append(r.recordBuf, b)
			// 	// r.state = rStateInQuotedField
			case rStateEndOfQuotedField:
				if rErr == nil {
					r.done = true
					r.parsingErr(ErrInvalidQuotedFieldEnding)
					return false
				}
			case rStateInLineComment:
				// r.state = rStateInLineComment
			}

			if rErr == nil {
				continue
			}
		}
		if rErr != nil {
			r.done = true
			if !errors.Is(rErr, io.EOF) {
				r.ioErr(rErr)
				return false
			}
			if size == 0 {
				return r.handleEOF()
			}
			// right here in the code is the only place where the runtime could loop back around where done = true and the last character
			// has been processed
		}

		switch r.state {
		case rStateStartOfDoc:
			if c == utf8ByteOrderMarker {
				if r.removeByteOrderMarker {
					r.state = rStateStartOfRecord

					// checking just in case EOF was encountered earlier
					// with a size > 0
					goto CONTINUE_UNLESS_DONE
				}
			} else if r.errOnNoByteOrderMarker {
				r.done = true
				r.parsingErr(ErrNoByteOrderMarker)
				return false
			}

			r.state = rStateStartOfRecord
			fallthrough
		case rStateStartOfRecord:
			switch c {
			case r.fieldSeparator:
				r.fieldLengths = append(r.fieldLengths, 0)
				// r.fieldStart = len(r.recordBuf)
				r.state = rStateStartOfField
				r.fieldIndex++
			default:
				if r.isRecordSeparator(c) {
					r.fieldLengths = append(r.fieldLengths, 0)
					// r.fieldStart = len(r.recordBuf)
					// r.state = rStateStartOfRecord
					if r.checkNumFields() {
						r.fieldIndex = 0
						r.recordIndex++
						return true
					}
					return false
				}
				if r.quoteSet && c == r.quote {
					r.state = rStateInQuotedField
				} else if (!r.afterStartOfRecords || r.commentsAllowedAfterStartOfRecords) && r.commentSet && c == r.comment {
					r.state = rStateInLineComment
				} else {
					r.recordBuf = append(r.recordBuf, []byte(string(c))...)
					r.state = rStateInField
				}
			}
		case rStateInQuotedField:
			switch c {
			case r.quote:
				r.state = rStateEndOfQuotedField
			default:
				r.recordBuf = append(r.recordBuf, []byte(string(c))...)
				// r.state = rStateInQuotedField
			}
		case rStateEndOfQuotedField:
			switch c {
			case r.fieldSeparator:
				r.fieldLengths = append(r.fieldLengths, len(r.recordBuf)-r.fieldStart)
				r.fieldStart = len(r.recordBuf)
				if r.fieldNumOverflow() {
					return false
				}
				r.state = rStateStartOfField
				r.fieldIndex++
			case r.quote:
				r.recordBuf = append(r.recordBuf, []byte(string(r.quote))...)
				r.state = rStateInQuotedField
			default:
				if r.isRecordSeparator(c) {
					r.fieldLengths = append(r.fieldLengths, len(r.recordBuf)-r.fieldStart)
					r.fieldStart = len(r.recordBuf)
					r.state = rStateStartOfRecord
					if r.checkNumFields() {
						r.fieldIndex = 0
						r.recordIndex++
						return true
					}
					return false
				}
				r.done = true
				r.parsingErr(ErrInvalidQuotedFieldEnding)
				return false
			}
		case rStateStartOfField:
			switch c {
			case r.fieldSeparator:
				r.fieldLengths = append(r.fieldLengths, 0)
				// r.fieldStart = len(r.recordBuf)
				if r.fieldNumOverflow() {
					return false
				}
				// r.state = rStateStartOfField
				r.fieldIndex++
			default:
				if r.isRecordSeparator(c) {
					r.fieldLengths = append(r.fieldLengths, 0)
					// r.fieldStart = len(r.recordBuf)
					r.state = rStateStartOfRecord
					if r.checkNumFields() {
						r.fieldIndex = 0
						r.recordIndex++
						return true
					}
					return false
				}
				if r.quoteSet && c == r.quote {
					r.state = rStateInQuotedField
				} else {
					r.recordBuf = append(r.recordBuf, []byte(string(c))...)
					r.state = rStateInField
				}
			}
		case rStateInField:
			switch c {
			case r.fieldSeparator:
				r.fieldLengths = append(r.fieldLengths, len(r.recordBuf)-r.fieldStart)
				r.fieldStart = len(r.recordBuf)
				if r.fieldNumOverflow() {
					return false
				}
				r.state = rStateStartOfField
				r.fieldIndex++
			default:
				if r.isRecordSeparator(c) {
					r.fieldLengths = append(r.fieldLengths, len(r.recordBuf)-r.fieldStart)
					r.fieldStart = len(r.recordBuf)
					r.state = rStateStartOfRecord
					if r.checkNumFields() {
						r.fieldIndex = 0
						r.recordIndex++
						return true
					}
					return false
				}
				r.recordBuf = append(r.recordBuf, []byte(string(c))...)
				// r.state = rStateInField
			}
		case rStateInLineComment:
			if r.isRecordSeparator(c) {
				r.state = rStateStartOfRecord
				// r.recordIndex++ // not valid in this case because the previous state was not a record

				// checking just in case EOF was encountered earlier
				// with a size > 0
				goto CONTINUE_UNLESS_DONE
			}
		}

	CONTINUE_UNLESS_DONE:
		if r.done {
			break
		}
	}

	return r.checkNumFields()
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
