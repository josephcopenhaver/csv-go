package csv

import (
	"math"
	"strconv"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestFieldWriterMarshalText(t *testing.T) {
	t.Parallel()

	is := assert.New(t)

	// this function verifies serialization invariants hold over time

	//
	// Invalid
	//

	{
		var f FieldWriter
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.NotNil(err)
		is.Nil(v)
		is.Equal(ErrInvalidFieldWriter, err)
	}

	fw := FieldWriters()

	//
	// Bytes
	//

	{
		f := fw.Bytes([]byte(``))
		is.True(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(``, string(v))
	}

	{
		f := fw.Bytes([]byte(`9999999999999999999999999999999999999999999999999`))
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`9999999999999999999999999999999999999999999999999`, string(v))
	}

	// rune
	{
		f := fw.Rune('"')
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`"`, string(v))
	}

	// invalid rune
	{
		f := fw.Rune(0x808080)
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.ErrorIs(err, ErrInvalidRune)
		is.Nil(v)
	}

	//
	// String
	//

	{
		f := fw.String(``)
		is.True(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(``, string(v))
	}

	{
		f := fw.String(`9999999999999999999999999999999999999999999999999`)
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`9999999999999999999999999999999999999999999999999`, string(v))
	}

	//
	// Int
	//

	// Int(0)
	{
		f := fw.Int(0)
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`0`, string(v))
	}

	// Int(-1)
	{
		f := fw.Int(-1)
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`-1`, string(v))
	}

	// Int(math.MinInt64)
	{
		f := fw.Int(math.MinInt64)
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(strconv.FormatInt(math.MinInt64, 10), string(v))
	}

	// Int(math.MaxInt64)
	{
		f := fw.Int(math.MaxInt64)
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(strconv.FormatInt(math.MaxInt64, 10), string(v))
	}

	// Int(math.MaxUint64)
	{
		x := uint64(math.MaxUint64)
		f := fw.Int(int(x))
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`-1`, string(v))
	}

	//
	// Int64
	//

	// Int64(0)
	{
		f := fw.Int64(0)
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`0`, string(v))
	}

	// Int64(-1)
	{
		f := fw.Int64(-1)
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`-1`, string(v))
	}

	// Int64(math.MinInt64)
	{
		f := fw.Int64(math.MinInt64)
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(strconv.FormatInt(math.MinInt64, 10), string(v))
	}

	// Int64(math.MaxInt64)
	{
		f := fw.Int64(math.MaxInt64)
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(strconv.FormatInt(math.MaxInt64, 10), string(v))
	}

	// Int64(math.MaxUint64)
	{
		x := uint64(math.MaxUint64)
		f := fw.Int64(int64(x))
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`-1`, string(v))
	}

	//
	// Uint64
	//

	// Uint64(uint64(-1))
	{
		x := -1
		f := fw.Uint64(uint64(x))
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(strconv.FormatUint(math.MaxUint64, 10), string(v))
	}

	// Uint64(0)
	{
		f := fw.Uint64(0)
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`0`, string(v))
	}

	// Uint64(math.MaxUint64)
	{
		f := fw.Uint64(math.MaxUint64)
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(strconv.FormatUint(math.MaxUint64, 10), string(v))
	}

	//
	// Time
	//

	// Time(time.Time{})
	{
		f := fw.Time(time.Time{})
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`0001-01-01T00:00:00Z`, string(v))
	}

	// Time(time.Time{}.Add(1))
	{
		f := fw.Time(time.Time{}.Add(1))
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`0001-01-01T00:00:00.000000001Z`, string(v))
	}

	// Time(time.Time{}.Add(1).In(<+1 hour timezone>))
	{
		f := fw.Time(time.Time{}.Add(1).In(time.FixedZone("unused", int((time.Hour*1)/time.Second))))
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`0001-01-01T01:00:00.000000001+01:00`, string(v))
	}

	//
	// bool
	//

	// bool(true)
	{
		f := fw.Bool(true)
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`1`, string(v))
	}

	// bool(false)
	{
		f := fw.Bool(false)
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`0`, string(v))
	}

	//
	// Duration
	//

	// Duration(-1)
	{
		f := fw.Duration(-1)
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`-1`, string(v))
	}

	// Duration(1)
	{
		f := fw.Duration(1)
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`1`, string(v))
	}

	//
	// Float64
	//

	// Float64(0.01)
	{
		f := fw.Float64(0.01)
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`0.01`, string(v))
	}

	// Float64(-0.01)
	{
		f := fw.Float64(-0.01)
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`-0.01`, string(v))
	}

	// Float64(math.NaN())
	{
		f := fw.Float64(math.NaN())
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`NaN`, string(v))
	}

	// Float64(0.0)
	{
		f := fw.Float64(0.0)
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`0`, string(v))
	}

	// Float64(math.Inf(1))
	{
		f := fw.Float64(math.Inf(1))
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`+Inf`, string(v))
	}

	// Float64(math.Inf(0))
	{
		f := fw.Float64(math.Inf(0))
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`+Inf`, string(v))
	}

	// Float64(math.Inf(-1))
	{
		f := fw.Float64(math.Inf(-1))
		is.False(f.isZeroLen())
		v, err := f.MarshalText()
		is.Nil(err)
		is.Equal(`-Inf`, string(v))
	}
}

