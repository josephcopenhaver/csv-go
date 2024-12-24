package csv

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
	"unsafe"
)

// TODO: add an option to add a starting utf8 byte order marker

var (
	ErrRowNilOrEmpty   = errors.New("row is nil or empty")
	ErrNonUTF8InRecord = errors.New("non-utf8 characters present in record")
	ErrWriterClosed    = errors.New("writer closed")
	ErrHeaderWritten   = errors.New("header already written")

	ErrNonUTF8InComment = errors.New("TODO") // TODO

	errNonUTF8InRecordNonCRLFMode = fmt.Errorf("non-CRLF mode: %w", ErrNonUTF8InRecord)
)

type wCfg struct {
	writer         io.Writer
	recordSep      [2]rune
	numFields      int
	fieldSeparator rune
	quote          rune
	escape         rune
	numFieldsSet   bool
	errOnNonUTF8   bool
	escapeSet      bool
	recordSepLen   int8
}

type WriterOption func(*wCfg)

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

func (cfg *wCfg) validate() error {
	if cfg.writer == nil {
		return errors.New("writer must not be nil")
	}

	if cfg.recordSepLen == -1 {
		return ErrBadRecordSeparator
	}

	if cfg.numFieldsSet && cfg.numFields <= 0 {
		return errors.New("num fields must be greater than zero")
	}

	// note: not letting field separators, quotes, or escapes be newline characters is
	// opinionated but it is a good practice leaning towards simplicity.
	//
	// I am open to a PR that makes this more flexible.

	if !validUtf8Rune(cfg.fieldSeparator) || isNewlineRuneForWrite(cfg.fieldSeparator) {
		return errors.New("invalid field separator value")
	}

	if !validUtf8Rune(cfg.quote) || isNewlineRuneForWrite(cfg.fieldSeparator) {
		return errors.New("invalid quote value")
	}

	if cfg.escapeSet && (!validUtf8Rune(cfg.escape) || isNewlineRuneForWrite(cfg.fieldSeparator)) {
		return errors.New("invalid escape value")
	}

	if cfg.fieldSeparator == cfg.quote {
		return errors.New("invalid field separator and quote combination")
	}

	if cfg.escapeSet {
		if cfg.fieldSeparator == cfg.escape {
			return errors.New("invalid field separator and escape combination")
		}

		if cfg.quote == cfg.escape {
			return errors.New("invalid quote and escape combination")
		}
	}

	return nil
}

type Writer struct {
	writeRow func(row []string) (int, error)
	// scanField will scan the input byte slice
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
	fieldBuf                []byte
	recordBuf               []byte
	escapedQuote            [utf8.UTFMax * 2]byte
	escapedEscape           [utf8.UTFMax * 2]byte
	recordSepBytes          [utf8.UTFMax]byte
	numFields               int
	writer                  io.Writer
	err                     error
	quote, fieldSep, escape rune
	recordSep               [2]rune
	recordSepLen            int8
	escapedQuoteByteLen     int8
	escapedEscapeByteLen    int8
	recordSepByteLen        int8
	escapeSet               bool
	errOnNonUTF8            bool
	headersWritten          bool
}

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

	var escapedQuote [utf8.UTFMax * 2]byte
	var escapedEscape [utf8.UTFMax * 2]byte
	var recordSepBytes [utf8.UTFMax]byte
	var escapedQuoteByteLen, escapedEscapeByteLen, recordSepByteLen int8
	if cfg.escapeSet {
		n := utf8.EncodeRune(escapedQuote[:], cfg.escape)
		n += utf8.EncodeRune(escapedQuote[n:], cfg.quote)
		escapedQuoteByteLen = int8(n)

		n = utf8.EncodeRune(escapedEscape[:], cfg.escape)
		n += utf8.EncodeRune(escapedEscape[n:], cfg.escape)
		escapedEscapeByteLen = int8(n)
	} else {
		n := utf8.EncodeRune(escapedQuote[:], cfg.quote)
		n += utf8.EncodeRune(escapedQuote[n:], cfg.quote)
		escapedQuoteByteLen = int8(n)
	}

	{
		n := utf8.EncodeRune(recordSepBytes[:], cfg.recordSep[0])
		if cfg.recordSepLen == 2 {
			n += utf8.EncodeRune(recordSepBytes[n:], cfg.recordSep[1])
		}
		recordSepByteLen = int8(n)
	}

	w := &Writer{
		numFields:            cfg.numFields,
		writer:               cfg.writer,
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
	}

	w.processField = w.processFieldFunc()

	w.writeRow = w.defaultWriteRow

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
// It will never attempt to close the underlying writer
// instance.
func (w *Writer) Close() error {
	w.setErr(ErrWriterClosed)
	return nil
}

