package csv

import (
	"unicode/utf8"
)

type writeBuffer struct {
	recordBuf            []byte
	escapeControlRuneSet runeSet4
	escapedQuoteSeq      twoRuneEncoder
	escapedEscapeSeq     twoRuneEncoder
	quote                rune
}

// setRecordBuf should only be called when the record buf has been appended to
// and might have been reallocated as a result and clear mem on free is enabled.
//
// This function will clear the old buffer if it is no longer being utilized.
func (wb *writeBuffer) setRecordBuf(p []byte) {
	old := wb.recordBuf
	wb.recordBuf = p

	if cap(old) == 0 {
		return
	}
	old = old[:cap(old)]

	if &old[0] == &p[0] {
		return
	}

	clear(old)
}

// TODO: migrate load* functions back to code generation

// loadQF_memclearOff is called after a
// quote, escape, or csv format sensitive character is found in the field data.
// The parent context will handle wrapping the field in quotes and communicate to this function where to
// start scanning in the source for characters to escape. The parent context will not write any part of
// the source to the record staging zone.
//
// Essentially the function picks up after the parent context starts a quoting process which the parent
// will also complete.
func (wb *writeBuffer) loadQF_memclearOff(p []byte, scanIdx int) {
	r, n, i := wb.escapeControlRuneSet.indexAnyRuneLenInBytes(p[scanIdx:])
	if i == -1 {
		wb.recordBuf = append(wb.recordBuf, p...)
		return
	}
	scanIdx += i

	//
	// found a control rune of some kind that must be escaped
	//

	wb.recordBuf = append(wb.recordBuf, p[:scanIdx]...)

	for {
		scanIdx += int(n)

		if wb.quote == r {
			wb.recordBuf = wb.escapedQuoteSeq.appendText(wb.recordBuf)
		} else {
			wb.recordBuf = wb.escapedEscapeSeq.appendText(wb.recordBuf)
		}

		r, n, i = wb.escapeControlRuneSet.indexAnyRuneLenInBytes(p[scanIdx:])
		if i == -1 {
			wb.recordBuf = append(wb.recordBuf, p[scanIdx:]...)
			return
		}

		prevIdx := scanIdx
		scanIdx += i
		wb.recordBuf = append(wb.recordBuf, p[prevIdx:scanIdx]...)
	}
}

// loadQFWithCheckUTF8_memclearOff performs the same duties as loadQField_memclearOff and in a much more expensive
// scan operation also validates that the field contents are valid utf8 sequences.
func (wb *writeBuffer) loadQFWithCheckUTF8_memclearOff(p []byte, scanIdx int) error {
	var loadIdx, n int
	var r rune
	for {
		if scanIdx >= len(p) {
			wb.recordBuf = append(wb.recordBuf, p[loadIdx:]...)
			return nil
		}

		if b := p[scanIdx]; b < utf8.RuneSelf {
			if !wb.escapeControlRuneSet.containsSingleByteRune(b) {
				scanIdx++
				continue
			}
			r = rune(b)
			n = 1
		} else if r, n = utf8.DecodeRune(p[scanIdx:]); n == 1 {
			return ErrNonUTF8InRecord
		} else if !wb.escapeControlRuneSet.containsMBRune(r) {
			scanIdx += n
			continue
		}

		//
		// found a control rune of some kind that must be escaped
		//

		wb.recordBuf = append(wb.recordBuf, p[loadIdx:scanIdx]...)

		scanIdx += n
		loadIdx = scanIdx

		if wb.quote == r {
			wb.recordBuf = wb.escapedQuoteSeq.appendText(wb.recordBuf)
			continue
		}

		wb.recordBuf = wb.escapedEscapeSeq.appendText(wb.recordBuf)
	}
}

// loadStrQF_memclearOff is called after a
// quote, escape, or csv format sensitive character is found in the field data.
// The parent context will handle wrapping the field in quotes and communicate to this function where to
// start scanning in the source for characters to escape. The parent context will not write any part of
// the source to the record staging zone.
//
// Essentially the function picks up after the parent context starts a quoting process which the parent
// will also complete.
func (wb *writeBuffer) loadStrQF_memclearOff(s string, scanIdx int) {
	r, n, i := wb.escapeControlRuneSet.indexAnyRuneLenInString(s[scanIdx:])
	if i == -1 {
		wb.recordBuf = append(wb.recordBuf, s...)
		return
	}
	scanIdx += i

	//
	// found a control rune of some kind that must be escaped
	//

	wb.recordBuf = append(wb.recordBuf, s[:scanIdx]...)

	for {
		scanIdx += int(n)

		if wb.quote == r {
			wb.recordBuf = wb.escapedQuoteSeq.appendText(wb.recordBuf)
		} else {
			wb.recordBuf = wb.escapedEscapeSeq.appendText(wb.recordBuf)
		}

		r, n, i = wb.escapeControlRuneSet.indexAnyRuneLenInString(s[scanIdx:])
		if i == -1 {
			wb.recordBuf = append(wb.recordBuf, s[scanIdx:]...)
			return
		}

		prevIdx := scanIdx
		scanIdx += i
		wb.recordBuf = append(wb.recordBuf, s[prevIdx:scanIdx]...)
	}
}

