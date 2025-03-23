package csv

import (
	"unicode/utf8"
	"unsafe"
)

func (w *Writer) processField_escapeUnset_quoteUnforced_memclearDisabled(v []byte) (int, error) {
	var si, i, di int
	var r rune

	for {
		r, di = utf8.DecodeRune(v[i:])
		switch di {
		case 0:
			return -1, nil
		case 1:
			if r == utf8.RuneError {
				if w.errOnNonUTF8 {
					return -1, ErrNonUTF8InRecord
				}

				i += di
				continue
			}
		}

		switch r {
		case w.quote:
			w.fieldBuf = append(w.fieldBuf, v[:i]...)
			w.fieldBuf = append(w.fieldBuf, w.escapedQuote[:w.escapedQuoteByteLen]...)

			i += di
			si = i
		default:
			i += di

			if !w.runeRequiresQuotes(r) {
				continue
			}
		}

		break
	}

	si2, err := w.escapeChars_escapeDisabled_memclearDisabled(v[si:], i-si)
	if err != nil {
		return -1, err
	}

	return si + si2, nil
}

func (w *Writer) processField_escapeSet_quoteUnforced_memclearDisabled(v []byte) (int, error) {
	var si, i, di int
	var r rune

	for {
		r, di = utf8.DecodeRune(v[i:])
		switch di {
		case 0:
			return -1, nil
		case 1:
			if r == utf8.RuneError {
				if w.errOnNonUTF8 {
					return -1, ErrNonUTF8InRecord
				}

				i += di
				continue
			}
		}

		switch r {
		case w.quote:
			w.fieldBuf = append(w.fieldBuf, v[:i]...)
			w.fieldBuf = append(w.fieldBuf, w.escapedQuote[:w.escapedQuoteByteLen]...)

			i += di
			si = i
		case w.escape:
			w.fieldBuf = append(w.fieldBuf, v[:i]...)
			w.fieldBuf = append(w.fieldBuf, w.escapedEscape[:w.escapedEscapeByteLen]...)

			i += di
			si = i
		default:
			i += di

			if !w.runeRequiresQuotes(r) {
				continue
			}
		}

		break
	}

	si2, err := w.escapeChars_escapeEnabled_memclearDisabled(v[si:], i-si)
	if err != nil {
		return -1, err
	}

	return si + si2, nil
}

func (w *Writer) processField_escapeUnset_quoteForced_memclearDisabled(v []byte) (int, error) {

	n, err := w.escapeChars_escapeDisabled_memclearDisabled(v, 0)
	if err != nil {
		return -1, err
	}

	return n, nil
}

func (w *Writer) processField_escapeSet_quoteForced_memclearDisabled(v []byte) (int, error) {

	n, err := w.escapeChars_escapeEnabled_memclearDisabled(v, 0)
	if err != nil {
		return -1, err
	}

	return n, nil
}

func (w *Writer) processField_escapeUnset_quoteUnforced_memclearEnabled(v []byte) (int, error) {
	var si, i, di int
	var r rune

	for {
		r, di = utf8.DecodeRune(v[i:])
		switch di {
		case 0:
			return -1, nil
		case 1:
			if r == utf8.RuneError {
				if w.errOnNonUTF8 {
					return -1, ErrNonUTF8InRecord
				}

				i += di
				continue
			}
		}

		switch r {
		case w.quote:
			w.appendField(v[:i], w.escapedQuote[:w.escapedQuoteByteLen])

			i += di
			si = i
		default:
			i += di

			if !w.runeRequiresQuotes(r) {
				continue
			}
		}

		break
	}

	si2, err := w.escapeChars_escapeDisabled_memclearEnabled(v[si:], i-si)
	if err != nil {
		return -1, err
	}

	return si + si2, nil
}