func TestFieldWriterAppendMinInt(t *testing.T) {
	t.Parallel()

	// case: min int will always serialize to 20 characters
	// and a buffer is never reallocated when it is at these bounds

	is := assert.New(t)
	fw := FieldWriters()
	buf := make([]byte, 0, 20)

	// valid bounds
	{
		f := fw.Int(math.MinInt)
		v, err := f.AppendText(buf)
		is.Nil(err)
		is.Equal(strconv.Itoa(math.MinInt), string(v))
		is.True(&(buf[:1][0]) == &v[0])
	}

	// bounds too short
	{
		f := fw.Int(math.MinInt)
		v, err := f.AppendText(buf[: 0 : cap(buf)-1])
		is.Nil(err)
		is.Equal(strconv.Itoa(math.MinInt), string(v))
		is.False(&(buf[:1][0]) == &v[0])
	}
}

func TestFieldWriterAppendMaxInt(t *testing.T) {
	t.Parallel()

	// case: max int will always serialize to 19 characters
	// and a buffer is never reallocated when it is at these bounds

	is := assert.New(t)
	fw := FieldWriters()
	buf := make([]byte, 0, 19)

	// valid bounds
	{
		f := fw.Int(math.MaxInt)
		v, err := f.AppendText(buf)
		is.Nil(err)
		is.Equal(strconv.Itoa(math.MaxInt), string(v))
		is.True(&(buf[:1][0]) == &v[0])
	}

	// bounds too short
	{
		f := fw.Int(math.MaxInt)
		v, err := f.AppendText(buf[: 0 : cap(buf)-1])
		is.Nil(err)
		is.Equal(strconv.Itoa(math.MaxInt), string(v))
		is.False(&(buf[:1][0]) == &v[0])
	}
}

func TestFieldWriterAppendMinInt64(t *testing.T) {
	t.Parallel()

	// case: min int will always serialize to 20 characters
	// and a buffer is never reallocated when it is at these bounds

	is := assert.New(t)
	fw := FieldWriters()
	buf := make([]byte, 0, 20)

	// valid bounds
	{
		f := fw.Int64(math.MinInt64)
		v, err := f.AppendText(buf)
		is.Nil(err)
		is.Equal(strconv.Itoa(math.MinInt64), string(v))
		is.True(&(buf[:1][0]) == &v[0])
	}

	// bounds too short
	{
		f := fw.Int64(math.MinInt64)
		v, err := f.AppendText(buf[: 0 : cap(buf)-1])
		is.Nil(err)
		is.Equal(strconv.Itoa(math.MinInt64), string(v))
		is.False(&(buf[:1][0]) == &v[0])
	}
}

func TestFieldWriterAppendMaxInt64(t *testing.T) {
	t.Parallel()

	// case: max int will always serialize to 19 characters
	// and a buffer is never reallocated when it is at these bounds

	is := assert.New(t)
	fw := FieldWriters()
	buf := make([]byte, 0, 19)

	// valid bounds
	{
		f := fw.Int64(math.MaxInt64)
		v, err := f.AppendText(buf)
		is.Nil(err)
		is.Equal(strconv.Itoa(math.MaxInt64), string(v))
		is.True(&(buf[:1][0]) == &v[0])
	}

	// bounds too short
	{
		f := fw.Int64(math.MaxInt64)
		v, err := f.AppendText(buf[: 0 : cap(buf)-1])
		is.Nil(err)
		is.Equal(strconv.Itoa(math.MaxInt64), string(v))
		is.False(&(buf[:1][0]) == &v[0])
	}
}

