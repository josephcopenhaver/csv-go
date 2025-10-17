package csv

import (
	"bytes"
	"errors"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
	"unsafe"
)

// TODO: ideally the `func (w *Writer) loadQF_*` functions should be inlined

// TODO: not all runtime flags are mutable through the writer instance lifecycle
//
// for those that are immutable they should be expanded into local values which
// are not on the heap to avoid reloads of stack data and reuse of a register

// TODO: some field types are guaranteed to never have overlap with csv encoding characters
// (quote, escape, record separator, and field separator) such as the types that fit in 64 bits
// for reasonably most values of csv configuration formats
//
// it is possible to skip the processing of those serialized byte streams for csv encoding
// that currently happens

var (
	//
	// writer specific errors
	//

	ErrWriteHeaderFailed         = errors.New("write header failed")
	ErrRowNilOrEmpty             = errors.New("row is nil or empty")
	ErrNonUTF8InRecord           = errors.New("non-utf8 characters in record")
	ErrNonUTF8InComment          = errors.New("non-utf8 characters in comment")
	ErrWriterClosed              = errors.New("writer closed")
	ErrHeaderWritten             = errors.New("header already written")
	ErrInvalidFieldCountInRecord = errors.New("invalid field count in record")
	ErrInvalidFieldWriter        = errors.New("invalid field writer")
	ErrInvalidRune               = errors.New("invalid rune")
)

type wFlag uint8

const (
	wFlagErrOnNonUTF8 wFlag = 1 << iota
	wFlagControlRuneOverlap
	wFlagEscapeSet
	wFlagForceQuoteFirstField

	//
	// low access rate flags
	//

	wFlagCommentSet
	wFlagHeaderWritten
	wFlagClosed
	wFlagClearMemoryAfterFree
)

const (
	u64signBitMask uint64 = 0x8000000000000000

	maxLenSerializedTime    = 35
	maxLenSerializedBool    = 1
	maxLenSerializedFloat64 = 24
	// boundedFieldWritersMaxByteLen is the maximum output byte length of the fieldWriter types that serialize to bounded byte sizes (so types other than bytes, and string plus their UTF8 variants for the case of bytes and string)
	//
	// it should match the maxLenSerializedTime value
	boundedFieldWritersMaxByteLen = 35

	invalidRuneUTF8Encoded = 0xEFBFBD
	// note that if a utf8.RuneError rune is supplied the offset is 1 and not 2
	//
	// 2 here is completely invalid and obviously so from an encoding perspective,
	// so we use it to indicate that the rune supplied to the writer is definitely
	// invalid
	invalidRuneUTF8EncodedWithOffset = ((2 << (8 * 4)) | invalidRuneUTF8Encoded)

	// fieldWriterTypesRuneList contains all non rune, bytes, and string type output byte values which can permute into various combinations
	fieldWriterTypesRuneList = "-:.+0123456789aefInNTZ" // 0-9, float, NaN, Inf, time
)

type wFieldKind uint8

const (
	_ wFieldKind = iota
	wfkBytes
	wfkString
	wfkInt
	wfkInt64
	wfkDuration
	wfkUint64
	wfkTime
	wfkRune
	wfkBool
	wfkFloat64
)

type FieldWriter struct {
	kind  wFieldKind
	bytes []byte
	time  time.Time
	str   string
	// _64_bits holds data for kinds that can be expressed within the 64 bits of uint64:
	// int, int64, duration, uint64, rune, bool, and float64
	_64_bits uint64
}

func (w *FieldWriter) isZeroLen() bool {
	switch w.kind {
	case wfkBytes:
		return len(w.bytes) == 0
	case wfkString:
		return len(w.str) == 0
	case wfkInt, wfkInt64, wfkDuration, wfkUint64, wfkTime, wfkRune, wfkBool, wfkFloat64:
		return false
	default:
		// I reserve the right to panic here in the future should I wish to.
		return false
	}
}

func (w *FieldWriter) startsWithRune(buf []byte, r rune) bool {
	p := []byte(string(r))

	switch w.kind {
	case wfkBytes:
		if len(w.bytes) == 0 {
			return false
		}

		return bytes.HasPrefix(w.bytes, p)
	case wfkString:
		s := w.str
		if len(s) == 0 {
			return false
		}

		b := unsafe.Slice(unsafe.StringData(s), len(s))

		return bytes.HasPrefix(b, p)
	case wfkInt, wfkInt64, wfkDuration, wfkUint64, wfkTime, wfkRune, wfkBool, wfkFloat64:
		var err error
		buf, err = w.AppendText(buf)
		return (err == nil && bytes.HasPrefix(buf, p))
	default:
		// I reserve the right to panic here in the future should I wish to.
		return false
	}
}

func (w *FieldWriter) runeAppendText(b []byte) ([]byte, error) {
	r := w._64_bits
	if r == invalidRuneUTF8EncodedWithOffset {
		return nil, ErrInvalidRune
	}

	var buf [utf8.UTFMax]byte
	buf[3] = byte(r)
	buf[2] = byte(r >> (8 * 1))
	buf[1] = byte(r >> (8 * 2))
	buf[0] = byte(r >> (8 * 3))
	offset := uint8(r >> (8 * 4))

	return append(b, buf[offset:]...), nil
}