func (w *Writer) processField_escapeSet_quoteUnforced_memclearEnabled(v []byte) (int, error) {
	var si, i, di int
	var r rune

	for {
		r, di = utf8.DecodeRune(v[i:])
		switch di {
		case 0:
			return -1, nil
		case 1:
			if r == utf8.RuneError {
				if w.errOnNonUTF8 {
					return -1, ErrNonUTF8InRecord
				}

				i += di
				continue
			}
		}

		switch r {
		case w.quote:
			w.appendField(v[:i], w.escapedQuote[:w.escapedQuoteByteLen])

			i += di
			si = i
		case w.escape:
			w.appendField(v[:i], w.escapedEscape[:w.escapedEscapeByteLen])

			i += di
			si = i
		default:
			i += di

			if !w.runeRequiresQuotes(r) {
				continue
			}
		}

		break
	}

	si2, err := w.escapeChars_escapeEnabled_memclearEnabled(v[si:], i-si)
	if err != nil {
		return -1, err
	}

	return si + si2, nil
}

func (w *Writer) processField_escapeUnset_quoteForced_memclearEnabled(v []byte) (int, error) {

	n, err := w.escapeChars_escapeDisabled_memclearEnabled(v, 0)
	if err != nil {
		return -1, err
	}

	return n, nil
}

func (w *Writer) processField_escapeSet_quoteForced_memclearEnabled(v []byte) (int, error) {

	n, err := w.escapeChars_escapeEnabled_memclearEnabled(v, 0)
	if err != nil {
		return -1, err
	}

	return n, nil
}

