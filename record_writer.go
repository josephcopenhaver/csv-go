package csv

import (
	"time"
)

// RecordWriter instances must always have life-cycles
// that end with calls to Write and/or Abort.
type RecordWriter struct {
	err       error
	nextField int
	numFields int
	bitFlags  wFlag
	writeBuffer
	w *Writer
}

func (w *Writer) NewRecord() *RecordWriter {
	wb := w.writeBuffer
	w.recordBuf = nil

	if (w.bitFlags & wFlagForceQuoteFirstField) != 0 {
		panic("improper concurrent access detected on record creation")
	}

	return &RecordWriter{
		err:         w.err,
		bitFlags:    w.bitFlags,
		writeBuffer: wb,
		w:           w,
	}
}

func (rw *RecordWriter) Err() error {
	return rw.err
}

func (rw *RecordWriter) Abort() bool {
	if rw.w.err == nil && (rw.bitFlags&wFlagClosed) == 0 {
		rw.nextField = 0
		rw.bitFlags |= wFlagClosed
		rw.err = nil
		if recordBuf := rw.writeBuffer.recordBuf; recordBuf != nil {
			rw.writeBuffer.recordBuf = nil
			rw.w.recordBuf, recordBuf = recordBuf, rw.w.recordBuf
			if recordBuf != nil {
				if (rw.bitFlags & wFlagClearMemoryAfterFree) != 0 {
					clear(recordBuf[:cap(recordBuf)])
				}
				panic("improper concurrent access detected during abort")
			}
		}
		return true
	}

	return false
}

func (rw *RecordWriter) Bytes(p []byte) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		return rw.bytes_memclearOff(p, rw.bitFlags)
	}

	return rw.bytes_memclearOn(p, rw.bitFlags)
}

func (rw *RecordWriter) UncheckedUTF8Bytes(p []byte) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		return rw.bytes_memclearOff(p, rw.bitFlags&(^wFlagErrOnNonUTF8))
	}

	return rw.bytes_memclearOn(p, rw.bitFlags&(^wFlagErrOnNonUTF8))
}

func (rw *RecordWriter) String(s string) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		return rw.string_memclearOff(s, rw.bitFlags)
	}

	return rw.string_memclearOn(s, rw.bitFlags)
}

func (rw *RecordWriter) UncheckedUTF8String(s string) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		return rw.string_memclearOff(s, rw.bitFlags&(^wFlagErrOnNonUTF8))
	}

	return rw.string_memclearOn(s, rw.bitFlags&(^wFlagErrOnNonUTF8))
}

func (rw *RecordWriter) Int64(i int64) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		return rw.int64_memclearOff(i)
	}

	return rw.int64_memclearOn(i)
}

func (rw *RecordWriter) Int(i int) *RecordWriter {
	return rw.Int64(int64(i))
}

func (rw *RecordWriter) Duration(d time.Duration) *RecordWriter {
	return rw.Int64(int64(d))
}

func (rw *RecordWriter) Uint64(i uint64) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		return rw.uint64_memclearOff(i)
	}

	return rw.uint64_memclearOn(i)
}

func (rw *RecordWriter) Time(t time.Time) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		return rw.time_memclearOff(t)
	}

	return rw.time_memclearOn(t)
}

func (rw *RecordWriter) Bool(b bool) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		return rw.bool_memclearOff(b)
	}

	return rw.bool_memclearOn(b)
}

func (rw *RecordWriter) Float64(f float64) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		return rw.float64_memclearOff(f)
	}

	return rw.float64_memclearOn(f)
}

func (rw *RecordWriter) Rune(r rune) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		return rw.rune_withCheckUTF8_memclearOff(r)
	}

	return rw.rune_withCheckUTF8_memclearOn(r)
}

func (rw *RecordWriter) UncheckedUTF8Rune(r rune) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		return rw.rune_memclearOff(r)
	}

	return rw.rune_memclearOn(r)
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