func (w *FieldWriter) AppendText(b []byte) ([]byte, error) {

	switch w.kind {
	case wfkBytes:
		return append(b, w.bytes...), nil
	case wfkString:
		return append(b, w.str...), nil
	case wfkInt, wfkInt64, wfkDuration:
		return strconv.AppendInt(b, int64(w._64_bits), 10), nil
	case wfkUint64:
		return strconv.AppendUint(b, uint64(w._64_bits), 10), nil
	case wfkTime:
		return w.time.AppendFormat(b, time.RFC3339Nano), nil
	case wfkRune:
		return w.runeAppendText(b)
	case wfkBool:
		boolAsByte := byte('0') + byte(w._64_bits)
		return append(b, boolAsByte), nil
	case wfkFloat64:
		return strconv.AppendFloat(b, math.Float64frombits(w._64_bits), 'g', -1, 64), nil
	}

	return nil, ErrInvalidFieldWriter
}

func (w *FieldWriter) MarshalText() (text []byte, err error) {
	var b []byte

	switch w.kind {
	case wfkBytes:
		// Note: don't be tempted to just return the inner buffer here
		// there is no contract with the calling context to never modify the slice it gets in return
		n := len(w.bytes)
		if n == 0 {
			return nil, nil
		}
		b = make([]byte, 0, n)
	case wfkString:
		// Note: don't be tempted to just return the inner buffer here
		// there is no contract with the calling context to never modify the slice it gets in return
		n := len(w.str)
		if n == 0 {
			return nil, nil
		}
		b = make([]byte, 0, n)
	case wfkInt, wfkInt64, wfkDuration:
		// base10ByteLen at most would be 20 if negative, 19 if positive
		//
		// might be better off making a buffer pool of a fixed size
		// but I leave that for callers of AppendText to implement
		input := w._64_bits
		var signAdjustment int
		if input&u64signBitMask != 0 {
			input = uint64(-int64(w._64_bits))
			signAdjustment = 1
		}
		base10ByteLen := decLenU64(input) + signAdjustment

		b = make([]byte, 0, base10ByteLen)
	case wfkUint64:
		// base10ByteLen at most would be 20 (always positive)
		//
		// might be better off making a buffer pool of a fixed size
		// but I leave that for callers of AppendText to implement
		base10ByteLen := decLenU64(w._64_bits)

		b = make([]byte, 0, base10ByteLen)
	case wfkTime:
		b = make([]byte, 0, maxLenSerializedTime)
	case wfkRune:
		// would be at most utf8.UTFMax (4) bytes in length and never empty
		b = make([]byte, 0, utf8.UTFMax)
	case wfkBool:
		b = make([]byte, 0, maxLenSerializedBool)
	case wfkFloat64:
		b = make([]byte, 0, maxLenSerializedFloat64)
	}

	return w.AppendText(b)
}

type FieldWriterFactory struct{}

func (FieldWriterFactory) Bytes(p []byte) FieldWriter {
	return FieldWriter{
		kind:  wfkBytes,
		bytes: p,
	}
}

func (FieldWriterFactory) UTF8Bytes(p []byte) FieldWriter {
	return FieldWriter{
		kind:     wfkBytes,
		bytes:    p,
		_64_bits: 1,
	}
}

func (FieldWriterFactory) String(s string) FieldWriter {
	return FieldWriter{
		kind: wfkString,
		str:  s,
	}
}

func (FieldWriterFactory) UTF8String(s string) FieldWriter {
	return FieldWriter{
		kind:     wfkString,
		str:      s,
		_64_bits: 1,
	}
}

func (FieldWriterFactory) Int(i int) FieldWriter {
	return FieldWriter{
		kind:     wfkInt,
		_64_bits: uint64(i),
	}
}

func (FieldWriterFactory) Int64(i int64) FieldWriter {
	return FieldWriter{
		kind:     wfkInt64,
		_64_bits: uint64(i),
	}
}

func (FieldWriterFactory) Uint64(i uint64) FieldWriter {
	return FieldWriter{
		kind:     wfkUint64,
		_64_bits: i,
	}
}

func (FieldWriterFactory) Time(t time.Time) FieldWriter {
	return FieldWriter{
		kind: wfkTime,
		time: t,
	}
}

// Rune value must be a valid utf8 rune value otherwise
// attempting to write the rune will result in an
// ErrInvalidRune error.
func (FieldWriterFactory) Rune(r rune) FieldWriter {
	numBytes := utf8.RuneLen(r)
	if numBytes == -1 {
		return FieldWriter{
			kind:     wfkRune,
			_64_bits: invalidRuneUTF8EncodedWithOffset,
		}
	}

	var buf [utf8.UTFMax]byte
	offset := uint8(utf8.UTFMax - numBytes)
	utf8.EncodeRune(buf[offset:], r)

	v := (uint64(offset) << (8 * 4)) |
		(uint64(buf[0]) << (8 * 3)) |
		(uint64(buf[1]) << (8 * 2)) |
		(uint64(buf[2]) << (8 * 1)) |
		uint64(buf[3])

	return FieldWriter{
		kind:     wfkRune,
		_64_bits: v,
	}
}

