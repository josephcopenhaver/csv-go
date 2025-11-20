package csv

import (
	"errors"
	"time"
)

var (
	// ErrRecordWriterClosed is returned by a call to Err() or Write() should you have attempted to use a record writer that was already used to perform a write or the staged writes were rolled back via a call to Rollback().
	//
	// Note that ErrRecordWritten is a sub-class of this error.
	ErrRecordWriterClosed = errors.New("record writer closed")

	// ErrRecordWritten is returned from the call to Err() after a write succeeds.
	//
	// It is a sub-class of the error ErrRecordWriterClosed.
	//
	// In general this error should not be used to validate a write - instead use the result of the Write call.
	ErrRecordWritten = error(&errRecordWritten{})
)

type errRecordWritten struct{}

func (e *errRecordWritten) Error() string {
	return "record written"
}

func (e *errRecordWritten) Is(target error) bool {
	return target == ErrRecordWritten || errors.Is(ErrRecordWriterClosed, target)
}

// RecordWriter instances must always have life-cycles
// that end with calls to Write and/or Rollback.
type RecordWriter struct {
	err       error
	nextField int
	numFields int
	bitFlags  wFlag
	writeBuffer
	w *Writer
}

func (w *Writer) NewRecord() *RecordWriter {
	var wb writeBuffer
	err := w.err
	bitFlags := w.bitFlags

	if (bitFlags & wFlagForceQuoteFirstField) != 0 {
		panic("improper concurrent access detected on record creation")
	}

	if err == nil && (bitFlags&wFlagClosed) == 0 {
		wb = w.writeBuffer
		w.recordBuf = nil
	} else {
		if err == nil {
			err = ErrWriterClosed
		}
		bitFlags |= wFlagClosed
	}

	return &RecordWriter{
		err:         err,
		bitFlags:    bitFlags,
		writeBuffer: wb,
		w:           w,
	}
}

func (rw *RecordWriter) Err() error {
	return rw.err
}

func (rw *RecordWriter) incrementNextField() {
	nextField := rw.nextField + 1
	if nextField <= 0 {
		panic("too many fields: integer overflow")
	}
	rw.nextField = nextField
}

func (rw *RecordWriter) abort(err error) {
	if rw.err == nil {
		rw.err = err
	}

	if (rw.bitFlags & wFlagClosed) != 0 {
		return
	}
	rw.bitFlags |= wFlagClosed

	recordBuf := rw.recordBuf
	if recordBuf == nil {
		return
	}
	rw.recordBuf = nil

	if rw.w.err != nil {
		if (rw.bitFlags & wFlagClearMemoryAfterFree) != 0 {
			clear(recordBuf[:cap(recordBuf)])
		}
		return
	}

	rw.w.recordBuf, recordBuf = recordBuf, rw.w.recordBuf
	if recordBuf != nil {
		if (rw.bitFlags & wFlagClearMemoryAfterFree) != 0 {
			clear(recordBuf[:cap(recordBuf)])
		}
		panic("improper concurrent access detected during record writer rollback")
	}
}

func (rw *RecordWriter) Rollback() {
	rw.abort(ErrRecordWriterClosed)
}

func (rw *RecordWriter) bytes(p []byte) {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if !rw.preflightCheck_memclearOff() {
			return
		}

		rw.bytes_memclearOff(p, false)
		return
	}

	if !rw.preflightCheck_memclearOn() {
		return
	}

	rw.bytes_memclearOn(p, false)
}

func (rw *RecordWriter) Bytes(p []byte) *RecordWriter {
	rw.bytes(p)
	return rw
}

func (rw *RecordWriter) uncheckedUTF8Bytes(p []byte) {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if !rw.preflightCheck_memclearOff() {
			return
		}

		rw.bytes_memclearOff(p, true)
		return
	}

	if !rw.preflightCheck_memclearOn() {
		return
	}

	rw.bytes_memclearOn(p, true)
}

func (rw *RecordWriter) UncheckedUTF8Bytes(p []byte) *RecordWriter {
	rw.uncheckedUTF8Bytes(p)
	return rw
}

