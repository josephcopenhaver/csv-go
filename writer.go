package csv

import (
	"errors"
	"fmt"
	"io"
	"unicode/utf8"
	"unsafe"
)

// TODO: add an option to add a starting utf8 byte order marker

var (
	ErrRowNilOrEmpty   = errors.New("row is nil or empty")
	ErrNonUTF8InRecord = errors.New("non-utf8 characters present in record")
	ErrWriterClosed    = errors.New("writer closed")

	errNonUTF8InRecordNonCRLFMode = fmt.Errorf("non-CRLF mode: %w", ErrNonUTF8InRecord)
)

type Writer struct {
	// doubleQuotesInField will scan the input byte slice
	//
	// if the slice contains a quote character then it is doubled
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
	doubleQuotesInField func([]byte) (int, error)
	fieldBuf            []byte
	recordBuf           []byte
	doubleQuote         [utf8.UTFMax * 2]byte
	recordSepBytes      [utf8.UTFMax]byte
	numFields           int
	writer              io.Writer
	err                 error
	quote, fieldSep     rune
	recordSep           [2]rune
	recordSepLen        int8
	doubleQuoteByteLen  int8
	recordSepByteLen    int8
	errOnNonUTF8        bool
}

type WriterOption func(*wCfg)

type writerOpts struct{}

func WriterOpts() writerOpts {
	return writerOpts{}
}