func (FieldWriterFactory) Bool(b bool) FieldWriter {
	result := FieldWriter{
		kind: wfkBool,
	}

	if b {
		result._64_bits = 1
	}

	return result
}

func (FieldWriterFactory) Duration(d time.Duration) FieldWriter {
	return FieldWriter{
		kind:     wfkDuration,
		_64_bits: uint64(d),
	}
}

func (FieldWriterFactory) Float64(f float64) FieldWriter {
	return FieldWriter{
		kind:     wfkFloat64,
		_64_bits: math.Float64bits(f),
	}
}

func FieldWriters() FieldWriterFactory {
	return FieldWriterFactory{}
}

type writeIOErr struct {
	err error
}

func (e writeIOErr) Is(target error) bool {
	return errors.Is(ErrIO, target) || errors.Is(e.err, target)
}

func (e writeIOErr) Error() string {
	return "io error: " + e.err.Error()
}

type wCfg struct {
	writer                     io.Writer
	initialRecordBufferSize    int
	recordBuf                  []byte
	recordSep                  [2]rune
	numFields                  int
	fieldSeparator             rune
	quote                      rune
	escape                     rune
	comment                    rune
	initialRecordBufferSizeSet bool
	recordBufSet               bool
	numFieldsSet               bool
	errOnNonUTF8               bool
	escapeSet                  bool
	commentSet                 bool
	recordSepRuneLen           int8
	clearMemoryAfterFree       bool
	fwOverlapsControlRunes     bool
}

type WriterOption func(*wCfg)

// WriterOptions should never be instantiated manually
//
// Instead call WriterOpts()
//
// This is only exported to allow godocs to discover the exported methods.
//
// WriterOptions will never have exported members and the zero value is not
// part of the semver guarantee. Instantiate it incorrectly at your own peril.
//
// Calling the function is a nop that is compiled away anyways, you will not
// optimize anything at all. Use WriterOpts()!
type WriterOptions struct{}

func WriterOpts() WriterOptions {
	return WriterOptions{}
}

