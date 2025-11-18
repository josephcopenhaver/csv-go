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
	recordBuf []byte
	w         *Writer
}

func (w *Writer) NewRecord() *RecordWriter {
	recordBuf := w.recordBuf
	w.recordBuf = nil

	return &RecordWriter{
		err:       w.err,
		bitFlags:  w.bitFlags,
		recordBuf: recordBuf,
		w:         w,
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
		if recordBuf := rw.recordBuf; recordBuf != nil {
			rw.recordBuf = nil
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
		return rw.bytes_memclearOff(p)
	}

	return rw.bytes_memclearOn(p)
}

func (rw *RecordWriter) String(s string) *RecordWriter {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		return rw.string_memclearOff(s)
	}

	return rw.string_memclearOn(s)
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

func (rw *RecordWriter) setRecordBuf(p []byte) {
	old := rw.recordBuf
	rw.recordBuf = p

	if cap(old) == 0 {
		return
	}
	old = old[:cap(old)]

	if &old[0] == &p[0] {
		return
	}

	clear(old)
}

func (rw *RecordWriter) appendStrRec(s string) {
	appendStrAndClear(&rw.recordBuf, s)
}

func (rw *RecordWriter) appendRec(p []byte) {
	appendAndClear(&rw.recordBuf, p)
}
