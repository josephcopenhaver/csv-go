package csv

import (
	"errors"
	"time"
)

var (
	// ErrRecordWriterClosed is returned by a call to Err() or Write() when you attempt to use a record writer that has already completed its lifecycle. A record writer lifecycle is complete when Rollback is called or Write is called.
	//
	// Note that ErrRecordWritten is a sub-class of this error.
	ErrRecordWriterClosed = errors.New("record writer closed")

	// ErrRecordWritten is returned from the call to Err() after a write succeeds.
	//
	// It is a subclass of the error ErrRecordWriterClosed.
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

	if (bitFlags & (wFlagForceQuoteFirstField | wFlagRecordBuffCheckedOut)) != 0 {
		// a pure write lifecycle flag which cannot be set outside of
		// a write context was set
		panic("invalid concurrent access detected on record creation")
	}

	if err == nil && (bitFlags&wFlagClosed) == 0 {
		wb = w.writeBuffer
		w.recordBuf = nil
		w.bitFlags |= wFlagRecordBuffCheckedOut
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
		rw.w.bitFlags &= (^wFlagRecordBuffCheckedOut)
		return
	}
	rw.recordBuf = nil

	if rw.w.err != nil {
		if (rw.bitFlags & wFlagClearMemoryAfterFree) != 0 {
			clear(recordBuf[:cap(recordBuf)])
		}
		return
	}

	rw.w.recordBuf = recordBuf
	rw.w.bitFlags &= (^wFlagRecordBuffCheckedOut)
}

func (rw *RecordWriter) Rollback() {
	rw.abort(ErrRecordWriterClosed)
}

func (rw *RecordWriter) Bytes(p []byte) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if rw.preflightCheck_memclearOff() {
			rw.bytes_memclearOff(p, false)
		}
		return rw
	}

	if rw.preflightCheck_memclearOn() {
		rw.bytes_memclearOn(p, false)
	}
	return rw
}

func (rw *RecordWriter) UncheckedUTF8Bytes(p []byte) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if rw.preflightCheck_memclearOff() {
			rw.bytes_memclearOff(p, true)
		}
		return rw
	}

	if rw.preflightCheck_memclearOn() {
		rw.bytes_memclearOn(p, true)
	}
	return rw
}

func (rw *RecordWriter) String(s string) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if rw.preflightCheck_memclearOff() {
			rw.string_memclearOff(s, false)
		}
		return rw
	}

	if rw.preflightCheck_memclearOn() {
		rw.string_memclearOn(s, false)
	}
	return rw
}

func (rw *RecordWriter) UncheckedUTF8String(s string) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if rw.preflightCheck_memclearOff() {
			rw.string_memclearOff(s, true)
		}
		return rw
	}

	if rw.preflightCheck_memclearOn() {
		rw.string_memclearOn(s, true)
	}
	return rw
}

func (rw *RecordWriter) Int64(i int64) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if rw.preflightCheck_memclearOff() {
			rw.int64_memclearOff(i)
		}
		return rw
	}

	if rw.preflightCheck_memclearOn() {
		rw.int64_memclearOn(i)
	}
	return rw
}

func (rw *RecordWriter) Int(i int) *RecordWriter {
	return rw.Int64(int64(i))
}

func (rw *RecordWriter) Duration(d time.Duration) *RecordWriter {
	return rw.Int64(int64(d))
}

func (rw *RecordWriter) Uint64(i uint64) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if rw.preflightCheck_memclearOff() {
			rw.uint64_memclearOff(i)
		}
		return rw
	}

	if rw.preflightCheck_memclearOn() {
		rw.uint64_memclearOn(i)
	}
	return rw
}

func (rw *RecordWriter) Time(t time.Time) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if rw.preflightCheck_memclearOff() {
			rw.time_memclearOff(t)
		}
		return rw
	}

	if rw.preflightCheck_memclearOn() {
		rw.time_memclearOn(t)
	}
	return rw
}

func (rw *RecordWriter) Bool(b bool) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if rw.preflightCheck_memclearOff() {
			rw.bool_memclearOff(b)
		}
		return rw
	}

	if rw.preflightCheck_memclearOn() {
		rw.bool_memclearOn(b)
	}
	return rw
}

func (rw *RecordWriter) Float64(f float64) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if rw.preflightCheck_memclearOff() {
			rw.float64_memclearOff(f)
		}
		return rw
	}

	if rw.preflightCheck_memclearOn() {
		rw.float64_memclearOn(f)
	}
	return rw
}

func (rw *RecordWriter) Rune(r rune) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if rw.preflightCheck_memclearOff() {
			rw.rune_withCheckUTF8_memclearOff(r)
		}
		return rw
	}

	if rw.preflightCheck_memclearOn() {
		rw.rune_withCheckUTF8_memclearOn(r)
	}
	return rw
}

func (rw *RecordWriter) UncheckedUTF8Rune(r rune) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		if rw.preflightCheck_memclearOff() {
			rw.rune_memclearOff(r)
		}
		return rw
	}

	if rw.preflightCheck_memclearOn() {
		rw.rune_memclearOn(r)
	}
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
