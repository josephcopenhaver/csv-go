package csv

import (
	"bytes"
	"errors"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
	"unsafe"
)

var (
	ErrInvalidFieldWriter = errors.New("invalid field writer")
)

const (
	u64signBitMask uint64 = 0x8000000000000000

	maxLenSerializedTime    = 35
	maxLenSerializedBool    = 1
	maxLenSerializedFloat64 = 24
	// boundedFieldWritersMaxByteLen is the maximum output byte length of the fieldWriter types that serialize to bounded byte sizes (so types other than bytes, and string plus their UTF8 variants for the case of bytes and string)
	//
	// it should match the maxLenSerializedTime value
	boundedFieldWritersMaxByteLen = 35

	invalidRuneUTF8Encoded = 0xEFBFBD
	// note that if a utf8.RuneError rune is supplied the offset is 1 and not 2
	//
	// 2 here is completely invalid and obviously so from an encoding perspective,
	// so we use it to indicate that the rune supplied to the writer is definitely
	// invalid
	invalidRuneUTF8EncodedWithOffset = ((2 << (8 * 4)) | invalidRuneUTF8Encoded)

	// fieldWriterTypesRuneList contains all non rune, bytes, and string type output byte values which can permute into various combinations
	fieldWriterTypesRuneList = "-:.+0123456789aefInNTZ" // 0-9, float, NaN, Inf, time
)

type wFieldKind uint8

const (
	_ wFieldKind = iota
	wfkBytes
	wfkString
	wfkInt
	wfkInt64
	wfkDuration
	wfkUint64
	wfkTime
	wfkRune
	wfkBool
	wfkFloat64
)

type FieldWriter struct {
	kind  wFieldKind
	bytes []byte
	time  time.Time
	str   string
	// _64_bits holds data for kinds that can be expressed within the 64 bits of uint64:
	// int, int64, duration, uint64, rune, bool, and float64
	//
	// for []byte and string types if the first bit is set the serialization will not check the byte stream has valid UTF8 in it. No other bits will be set in this attribute when the sub-type is one of those two.
	_64_bits uint64
}

func (w *FieldWriter) isZeroLen() bool {
	switch w.kind {
	case wfkBytes:
		return len(w.bytes) == 0
	case wfkString:
		return len(w.str) == 0
	case wfkInt, wfkInt64, wfkDuration, wfkUint64, wfkTime, wfkRune, wfkBool, wfkFloat64:
		return false
	default:
		// I reserve the right to panic here in the future should I wish to.
		return false
	}
}

func (w *FieldWriter) startsWithRune(buf []byte, r rune) bool {
	p := []byte(string(r))

	switch w.kind {
	case wfkBytes:
		if len(w.bytes) == 0 {
			return false
		}

		return bytes.HasPrefix(w.bytes, p)
	case wfkString:
		s := w.str
		if len(s) == 0 {
			return false
		}

		b := unsafe.Slice(unsafe.StringData(s), len(s))

		return bytes.HasPrefix(b, p)
	case wfkInt, wfkInt64, wfkDuration, wfkUint64, wfkTime, wfkRune, wfkBool, wfkFloat64:
		var err error
		buf, err = w.AppendText(buf)
		return (err == nil && bytes.HasPrefix(buf, p))
	default:
		// I reserve the right to panic here in the future should I wish to.
		return false
	}
}

func (w *FieldWriter) runeAppendText(b []byte) ([]byte, error) {
	r := w._64_bits
	if r == invalidRuneUTF8EncodedWithOffset {
		return nil, ErrInvalidRune
	}

	var buf [utf8.UTFMax]byte
	buf[3] = byte(r)
	buf[2] = byte(r >> (8 * 1))
	buf[1] = byte(r >> (8 * 2))
	buf[0] = byte(r >> (8 * 3))
	offset := uint8(r >> (8 * 4))

	return append(b, buf[offset:]...), nil
}

func (w *FieldWriter) AppendText(b []byte) ([]byte, error) {

	switch w.kind {
	case wfkBytes:
		return append(b, w.bytes...), nil
	case wfkString:
		return append(b, w.str...), nil
	case wfkInt, wfkInt64, wfkDuration:
		return strconv.AppendInt(b, int64(w._64_bits), 10), nil
	case wfkUint64:
		return strconv.AppendUint(b, uint64(w._64_bits), 10), nil
	case wfkTime:
		return w.time.AppendFormat(b, time.RFC3339Nano), nil
	case wfkRune:
		return w.runeAppendText(b)
	case wfkBool:
		boolAsByte := byte('0') + byte(w._64_bits)
		return append(b, boolAsByte), nil
	case wfkFloat64:
		return strconv.AppendFloat(b, math.Float64frombits(w._64_bits), 'g', -1, 64), nil
	}

	return nil, ErrInvalidFieldWriter
}