type whCfg struct {
	headers         []string
	commentLines    []string
	commentRune     rune
	trimHeaders     bool
	headersSet      bool
	commentRuneSet  bool
	commentLinesSet bool
}

type WriteHeaderOption func(*whCfg)

type WriteHeaderOptions struct{}

func WriteHeaderOpts() WriteHeaderOptions {
	return WriteHeaderOptions{}
}

func (WriteHeaderOptions) TrimHeaders(b bool) WriteHeaderOption {
	return func(cfg *whCfg) {
		cfg.trimHeaders = b
	}
}

func (WriteHeaderOptions) Headers(h []string) WriteHeaderOption {
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
		if w.escapeSet && w.escape == cfg.commentRune {
			return errors.New("invalid escape and comment rune combination")
		}
	}

	if cfg.commentLinesSet && !cfg.commentRuneSet {
		return errors.New("comment lines require a comment rune as well")
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

	if w.headersWritten {
		return result, ErrHeaderWritten
	}
	w.headersWritten = true

	var cfg whCfg
	for _, f := range options {
		f(&cfg)
	}

	if err := cfg.validate(w); err != nil {
		return result, errors.Join(ErrBadConfig, err)
	}

	if cfg.commentLinesSet {
		var buf bytes.Buffer

		if !cfg.commentRuneSet {
			return result, ErrBadConfig // TODO shape
		}

		// TODO: explain via comments why no newline sequences are allowed in comments
		// hint: makes for predictable parsing

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
				line := unsafe.Slice(unsafe.StringData(commentBuf[0]), len(commentBuf[0]))

				buf.WriteRune(cfg.commentRune)
				buf.WriteRune(' ')

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
						for _, v := range w.recordSep[:w.recordSepByteLen] {
							buf.WriteRune(v)
						}
						buf.WriteRune(cfg.commentRune)
						buf.WriteRune(' ')

						line = line[s:]
						continue
					}

					buf.WriteRune(r)
					line = line[s:]
				}

				for _, v := range w.recordSep[:w.recordSepByteLen] {
					buf.WriteRune(v)
				}

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
			return result, err
		}
	}

	if !cfg.headersSet {
		return result, nil
	}

	headers := cfg.headers
	if cfg.trimHeaders {
		for i := range headers {
			headers[i] = strings.TrimSpace(headers[i])
		}
	}

	// TODO: if providing buffer size hints, then it's likely that the headers should use a different size hint

	n, err := w.WriteRow(headers...)
	result += n
	return result, err
}

func (w *Writer) writeField(input string) error {
	if input == "" {
		return nil
	}
	defer func() {
		w.fieldBuf = w.fieldBuf[:0]
	}()

	// v here is immutable
	//
	// unsafe may look concerning and scary, and it can be,
	// however in this case we're never writing to the slice
	// created here which is stored within `v`
	//
	// since strings are immutable as well this is actually a safe
	// usage of the unsafe package to avoid an allocation we're
	// just going to read from and then throw away before this
	// returns
	v := unsafe.Slice(unsafe.StringData(input), len(input))

	si, err := w.processField(v)
	if err != nil {

		return err
	} else if si == -1 {
		// w.fieldBuf is guaranteed to be empty on this code path
		//
		// use v instead
		w.recordBuf = append(w.recordBuf, v...)

		return nil
	}

	// w.fieldBuf might have a len greater than zero on this code path
	// if it does then use it

	w.recordBuf = append(w.recordBuf, []byte(string(w.quote))...)
	if len(w.fieldBuf) > 0 {
		w.recordBuf = append(w.recordBuf, w.fieldBuf...)
		w.recordBuf = append(w.recordBuf, v[si:]...)
	} else {
		w.recordBuf = append(w.recordBuf, v...)
	}
	w.recordBuf = append(w.recordBuf, []byte(string(w.quote))...)

	return nil
}