func (w *Writer) escapeChars_escapeDisabled_memclearDisabled(v []byte, i int) (int, error) {
	var si, di int
	var r rune

	for {
		r, di = utf8.DecodeRune(v[i:])
		switch di {
		case 0:
			return si, nil
		case 1:
			if r == utf8.RuneError {
				if w.errOnNonUTF8 {
					return 0, ErrNonUTF8InRecord
				}

				i += di
				continue
			}
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

func (w *Writer) escapeChars_escapeEnabled_memclearDisabled(v []byte, i int) (int, error) {
	var si, di int
	var r rune

	for {
		r, di = utf8.DecodeRune(v[i:])
		switch di {
		case 0:
			return si, nil
		case 1:
			if r == utf8.RuneError {
				if w.errOnNonUTF8 {
					return 0, ErrNonUTF8InRecord
				}

				i += di
				continue
			}
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

func (w *Writer) escapeChars_escapeDisabled_memclearEnabled(v []byte, i int) (int, error) {
	var si, di int
	var r rune

	for {
		r, di = utf8.DecodeRune(v[i:])
		switch di {
		case 0:
			return si, nil
		case 1:
			if r == utf8.RuneError {
				if w.errOnNonUTF8 {
					return 0, ErrNonUTF8InRecord
				}

				i += di
				continue
			}
		}

		switch r {
		case w.quote:
			w.appendField(v[si:i], w.escapedQuote[:w.escapedQuoteByteLen])

			i += di
			si = i
		default:
			i += di
		}
	}
}

func (w *Writer) escapeChars_escapeEnabled_memclearEnabled(v []byte, i int) (int, error) {
	var si, di int
	var r rune

	for {
		r, di = utf8.DecodeRune(v[i:])
		switch di {
		case 0:
			return si, nil
		case 1:
			if r == utf8.RuneError {
				if w.errOnNonUTF8 {
					return 0, ErrNonUTF8InRecord
				}

				i += di
				continue
			}
		}

		switch r {
		case w.quote:
			w.appendField(v[si:i], w.escapedQuote[:w.escapedQuoteByteLen])

			i += di
			si = i
		case w.escape:
			w.appendField(v[si:i], w.escapedEscape[:w.escapedEscapeByteLen])

			i += di
			si = i
		default:
			i += di
		}
	}
}

func (w *Writer) writeRow_memclearDisabled(row []string) (int, error) {
	defer func() {
		w.recordBuf = w.recordBuf[:0]
	}()

	if len(row) == 0 {
		return 0, ErrRowNilOrEmpty
	}

	if w.numFields != len(row) {
		if w.numFields != -1 {
			return 0, ErrInvalidFieldCountInRecord
		}

		w.numFields = len(row)
	}

	if len(row) == 1 && row[0] == "" {
		// This is a safety feature that makes the document slightly more durable to being edited.
		// If we could guarantee that the "record terminator" is never removed by accident via
		// "whitespace removal" of editors then this is extra work with no benefit. If this ever
		// becomes disable-allowed then I would still default it to enabled behavior.

		// note that this creates quite a bit of extra characters at times
		// ideally only the last row would have this escaping as most parsers
		// would understand the rows in-between as empty-value cells
		//
		// doing this would require that we buffer the last written line
		// and either add a close or flush function we expect persons to call
		//
		// but then again this only affects tables where there is one and only one attribute that is often an empty string
		//
		// seems like an odd path to optimize for, but we could
		w.writeDoubleQuotesForRecord()
	} else {
		if err := w.writeField_memclearDisabled(w.processFirstField, row[0]); err != nil {
			return 0, err
		}

		for _, v := range row[1:] {

			// write field separator
			w.recordBuf = append(w.recordBuf, []byte(string(w.fieldSep))...)

			if err := w.writeField_memclearDisabled(w.processField, v); err != nil {
				return 0, err
			}
		}
	}

	w.recordBuf = append(w.recordBuf, w.recordSepBytes[:w.recordSepByteLen]...)

	w.recordWritten = true
	n, err := w.writer.Write(w.recordBuf)
	if err != nil {
		err := writeIOErr{err}
		w.setErr(err)
		return n, err
	}

	return n, nil
}

func (w *Writer) writeDoubleQuotesForRecord_memclearDisabled() {
	w.recordBuf = append(w.recordBuf, w.twoQuotes[:w.twoQuotesByteLen]...)
}

func (w *Writer) writeField_memclearDisabled(processField func([]byte) (int, error), input string) error {
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
	//
	// It will also never be called if the len is zero,
	// just as an extra precaution.
	v := unsafe.Slice(unsafe.StringData(input), len(input))

	si, err := processField(v)
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

func (w *Writer) writeRow_memclearEnabled(row []string) (int, error) {
	defer func() {
		w.recordBuf = w.recordBuf[:0]
	}()

	if len(row) == 0 {
		return 0, ErrRowNilOrEmpty
	}

	if w.numFields != len(row) {
		if w.numFields != -1 {
			return 0, ErrInvalidFieldCountInRecord
		}

		w.numFields = len(row)
	}

	if len(row) == 1 && row[0] == "" {
		// This is a safety feature that makes the document slightly more durable to being edited.
		// If we could guarantee that the "record terminator" is never removed by accident via
		// "whitespace removal" of editors then this is extra work with no benefit. If this ever
		// becomes disable-allowed then I would still default it to enabled behavior.

		// note that this creates quite a bit of extra characters at times
		// ideally only the last row would have this escaping as most parsers
		// would understand the rows in-between as empty-value cells
		//
		// doing this would require that we buffer the last written line
		// and either add a close or flush function we expect persons to call
		//
		// but then again this only affects tables where there is one and only one attribute that is often an empty string
		//
		// seems like an odd path to optimize for, but we could
		w.writeDoubleQuotesForRecord()
	} else {
		if err := w.writeField_memclearEnabled(w.processFirstField, row[0]); err != nil {
			return 0, err
		}

		for _, v := range row[1:] {

			// write field separator
			w.appendRec([]byte(string(w.fieldSep)))

			if err := w.writeField_memclearEnabled(w.processField, v); err != nil {
				return 0, err
			}
		}
	}

	w.appendRec(w.recordSepBytes[:w.recordSepByteLen])

	w.recordWritten = true
	n, err := w.writer.Write(w.recordBuf)
	if err != nil {
		err := writeIOErr{err}
		w.setErr(err)
		return n, err
	}

	return n, nil
}

func (w *Writer) writeDoubleQuotesForRecord_memclearEnabled() {
	w.appendRec(w.twoQuotes[:w.twoQuotesByteLen])
}

func (w *Writer) writeField_memclearEnabled(processField func([]byte) (int, error), input string) error {
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
	//
	// It will also never be called if the len is zero,
	// just as an extra precaution.
	v := unsafe.Slice(unsafe.StringData(input), len(input))

	si, err := processField(v)
	if err != nil {

		return err
	} else if si == -1 {
		// w.fieldBuf is guaranteed to be empty on this code path
		//
		// use v instead
		w.appendRec(v)

		return nil
	}

	// w.fieldBuf might have a len greater than zero on this code path
	// if it does then use it

	w.appendRec([]byte(string(w.quote)))
	if len(w.fieldBuf) > 0 {
		w.appendRec(w.fieldBuf, v[si:])

	} else {
		w.appendRec(v)
	}
	w.appendRec([]byte(string(w.quote)))

	return nil
}
