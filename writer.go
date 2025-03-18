package csv

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"unicode/utf8"
	"unsafe"
)

var (
	//
	// writer specific errors
	//

	ErrRowNilOrEmpty             = errors.New("row is nil or empty")
	ErrNonUTF8InRecord           = errors.New("non-utf8 characters in record")
	ErrNonUTF8InComment          = errors.New("non-utf8 characters in comment")
	ErrWriterClosed              = errors.New("writer closed")
	ErrHeaderWritten             = errors.New("header already written")
	ErrInvalidFieldCountInRecord = errors.New("invalid field count in record")
)

type writeIOErr struct {
	err error
}

func (e writeIOErr) Unwrap() []error {
	return []error{ErrIO, e.err}
}

func (e writeIOErr) Error() string {
	return "io error: " + e.err.Error()
}

type wCfg struct {
	writer               io.Writer
	recordSep            [2]rune
	numFields            int
	fieldSeparator       rune
	quote                rune
	escape               rune
	numFieldsSet         bool
	errOnNonUTF8         bool
	escapeSet            bool
	recordSepLen         int8
	clearMemoryAfterFree bool
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
		return badRecordSeparatorWConfig
	}
	if n1+n2 == len(v) && r1 == asciiCarriageReturn && r2 == asciiLineFeed {
		return func(cfg *wCfg) {
			cfg.recordSep[0] = r1
			cfg.recordSep[1] = r2
			cfg.recordSepLen = 2
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

func (cfg *wCfg) validate() error {
	if cfg.writer == nil {
		return errors.New("nil writer")
	}

	if cfg.recordSepLen == 0 {
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

	return nil
}

type Writer struct {
	writeDoubleQuotesForRecord func()
	writeRow                   func(row []string) (int, error)
	// processField will scan the input byte slice
	//
	// if the slice contains a quote character then it is escaped
	// and the writes occur to the w.fieldBuf slice.
	//
	// if the slice contains an escape character then it is escaped
	// and the writes occur to the w.fieldBuf slice.
	//
	// returns a value greater than -1 if the contents of w.fieldBuf
	// or if the original input slice should be wrapped in quotes.
	//
	// if it returns -1 then w.fieldBuf has not changed and input
	// contents do not require escaping.
	//
	// to determine if w.fieldBuf or the original input slice should be used
	// when serializing simply check the size of w.fieldBuf. if it is has a
	// length greater than zero then w.fieldBuf should be used
	//
	// A portion of the input slice may need to still be copied to the record buffer
	// as well after calling this function. That slice starts at the returned index.
	processField            func([]byte) (int, error)
	processFirstField       func([]byte) (int, error)
	fieldBuf                []byte
	recordBuf               []byte
	twoQuotes               [utf8.UTFMax * 2]byte
	escapedQuote            [utf8.UTFMax * 2]byte
	escapedEscape           [utf8.UTFMax * 2]byte
	recordSepBytes          [utf8.UTFMax]byte
	numFields               int
	writer                  io.Writer
	err                     error
	quote, fieldSep, escape rune
	recordSep               [2]rune
	recordSepLen            int8
	twoQuotesByteLen        int8
	escapedQuoteByteLen     int8
	escapedEscapeByteLen    int8
	recordSepByteLen        int8
	escapeSet               bool
	errOnNonUTF8            bool
	headerWritten           bool
	recordWritten           bool
	clearMemoryAfterFree    bool
	closed                  bool
}

// NewWriter creates a new instance of a CSV writer which is not safe for concurrent reads.
func NewWriter(options ...WriterOption) (*Writer, error) {

	cfg := wCfg{
		numFields:      -1,
		quote:          '"',
		recordSep:      [2]rune{asciiLineFeed, 0},
		recordSepLen:   1,
		fieldSeparator: ',',
		errOnNonUTF8:   true,
	}

	for _, f := range options {
		f(&cfg)
	}

	if err := cfg.validate(); err != nil {
		return nil, errors.Join(ErrBadConfig, err)
	}

	var twoQuotes [utf8.UTFMax * 2]byte
	var escapedQuote [utf8.UTFMax * 2]byte
	var escapedEscape [utf8.UTFMax * 2]byte
	var recordSepBytes [utf8.UTFMax]byte
	var escapedQuoteByteLen, escapedEscapeByteLen, recordSepByteLen, twoQuotesByteLen int8
	if cfg.escapeSet {
		n := utf8.EncodeRune(escapedQuote[:], cfg.escape)
		n += utf8.EncodeRune(escapedQuote[n:], cfg.quote)
		escapedQuoteByteLen = int8(n)

		n = utf8.EncodeRune(escapedEscape[:], cfg.escape)
		copy(escapedEscape[n:], escapedEscape[:n])
		n <<= 1
		escapedEscapeByteLen = int8(n)
	} else {
		n := utf8.EncodeRune(escapedQuote[:], cfg.quote)
		copy(escapedQuote[n:], escapedQuote[:n])
		n <<= 1
		escapedQuoteByteLen = int8(n)
	}

	{
		n := utf8.EncodeRune(recordSepBytes[:], cfg.recordSep[0])
		if cfg.recordSepLen == 2 {
			n += utf8.EncodeRune(recordSepBytes[n:], cfg.recordSep[1])
		}
		recordSepByteLen = int8(n)
	}

	{
		n := utf8.EncodeRune(twoQuotes[:], cfg.quote)
		copy(twoQuotes[n:], twoQuotes[:n])
		n <<= 1
		twoQuotesByteLen = int8(n)
	}

	w := &Writer{
		numFields:            cfg.numFields,
		writer:               cfg.writer,
		twoQuotes:            twoQuotes,
		twoQuotesByteLen:     twoQuotesByteLen,
		escapedEscape:        escapedEscape,
		escapedEscapeByteLen: escapedEscapeByteLen,
		escapedQuote:         escapedQuote,
		escapedQuoteByteLen:  escapedQuoteByteLen,
		recordSepBytes:       recordSepBytes,
		recordSepByteLen:     recordSepByteLen,
		quote:                cfg.quote,
		recordSep:            cfg.recordSep,
		recordSepLen:         cfg.recordSepLen,
		fieldSep:             cfg.fieldSeparator,
		errOnNonUTF8:         cfg.errOnNonUTF8,
		escape:               cfg.escape,
		escapeSet:            cfg.escapeSet,
		clearMemoryAfterFree: cfg.clearMemoryAfterFree,
	}

	if w.clearMemoryAfterFree {
		w.writeRow = w.writeRow_memclearEnabled
		w.writeDoubleQuotesForRecord = w.writeDoubleQuotesForRecord_memclearEnabled
	} else {
		w.writeRow = w.writeRow_memclearDisabled
		w.writeDoubleQuotesForRecord = w.writeDoubleQuotesForRecord_memclearDisabled
	}

	w.processField = w.processFieldFunc(false)
	w.processFirstField = w.processField

	{
		f := w.writeRow
		w.writeRow = func(row []string) (int, error) {
			w.writeRow = f
			w.headerWritten = true
			return w.writeRow(row)
		}
	}

	return w, nil
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
	if w.closed {
		return nil
	}
	w.closed = true

	w.setErr(ErrWriterClosed)

	if w.clearMemoryAfterFree {
		for _, v := range [][]byte{w.fieldBuf, w.recordBuf} {
			v := v[:cap(v)]
			for i := range v {
				v[i] = 0
			}
		}
	}

	return nil
}

func (w *Writer) appendField(bufs ...[]byte) {
	for _, v := range bufs {
		old := w.fieldBuf

		w.fieldBuf = append(w.fieldBuf, v...)

		if cap(old) == 0 {
			continue
		}

		old = old[:cap(old)]

		if &old[0] == &(w.fieldBuf[:1])[0] {
			continue
		}

		for i := range old {
			old[i] = 0
		}
	}
}

func (w *Writer) appendRec(bufs ...[]byte) {
	for _, v := range bufs {
		old := w.recordBuf

		w.recordBuf = append(w.recordBuf, v...)

		if cap(old) == 0 {
			continue
		}

		old = old[:cap(old)]

		if &old[0] == &(w.recordBuf[:1])[0] {
			continue
		}

		for i := range old {
			old[i] = 0
		}
	}
}

type whCfg struct {
	headers         []string
	commentLines    []string
	commentRune     rune
	trimHeaders     bool
	headersSet      bool
	commentRuneSet  bool
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
		cfg.commentRune = r
		cfg.commentRuneSet = true
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

	if cfg.commentRuneSet {

		// note: not letting comment runes be newline characters is
		// opinionated but it is a good practice leaning towards simplicity.
		//
		// I am open to a PR that makes this more flexible.

		if !validUtf8Rune(cfg.commentRune) || isNewlineRuneForWrite(cfg.commentRune) {
			return errors.New("invalid comment rune")
		}

		if w.quote == cfg.commentRune {
			return errors.New("invalid quote and comment rune combination")
		}

		if w.fieldSep == cfg.commentRune {
			return errors.New("invalid field separator and comment rune combination")
		}

		if w.escapeSet && w.escape == cfg.commentRune {
			return errors.New("invalid escape and comment rune combination")
		}
	} else if cfg.commentLinesSet {
		return errors.New("comment lines require a comment rune")
	}

	// positive path remaining actions:

	if cfg.headersSet && w.numFields == -1 {
		w.numFields = len(cfg.headers)
	}

	return nil
}

func (w *Writer) WriteHeader(options ...WriteHeaderOption) (int, error) {
	var result int
	if err := w.err; err != nil {
		return result, err
	}

	if w.headerWritten {
		return result, ErrHeaderWritten
	}
	w.headerWritten = true

	cfg := whCfg{
		// no defaults
	}
	for _, f := range options {
		f(&cfg)
	}

	if err := cfg.validate(w); err != nil {
		return result, errors.Join(ErrBadConfig, err)
	}

	if cfg.includeBOM {
		utf8BOM := []byte{0xEF, 0xBB, 0xBF}

		n, err := w.writer.Write(utf8BOM)
		result += n
		if err != nil {
			err := writeIOErr{err}
			w.setErr(err)
			return result, err
		}
	}

	if cfg.commentLinesSet {

		var buf bytes.Buffer
		if w.clearMemoryAfterFree {
			defer func() {
				buf.Reset()
				b := buf.Bytes()
				b = b[:cap(b)]
				for i := range b {
					b[i] = 0
				}
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

				buf.WriteRune(cfg.commentRune)
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
							if w.errOnNonUTF8 {
								return result, ErrNonUTF8InComment
							}

							buf.WriteByte(line[0])

							line = line[s:]
							continue
						}

						if isNewlineRuneForWrite(r) {
							buf.Write(w.recordSepBytes[:w.recordSepByteLen])
							buf.WriteRune(cfg.commentRune)
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

		n, err := io.Copy(w.writer, &buf)
		result += int(n)
		if err != nil {
			err := writeIOErr{err}
			w.setErr(err)
			return result, err
		}

		// these slip closures handle the case where the first rune of the
		// first column of the first record match the comment rune when a
		// comment has been written to the writer so when being read the
		// data record is not interpreted as a comment by mistake
		{
			prevProcessFirstField := w.processFirstField
			prevWriteDoubleQuotesForRecord := w.writeDoubleQuotesForRecord
			commentRune := cfg.commentRune
			fieldQuoter := w.processFieldFunc(true)

			w.processFirstField = func(v []byte) (int, error) {
				if !w.recordWritten {
					if r, _ := utf8.DecodeRune(v); r == commentRune {
						return fieldQuoter(v)
					}

					return prevProcessFirstField(v)
				}

				w.processFirstField = prevProcessFirstField
				w.writeDoubleQuotesForRecord = prevWriteDoubleQuotesForRecord
				return w.processFirstField(v)
			}

			w.writeDoubleQuotesForRecord = func() {
				if !w.recordWritten {
					prevWriteDoubleQuotesForRecord()
					return
				}

				w.processFirstField = prevProcessFirstField
				w.writeDoubleQuotesForRecord = prevWriteDoubleQuotesForRecord
				w.writeDoubleQuotesForRecord()
			}
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
	return result, err
}

func (w *Writer) processFieldFunc(forceQuote bool) func(v []byte) (int, error) {
	if w.escapeSet {
		if forceQuote {
			if w.clearMemoryAfterFree {
				return w.processField_escapeSet_quoteForced_memclearEnabled
			}
			return w.processField_escapeSet_quoteForced_memclearDisabled
		}

		if w.clearMemoryAfterFree {
			return w.processField_escapeSet_quoteUnforced_memclearEnabled
		}
		return w.processField_escapeSet_quoteUnforced_memclearDisabled
	}

	if forceQuote {
		if w.clearMemoryAfterFree {
			return w.processField_escapeUnset_quoteForced_memclearEnabled
		}
		return w.processField_escapeUnset_quoteForced_memclearDisabled
	}

	if w.clearMemoryAfterFree {
		return w.processField_escapeUnset_quoteUnforced_memclearEnabled
	}

	return w.processField_escapeUnset_quoteUnforced_memclearDisabled
}

func (w *Writer) WriteRow(row ...string) (int, error) {
	return w.writeRow(row)
}

func (w *Writer) runeRequiresQuotes(r rune) bool {
	switch r {
	case w.fieldSep:
		return true
	default:
		return isNewlineRuneForWrite(r)
	}
}

func (w *Writer) setErr(err error) {
	w.err = err
	w.writeRow = func(row []string) (int, error) {
		return 0, w.err
	}
}

//
// helpers
//

func badRecordSeparatorWConfig(cfg *wCfg) {
	cfg.recordSepLen = 0
}

func isNewlineRuneForWrite(c rune) bool {
	switch c {
	case asciiCarriageReturn, asciiLineFeed, asciiVerticalTab, asciiFormFeed, utf8NextLine, utf8LineSeparator:
		return true
	}
	return false
}