func (w *Writer) processFieldFunc() func(v []byte) (int, error) {
	var runeRequiresQuotes func(r rune) bool

	if w.escapeSet {

		if w.recordSepLen == 2 || isNewlineRuneForWrite(w.recordSep[0]) {
			// record sep is a newline char on this path
			//
			// so no need to check for it
			runeRequiresQuotes = w.runeRequiresQuotesWithEscapeCheck
		} else {
			runeRequiresQuotes = w.runeRequiresQuotesWithRecordSepCheckWithEscapeCheck
		}

		return func(v []byte) (int, error) {
			var si, i, di int
			var r rune

			for {
				r, di = utf8.DecodeRune(v[i:])
				if di == 0 {
					return -1, nil
				}
				if di == 1 && r == utf8.RuneError {
					if w.errOnNonUTF8 {
						return -1, errNonUTF8InRecordNonCRLFMode
					}

					i += di
					continue
				}

				switch r {
				case w.quote:
					w.fieldBuf = append(w.fieldBuf, v[:i]...)
					w.fieldBuf = append(w.fieldBuf, w.escapedQuote[:w.escapedQuoteByteLen]...)

					i += di
					si = i
					break
				case w.escape:
					w.fieldBuf = append(w.fieldBuf, v[:i]...)
					w.fieldBuf = append(w.fieldBuf, w.escapedEscape[:w.escapedEscapeByteLen]...)

					i += di
					si = i
					break
				}

				i += di

				if runeRequiresQuotes(r) {
					break
				}
			}

			si2, err := w.escapeCharsWithEscape(v[si:], i-si)
			if err != nil {
				return -1, err
			}

			return si + si2, nil
		}
	}

	if w.recordSepLen == 2 || isNewlineRuneForWrite(w.recordSep[0]) {
		// record sep is a newline char on this path
		//
		// so no need to check for it
		runeRequiresQuotes = w.runeRequiresQuotes
	} else {
		runeRequiresQuotes = w.runeRequiresQuotesWithRecordSepCheck
	}

	return func(v []byte) (int, error) {
		var si, i, di int
		var r rune

		for {
			r, di = utf8.DecodeRune(v[i:])
			if di == 0 {
				return -1, nil
			}
			if di == 1 && r == utf8.RuneError {
				if w.errOnNonUTF8 {
					return -1, errNonUTF8InRecordNonCRLFMode
				}

				i += di
				continue
			}

			if r == w.quote {
				w.fieldBuf = append(w.fieldBuf, v[:i]...)
				w.fieldBuf = append(w.fieldBuf, w.escapedQuote[:w.escapedQuoteByteLen]...)

				i += di
				si = i
				break
			}

			i += di

			if runeRequiresQuotes(r) {
				break
			}
		}

		si2, err := w.escapeChars(v[si:], i-si)
		if err != nil {
			return -1, err
		}

		return si + si2, nil
	}
}

func (w *Writer) escapeCharsWithEscape(v []byte, i int) (int, error) {
	var si, di int
	var r rune

	for {
		r, di = utf8.DecodeRune(v[i:])
		if di == 0 {
			return si, nil
		}

		if di == 1 && r == utf8.RuneError {
			if w.errOnNonUTF8 {
				return 0, ErrNonUTF8InRecord
			}

			i += di
			continue
		}

		switch r {
		case w.quote:
			w.fieldBuf = append(w.fieldBuf, v[si:i]...)
			w.fieldBuf = append(w.fieldBuf, w.escapedQuote[:w.escapedQuoteByteLen]...)

			i += di
			si = i
		case w.escape:
			w.fieldBuf = append(w.fieldBuf, v[si:i]...)
			w.fieldBuf = append(w.fieldBuf, w.escapedEscape[:w.escapedEscapeByteLen]...)

			i += di
			si = i
		default:
			i += di
		}
	}
}

