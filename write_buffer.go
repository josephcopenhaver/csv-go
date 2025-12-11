package csv

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