// loadStrQFWithCheckUTF8_memclearOff performs the same duties as loadStrQField_memclearOff and in a much more expensive
// scan operation also validates that the field contents are valid utf8 sequences.
func (wb *writeBuffer) loadStrQFWithCheckUTF8_memclearOff(s string, scanIdx int) error {
	var loadIdx, n int
	var r rune
	for {
		if scanIdx >= len(s) {
			wb.recordBuf = append(wb.recordBuf, s[loadIdx:]...)
			return nil
		}

		if b := s[scanIdx]; b < utf8.RuneSelf {
			if !wb.escapeControlRuneSet.containsSingleByteRune(b) {
				scanIdx++
				continue
			}
			r = rune(b)
			n = 1
		} else if r, n = utf8.DecodeRuneInString(s[scanIdx:]); n == 1 {
			return ErrNonUTF8InRecord
		} else if !wb.escapeControlRuneSet.containsMBRune(r) {
			scanIdx += n
			continue
		}

		//
		// found a control rune of some kind that must be escaped
		//

		wb.recordBuf = append(wb.recordBuf, s[loadIdx:scanIdx]...)

		scanIdx += n
		loadIdx = scanIdx

		if wb.quote == r {
			wb.recordBuf = wb.escapedQuoteSeq.appendText(wb.recordBuf)
			continue
		}

		wb.recordBuf = wb.escapedEscapeSeq.appendText(wb.recordBuf)
	}
}

// loadQF_memclearOn is called after a
// quote, escape, or csv format sensitive character is found in the field data.
// The parent context will handle wrapping the field in quotes and communicate to this function where to
// start scanning in the source for characters to escape. The parent context will not write any part of
// the source to the record staging zone.
//
// Essentially the function picks up after the parent context starts a quoting process which the parent
// will also complete.
func (wb *writeBuffer) loadQF_memclearOn(p []byte, scanIdx int) {
	r, n, i := wb.escapeControlRuneSet.indexAnyRuneLenInBytes(p[scanIdx:])
	if i == -1 {
		wb.appendRec(p)
		return
	}
	scanIdx += i

	//
	// found a control rune of some kind that must be escaped
	//

	wb.appendRec(p[:scanIdx])

	for {
		scanIdx += int(n)

		if wb.quote == r {
			wb.setRecordBuf(wb.escapedQuoteSeq.appendText(wb.recordBuf))
		} else {
			wb.setRecordBuf(wb.escapedEscapeSeq.appendText(wb.recordBuf))
		}

		r, n, i = wb.escapeControlRuneSet.indexAnyRuneLenInBytes(p[scanIdx:])
		if i == -1 {
			wb.appendRec(p[scanIdx:])
			return
		}

		prevIdx := scanIdx
		scanIdx += i
		wb.appendRec(p[prevIdx:scanIdx])
	}
}

// loadQFWithCheckUTF8_memclearOn performs the same duties as loadQField_memclearOn and in a much more expensive
// scan operation also validates that the field contents are valid utf8 sequences.
func (wb *writeBuffer) loadQFWithCheckUTF8_memclearOn(p []byte, scanIdx int) error {
	var loadIdx, n int
	var r rune
	for {
		if scanIdx >= len(p) {
			wb.appendRec(p[loadIdx:])
			return nil
		}

		if b := p[scanIdx]; b < utf8.RuneSelf {
			if !wb.escapeControlRuneSet.containsSingleByteRune(b) {
				scanIdx++
				continue
			}
			r = rune(b)
			n = 1
		} else if r, n = utf8.DecodeRune(p[scanIdx:]); n == 1 {
			return ErrNonUTF8InRecord
		} else if !wb.escapeControlRuneSet.containsMBRune(r) {
			scanIdx += n
			continue
		}

		//
		// found a control rune of some kind that must be escaped
		//

		wb.appendRec(p[loadIdx:scanIdx])

		scanIdx += n
		loadIdx = scanIdx

		if wb.quote == r {
			wb.setRecordBuf(wb.escapedQuoteSeq.appendText(wb.recordBuf))
			continue
		}

		wb.setRecordBuf(wb.escapedEscapeSeq.appendText(wb.recordBuf))
	}
}