func (w *Writer) escapeChars(v []byte, i int) (int, error) {
	var si, di int
	var r rune

	for {
		r, di = utf8.DecodeRune(v[i:])
		if di == 0 {
			return si, nil
		}

		if di == 1 && r == utf8.RuneError {
			if w.errOnNonUTF8 {
				return 0, ErrNonUTF8InRecord
			}

			i += di
			continue
		}

		switch r {
		case w.quote:
			w.fieldBuf = append(w.fieldBuf, v[si:i]...)
			w.fieldBuf = append(w.fieldBuf, w.escapedQuote[:w.escapedQuoteByteLen]...)

			i += di
			si = i
		default:
			i += di
		}
	}
}

func (w *Writer) WriteRow(row ...string) (int, error) {
	return w.writeRow(row)
}

func (w *Writer) defaultWriteRow(row []string) (int, error) {
	defer func() {
		w.recordBuf = w.recordBuf[:0]
	}()

	if len(row) == 0 {
		return 0, ErrRowNilOrEmpty
	}

	if w.numFields != len(row) {
		if w.numFields != -1 {
			return 0, errors.New("incorrect number of fields")
		}

		w.numFields = len(row)
	}

	if len(row) == 1 && row[0] == "" {
		// note that this creates quite a bit of extra characters at times
		// ideally only the last row would have this escaping
		//
		// doing this would require that we buffer the last written line
		// and either add a close or flush function we expect persons to call
		//
		// but then again this only affects tables where there is one and only one attribute that is often an empty string
		//
		// what u doing there with that confusing gibberish mate?
		w.recordBuf = append(w.recordBuf, []byte(string(w.quote))...)
		w.recordBuf = append(w.recordBuf, []byte(string(w.quote))...)
	} else {
		if err := w.writeField(row[0]); err != nil {
			return 0, err
		}

		for _, v := range row[1:] {

			// write field separator
			w.recordBuf = append(w.recordBuf, []byte(string(w.fieldSep))...)

			if err := w.writeField(v); err != nil {
				return 0, err
			}
		}
	}

	w.recordBuf = append(w.recordBuf, w.recordSepBytes[:w.recordSepByteLen]...)

	n, err := w.writer.Write(w.recordBuf)
	if err != nil {
		w.setErr(err)
		return n, err
	}

	return n, nil
}

func (w *Writer) setErr(err error) {
	w.err = err
	w.writeRow = func(row []string) (int, error) {
		return 0, w.err
	}
}

func (w *Writer) runeRequiresQuotes(r rune) bool {
	return r == w.fieldSep || isNewlineRuneForWrite(r)
}

func (w *Writer) runeRequiresQuotesWithRecordSepCheck(r rune) bool {
	switch r {
	case w.fieldSep:
		return true
	case w.recordSep[0]:
		// record sep is not a newline char, that is why it is here
		//
		// should we require record sep always be a newline rune then
		// this becomes dead code
		return true
	default:
		return isNewlineRuneForWrite(r)
	}
}

func (w *Writer) runeRequiresQuotesWithEscapeCheck(r rune) bool {
	switch r {
	case w.fieldSep:
		return true
	case w.escape:
		return true
	default:
		return isNewlineRuneForWrite(r)
	}
}

func (w *Writer) runeRequiresQuotesWithRecordSepCheckWithEscapeCheck(r rune) bool {
	switch r {
	case w.fieldSep:
		return true
	case w.escape:
		return true
	case w.recordSep[0]:
		// record sep is not a newline char, that is why it is here
		//
		// should we require record sep always be a newline rune then
		// this becomes dead code
		return true
	default:
		return isNewlineRuneForWrite(r)
	}
}

//
// helpers
//

func badRecordSeparatorWConfig(cfg *wCfg) {
	cfg.recordSepLen = -1
}

func isNewlineRuneForWrite(c rune) bool {
	switch c {
	case asciiCarriageReturn, asciiLineFeed, asciiVerticalTab, asciiFormFeed, utf8NextLine, utf8LineSeparator:
		return true
	}
	return false
}
