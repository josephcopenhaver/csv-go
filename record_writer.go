package csv

import (
	"errors"
	"time"
	"unicode/utf8"
)

var (
	// ErrRecordWriterClosed indicates that the RecordWriter has completed its lifecycle and cannot be used for further writing.
	// It does not indicate that a record was successfully written or not;
	// it only indicates that the RecordWriter instance is no longer usable for writing and is ready for garbage collection.
	//
	// For write success status, check the error return value from Write.
	ErrRecordWriterClosed = errors.New("record writer closed")
)

// A previous iteration of the RecordWriter struct had additional behavior that copied the writeBuffer
// from the parent Writer instance into the RecordWriter instance to allow for more
// isolated write operations. The intent was to reduce complexity around managing
// the writeBuffer state in the parent Writer instance during the RecordWriter lifecycle
// as well as reuse the behaviors defined on the writeBuffer type.
//
// However, the recordBuf attribute was the only part of the writeBuffer that had contents updated
// and might be reallocated during record assembly. In addition, we were explicitly copying the recordBuf
// back to the parent Writer instance writeBuffer at the end of the RecordWriter lifecycle. Given that,
// the additional complexity of managing two writeBuffer instances per active RecordWriter, and the
// minor decrease in performance it caused on each record write I've opted to no longer try to manage
// a separate writeBuffer or recordBuf instance.
//
// It did operate as a self-documenting mechanism that indicated exactly what parts of writer were being
// "checked out" and back in during the RecordWriter lifecycle, but that is not sufficient justification
// when instead an explicit comment block such as this one can be used to explain lifecycle concerns and data flow
// much more clearly/exhaustively alongside the actual implementation.
//
// A RecordWriter is used to assemble and write a single CSV record to an underlying io.Writer managed by the csv.Writer
// instance. The csv.Writer instance manages the overall CSV writing process including configuration options, buffering,
// and error handling. The RecordWriter locks the csv.Writer instance for its lifecycle to prevent concurrent writes
// that could corrupt the output or lead to inconsistent state within the csv.Writer. A RecordWriter technically has full
// access to the parent csv.Writer instance during its lifecycle, but it will only modify the recordBuf attribute of the
// writeBuffer within the parent csv.Writer instance and the bitFlags attribute to manage lifecycle state. If the number
// of columns is being discovered it will also update the numFields attribute when the first record is written. To perform
// its duties, the RecordWriter provides methods to append various data types as fields to the current record being
// assembled. Once all desired fields have been appended, the Write method flushes the assembled record to the underlying
// io.Writer and releases the lock on the parent csv.Writer instance, allowing for additional writing operations to proceed.
//
// A RecordWriter requires read access of all the contents of the parent csv.Writer's writeBuffer to properly
// format and write the CSV record. It will only ever change the Writer's numFields, writeBuffer.recordBuf, and
// bitFlags attributes.

// RecordWriter instances must always have life-cycles
// that end with calls to Write and/or Rollback.
//
// Failure to do so will leave the parent writer in a locked state.
type RecordWriter struct {
	err       error
	bitFlags  wFlag
	nextField int
	w         *Writer
}

// NewRecord creates a new RecordWriter instance associated with
// the parent csv Writer instance.
//
// The returned RecordWriter instance must have its lifecycle
// ended with a call to Write and/or Rollback.
//
// Concurrent calls to NewRecord are not supported.
//
// While the RecordWriter is active, the parent Writer instance
// is locked from additional writing until the RecordWriter's
// lifecycle is ended with a call to Write or Rollback.
//
// If another RecordWriter is already active, NewRecord will return
// a nil RecordWriter and ErrWriterNotReady.
//
// If the parent Writer instance is in an error state or closed,
// NewRecord will return a nil RecordWriter and the existing error.
func (w *Writer) NewRecord() (*RecordWriter, error) {

	// short circuit if the parent writer is already in an error state
	if err := w.err; err != nil {
		return nil, err
	}

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

	if (bitFlags & (wFlagRecordBuffCheckedOut | wFlagForceQuoteFirstField | wFlagClosed)) != 0 {
		// note that it should be impossible for wFlagClosed to be set here
		// while err is nil, but just in case we're going to check for it
		// since this is a slow/unlikely path anyways.
		//
		// wFlagClosed indicates that the Writer is no longer usable
		//
		// wFlagForceQuoteFirstField indicates that
		// a pure write lifecycle flag which cannot be set outside of
		// a write context was set
		//
		// wFlagRecordBuffCheckedOut indicates that another
		// RecordWriter is already active
		//
		// both the last two conditions indicate invalid concurrent access
		if (bitFlags & wFlagClosed) != 0 {
			return nil, ErrWriterClosed
		}
		return nil, ErrWriterNotReady
	}
	w.bitFlags |= wFlagRecordBuffCheckedOut

	return &RecordWriter{
		bitFlags: bitFlags,
		w:        w,
	}, nil
}