func TestFieldWriterAppendMinDuration(t *testing.T) {
	t.Parallel()

	// case: min int will always serialize to 20 characters
	// and a buffer is never reallocated when it is at these bounds

	is := assert.New(t)
	fw := FieldWriters()
	buf := make([]byte, 0, 20)

	// valid bounds
	{
		f := fw.Duration(math.MinInt)
		v, err := f.AppendText(buf)
		is.Nil(err)
		is.Equal(strconv.Itoa(math.MinInt), string(v))
		is.True(&(buf[:1][0]) == &v[0])
	}

	// bounds too short
	{
		f := fw.Duration(math.MinInt)
		v, err := f.AppendText(buf[: 0 : cap(buf)-1])
		is.Nil(err)
		is.Equal(strconv.Itoa(math.MinInt), string(v))
		is.False(&(buf[:1][0]) == &v[0])
	}
}

func TestFieldWriterAppendMaxDuration(t *testing.T) {
	t.Parallel()

	// case: max int will always serialize to 19 characters
	// and a buffer is never reallocated when it is at these bounds

	is := assert.New(t)
	fw := FieldWriters()
	buf := make([]byte, 0, 19)

	// valid bounds
	{
		f := fw.Duration(math.MaxInt)
		v, err := f.AppendText(buf)
		is.Nil(err)
		is.Equal(strconv.Itoa(math.MaxInt), string(v))
		is.True(&(buf[:1][0]) == &v[0])
	}

	// bounds too short
	{
		f := fw.Duration(math.MaxInt)
		v, err := f.AppendText(buf[: 0 : cap(buf)-1])
		is.Nil(err)
		is.Equal(strconv.Itoa(math.MaxInt), string(v))
		is.False(&(buf[:1][0]) == &v[0])
	}
}

func TestFieldWriterAppendMinUInt64(t *testing.T) {
	t.Parallel()

	// case: min uint will always serialize to 1 character
	// and a buffer is never reallocated when it is at these bounds

	is := assert.New(t)
	fw := FieldWriters()
	buf := make([]byte, 0, 1)

	// valid bounds
	{
		f := fw.Uint64(0)
		v, err := f.AppendText(buf)
		is.Nil(err)
		is.Equal(strconv.Itoa(0), string(v))
		is.True(&(buf[:1][0]) == &v[0])
	}

	// bounds too short
	{
		f := fw.Uint64(0)
		v, err := f.AppendText(buf[: 0 : cap(buf)-1])
		is.Nil(err)
		is.Equal(strconv.Itoa(0), string(v))
		is.False(&(buf[:1][0]) == &v[0])
	}
}

func TestFieldWriterAppendMaxUint64(t *testing.T) {
	t.Parallel()

	// case: max uint64 will always serialize to 20 characters
	// and a buffer is never reallocated when it is at these bounds

	is := assert.New(t)
	fw := FieldWriters()
	buf := make([]byte, 0, 20)

	// valid bounds
	{
		f := fw.Uint64(math.MaxUint64)
		v, err := f.AppendText(buf)
		is.Nil(err)
		is.Equal(strconv.FormatUint(math.MaxUint64, 10), string(v))
		is.True(&(buf[:1][0]) == &v[0])
	}

	// bounds too short
	{
		f := fw.Uint64(math.MaxUint64)
		v, err := f.AppendText(buf[: 0 : cap(buf)-1])
		is.Nil(err)
		is.Equal(strconv.FormatUint(math.MaxUint64, 10), string(v))
		is.False(&(buf[:1][0]) == &v[0])
	}
}

func TestFieldWriterAppendMaxLenTime(t *testing.T) {
	t.Parallel()

	// case: max serialized length time will always serialize to 35 characters
	// and a buffer is never reallocated when it is at these bounds

	is := assert.New(t)
	fw := FieldWriters()
	buf := make([]byte, 0, 35)

	maxLenSerializedTime := time.Time{}.Add(1).In(time.FixedZone("unused", int(time.Hour/time.Second)))

	// valid bounds
	{
		f := fw.Time(maxLenSerializedTime)
		v, err := f.AppendText(buf)
		is.Nil(err)
		is.Equal(maxLenSerializedTime.Format(time.RFC3339Nano), string(v))
		is.True(&(buf[:1][0]) == &v[0])
	}

	// bounds too short
	{
		f := fw.Time(maxLenSerializedTime)
		v, err := f.AppendText(buf[: 0 : cap(buf)-1])
		is.Nil(err)
		is.Equal(maxLenSerializedTime.Format(time.RFC3339Nano), string(v))
		is.False(&(buf[:1][0]) == &v[0])
	}
}