func (w *FieldWriter) MarshalText() (text []byte, err error) {
	var b []byte

	switch w.kind {
	case wfkBytes:
		// Note: don't be tempted to just return the inner buffer here
		// there is no contract with the calling context to never modify the slice it gets in return
		n := len(w.bytes)
		if n == 0 {
			return nil, nil
		}
		b = make([]byte, 0, n)
	case wfkString:
		// Note: don't be tempted to just return the inner buffer here
		// there is no contract with the calling context to never modify the slice it gets in return
		n := len(w.str)
		if n == 0 {
			return nil, nil
		}
		b = make([]byte, 0, n)
	case wfkInt, wfkInt64, wfkDuration:
		// base10ByteLen at most would be 20 if negative, 19 if positive
		//
		// might be better off making a buffer pool of a fixed size
		// but I leave that for callers of AppendText to implement
		input := w._64_bits
		var signAdjustment int
		if input&u64signBitMask != 0 {
			input = uint64(-int64(w._64_bits))
			signAdjustment = 1
		}
		base10ByteLen := decLenU64(input) + signAdjustment

		b = make([]byte, 0, base10ByteLen)
	case wfkUint64:
		// base10ByteLen at most would be 20 (always positive)
		//
		// might be better off making a buffer pool of a fixed size
		// but I leave that for callers of AppendText to implement
		base10ByteLen := decLenU64(w._64_bits)

		b = make([]byte, 0, base10ByteLen)
	case wfkTime:
		b = make([]byte, 0, maxLenSerializedTime)
	case wfkRune:
		// would be at most utf8.UTFMax (4) bytes in length and never empty
		b = make([]byte, 0, utf8.UTFMax)
	case wfkBool:
		b = make([]byte, 0, maxLenSerializedBool)
	case wfkFloat64:
		b = make([]byte, 0, maxLenSerializedFloat64)
	}

	return w.AppendText(b)
}

type FieldWriterFactory struct{}

func (FieldWriterFactory) Bytes(p []byte) FieldWriter {
	return FieldWriter{
		kind:  wfkBytes,
		bytes: p,
	}
}

// UncheckedUTF8Bytes serializes the same way as Bytes except that
// the content is not validated for utf8 compliance in any way.
//
// Please consider this to be a micro optimization and prefer Bytes
// instead should there be any uncertainty in the encoding of the
// byte contents.
func (FieldWriterFactory) UncheckedUTF8Bytes(p []byte) FieldWriter {
	return FieldWriter{
		kind:     wfkBytes,
		bytes:    p,
		_64_bits: 1,
	}
}

func (FieldWriterFactory) String(s string) FieldWriter {
	return FieldWriter{
		kind: wfkString,
		str:  s,
	}
}

// UncheckedUTF8String serializes the same way as String except that
// the content is not validated for utf8 compliance in any way.
//
// Please consider this to be a micro optimization and prefer String
// instead should there be any uncertainty in the encoding of the
// byte contents.
func (FieldWriterFactory) UncheckedUTF8String(s string) FieldWriter {
	return FieldWriter{
		kind:     wfkString,
		str:      s,
		_64_bits: 1,
	}
}

func (FieldWriterFactory) Int(i int) FieldWriter {
	return FieldWriter{
		kind:     wfkInt,
		_64_bits: uint64(i),
	}
}

func (FieldWriterFactory) Int64(i int64) FieldWriter {
	return FieldWriter{
		kind:     wfkInt64,
		_64_bits: uint64(i),
	}
}

func (FieldWriterFactory) Uint64(i uint64) FieldWriter {
	return FieldWriter{
		kind:     wfkUint64,
		_64_bits: i,
	}
}

func (FieldWriterFactory) Time(t time.Time) FieldWriter {
	return FieldWriter{
		kind: wfkTime,
		time: t,
	}
}

// Rune value must be a valid utf8 rune value otherwise
// attempting to write the rune will result in an
// ErrInvalidRune error.
func (FieldWriterFactory) Rune(r rune) FieldWriter {
	numBytes := utf8.RuneLen(r)
	if numBytes == -1 {
		return FieldWriter{
			kind:     wfkRune,
			_64_bits: invalidRuneUTF8EncodedWithOffset,
		}
	}

	var buf [utf8.UTFMax]byte
	offset := uint8(utf8.UTFMax - numBytes)
	utf8.EncodeRune(buf[offset:], r)

	v := (uint64(offset) << (8 * 4)) |
		(uint64(buf[0]) << (8 * 3)) |
		(uint64(buf[1]) << (8 * 2)) |
		(uint64(buf[2]) << (8 * 1)) |
		uint64(buf[3])

	return FieldWriter{
		kind:     wfkRune,
		_64_bits: v,
	}
}

func (FieldWriterFactory) Bool(b bool) FieldWriter {
	result := FieldWriter{
		kind: wfkBool,
	}

	if b {
		result._64_bits = 1
	}

	return result
}

func (FieldWriterFactory) Duration(d time.Duration) FieldWriter {
	return FieldWriter{
		kind:     wfkDuration,
		_64_bits: uint64(d),
	}
}

func (FieldWriterFactory) Float64(f float64) FieldWriter {
	return FieldWriter{
		kind:     wfkFloat64,
		_64_bits: math.Float64bits(f),
	}
}

func FieldWriters() FieldWriterFactory {
	return FieldWriterFactory{}
}

func isFieldWriterRune(runeList []rune) bool {
	for _, r := range runeList {
		if strings.ContainsRune(fieldWriterTypesRuneList, r) {
			return true
		}
	}

	return false
}
