package csv

import (
	"errors"
	"io"
	"strings"
	"unicode/utf8"
)

type Writer struct {
	buf             []byte
	numFields       int
	writer          io.Writer
	err             error
	escapeQuote     string
	escape          rune
	quote, fieldSep rune
	recordSep       []rune
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

	return &Writer{
		numFields:   cfg.numFields,
		writer:      cfg.writer,
		escapeQuote: string([]rune{cfg.escape, cfg.quote}),
		escape:      cfg.escape, // TODO: find out what to do with the escape value algorithmically or remove it
		quote:       cfg.quote,
		recordSep:   cfg.recordSep,
		fieldSep:    cfg.fieldSeparator,
	}, nil
}

func (w *Writer) writeField(v string) {

	hasQuote, hasSep := w.hasQuoteOrSep(v)

	if hasQuote {
		// TODO: ? Should the replace operation happen at the same time as the scan for quotes or separators?
		v = strings.ReplaceAll(v, string(w.quote), w.escapeQuote)
	} else if !hasSep {
		w.buf = append(w.buf, []byte(v)...)
		return
	}

	w.buf = append(w.buf, []byte(string(w.quote))...)
	w.buf = append(w.buf, []byte(v)...)
	w.buf = append(w.buf, []byte(string(w.quote))...)
}

func (w *Writer) hasQuoteOrSep(input string) (bool, bool) {
	v := []byte(input)
	var hasQuote, hasSep = false, false

	i := 0
	di := 0
	var r rune

	if len(w.recordSep) == 1 {
		for {
			r, di = utf8.DecodeRune(v[i:])
			if di == 0 {
				return hasQuote, hasSep
			}
			if di == 1 && r == utf8.RuneError {
				i += di
				continue
			}

			if r == w.quote {
				// TODO: ensure no overlap possible between quote and sep values
				hasQuote = true
				break
			}

			if r == w.fieldSep {
				hasSep = true
				break
			}

			if r == w.recordSep[0] {
				hasSep = true
				break
			}

			i += di
		}

		if !hasQuote {
			for {
				r, di = utf8.DecodeRune(v[i:])
				if di == 0 {
					return hasQuote, hasSep
				}
				if di == 1 && r == utf8.RuneError {
					i += di
					continue
				}

				if r == w.quote {
					hasQuote = true
					break
				}

				i += di
			}
		} else if !hasSep {
			for {
				r, di = utf8.DecodeRune(v[i:])
				if di == 0 {
					return hasQuote, hasSep
				}
				if di == 1 && r == utf8.RuneError {
					i += di
					continue
				}

				if r == w.fieldSep {
					hasSep = true
					break
				}

				if r == w.recordSep[0] {
					hasSep = true
					break
				}

				i += di
			}
		}

		return hasQuote, hasSep
	}

	var lastRune rune
	var lastRuneSet bool

	for {
		r, di = utf8.DecodeRune(v[i:])
		if di == 0 {
			return hasQuote, hasSep
		}
		if di == 1 && r == utf8.RuneError {
			i += di
			lastRuneSet = false
			continue
		}

		if r == w.quote {
			// TODO: ensure no overlap possible between quote and sep values
			hasQuote = true
			break
		}

		if r == w.fieldSep {
			hasSep = true
			break
		}

		if lastRuneSet && lastRune == w.recordSep[0] && r == w.recordSep[1] {
			hasSep = true
			break
		}

		lastRuneSet = true
		lastRune = r
		i += di
	}

	if !hasQuote {
		for {
			r, di = utf8.DecodeRune(v[i:])
			if di == 0 {
				return hasQuote, hasSep
			}
			if di == 1 && r == utf8.RuneError {
				i += di
				continue
			}

			if r == w.quote {
				hasQuote = true
				break
			}

			i += di
		}
	} else if !hasSep {
		lastRuneSet = false

		for {
			r, di = utf8.DecodeRune(v[i:])
			if di == 0 {
				return hasQuote, hasSep
			}
			if di == 1 && r == utf8.RuneError {
				i += di
				lastRuneSet = false
				continue
			}

			if r == w.fieldSep {
				hasSep = true
				break
			}

			if lastRuneSet && lastRune == w.recordSep[0] && r == w.recordSep[1] {
				hasSep = true
				break
			}

			lastRuneSet = true
			lastRune = r
			i += di
		}
	}

	return hasQuote, hasSep
}

func (w *Writer) WriteRow(row []string) (int, error) {
	if err := w.err; err != nil {
		return 0, err
	}
	defer func() {
		w.buf = w.buf[:0]
	}()

	if len(row) == 0 {
		panic("tried to write a nil or empty row")
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
		w.buf = append(w.buf, []byte(string(w.quote))...)
		w.buf = append(w.buf, []byte(string(w.quote))...)
	} else {
		w.writeField(row[0])

		for _, v := range row[1:] {

			// write field separator
			w.buf = append(w.buf, []byte(string(w.fieldSep))...)

			w.writeField(v)
		}
	}

	w.buf = append(w.buf, []byte(string(w.recordSep))...)

	n, err := w.writer.Write(w.buf)
	if err != nil {
		w.err = err
		return n, err
	}

	return n, nil
}