// loadStrQF_memclearOn is called after a
// quote, escape, or csv format sensitive character is found in the field data.
// The parent context will handle wrapping the field in quotes and communicate to this function where to
// start scanning in the source for characters to escape. The parent context will not write any part of
// the source to the record staging zone.
//
// Essentially the function picks up after the parent context starts a quoting process which the parent
// will also complete.
func (wb *writeBuffer) loadStrQF_memclearOn(s string, scanIdx int) {
	r, n, i := wb.escapeControlRuneSet.indexAnyRuneLenInString(s[scanIdx:])
	if i == -1 {
		wb.appendStrRec(s)
		return
	}
	scanIdx += i

	//
	// found a control rune of some kind that must be escaped
	//

	wb.appendStrRec(s[:scanIdx])

	for {
		scanIdx += int(n)

		if wb.quote == r {
			wb.setRecordBuf(wb.escapedQuoteSeq.appendText(wb.recordBuf))
		} else {
			wb.setRecordBuf(wb.escapedEscapeSeq.appendText(wb.recordBuf))
		}

		r, n, i = wb.escapeControlRuneSet.indexAnyRuneLenInString(s[scanIdx:])
		if i == -1 {
			wb.appendStrRec(s[scanIdx:])
			return
		}

		prevIdx := scanIdx
		scanIdx += i
		wb.appendStrRec(s[prevIdx:scanIdx])
	}
}

// loadStrQFWithCheckUTF8_memclearOn performs the same duties as loadStrQField_memclearOn and in a much more expensive
// scan operation also validates that the field contents are valid utf8 sequences.
func (wb *writeBuffer) loadStrQFWithCheckUTF8_memclearOn(s string, scanIdx int) error {
	var loadIdx, n int
	var r rune
	for {
		if scanIdx >= len(s) {
			wb.appendStrRec(s[loadIdx:])
			return nil
		}

		if b := s[scanIdx]; b < utf8.RuneSelf {
			if !wb.escapeControlRuneSet.containsSingleByteRune(b) {
				scanIdx++
				continue
			}
			r = rune(b)
			n = 1
		} else if r, n = utf8.DecodeRuneInString(s[scanIdx:]); n == 1 {
			return ErrNonUTF8InRecord
		} else if !wb.escapeControlRuneSet.containsMBRune(r) {
			scanIdx += n
			continue
		}

		//
		// found a control rune of some kind that must be escaped
		//

		wb.appendStrRec(s[loadIdx:scanIdx])

		scanIdx += n
		loadIdx = scanIdx

		if wb.quote == r {
			wb.setRecordBuf(wb.escapedQuoteSeq.appendText(wb.recordBuf))
			continue
		}

		wb.setRecordBuf(wb.escapedEscapeSeq.appendText(wb.recordBuf))
	}
}

// appendRec appends bytes from a slice to the current record buffer
// and will clear any memory released to the GC/OS.
func (wb *writeBuffer) appendRec(p []byte) {
	appendAndClear(&wb.recordBuf, p)
}

// appendStrRec appends bytes from a string to the current record buffer
// and will clear any memory released to the GC/OS.
func (wb *writeBuffer) appendStrRec(s string) {
	appendStrAndClear(&wb.recordBuf, s)
}

//
// helpers
//

// appendAndClear appends bytes from a slice to a buffer
// and will clear any memory released to the GC/OS.
//
// It should be small and simple enough to always be inlined by the compiler.
func appendAndClear(dstPtr *[]byte, p []byte) {
	n := len(p)
	if n == 0 {
		return
	}

	dst := *dstPtr
	dstCap := cap(dst)
	dstFree := dstCap - len(dst)

	*dstPtr = append(dst, p...)

	if dstFree >= n {
		// no reallocation occurred
		return
	}

	// a reallocation definitely occurred
	// clear the old contents within `dst`

	clear(dst[:dstCap])
}

// appendStrAndClear appends bytes from a string to a buffer
// and will clear any memory released to the GC/OS.
//
// It should be small and simple enough to always be inlined by the compiler.
func appendStrAndClear(dstPtr *[]byte, s string) {
	n := len(s)
	if n == 0 {
		return
	}

	dst := *dstPtr
	dstCap := cap(dst)
	dstFree := dstCap - len(dst)

	*dstPtr = append(dst, s...)

	if dstFree >= n {
		// no reallocation occurred
		return
	}

	// a reallocation definitely occurred
	// clear the old contents within `dst`

	clear(dst[:dstCap])
}
