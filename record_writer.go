package csv

import (
	"errors"
	"time"
)

var (
	// ErrRecordWriterClosed is returned by a call to Err() or Write() when you attempt to use a record writer that has already completed its lifecycle. A record writer lifecycle is complete when Rollback is called or Write is called.
	//
	// Note that ErrRecordWritten is a subclass of this error.
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
//
// Failure to do so will leave the parent writer in a locked state.
type RecordWriter struct {
	err       error
	nextField int
	numFields int
	bitFlags  wFlag
	writeBuffer
	w *Writer
}

// NewRecord creates a new RecordWriter instance associated with
// the parent csv Writer instance.
//
// The returned RecordWriter instance must have its lifecycle
// ended with a call to Write or Rollback.
//
// Concurrent calls to NewRecord will panic and are not supported.
//
// If the parent Writer instance is closed, the returned RecordWriter
// will have its Err() method return ErrWriterClosed.
//
// While the RecordWriter is active, the parent Writer instance
// is locked from additional writing until the RecordWriter's
// lifecycle is ended with a call to Write or Rollback.
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

// Err returns any error that has occurred during the lifecycle
// of the RecordWriter instance.
//
// If the RecordWriter has been closed through a call to
// Rollback, Err will return ErrRecordWriterClosed.
//
// If the RecordWriter has been successfully written
// through a call to Write, Err will return ErrRecordWritten
// which is a subclass of ErrRecordWriterClosed.
//
// Attempting to use a RecordWriter after its lifecycle has ended
// will force the Err method to return ErrRecordWriterClosed when
// it is next called.
//
// After a call to Write or Rollback, the RecordWriter
// instance cannot be used again for additional writing.
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

	if rw.w.err != nil {
		if (rw.bitFlags & wFlagClearMemoryAfterFree) != 0 {
			clear(recordBuf[:cap(recordBuf)])
		}
		return
	}

	rw.w.recordBuf = recordBuf
	rw.w.bitFlags &= (^wFlagRecordBuffCheckedOut)
}

// Rollback releases the csv Writer for additional writing through
// another RecordWriter instance or other means without flushing
// the record buffer to the io.Writer within the csv Writer
// instance.
//
// This RecordWriter cannot be used again after this call.
func (rw *RecordWriter) Rollback() {
	rw.abort(ErrRecordWriterClosed)
}

// Bytes appends a byte slice field to the current record.
//
// The byte slice is treated as UTF-8 encoded data and validated as such before writing
// unless the Writer was created with the DisableUTF8Validation option set to true.
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

// UncheckedUTF8Bytes appends a byte slice field to the current record
// in a similar manner to Bytes but skips UTF-8 validation.
//
// Please consider this to be a micro optimization and prefer Bytes
// instead should there be any uncertainty in the encoding of the
// byte contents.
//
// WARNING: Using this method with invalid UTF-8 data will produce invalid CSV output.
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

// String appends a string field to the current record.
//
// The string is treated as UTF-8 encoded data and validated as such before writing
// unless the Writer was created with the DisableUTF8Validation option set to true.
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

// UncheckedUTF8String appends a string field to the current record
// in a similar manner to String but skips UTF-8 validation.
//
// Please consider this to be a micro optimization and prefer String
// instead should there be any uncertainty in the encoding of the
// byte contents.
//
// WARNING: Using this method with invalid UTF-8 data will produce invalid CSV output.
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

// Int64 appends a base-10 encoded int64 field to the current record.
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

// Int appends a base-10 encoded int field to the current record.
func (rw *RecordWriter) Int(i int) *RecordWriter {
	return rw.Int64(int64(i))
}

// Duration appends a time.Duration field to the current record
// as its int64 nanosecond count base-10 string representation.
func (rw *RecordWriter) Duration(d time.Duration) *RecordWriter {
	return rw.Int64(int64(d))
}

// Uint64 appends a base-10 encoded uint64 field to the current record.
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

// Time appends a time.Time field to the current record
// as its string representation.
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

// Bool appends a bool field to the current record,
// where true = 1 and false = 0.
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

// Float64 appends a base-10 encoded float64 field to the current record
// using strconv.FormatFloat with fmt='g'.
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

// Rune appends a rune field to the current record.
//
// The rune value is treated as UTF-8 encoded data and validated as such before writing
// unless the Writer was created with the DisableUTF8Validation option set to true.
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

// Write flushes the record buffer to the io.Writer within the csv Writer
// instance and releases the csv Writer for additional writing through
// another RecordWriter instance or other means.
//
// This RecordWriter instance cannot be used again after this call.
//
// If the write is successful then this function will return a nil error,
// while the Err() method will return ErrRecordWritten.
//
// If the write errors, then the underlying error will be returned
// and the Err() method will return that same error.
func (rw *RecordWriter) Write() (int, error) {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		return rw.write_memclearOff()
	}

	return rw.write_memclearOn()
}