func (rw *RecordWriter) string(s string) {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if !rw.preflightCheck_memclearOff() {
			return
		}

		rw.string_memclearOff(s, false)
		return
	}

	if !rw.preflightCheck_memclearOn() {
		return
	}

	rw.string_memclearOn(s, false)
}

func (rw *RecordWriter) String(s string) *RecordWriter {
	rw.string(s)
	return rw
}

func (rw *RecordWriter) uncheckedUTF8String(s string) {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if !rw.preflightCheck_memclearOff() {
			return
		}

		rw.string_memclearOff(s, true)
		return
	}

	if !rw.preflightCheck_memclearOn() {
		return
	}

	rw.string_memclearOn(s, true)
}

func (rw *RecordWriter) UncheckedUTF8String(s string) *RecordWriter {
	rw.uncheckedUTF8String(s)
	return rw
}

func (rw *RecordWriter) int64(i int64) {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if !rw.preflightCheck_memclearOff() {
			return
		}

		rw.int64_memclearOff(i)
		return
	}

	if !rw.preflightCheck_memclearOn() {
		return
	}

	rw.int64_memclearOn(i)
}

func (rw *RecordWriter) Int64(i int64) *RecordWriter {
	rw.int64(i)
	return rw
}

func (rw *RecordWriter) Int(i int) *RecordWriter {
	return rw.Int64(int64(i))
}

func (rw *RecordWriter) Duration(d time.Duration) *RecordWriter {
	return rw.Int64(int64(d))
}

func (rw *RecordWriter) uint64(i uint64) {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if !rw.preflightCheck_memclearOff() {
			return
		}

		rw.uint64_memclearOff(i)
		return
	}

	if !rw.preflightCheck_memclearOn() {
		return
	}

	rw.uint64_memclearOn(i)
}

func (rw *RecordWriter) Uint64(i uint64) *RecordWriter {
	rw.uint64(i)
	return rw
}

func (rw *RecordWriter) time(t time.Time) {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if !rw.preflightCheck_memclearOff() {
			return
		}

		rw.time_memclearOff(t)
		return
	}

	if !rw.preflightCheck_memclearOn() {
		return
	}

	rw.time_memclearOn(t)
}

func (rw *RecordWriter) Time(t time.Time) *RecordWriter {
	rw.time(t)
	return rw
}

func (rw *RecordWriter) bool(b bool) {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if !rw.preflightCheck_memclearOff() {
			return
		}

		rw.bool_memclearOff(b)
		return
	}

	if !rw.preflightCheck_memclearOn() {
		return
	}

	rw.bool_memclearOn(b)
}

func (rw *RecordWriter) Bool(b bool) *RecordWriter {
	rw.bool(b)
	return rw
}

func (rw *RecordWriter) float64(f float64) {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if !rw.preflightCheck_memclearOff() {
			return
		}

		rw.float64_memclearOff(f)
		return
	}

	if !rw.preflightCheck_memclearOn() {
		return
	}

	rw.float64_memclearOn(f)
}

func (rw *RecordWriter) Float64(f float64) *RecordWriter {
	rw.float64(f)
	return rw
}

func (rw *RecordWriter) rune(r rune) {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if !rw.preflightCheck_memclearOff() {
			return
		}

		rw.rune_withCheckUTF8_memclearOff(r)
		return
	}

	if !rw.preflightCheck_memclearOn() {
		return
	}

	rw.rune_withCheckUTF8_memclearOn(r)
}

func (rw *RecordWriter) Rune(r rune) *RecordWriter {
	rw.rune(r)
	return rw
}

func (rw *RecordWriter) uncheckedUTF8Rune(r rune) {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if !rw.preflightCheck_memclearOff() {
			return
		}

		rw.rune_memclearOff(r)
		return
	}

	if !rw.preflightCheck_memclearOn() {
		return
	}

	rw.rune_memclearOn(r)
}

func (rw *RecordWriter) UncheckedUTF8Rune(r rune) *RecordWriter {
	rw.uncheckedUTF8Rune(r)
	return rw
}

// Writer ...
//
// The record cannot be used again after this call.
func (rw *RecordWriter) Write() (int, error) {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		return rw.write_memclearOff()
	}

	return rw.write_memclearOn()
}
