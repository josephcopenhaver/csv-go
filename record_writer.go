package csv

import (
	"bytes"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
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

func (rw *RecordWriter) preflightCheck() bool {
	if rw.err != nil {
		return false
	}

	if err := rw.w.err; err != nil {
		rw.err = err
		return false
	}

	if rw.nextField != 0 {
		rw.recordBuf = rw.w.fieldSepSeq.appendText(rw.recordBuf)
	} else {
		if (rw.bitFlags & wFlagClosed) != 0 {
			panic("must not reuse record references after they are written or aborted")
		}
		rw.bitFlags = rw.w.bitFlags
		rw.numFields = rw.w.numFields
		rw.recordBuf = rw.recordBuf[:0]
		if rw.w.comment != invalidControlRune && (rw.bitFlags&wFlagFirstRecordWritten) == 0 {
			rw.bitFlags |= wFlagForceQuoteFirstField
		}
	}

	nextField := rw.nextField + 1
	if nextField <= 0 {
		panic("too many fields: integer overflow")
	}
	rw.nextField = nextField

	return true
}

func (rw *RecordWriter) unsafeAppendUTF8FieldBytes(p []byte) {
	var i int
	if (rw.bitFlags&wFlagForceQuoteFirstField) == 0 || !bytes.HasPrefix(p, []byte(string(rw.w.comment))) {
		i = rw.w.controlRuneSet.indexAnyInBytes(p)
		if i == -1 {
			rw.recordBuf = append(rw.recordBuf, p...)
			return
		}
	}

	rw.recordBuf = rw.w.quoteSeq.appendText(rw.recordBuf)
	rw.w.loadQF_memclearOff(p, i)
	rw.recordBuf = rw.w.quoteSeq.appendText(rw.recordBuf)
}

func (rw *RecordWriter) Bytes(p []byte) *RecordWriter {
	if !rw.preflightCheck() {
		return rw
	}

	if (rw.bitFlags&wFlagErrOnNonUTF8) != 0 && !utf8.Valid(p) {
		rw.err = ErrNonUTF8InRecord
		return rw
	}

	var i int
	if (rw.bitFlags&wFlagForceQuoteFirstField) == 0 || !bytes.HasPrefix(p, []byte(string(rw.w.comment))) {
		i = rw.w.controlRuneSet.indexAnyInBytes(p)
		if i == -1 {
			rw.recordBuf = append(rw.recordBuf, p...)
			return rw
		}
	}

	rw.recordBuf = rw.w.quoteSeq.appendText(rw.recordBuf)
	rw.w.loadQF_memclearOff(p, i)
	rw.recordBuf = rw.w.quoteSeq.appendText(rw.recordBuf)

	return rw
}

func (rw *RecordWriter) String(s string) *RecordWriter {
	if !rw.preflightCheck() {
		return rw
	}

	if (rw.bitFlags&wFlagErrOnNonUTF8) != 0 && !utf8.ValidString(s) {
		rw.err = ErrNonUTF8InRecord
		return rw
	}

	var i int
	if (rw.bitFlags&wFlagForceQuoteFirstField) == 0 || !strings.HasPrefix(s, string(rw.w.comment)) {
		i = rw.w.controlRuneSet.indexAnyInString(s)
		if i == -1 {
			rw.recordBuf = append(rw.recordBuf, s...)
			return rw
		}
	}

	rw.recordBuf = rw.w.quoteSeq.appendText(rw.recordBuf)
	rw.w.loadStrQF_memclearOff(s, i)
	rw.recordBuf = rw.w.quoteSeq.appendText(rw.recordBuf)

	return rw
}

func (rw *RecordWriter) Int64(i int64) *RecordWriter {
	if !rw.preflightCheck() {
		return rw
	}

	if (rw.bitFlags & (wFlagControlRuneOverlap | wFlagForceQuoteFirstField)) == 0 {
		rw.recordBuf = strconv.AppendInt(rw.recordBuf, i, 10)
		return rw
	}

	rw.unsafeAppendUTF8FieldBytes(strconv.AppendInt(rw.w.fieldWriterBuf[:0], i, 10))
	return rw
}

func (rw *RecordWriter) Int(i int) *RecordWriter {
	return rw.Int64(int64(i))
}

func (rw *RecordWriter) Duration(d time.Duration) *RecordWriter {
	return rw.Int64(int64(d))
}

func (rw *RecordWriter) Uint64(i uint64) *RecordWriter {
	if !rw.preflightCheck() {
		return rw
	}

	if (rw.bitFlags & (wFlagControlRuneOverlap | wFlagForceQuoteFirstField)) == 0 {
		rw.recordBuf = strconv.AppendUint(rw.recordBuf, i, 10)
		return rw
	}

	rw.unsafeAppendUTF8FieldBytes(strconv.AppendUint(rw.w.fieldWriterBuf[:0], i, 10))
	return rw
}

func (rw *RecordWriter) Time(t time.Time) *RecordWriter {
	if !rw.preflightCheck() {
		return rw
	}

	if (rw.bitFlags & (wFlagControlRuneOverlap | wFlagForceQuoteFirstField)) == 0 {
		rw.recordBuf = t.AppendFormat(rw.recordBuf, time.RFC3339Nano)
		return rw
	}

	rw.unsafeAppendUTF8FieldBytes(t.AppendFormat(rw.w.fieldWriterBuf[:0], time.RFC3339Nano))
	return rw
}

func (rw *RecordWriter) Bool(b bool) *RecordWriter {
	if !rw.preflightCheck() {
		return rw
	}

	v := byte('0')
	if b {
		v += 1
	}
	if (rw.bitFlags & (wFlagControlRuneOverlap | wFlagForceQuoteFirstField)) == 0 {
		rw.recordBuf = append(rw.recordBuf, v)
	} else {
		rw.w.fieldWriterBuf[0] = v
		rw.unsafeAppendUTF8FieldBytes(rw.w.fieldWriterBuf[:1])
	}

	return rw
}

func (rw *RecordWriter) Float64(f float64) *RecordWriter {
	if !rw.preflightCheck() {
		return rw
	}

	if (rw.bitFlags & (wFlagControlRuneOverlap | wFlagForceQuoteFirstField)) == 0 {
		rw.recordBuf = strconv.AppendFloat(rw.recordBuf, f, 'g', -1, 64)
	} else {
		rw.unsafeAppendUTF8FieldBytes(strconv.AppendFloat(rw.w.fieldWriterBuf[:0], f, 'g', -1, 64))
	}

	return rw
}

func (rw *RecordWriter) Rune(r rune) *RecordWriter {
	if !rw.preflightCheck() {
		return rw
	}

	if !utf8.ValidRune(r) {
		rw.err = ErrInvalidRune
		return rw
	}

	if (rw.bitFlags & wFlagForceQuoteFirstField) == 0 {
		if r < utf8.RuneSelf {
			if !rw.w.controlRuneSet.containsSingleByteRune(byte(r)) {
				goto SIMPLE_APPEND
			}
		} else if !rw.w.controlRuneSet.containsMBRune(r) {
			goto SIMPLE_APPEND
		}
	}

	rw.unsafeAppendUTF8FieldBytes(utf8.AppendRune(rw.w.fieldWriterBuf[:0], r))
	return rw

SIMPLE_APPEND:
	rw.recordBuf = utf8.AppendRune(rw.recordBuf, r)
	return rw
}

// Writer ...
//
// The record cannot be used again after this call.
func (rw *RecordWriter) Write() (int, error) {
	if err := rw.err; err != nil {
		return 0, err
	}

	switch rw.nextField {
	case 0:
		err := ErrRowNilOrEmpty
		rw.err = err
		return 0, err
	case 1:
		if numFields := rw.numFields; numFields == -1 {
			rw.w.numFields = 1
		} else if numFields != 1 {
			err := errNotEnoughFields{numFields, 1}
			rw.err = err
			return 0, err
		}
		if len(rw.recordBuf) == 0 {
			rw.recordBuf = rw.w.quoteSeq.appendText(rw.recordBuf)
			rw.recordBuf = rw.w.quoteSeq.appendText(rw.recordBuf)
		}
	default:
		if numFields := rw.numFields; numFields == -1 {
			rw.w.numFields = rw.nextField
		} else if numFields != rw.nextField {
			if numFields > rw.nextField {
				err := errNotEnoughFields{numFields, rw.nextField}
				rw.err = err
				return 0, err
			} else {
				err := ErrInvalidFieldCountInRecord
				rw.err = err
				return 0, err
			}
		}
	}

	recordBuf := rw.w.recordSepSeq.appendText(rw.recordBuf)
	rw.recordBuf = nil
	rw.nextField = 0
	rw.bitFlags |= wFlagClosed
	rw.w.recordBuf, recordBuf = recordBuf, rw.w.recordBuf
	rw.w.bitFlags |= wFlagFirstRecordWritten

	if recordBuf != nil {
		if (rw.bitFlags & wFlagClearMemoryAfterFree) != 0 {
			clear(recordBuf[:cap(recordBuf)])
		}
		panic("improper concurrent access detected during write")
	}

	if (rw.w.bitFlags & wFlagClosed) != 0 {
		if (rw.bitFlags & wFlagClearMemoryAfterFree) != 0 {
			clear(rw.w.recordBuf[:cap(rw.w.recordBuf)])
		}
		err := ErrWriterClosed
		rw.err = err
		return 0, err
	}

	n, err := rw.w.writer.Write(rw.w.recordBuf)
	if err != nil {
		err = writeIOErr{err}
		rw.err = err
		if rw.w.err == nil {
			rw.w.setErr(err)
		}
	}
	return n, err
}
