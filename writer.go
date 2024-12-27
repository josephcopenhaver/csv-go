package csv

import (
	"errors"
	"fmt"
	"io"
	"unicode/utf8"
	"unsafe"
)

var (
	ErrRowNilOrEmpty   = errors.New("row is nil or empty")
	ErrNonUTF8InRecord = errors.New("non-utf8 characters present in record")
)

type Writer struct {
	// escapeQuotesInField will scan the input byte slice
	//
	// if the slice contains a quote character then it is escaped
	// and the writes occur to the w.fieldBuf slice.
	//
	// returns true if the contents of w.fieldBuf should be escaped
	// or if the original input slice should be wrapped in quotes.
	//
	// if it returns false then w.fieldBuf has not changed and input
	// contents do not require escaping.
	//
	// to determine if w.fieldBuf or the original input slice should be used
	// when serializing simply check the size of w.fieldBuf. if it is has a
	// length greater than zero then w.fieldBuf should be used
	//
	// A portion of the input slice may need to still be copied to the record buffer
	// as well after calling this function. That slice starts at the returned index.
	// which is only valid when quoting is required and w.fieldBuf has a length
	// greater than zero
	escapeQuotesInField func([]byte) (int, bool, error)
	fieldBuf            []byte
	recordBuf           []byte
	escapeQuote         []byte
	recordSepBytes      []byte
	numFields           int
	writer              io.Writer
	err                 error
	escape              rune
	quote, fieldSep     rune
	recordSep           []rune
}

type WriterOption func(*wCfg)

type writerOpts struct{}

func WriterOpts() writerOpts {
	return writerOpts{}
}