func (WriterOptions) RecordSeparator(s string) WriterOption {
	if len(s) == 0 {
		return badRecordSeparatorWConfig
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
		return badRecordSeparatorWConfig
	}
	if n1 == len(v) {
		// note: requiring record separators to be newline character is opinionated
		// but it is a good practice leaning towards simplicity.
		//
		// I am open to a PR that makes this more flexible.
		//
		// Note that if this is relaxed then runeRequiresQuotes funcs will need to
		// check for the non-newline rune record separators.

		if !isNewlineRuneForWrite(r1) {
			return badRecordSeparatorWConfig
		}

		return func(cfg *wCfg) {
			cfg.recordSep[0] = r1
			cfg.recordSepRuneLen = 1
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
		return badRecordSeparatorWConfig
	}
	if n1+n2 == len(v) && r1 == asciiCarriageReturn && r2 == asciiLineFeed {
		return func(cfg *wCfg) {
			cfg.recordSep[0] = r1
			cfg.recordSep[1] = r2
			cfg.recordSepRuneLen = 2
		}
	}

	return badRecordSeparatorWConfig
}

func (WriterOptions) FieldSeparator(v rune) WriterOption {
	return func(cfg *wCfg) {
		cfg.fieldSeparator = v
	}
}

func (WriterOptions) Quote(v rune) WriterOption {
	return func(cfg *wCfg) {
		cfg.quote = v
	}
}

func (WriterOptions) Writer(v io.Writer) WriterOption {
	return func(cfg *wCfg) {
		cfg.writer = v
	}
}

func (WriterOptions) NumFields(v int) WriterOption {
	return func(cfg *wCfg) {
		cfg.numFields = v
		cfg.numFieldsSet = true
	}
}

func (WriterOptions) ErrorOnNonUTF8(v bool) WriterOption {
	return func(cfg *wCfg) {
		cfg.errOnNonUTF8 = v
	}
}

func (WriterOptions) Escape(r rune) WriterOption {
	return func(cfg *wCfg) {
		cfg.escape = r
		cfg.escapeSet = true
	}
}

// ClearFreedDataMemory ensures that whenever a shared memory buffer
// that contains data goes out of scope that zero values are written
// to every byte within the buffer.
//
// This may significantly degrade performance and is recommended only
// for sensitive data or long-lived processes.
func (WriterOptions) ClearFreedDataMemory(b bool) WriterOption {
	return func(cfg *wCfg) {
		cfg.clearMemoryAfterFree = b
	}
}

// InitialRecordBufferSize is a hint to pre-allocate record buffer space once
// and reduce the number of re-allocations when processing fields to write.
//
// Please consider this to be a micro optimization in most circumstances just
// because it's not likely that most users will know the maximum total record
// size they wish to target / be under and it's generally a better practice
// to leave these details to the go runtime to coordinate via standard
// garbage collection.
func (WriterOptions) InitialRecordBufferSize(v int) WriterOption {
	return func(cfg *wCfg) {
		cfg.initialRecordBufferSize = v
		cfg.initialRecordBufferSizeSet = true
	}
}

// InitialRecordBuffer is a hint to pre-allocate record buffer space once
// externally and pipe it in to reduce the number of re-allocations when
// processing a writer and reuse it at a later time after the writer is closed.
//
// This option should generally not be used. It only exists to assist with
// processing large numbers of CSV files should memory be a clear constraint.
// There is no guarantee this buffer will always be used till the end of the
// csv Writer's lifecycle.
//
// Please consider this to be a micro optimization in most circumstances just because is tightens the usage
// contract of the csv Reader in ways most would not normally consider.
func (WriterOptions) InitialRecordBuffer(v []byte) WriterOption {
	return func(cfg *wCfg) {
		cfg.recordBuf = v
		cfg.recordBufSet = true
	}
}

// InitialFieldBufferSize is deprecated and no longer has any effect.
//
// Historically:
//
// InitialFieldBufferSize is a hint to pre-allocate field buffer space once
// and reduce the number of re-allocations when processing fields to write.
//
// Please consider this to be a micro optimization in most circumstances just
// because it's not likely that most users will know the maximum total field
// size they wish to target / be under and it's generally a better practice
// to leave these details to the go runtime to coordinate via standard
// garbage collection.
func (WriterOptions) InitialFieldBufferSize(v int) WriterOption {
	return func(cfg *wCfg) {
	}
}

// InitialFieldBuffer is deprecated and no longer has any effect.
//
// Historically:
//
// InitialFieldBuffer is a hint to pre-allocate field buffer space once
// externally and pipe it in to reduce the number of re-allocations when
// processing a writer and reuse it at a later time after the writer is closed.
//
// This option should generally not be used. It only exists to assist with
// processing large numbers of CSV files should memory be a clear constraint.
// There is no guarantee this buffer will always be used till the end of the
// csv Writer's lifecycle.
//
// Please consider this to be a micro optimization in most circumstances just because is tightens the usage
// contract of the csv Reader in ways most would not normally consider.
func (WriterOptions) InitialFieldBuffer(v []byte) WriterOption {
	return func(cfg *wCfg) {
	}
}

func (WriterOptions) CommentRune(r rune) WriterOption {
	return func(cfg *wCfg) {
		cfg.comment = r
		cfg.commentSet = true
	}
}

func (cfg *wCfg) validate() error {
	if cfg.writer == nil {
		return errors.New("nil writer")
	}

	if cfg.recordSepRuneLen == 0 {
		return errors.New("record separator can only be one valid utf8 newline rune long or \"\\r\\n\"")
	}

	if cfg.numFieldsSet && cfg.numFields <= 0 {
		return errors.New("num fields must be greater than zero")
	}

	if cfg.escapeSet && cfg.escape == cfg.quote {
		cfg.escapeSet = false
	}

	// note: not letting field separators, quotes, or escapes be newline characters is
	// opinionated but it is a good practice leaning towards simplicity.
	//
	// I am open to a PR that makes this more flexible.

	if !validUtf8Rune(cfg.fieldSeparator) || isNewlineRuneForWrite(cfg.fieldSeparator) {
		return errors.New("invalid field separator value")
	}

	if !validUtf8Rune(cfg.quote) || isNewlineRuneForWrite(cfg.quote) {
		return errors.New("invalid quote value")
	}

	if cfg.escapeSet && (!validUtf8Rune(cfg.escape) || isNewlineRuneForWrite(cfg.escape)) {
		return errors.New("invalid escape value")
	}

	if cfg.fieldSeparator == cfg.quote {
		return errors.New("invalid field separator and quote combination")
	}

	if cfg.escapeSet && cfg.fieldSeparator == cfg.escape {
		return errors.New("invalid field separator and escape combination")
	}

	{
		listStart := uint8(0)
		runeArray := [...]rune{cfg.escape, cfg.quote, cfg.fieldSeparator}
		if !cfg.escapeSet {
			listStart++
		}
		if isFieldWriterRune(runeArray[listStart:]...) {
			cfg.fwOverlapsControlRunes = true
		}
	}

	if cfg.recordBufSet && cfg.initialRecordBufferSizeSet {
		return errors.New("initial record buffer size cannot be specified when also setting the initial record buffer")
	}

	if cfg.initialRecordBufferSizeSet && cfg.initialRecordBufferSize < 0 {
		return errors.New("initial record buffer size must be greater than or equal to zero")
	}

	if cfg.commentSet {
		if err := isValidComment(cfg.comment, cfg.quote, cfg.fieldSeparator, cfg.escape, cfg.escapeSet); err != nil {
			return err
		}
	}

	return nil
}

type Writer struct {
	writeRow                func([]FieldWriter) (int, error)
	fieldWriters            []FieldWriter
	fieldWriterBuf          [boundedFieldWritersMaxByteLen]byte
	recordBuf               []byte
	controlRunes            string
	escapeControlRunes      string
	twoQuotes               [utf8.UTFMax * 2]byte
	escapedQuote            [utf8.UTFMax * 2]byte
	escapedEscape           [utf8.UTFMax * 2]byte
	fieldSepBytes           [utf8.UTFMax]byte
	recordSepBytes          [utf8.UTFMax]byte
	quoteBytes              [utf8.UTFMax]byte
	comment                 rune
	numFields               int
	writer                  io.Writer
	err                     error
	quote, fieldSep, escape rune
	quoteByteLen            int8
	fieldSepByteLen         int8
	recordSepRuneLen        int8
	twoQuotesByteLen        int8
	escapedQuoteByteLen     int8
	escapedEscapeByteLen    int8
	recordSepByteLen        int8
	bitFlags                wFlag
}

// NewWriter creates a new instance of a CSV writer which is not safe for concurrent reads.
func NewWriter(options ...WriterOption) (*Writer, error) {

	cfg := wCfg{
		numFields:        -1,
		quote:            '"',
		recordSep:        [2]rune{asciiLineFeed, 0},
		recordSepRuneLen: 1,
		fieldSeparator:   ',',
		errOnNonUTF8:     true,
	}

	for _, f := range options {
		f(&cfg)
	}

	if err := cfg.validate(); err != nil {
		return nil, errors.Join(ErrBadConfig, err)
	}

	var controlRunes, escapeControlRunes string
	var quoteBytes [utf8.UTFMax]byte
	var twoQuotes [utf8.UTFMax * 2]byte
	var escapedQuote [utf8.UTFMax * 2]byte
	var escapedEscape [utf8.UTFMax * 2]byte
	var fieldSepBytes, recordSepBytes [utf8.UTFMax]byte
	var escapedQuoteByteLen, escapedEscapeByteLen, recordSepByteLen, twoQuotesByteLen int8
	quoteByteLen := int8(utf8.EncodeRune(quoteBytes[:], cfg.quote))
	fieldSepByteLen := int8(utf8.EncodeRune(fieldSepBytes[:], cfg.fieldSeparator))
	if cfg.escapeSet {
		controlRunes = string([]rune{cfg.quote, cfg.escape, cfg.fieldSeparator}) + "\x0A\x0B\x0C\x0D\u0085\u2028"
		escapeControlRunes = string([]rune{cfg.quote, cfg.escape})

		n := utf8.EncodeRune(escapedQuote[:], cfg.escape)
		n += utf8.EncodeRune(escapedQuote[n:], cfg.quote)
		escapedQuoteByteLen = int8(n)

		n = utf8.EncodeRune(escapedEscape[:], cfg.escape)
		copy(escapedEscape[n:], escapedEscape[:n])
		n *= 2
		escapedEscapeByteLen = int8(n)
	} else {
		controlRunes = string([]rune{cfg.quote, cfg.fieldSeparator}) + "\x0A\x0B\x0C\x0D\u0085\u2028"
		escapeControlRunes = string(cfg.quote)

		n := utf8.EncodeRune(escapedQuote[:], cfg.quote)
		copy(escapedQuote[n:], escapedQuote[:n])
		n *= 2
		escapedQuoteByteLen = int8(n)
	}

	{
		n := utf8.EncodeRune(recordSepBytes[:], cfg.recordSep[0])
		if cfg.recordSepRuneLen == 2 {
			n += utf8.EncodeRune(recordSepBytes[n:], cfg.recordSep[1])
		}
		recordSepByteLen = int8(n)
	}

	{
		n := utf8.EncodeRune(twoQuotes[:], cfg.quote)
		copy(twoQuotes[n:], twoQuotes[:n])
		n *= 2
		twoQuotesByteLen = int8(n)
	}

	var recordBuf []byte
	if cfg.initialRecordBufferSizeSet {
		recordBuf = make([]byte, 0, cfg.initialRecordBufferSize)
	} else if cfg.recordBufSet {
		recordBuf = cfg.recordBuf[:0:len(cfg.recordBuf)]
	}

	var bitFlags wFlag
	if cfg.errOnNonUTF8 {
		bitFlags |= wFlagErrOnNonUTF8
	}
	if cfg.escapeSet {
		bitFlags |= wFlagEscapeSet
	}
	if cfg.clearMemoryAfterFree {
		bitFlags |= wFlagClearMemoryAfterFree
	}
	if cfg.commentSet {
		bitFlags |= wFlagCommentSet
		if strings.ContainsRune(fieldWriterTypesRuneList, cfg.comment) {
			bitFlags |= wFlagControlRuneOverlap
		}
	}

	w := &Writer{
		numFields:            cfg.numFields,
		writer:               cfg.writer,
		controlRunes:         controlRunes,
		escapeControlRunes:   escapeControlRunes,
		quoteBytes:           quoteBytes,
		quoteByteLen:         quoteByteLen,
		twoQuotes:            twoQuotes,
		twoQuotesByteLen:     twoQuotesByteLen,
		escapedEscape:        escapedEscape,
		escapedEscapeByteLen: escapedEscapeByteLen,
		escapedQuote:         escapedQuote,
		escapedQuoteByteLen:  escapedQuoteByteLen,
		fieldSepBytes:        fieldSepBytes,
		fieldSepByteLen:      fieldSepByteLen,
		recordSepBytes:       recordSepBytes,
		recordSepByteLen:     recordSepByteLen,
		comment:              cfg.comment,
		quote:                cfg.quote,
		recordSepRuneLen:     cfg.recordSepRuneLen,
		fieldSep:             cfg.fieldSeparator,
		escape:               cfg.escape,
		recordBuf:            recordBuf,
		bitFlags:             bitFlags,
	}

	w.setWriteRowStrategy(cfg.clearMemoryAfterFree, cfg.escapeSet)

	return w, nil
}

func (w *Writer) setWriteRowStrategy(clearMemoryAfterFree, escapeSet bool) {
	var f func([]FieldWriter) (int, error)

	if !clearMemoryAfterFree {
		f = w.writeRow_memclearOff
	} else {
		f = w.writeRow_memclearOn
	}

	w.writeRow = func(row []FieldWriter) (int, error) {
		w.writeRow = f
		w.bitFlags |= wFlagHeaderWritten

		if (w.bitFlags & wFlagCommentSet) != 0 {
			// detect if the first field begins with a comment sequence
			// and if so, set that the first field should be quoted
			// regardless of its type or content

			if row[0].startsWithRune(w.fieldWriterBuf[:0], w.comment) {
				w.bitFlags |= wFlagForceQuoteFirstField
				defer func() {
					w.bitFlags &= ^wFlagForceQuoteFirstField
				}()
			}
		}

		return w.writeRow(row)
	}
}

// Close should be called after writing all rows
// successfully to the underlying writer.
//
// Close currently always returns nil, but in the future
// it may not.
//
// Should any configuration options require post-flight
// checks they will be implemented here.
//
// It will never attempt to flush or close the underlying writer
// instance. That is left to the calling context.
func (w *Writer) Close() error {
	if (w.bitFlags & wFlagClosed) != 0 {
		return nil
	}
	w.bitFlags |= wFlagClosed

	w.setErr(ErrWriterClosed)

	if (w.bitFlags & wFlagClearMemoryAfterFree) != 0 {
		clear(w.recordBuf[:cap(w.recordBuf)])
		clear(w.fieldWriters)
		clear(w.fieldWriterBuf[:])
	}

	return nil
}

func (w *Writer) appendRec(buf []byte) {
	if len(buf) == 0 {
		return
	}

	old := w.recordBuf
	w.recordBuf = append(w.recordBuf, buf...)

	if cap(old) == 0 {
		return
	}
	old = old[:cap(old)]

	if &old[0] == &w.recordBuf[0] {
		return
	}

	clear(old)
}

// setRecordBuf should only be called when the record buf has been appended to
// and might have been reallocated as a result and clear mem on free is enabled.
//
// This function will clear the old buffer if it is no longer being utilized.
func (w *Writer) setRecordBuf(p []byte) {
	old := w.recordBuf
	w.recordBuf = p

	if cap(old) == 0 {
		return
	}
	old = old[:cap(old)]

	if &old[0] == &p[0] {
		return
	}

	clear(old)
}

type whCfg struct {
	headers         []string
	commentLines    []string
	comment         rune
	trimHeaders     bool
	headersSet      bool
	commentSet      bool
	commentLinesSet bool
	includeBOM      bool
}

type WriteHeaderOption func(*whCfg)

// WriteHeaderOptions should never be instantiated manually
//
// Instead call WriteHeaderOpts()
//
// This is only exported to allow godocs to discover the exported methods.
//
// WriteHeaderOptions will never have exported members and the zero value is not
// part of the semver guarantee. Instantiate it incorrectly at your own peril.
//
// Calling the function is a nop that is compiled away anyways, you will not
// optimize anything at all. Use WriteHeaderOpts()!
type WriteHeaderOptions struct{}

func WriteHeaderOpts() WriteHeaderOptions {
	return WriteHeaderOptions{}
}

func (WriteHeaderOptions) TrimHeaders(b bool) WriteHeaderOption {
	return func(cfg *whCfg) {
		cfg.trimHeaders = b
	}
}

func (WriteHeaderOptions) Headers(h ...string) WriteHeaderOption {
	return func(cfg *whCfg) {
		cfg.headers = h
		cfg.headersSet = true
	}
}

func (WriteHeaderOptions) CommentRune(r rune) WriteHeaderOption {
	return func(cfg *whCfg) {
		cfg.comment = r
		cfg.commentSet = true
	}
}

func (WriteHeaderOptions) CommentLines(s ...string) WriteHeaderOption {
	return func(cfg *whCfg) {
		cfg.commentLines = s
		cfg.commentLinesSet = true
	}
}

func (WriteHeaderOptions) IncludeByteOrderMarker(b bool) WriteHeaderOption {
	return func(cfg *whCfg) {
		cfg.includeBOM = b
	}
}

func (cfg *whCfg) validate(w *Writer) error {
	if cfg.headersSet {
		if len(cfg.headers) == 0 {
			return errors.New("headers length must be greater than zero")
		} else if w.numFields != -1 && w.numFields != len(cfg.headers) {
			return errors.New("headers length does not match number of fields")
		}
	}

	if !cfg.commentSet {
		if (w.bitFlags & wFlagCommentSet) != 0 {
			// loads the value from the parent writer context and
			// trusts the validation the writer already performed
			cfg.commentSet = true
			cfg.comment = w.comment
		} else if cfg.commentLinesSet {
			return errors.New("comment lines require a comment rune")
		}
	} else if (w.bitFlags & wFlagCommentSet) != 0 {
		return errors.New("comment rune cannot be specified when writing headers when the writer instance already has one specified")
	} else if err := isValidComment(cfg.comment, w.quote, w.fieldSep, w.escape, (w.bitFlags&wFlagEscapeSet) != 0); err != nil {
		return err
	} else {
		// note this block is a positive path state-altering action
		// anything added after this parent if context should similarly
		// be a positive path action that should never really fail

		// comment was specified when writing the header
		// but not in the writer config
		//
		// so load the writer with the updated config context

		if strings.ContainsRune(fieldWriterTypesRuneList, cfg.comment) {
			w.bitFlags |= wFlagControlRuneOverlap
		}

		w.bitFlags |= wFlagCommentSet
		w.comment = cfg.comment
	}

	// positive path remaining actions:
	// (note that the first one in the positive path was the else clause above)

	if cfg.headersSet && w.numFields == -1 {
		w.numFields = len(cfg.headers)
	}

	return nil
}

func (w *Writer) writeBOM() (int, error) {
	utf8BOMBytes := []byte{
		0xFF & (utf8ByteOrderMarker >> 16),
		0xFF & (utf8ByteOrderMarker >> 8),
		0xFF & utf8ByteOrderMarker,
	}

	n, err := w.writer.Write(utf8BOMBytes)
	if err != nil {
		err = errors.Join(ErrWriteHeaderFailed, writeIOErr{err})
		w.setErr(err)
	}
	return n, err
}

func (w *Writer) WriteHeader(options ...WriteHeaderOption) (int, error) {
	var result int
	if err := w.err; err != nil {
		return result, err
	}

	if (w.bitFlags & wFlagHeaderWritten) != 0 {
		return result, ErrHeaderWritten
	}
	w.bitFlags |= wFlagHeaderWritten

	cfg := whCfg{
		// no defaults
	}
	for _, f := range options {
		f(&cfg)
	}

	if err := cfg.validate(w); err != nil {
		return result, errors.Join(ErrBadConfig, err)
	}

	if cfg.commentLinesSet {
		var buf bytes.Buffer
		if (w.bitFlags & wFlagClearMemoryAfterFree) != 0 {
			defer func() {
				buf.Reset()
				b := buf.Bytes()
				clear(b[:cap(b)])
			}()
		}

		// note that while separate strings will be placed on separate lines, all newline
		// sequences will be converted to record separator newline sequences.
		//
		// This makes for predictable record separator discovery and parsing when reading.

		// TODO: don't write runes, write chunks of bytes where possible

		commentBufArr := [2]string{}
		var found bool

		for i := range cfg.commentLines {
			commentBuf := commentBufArr[:]
			commentBuf[0], commentBuf[1], found = strings.Cut(cfg.commentLines[i], "\r\n")
			if !found {
				commentBuf = commentBuf[:1]
			}
			for {

				buf.WriteRune(cfg.comment)
				buf.WriteRune(' ')

				if len(commentBuf[0]) > 0 {

					// line here is immutable
					//
					// unsafe may look concerning and scary, and it can be,
					// however in this case we're never writing to the slice
					// created here which is stored within `line`
					//
					// since strings are immutable as well this is actually a safe
					// usage of the unsafe package to avoid an allocation we're
					// just going to read from and then throw away before this
					// returns
					//
					// while line itself does get redefined via re-slicing this does not
					// change the internals of the memory block the slice itself points to.
					//
					// It will also never be called if the len is zero,
					// just as an extra precaution.
					line := unsafe.Slice(unsafe.StringData(commentBuf[0]), len(commentBuf[0]))

					for len(line) != 0 {
						r, s := utf8.DecodeRune(line)
						if s == 1 && r == utf8.RuneError {
							if (w.bitFlags & wFlagErrOnNonUTF8) != 0 {
								return result, ErrNonUTF8InComment
							}

							buf.WriteByte(line[0])

							line = line[s:]
							continue
						}

						if isNewlineRuneForWrite(r) {
							buf.Write(w.recordSepBytes[:w.recordSepByteLen])
							buf.WriteRune(cfg.comment)
							buf.WriteRune(' ')

							line = line[s:]
							continue
						}

						buf.WriteRune(r)
						line = line[s:]
					}
				}

				buf.Write(w.recordSepBytes[:w.recordSepByteLen])

				if len(commentBuf) < 2 {
					break
				}

				commentBuf[0], commentBuf[1], found = strings.Cut(commentBuf[1], "\r\n")
				if !found {
					commentBuf = commentBuf[:1]
				}
			}
		}

		if cfg.includeBOM {
			n, err := w.writeBOM()
			result += n
			if err != nil {
				return result, err
			}
		}

		n, err := io.Copy(w.writer, &buf)
		result += int(n)
		if err != nil {
			err := errors.Join(ErrWriteHeaderFailed, writeIOErr{err})
			w.setErr(err)
			return result, err
		}
	} else if cfg.includeBOM {
		n, err := w.writeBOM()
		result += n
		if err != nil {
			return result, err
		}
	}

	if !cfg.headersSet {
		return result, nil
	}

	headers := cfg.headers
	if cfg.trimHeaders {
		headers = make([]string, len(cfg.headers))
		for i := range cfg.headers {
			headers[i] = strings.TrimSpace(cfg.headers[i])
		}
	}

	// TODO: if providing buffer size hints, then it's likely that the headers should use a different size hint

	n, err := w.WriteRow(headers...)
	result += n
	if err != nil {
		v := errors.Join(ErrWriteHeaderFailed, err)
		err = v
		w.setErr(err)
	}

	return result, err
}

// WriteRow writes a vararg collection of strings as a csv record row.
//
// Each subsequent call to WriteRow, WriteFieldRow, or WriteFieldRowBorrowed should have the exact same slice length.
func (w *Writer) WriteRow(row ...string) (int, error) {
	if err := w.writeRowPreflightCheck(len(row)); err != nil {
		return 0, err
	}

	fieldWriters := w.fieldWriters
	if fieldWriters == nil {
		fieldWriters = make([]FieldWriter, len(row))
		w.fieldWriters = fieldWriters
	}

	fw := FieldWriters()

	for i := range row {
		fieldWriters[i] = fw.String(row[i])
	}

	return w.writeRow(fieldWriters)
}

// WriteFieldRow will take a vararg collection of FieldWriter instances and write them as a csv record row.
//
// Each subsequent call to WriteRow, WriteFieldRow, or WriteFieldRowBorrowed should have the exact same slice length.
//
// This call will copy the provided list of field writers to an internally maintained buffer for amortized access and removal of allocations due to the slice escaping.
//
// If the calling context maintains a reused slice of field writers per write iteration then consider instead using WriteFieldRowBorrowed
// if performance testing indicates that FieldWriter slice copying is a major contributing bottleneck for your case.
func (w *Writer) WriteFieldRow(row ...FieldWriter) (int, error) {
	if err := w.writeRowPreflightCheck(len(row)); err != nil {
		return 0, err
	}

	fieldWriters := w.fieldWriters
	if fieldWriters == nil {
		fieldWriters = make([]FieldWriter, len(row))
		w.fieldWriters = fieldWriters
	}

	copy(fieldWriters, row)

	return w.writeRow(fieldWriters)
}

// WriteFieldRowBorrowed is similar to WriteFieldRow except the slice of rows provided is expected to be externally maintained and reused.
// In such a case this function will be faster than WriteFieldRow, but really it should only be used if performance testing indicates copying of field writers that occurs in WriteFieldRow is a bottleneck
//
// Each subsequent call to WriteRow, WriteFieldRow, or WriteFieldRowBorrowed should have the exact same slice length.
func (w *Writer) WriteFieldRowBorrowed(row []FieldWriter) (int, error) {
	if err := w.writeRowPreflightCheck(len(row)); err != nil {
		return 0, err
	}

	return w.writeRow(row)
}

func (w *Writer) writeRowPreflightCheck(n int) (_err error) {
	// check if prior iterations left the writer in an errored state
	if err := w.err; err != nil {
		return err
	}

	// check if the number of fields to write is zero
	if n == 0 {
		// short circuit to return an error

		// a row write was attempted so even on the error path we
		// must not allow another write header attempt in any way
		w.bitFlags |= wFlagHeaderWritten

		return ErrRowNilOrEmpty
	}

	if v := w.numFields; v != n {
		if v != -1 {

			// a row write was attempted so even on the error path we
			// must not allow another write header attempt in any way
			w.bitFlags |= wFlagHeaderWritten

			return ErrInvalidFieldCountInRecord
		}
		w.numFields = n
	}

	return nil
}

func (w *Writer) setErr(err error) {
	w.err = err
}

//
// helpers
//

func badRecordSeparatorWConfig(cfg *wCfg) {
	cfg.recordSepRuneLen = 0
}

func isNewlineRuneForWrite(c rune) bool {
	switch c {
	case asciiCarriageReturn, asciiLineFeed, asciiVerticalTab, asciiFormFeed, utf8NextLine, utf8LineSeparator:
		return true
	}
	return false
}

func isFieldWriterRune(runeList ...rune) bool {
	for _, r := range runeList {
		if strings.ContainsRune(fieldWriterTypesRuneList, r) {
			return true
		}
	}

	return false
}

func isValidComment(comment, quote, fieldSep, escape rune, escapeSet bool) error {
	// note: not letting comment runes be newline characters is
	// opinionated but it is a good practice leaning towards simplicity.
	//
	// I am open to a PR that makes this more flexible.

	if !validUtf8Rune(comment) || isNewlineRuneForWrite(comment) {
		return errors.New("invalid comment rune")
	}

	if quote == comment {
		return errors.New("invalid quote and comment rune combination")
	}

	if fieldSep == comment {
		return errors.New("invalid field separator and comment rune combination")
	}

	if escapeSet && escape == comment {
		return errors.New("invalid escape and comment rune combination")
	}

	return nil
}

// intSumOverflowCheck will be kept super small
// so it can be easily inlined when used
func intSumOverflowCheck(sum, termAdded int) {
	if sum >= termAdded {
		return
	}

	panic(panicIntOverflow)
}