func TestFieldWriterAppendMaxFloat64(t *testing.T) {
	t.Parallel()

	// case: max serialized length of a positive float will not exceed 23 characters
	// and a buffer is never reallocated when it is at these bounds

	is := assert.New(t)
	fw := FieldWriters()
	buf := make([]byte, 0, 23)

	// valid bounds
	{
		f := fw.Float64(math.MaxFloat64)
		v, err := f.AppendText(buf)
		is.Nil(err)
		is.Equal(strconv.FormatFloat(math.MaxFloat64, 'g', -1, 64), string(v))
		is.True(&(buf[:1][0]) == &v[0])
	}

	// bounds too short
	{
		f := fw.Float64(math.MaxFloat64)
		v, err := f.AppendText(buf[: 0 : cap(buf)-1])
		is.Nil(err)
		is.Equal(strconv.FormatFloat(math.MaxFloat64, 'g', -1, 64), string(v))
		is.False(&(buf[:1][0]) == &v[0])
	}
}

func TestFieldWriterAppendMinFloat64(t *testing.T) {
	t.Parallel()

	// case: max serialized length of a negative float will not exceed 24 characters
	// and a buffer is never reallocated when it is at these bounds

	is := assert.New(t)
	fw := FieldWriters()
	buf := make([]byte, 0, 24)

	// valid bounds
	{
		f := fw.Float64(-math.MaxFloat64)
		v, err := f.AppendText(buf)
		is.Nil(err)
		is.Equal(strconv.FormatFloat(-math.MaxFloat64, 'g', -1, 64), string(v))
		is.True(&(buf[:1][0]) == &v[0])
	}

	// bounds too short
	{
		f := fw.Float64(-math.MaxFloat64)
		v, err := f.AppendText(buf[: 0 : cap(buf)-1])
		is.Nil(err)
		is.Equal(strconv.FormatFloat(-math.MaxFloat64, 'g', -1, 64), string(v))
		is.False(&(buf[:1][0]) == &v[0])
	}
}

func TestFieldWriterAppendBytes(t *testing.T) {
	t.Parallel()

	is := assert.New(t)
	fw := FieldWriters()

	const testStr = `9999999999999999999999999999999999999999999999999`
	buf := make([]byte, 0, len(testStr))

	// valid bounds
	{
		f := fw.Bytes([]byte(testStr))
		v, err := f.AppendText(buf)
		is.Nil(err)
		is.Equal(testStr, string(v))
		is.True(&(buf[:1][0]) == &v[0])
	}

	// bounds too short
	{
		f := fw.Bytes([]byte(testStr))
		v, err := f.AppendText(buf[: 0 : cap(buf)-1])
		is.Nil(err)
		is.Equal(testStr, string(v))
		is.False(&(buf[:1][0]) == &v[0])
	}
}

func TestFieldWriterAppendRune(t *testing.T) {
	t.Parallel()

	is := assert.New(t)
	fw := FieldWriters()

	// byte length of rune is known to be 1
	{
		buf := make([]byte, 0, 1)

		// valid bounds
		{
			f := fw.Rune('"')
			v, err := f.AppendText(buf)
			is.Nil(err)
			is.Equal(`"`, string(v))
			is.True(&(buf[:1][0]) == &v[0])
		}

		// bounds too short
		{
			f := fw.Rune('"')
			v, err := f.AppendText(buf[: 0 : cap(buf)-1])
			is.Nil(err)
			is.Equal(`"`, string(v))
			is.False(&(buf[:1][0]) == &v[0])
		}
	}

	// byte length of run is unknown
	{
		buf := make([]byte, 0, utf8.UTFMax)

		f := fw.Rune('"')
		v, err := f.AppendText(buf)
		is.Nil(err)
		is.Equal(`"`, string(v))
		is.True(&(buf[:1][0]) == &v[0])
	}
}

func TestFieldWriterAppendString(t *testing.T) {
	t.Parallel()

	is := assert.New(t)
	fw := FieldWriters()

	const testStr = `9999999999999999999999999999999999999999999999999`
	buf := make([]byte, 0, len(testStr))

	// valid bounds
	{
		f := fw.String(testStr)
		v, err := f.AppendText(buf)
		is.Nil(err)
		is.Equal(testStr, string(v))
		is.True(&(buf[:1][0]) == &v[0])
	}

	// bounds too short
	{
		f := fw.String(testStr)
		v, err := f.AppendText(buf[: 0 : cap(buf)-1])
		is.Nil(err)
		is.Equal(testStr, string(v))
		is.False(&(buf[:1][0]) == &v[0])
	}
}

func TestFieldWriterAppendInvalid(t *testing.T) {
	t.Parallel()

	is := assert.New(t)

	buf := make([]byte, 0, 1)

	f := FieldWriter{}
	v, err := f.AppendText(buf)
	is.Equal(ErrInvalidFieldWriter, err)
	is.Nil(v)
}