func (writerOpts) RecordSeparator(s string) WriterOption {

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

func (writerOpts) FieldSeparator(v rune) WriterOption {
	return func(cfg *wCfg) {
		cfg.fieldSeparator = v
	}
}

func (writerOpts) Quote(v rune) WriterOption {
	return func(cfg *wCfg) {
		cfg.quote = v
		cfg.quoteSet = true
	}
}

func (writerOpts) Writer(v io.Writer) WriterOption {
	return func(cfg *wCfg) {
		cfg.writer = v
	}
}

func (writerOpts) NumFields(v int) WriterOption {
	return func(cfg *wCfg) {
		cfg.numFields = v
		cfg.numFieldsSet = true
	}
}

func (writerOpts) TrimHeaders(v bool) WriterOption {
	return func(cfg *wCfg) {
		cfg.trimHeaders = v
	}
}

func (writerOpts) Headers(v []string) WriterOption {
	return func(cfg *wCfg) {
		cfg.headers = v
	}
}

func (writerOpts) DiscoverFieldCount(v bool) WriterOption {
	return func(cfg *wCfg) {
		cfg.discoverNumFields = v
	}
}

func (writerOpts) ErrorOnNonUTF8(v bool) WriterOption {
	return func(cfg *wCfg) {
		cfg.errOnNonUTF8 = v
	}
}

type wCfg struct {
	headers           []string
	writer            io.Writer
	recordSep         [2]rune
	numFields         int
	fieldSeparator    rune
	quote             rune
	discoverNumFields bool
	quoteSet          bool
	trimHeaders       bool
	numFieldsSet      bool
	errOnNonUTF8      bool
	recordSepLen      int8
}

func (cfg *wCfg) validate() error {
	if cfg.writer == nil {
		return errors.New("writer must not be nil")
	}

	// TODO: flush out further

	if cfg.recordSepLen == -1 {
		return ErrBadRecordSeparator
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

		if cfg.trimHeaders {
			// TODO: trim
		}
	}

	if cfg.discoverNumFields {
		// TODO: confirm behavior with the reader implementation
		if cfg.headers != nil {
			return errors.New("when headers are specified, field count discovery must not be enabled")
		}
		if cfg.numFieldsSet {
			return errors.New("when field count is specified, field count discovery must not be enabled")
		}
	}

	if !validUtf8Rune(cfg.fieldSeparator) {
		return errors.New("invalid field separator value")
	}

	if cfg.quoteSet {
		if !validUtf8Rune(cfg.quote) {
			return errors.New("invalid quote value")
		}

		if cfg.fieldSeparator == cfg.quote {
			return errors.New("invalid field separator and quote combination")
		}
	}

	if !cfg.quoteSet {
		cfg.quote = '"'
	}

	if !cfg.numFieldsSet && cfg.headers == nil {
		if cfg.discoverNumFields {
			cfg.numFields = -1 // let -1 indicate we need to discover the field count
		} else {
			return errors.New("must specify headers, the expected field count, or enable field count discovery")
		}
	}

	return nil
}

func NewWriter(options ...WriterOption) (*Writer, error) {

	cfg := wCfg{
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

	var doubleQuote [utf8.UTFMax * 2]byte
	var recordSepBytes [utf8.UTFMax]byte
	var doubleQuoteByteLen, recordSepByteLen int8
	{
		n := utf8.EncodeRune(doubleQuote[:], cfg.quote)
		n += utf8.EncodeRune(doubleQuote[n:], cfg.quote)
		doubleQuoteByteLen = int8(n)

		n = utf8.EncodeRune(recordSepBytes[:], cfg.recordSep[0])
		if cfg.recordSepLen == 2 {
			n += utf8.EncodeRune(recordSepBytes[n:], cfg.recordSep[1])
		}
		recordSepByteLen = int8(n)
	}

	w := &Writer{
		numFields:          cfg.numFields,
		writer:             cfg.writer,
		doubleQuote:        doubleQuote,
		doubleQuoteByteLen: doubleQuoteByteLen,
		recordSepBytes:     recordSepBytes,
		recordSepByteLen:   recordSepByteLen,
		quote:              cfg.quote,
		recordSep:          cfg.recordSep,
		recordSepLen:       cfg.recordSepLen,
		fieldSep:           cfg.fieldSeparator,
		errOnNonUTF8:       cfg.errOnNonUTF8,
	}

	if w.recordSepLen == 2 || isNewlineRuneForWrite(w.recordSep[0]) {
		w.doubleQuotesInField = w.doubleQuotesInFieldForNormalFieldSep
	} else {
		w.doubleQuotesInField = w.doubleQuotesInFieldForCustomFieldSep
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
// It will never attempt to close the underlying writer
// instance.
func (w *Writer) Close() error {
	w.err = ErrWriterClosed
	return nil
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

	si, err := w.doubleQuotesInField(v)
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

func (w *Writer) doubleQuotesInFieldForCustomFieldSep(v []byte) (int, error) {
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
			// TODO: ensure no overlap possible between quote and sep values
			w.fieldBuf = append(w.fieldBuf, v[:i]...)
			w.fieldBuf = append(w.fieldBuf, w.doubleQuote[:w.doubleQuoteByteLen]...)

			i += di
			si = i
			break
		}

		i += di

		if r == w.fieldSep || isNewlineRuneForWrite(r) || r == w.recordSep[0] {
			break
		}
	}

	si2, err := w.doubleQuotes(v[si:], i-si)
	if err != nil {
		return -1, err
	}

	return si + si2, nil
}

func (w *Writer) doubleQuotesInFieldForNormalFieldSep(v []byte) (int, error) {
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
			w.fieldBuf = append(w.fieldBuf, w.doubleQuote[:w.doubleQuoteByteLen]...)

			i += di
			si = i
			break
		}

		i += di

		if r == w.fieldSep || isNewlineRuneForWrite(r) {
			break
		}
	}

	si2, err := w.doubleQuotes(v[si:], i-si)
	if err != nil {
		return -1, err
	}

	return si + si2, nil
}

func (w *Writer) doubleQuotes(v []byte, i int) (int, error) {
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

		if r != w.quote {
			i += di

			continue
		}

		w.fieldBuf = append(w.fieldBuf, v[si:i]...)
		w.fieldBuf = append(w.fieldBuf, w.doubleQuote[:w.doubleQuoteByteLen]...)

		i += di
		si = i
	}
}

func (w *Writer) WriteRow(row []string) (int, error) {
	if err := w.err; err != nil {
		return 0, err
	}
	defer func() {
		w.recordBuf = w.recordBuf[:0]
	}()

	if len(row) == 0 {
		return 0, ErrRowNilOrEmpty
	}

	if w.numFields == -1 {
		w.numFields = len(row)
	} else if w.numFields != len(row) {
		return 0, errors.New("incorrect number of fields")
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
		w.err = err
		return n, err
	}

	return n, nil
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