func (writerOpts) RecordSeparator(v string) WriterOption {
	return func(cfg *wCfg) {
		cfg.recordSepStr = v
		cfg.recordSepStrSet = true
	}
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

func (writerOpts) Escape(v rune) WriterOption {
	return func(cfg *wCfg) {
		cfg.escape = v
		cfg.escapeSet = true
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

func (writerOpts) FieldCount(v int) WriterOption {
	return func(cfg *wCfg) {
		cfg.numFields = v
		cfg.numFieldsSet = true
	}
}

type wCfg struct {
	headers           []string
	writer            io.Writer
	recordSepStr      string
	recordSep         []rune
	numFields         int
	fieldSeparator    rune
	quote             rune
	escape            rune
	discoverNumFields bool
	recordSepStrSet   bool
	quoteSet          bool
	escapeSet         bool
	trimHeaders       bool
	numFieldsSet      bool
}

func (cfg *wCfg) validate() error {
	if cfg.writer == nil {
		return errors.New("writer must not be nil")
	}

	// TODO: flush out further

	if cfg.recordSepStrSet {
		s := cfg.recordSepStr
		cfg.recordSepStr = ""

		numBytes := len(s)
		if numBytes == 0 {
			return ErrBadRecordSeparator
		}

		r1, n1 := utf8.DecodeRuneInString(s)
		if n1 == numBytes {
			cfg.recordSep = []rune{r1}
		} else {

			r2, n2 := utf8.DecodeRuneInString(s[n1:])
			if n1+n2 != numBytes || (r1 != asciiCarriageReturn || r2 != asciiLineFeed) {
				return ErrBadRecordSeparator
			}

			cfg.recordSep = []rune{r1, r2}
		}
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
		// if escape would behave just like quote alone
		// then just have quote set
		if cfg.escapeSet && cfg.escape == cfg.quote {
			cfg.escapeSet = false
		}

		if cfg.fieldSeparator == cfg.quote {
			return errors.New("invalid field separator and quote combination")
		}
	}

	if cfg.escapeSet {
		if !validUtf8Rune(cfg.escape) {
			return errors.New("invalid escape value")
		}

		if cfg.fieldSeparator == cfg.escape {
			return errors.New("invalid field separator and escape combination")
		}

		if !cfg.quoteSet {
			return errors.New("escape can only be specified when quote is also specified")
		}
	}

	if !cfg.quoteSet && !cfg.escapeSet {
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
		recordSep:      []rune{asciiLineFeed},
		fieldSeparator: ',',
	}

	for _, f := range options {
		f(&cfg)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	escape := cfg.quote
	if cfg.escapeSet {
		escape = cfg.escape
	}

	var escapeQuote, recordSepBytes []byte
	{
		n := utf8.RuneLen(escape) + utf8.RuneLen(cfg.quote) + utf8.RuneLen(cfg.recordSep[0])
		if len(cfg.recordSep) == 2 {
			n += utf8.RuneLen(cfg.recordSep[1])
		}
		buf := make([]byte, n)
		n = utf8.EncodeRune(buf, escape)
		n += utf8.EncodeRune(buf[n:], cfg.quote)
		sn := n

		escapeQuote = buf[:n]

		n += utf8.EncodeRune(buf[n:], cfg.recordSep[0])
		if len(cfg.recordSep) == 2 {
			n += utf8.EncodeRune(buf[n:], cfg.recordSep[1])
		}

		recordSepBytes = buf[sn:n]
	}

	w := &Writer{
		numFields:      cfg.numFields,
		writer:         cfg.writer,
		escapeQuote:    escapeQuote,
		recordSepBytes: recordSepBytes,
		escape:         cfg.escape, // TODO: find out what to do with the escape value algorithmically or remove it
		quote:          cfg.quote,
		recordSep:      cfg.recordSep,
		fieldSep:       cfg.fieldSeparator,
	}

	if len(w.recordSep) == 2 {
		w.escapeQuotesInField = w.escapeQuotesInFieldForCRLFRecordSep
	} else {
		w.escapeQuotesInField = w.escapeQuotesInFieldForNonCRLFRecordSep
	}

	return w, nil
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
	// unsafe make look concerning and scary, and it can be,
	// however in this case we're never writing to the slice
	// created here which is stored within `v`
	//
	// since strings are immutable as well this is actually a safe
	// usage of the unsafe package to avoid an allocation we're
	// just going to read from and then throw away before this
	// returns
	v := unsafe.Slice(unsafe.StringData(input), len(input))

	si, needsQuoting, err := w.escapeQuotesInField(v)
	if err != nil {

		return err
	} else if !needsQuoting {
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

func (w *Writer) escapeQuotesInFieldForNonCRLFRecordSep(v []byte) (int, bool, error) {
	var si, i, di int
	var r rune

	for {
		r, di = utf8.DecodeRune(v[i:])
		if di == 0 {
			return -1, false, nil
		}
		if di == 1 && r == utf8.RuneError {

			// i += di
			// continue
			return -1, false, fmt.Errorf("non-CRLF mode: %w", ErrNonUTF8InRecord)
		}

		if r == w.quote {
			// TODO: ensure no overlap possible between quote and sep values
			w.fieldBuf = append(w.fieldBuf, v[:i]...)
			w.fieldBuf = append(w.fieldBuf, w.escapeQuote...)

			i += di
			si = i
			break
		}

		i += di

		if r == w.fieldSep || r == w.recordSep[0] {
			break
		}
	}

	si2, err := w.escapeQuotes(v[si:], i-si)
	if err != nil {
		return -1, false, err
	}

	return si + si2, true, nil
}

func (w *Writer) escapeQuotesInFieldForCRLFRecordSep(v []byte) (int, bool, error) {
	var si, i, di int
	var r, prevRune rune

	for {
		r, di = utf8.DecodeRune(v[i:])
		if di == 0 {
			return -1, false, nil
		}
		if di == 1 && r == utf8.RuneError {
			// lastRuneSet = false

			// i += di
			// continue
			return -1, false, fmt.Errorf("CRLF mode: %w", ErrNonUTF8InRecord)
		}

		if r == w.quote {
			w.fieldBuf = append(w.fieldBuf, v[:i]...)
			w.fieldBuf = append(w.fieldBuf, w.escapeQuote...)

			i += di
			si = i
			break
		}

		i += di

		if r == w.fieldSep {
			break
		}

		prevRuneIsSet := (i > di)
		if prevRuneIsSet && prevRune == w.recordSep[0] && r == w.recordSep[1] {
			break
		}

		prevRune = r
	}

	si2, err := w.escapeQuotes(v[si:], i-si)
	if err != nil {
		return -1, false, err
	}

	return si + si2, true, nil
}

func (w *Writer) escapeQuotes(v []byte, i int) (int, error) {
	var si, di int
	var r rune

	for {
		r, di = utf8.DecodeRune(v[i:])
		if di == 0 {
			return si, nil
		}

		if di == 1 && r == utf8.RuneError {

			// i += di
			// continue
			return 0, ErrNonUTF8InRecord
		}

		if r != w.quote {
			i += di

			continue
		}

		w.fieldBuf = append(w.fieldBuf, v[si:i]...)
		w.fieldBuf = append(w.fieldBuf, w.escapeQuote...)

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

	w.recordBuf = append(w.recordBuf, w.recordSepBytes...)

	n, err := w.writer.Write(w.recordBuf)
	if err != nil {
		w.err = err
		return n, err
	}

	return n, nil
}
