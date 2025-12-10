package csv

import (
	"errors"
	"time"
)

var (
	// ErrRecordWriterClosed indicates that the RecordWriter has completed its lifecycle and cannot be used for further writing.
	ErrRecordWriterClosed = errors.New("record writer closed")
)

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

	// Checking that the wFlagForceQuoteFirstField flag is not set because
	// it is a pure write lifecycle flag which cannot be set outside of
	// a write context.
	//
	// If it is set here, it indicates that another RecordWriter
	// or write process is active and has not yet been written or rolled
	// back.
	//
	// This is a concurrent access violation.
	//
	// In addition, the RecordWriter uses the flag to indicate that the
	// first field of the first record is being written and under certain
	// circumstances the field may need to be "force quoted" to ensure
	// valid CSV file output given additions to the CSV specification such
	// as header comment lines and properly managing records of one column
	// length.

	if (bitFlags & (wFlagRecordBuffCheckedOut | wFlagForceQuoteFirstField)) != 0 {
		// wFlagForceQuoteFirstField indicates that
		// a pure write lifecycle flag which cannot be set outside of
		// a write context was set
		//
		// wFlagRecordBuffCheckedOut indicates that another
		// RecordWriter is already active
		//
		// both conditions indicate invalid concurrent access
		panic("invalid concurrent access detected on record creation")
	}

	if err == nil && (bitFlags&wFlagClosed) == 0 {
		wb = w.writeBuffer
		w.bitFlags |= wFlagRecordBuffCheckedOut
	} else {
		// note that it should be impossible for err to be nil here
		// while wFlagClosed is also set, but just in case...
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
// Once Rollback or Write has been called, Err will always
// return a non-nil error value. If no error occurred during
// the RecordWriter lifecycle, Err will return
// ErrRecordWriterClosed. ErrRecordWriterClosed does not
// indicate that a record was successfully written;
// it only indicates that the RecordWriter instance
// is no longer usable for writing. For write success status,
// check the error return value from Write.
func (rw *RecordWriter) Err() error {
	return rw.err
}

func (rw *RecordWriter) abort(err error) {
	// Ensure that the internal error state is non-nil
	// from this point forward given the provided err
	// is never nil.
	if rw.err == nil {
		rw.err = err
	}

	if (rw.bitFlags & wFlagClosed) != 0 {
		return
	}
	rw.bitFlags |= wFlagClosed
	recordBuf := rw.recordBuf

	if rw.w.err != nil {
		// the parent writer context is already functionally closed
		// so just clear the record buffer if needed and return
		if (rw.bitFlags & wFlagClearMemoryAfterFree) != 0 {
			clear(recordBuf[:cap(recordBuf)])
		}
		return
	}

	// signal to the parent writer context that it can create
	// new RecordWriter instances and in general take over writing
	// duties again

	rw.w.recordBuf = recordBuf
	rw.w.bitFlags &= (^wFlagRecordBuffCheckedOut)
}

// Rollback releases the csv Writer for additional writing through
// another RecordWriter instance or other means without flushing
// the record buffer to the io.Writer within the csv Writer
// instance.
//
// This RecordWriter instance cannot be used for further writing
// after this call.
//
// This function is always safe to call even if the RecordWriter
// instance is already closed (write success, errored, or skipped),
// in some error state, or previously rolled back.
//
// If a previous Write call was attempted, Rollback will have no
// meaningful effect of any kind. Same applies if the RecordWriter
// instance was previously rolled back.
//
// Calling rollback will not change any existing error state in the
// RecordWriter instance should it be non-nil. If the error state
// is nil prior to calling Rollback, it will be set to
// ErrRecordWriterClosed. The error state can be retrieved
// through the Err() method.
func (rw *RecordWriter) Rollback() {
	rw.abort(ErrRecordWriterClosed)
}

// Bytes appends a byte slice field to the current record.
//
// The byte slice is treated as UTF-8 encoded data and validated as such before writing
// unless the Writer was created with the ErrorOnNonUTF8 option set to false.
//
// If the byte slice contains invalid UTF-8 and UTF-8 validation is enabled, the RecordWriter
// instance will enter an error state retrievable through the Err() method or eventually
// observable through a terminating Write call.
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
// unless the Writer was created with the ErrorOnNonUTF8 option set to false.
//
// If the string is invalid UTF-8 and UTF-8 validation is enabled, the RecordWriter
// instance will enter an error state retrievable through the Err() method or eventually
// observable through a terminating Write call.
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
// unless the Writer was created with the ErrorOnNonUTF8 option set to false.
//
// If the rune is invalid UTF-8 and UTF-8 validation is enabled, the RecordWriter
// instance will enter an error state retrievable through the Err() method or eventually
// observable through a terminating Write call.
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
// If the write is successful then this function will return a nil error.
// It is the only opportunity to retrieve such a "write success"
// status from the RecordWriter instance.
//
// This RecordWriter instance cannot be used for further writing
// after this call.
func (rw *RecordWriter) Write() (int, error) {
	if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
		return rw.write_memclearOff()
	}

	return rw.write_memclearOn()
}