// MustNewRecord is like NewRecord but panics if a RecordWriter cannot be created.
//
// It panics in two situations:
//   - when the Writer is no longer usable because it has already observed an
//     error or has been closed (I/O or lifecycle error), or
//   - when a previous RecordWriter is still active and has not been finalized
//     with Write or Rollback (programmer misuse).
//
// This helper is intended for applications that choose to treat an unusable
// Writer (closed, already errored, or with an active RecordWriter) as a fatal
// condition. Callers that need to recover from these states should use
// NewRecord and handle its error result instead. When using MustNewRecord, the
// caller is responsible for only invoking it on Writer instances that have not
// yet observed an error and that do not currently have an active RecordWriter.
func (w *Writer) MustNewRecord() *RecordWriter {
	rw, err := w.NewRecord()
	if err != nil {
		panic(err)
	}

	return rw
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
	recordBuf := rw.w.recordBuf

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
	if rw.err == nil {
		if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
			if rw.preflightCheck_memclearOff() {
				rw.bytes_memclearOff(p, false)
			}
			return rw
		}

		if rw.preflightCheck_memclearOn() {
			rw.bytes_memclearOn(p, false)
		}
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
	if rw.err == nil {
		if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
			if rw.preflightCheck_memclearOff() {
				rw.bytes_memclearOff(p, true)
			}
			return rw
		}

		if rw.preflightCheck_memclearOn() {
			rw.bytes_memclearOn(p, true)
		}
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
	if rw.err == nil {
		if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
			if rw.preflightCheck_memclearOff() {
				rw.string_memclearOff(s, false)
			}
			return rw
		}

		if rw.preflightCheck_memclearOn() {
			rw.string_memclearOn(s, false)
		}
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
	if rw.err == nil {
		if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
			if rw.preflightCheck_memclearOff() {
				rw.string_memclearOff(s, true)
			}
			return rw
		}

		if rw.preflightCheck_memclearOn() {
			rw.string_memclearOn(s, true)
		}
	}
	return rw
}

// Int64 appends a base-10 encoded int64 field to the current record.
func (rw *RecordWriter) Int64(i int64) *RecordWriter {
	if rw.err == nil {
		if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
			if rw.preflightCheck_memclearOff() {
				rw.int64_memclearOff(i)
			}
			return rw
		}

		if rw.preflightCheck_memclearOn() {
			rw.int64_memclearOn(i)
		}
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
	if rw.err == nil {
		if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
			if rw.preflightCheck_memclearOff() {
				rw.uint64_memclearOff(i)
			}
			return rw
		}

		if rw.preflightCheck_memclearOn() {
			rw.uint64_memclearOn(i)
		}
	}
	return rw
}

// Time appends a time.Time field to the current record
// as its string representation.
func (rw *RecordWriter) Time(t time.Time) *RecordWriter {
	if rw.err == nil {
		if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
			if rw.preflightCheck_memclearOff() {
				rw.time_memclearOff(t)
			}
			return rw
		}

		if rw.preflightCheck_memclearOn() {
			rw.time_memclearOn(t)
		}
	}
	return rw
}

// Bool appends a bool field to the current record,
// where true = 1 and false = 0.
func (rw *RecordWriter) Bool(b bool) *RecordWriter {
	if rw.err == nil {
		if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
			if rw.preflightCheck_memclearOff() {
				rw.bool_memclearOff(b)
			}
			return rw
		}

		if rw.preflightCheck_memclearOn() {
			rw.bool_memclearOn(b)
		}
	}
	return rw
}

// Float64 appends a base-10 encoded float64 field to the current record
// using strconv.FormatFloat with fmt='g'.
func (rw *RecordWriter) Float64(f float64) *RecordWriter {
	if rw.err == nil {
		if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
			if rw.preflightCheck_memclearOff() {
				rw.float64_memclearOff(f)
			}
			return rw
		}

		if rw.preflightCheck_memclearOn() {
			rw.float64_memclearOn(f)
		}
	}
	return rw
}

// Rune appends a rune field to the current record.
//
// The rune value is treated as UTF-8 encoded data and validated as such before writing.
//
// Please note that only valid runes can be written and attempting to write anything
// else will lead to an ErrInvalidRune error state in the RecordWriter instance.
// The error can be retrieved through the Err() method or eventually observable through
// a terminating Write call.
//
// The Writer option ErrorOnNonUTF8 does not affect this behavior!
func (rw *RecordWriter) Rune(r rune) *RecordWriter {
	if rw.err == nil {
		if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
			if rw.preflightCheck_memclearOff() {
				if !utf8.ValidRune(r) {
					rw.abort(ErrInvalidRune)
					return rw
				}

				rw.rune_memclearOff(r)
			}
			return rw
		}

		if rw.preflightCheck_memclearOn() {
			if !utf8.ValidRune(r) {
				rw.abort(ErrInvalidRune)
				return rw
			}

			rw.rune_memclearOn(r)
		}
	}
	return rw
}

// UncheckedUTF8Rune appends a rune field to the current record similarly to Rune but skips rune validation.
//
// It will not set an internal error state if a rune cannot be normally encoded to UTF8. Instead, invalid UTF-8 runes will be encoded as the UTF-8 replacement character.
//
// Please consider this to be a micro optimization and prefer Rune
// instead should there be any uncertainty in the rune value being a valid
// utf8 encodable value.
//
// WARNING: Invalid UTF-8 runes will be encoded as the UTF-8 replacement character.
func (rw *RecordWriter) UncheckedUTF8Rune(r rune) *RecordWriter {
	if rw.err == nil {
		if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
			if rw.preflightCheck_memclearOff() {
				rw.rune_memclearOff(r)
			}
			return rw
		}

		if rw.preflightCheck_memclearOn() {
			rw.rune_memclearOn(r)
		}
	}
	return rw
}

// Empty appends an empty (zero-length) field to the current record.
//
// It does not clear or reset the record or RecordWriter; it only appends a new
// empty field. To abandon the current record, use Rollback.
//
// This function is useful when the calling context is implementing a custom
// serialization scheme and needs to explicitly write empty fields. It is
// functionally equivalent to writing an empty string field through the String
// or Bytes methods but faster and more explicit in intent.
func (rw *RecordWriter) Empty() *RecordWriter {
	if rw.err == nil {
		if (rw.bitFlags & wFlagClearMemoryAfterFree) == 0 {
			rw.preflightCheck_memclearOff()
			return rw
		}

		rw.preflightCheck_memclearOn()
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
