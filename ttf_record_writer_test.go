package csv_test

import (
	"bytes"
	"math"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/josephcopenhaver/csv-go/v3"
	"github.com/stretchr/testify/assert"
)

type zeroType uint8

const (
	zeroBytesKey zeroType = iota + 1
	zeroUncheckedUTF8BytesKey
	zeroSBRune
	zeroMBRune
)

var (
	resOnlyCheckNoErr = wrRes{n: -1}
)

var fieldValuesByType = func() map[any][]any {
	m := map[any][]any{}

	m[int(0)] = []any{
		int(-1),
		int(0),
		int(1),
	}

	m[int64(0)] = []any{
		int64(-1),
		int64(0),
		int64(1),
	}

	m[uint64(0)] = []any{
		uint64(0),
		uint64(1),
		uint64(math.MaxUint64),
	}

	m[time.Duration(0)] = []any{
		time.Duration(-1),
		time.Duration(0),
		time.Duration(1),
	}

	m[""] = []any{
		"hello",
		"x",
		"goodbye",
	}

	m[uncheckedUTF8String("")] = []any{
		uncheckedUTF8String("hello"),
		uncheckedUTF8String("x"),
		uncheckedUTF8String("goodbye"),
	}

	m[zeroBytesKey] = []any{
		[]byte("hello"),
		[]byte("x"),
		[]byte("goodbye"),
	}

	m[zeroUncheckedUTF8BytesKey] = []any{
		uncheckedUTF8Bytes("hello"),
		uncheckedUTF8Bytes("x"),
		uncheckedUTF8Bytes("goodbye"),
	}

	m[false] = []any{false, true}

	m[float64(0)] = []any{
		float64(math.Inf(-1)),
		float64(-1),
		float64(0),
		float64(1),
		float64(math.Inf(0)),
	}

	sbRunes := []any{
		rune('A'),
		rune('Z'),
		rune('a'),
		rune('z'),
	}
	m[zeroSBRune] = sbRunes

	mbRunes := []any{
		rune('\U0001F600'),
		rune('\U0001F600' + 1),
		rune('\U0001F600' + 2),
		rune('\U0001F600' + 3),
	}
	m[zeroMBRune] = mbRunes

	m[rune(0)] = append(append([]any(nil), sbRunes...), mbRunes...)

	m[time.Time{}] = []any{
		time.Unix(0, 0).UTC(),
		time.Unix(1763670021, 0).UTC(),
		time.Unix(1763670021, 999999999).UTC(),
	}

	return m
}()

var fieldTypes = func() []any {
	v := make([]any, 0, len(fieldValuesByType))
	for k := range fieldValuesByType {
		v = append(v, k)
	}

	return v
}()

type uncheckedUTF8Bytes []byte
type uncheckedUTF8String string

type wrErrData struct {
	errAs            []any
	errAsNot         []any
	errIs            []error
	errIsNot         []error
	errStr           string
	errStrMsgAndArgs []any
	hasErr           bool
}

type wrRes struct {
	s      string
	fields []string
	n      int
	wrErrData
}

type rwf struct {
	f any
	wrErrData
}

type functionalRecordWriterTestCase struct {
	when, then      string
	selfInit        func(*functionalRecordWriterTestCase)
	newOpts         []csv.WriterOption
	newOptsF        func() []csv.WriterOption
	numRandomFields int
	fields          []rwf
	res             wrRes
	afterTest       func(*testing.T)
	sep             string
}

func (tc *functionalRecordWriterTestCase) clone() *functionalRecordWriterTestCase {
	ctc := *tc
	ctc.newOpts = slices.Clone(ctc.newOpts)
	ctc.fields = slices.Clone(ctc.fields)

	return &ctc
}

func (tc *functionalRecordWriterTestCase) Run(t *testing.T) {
	t.Helper()

	f := func(options ...func(*functionalRecordWriterTestCase)) func(*testing.T, any) {
		t.Helper()
		return func(t *testing.T, fieldType any) {
			t.Helper()

			tc := tc.clone()

			for _, f := range options {
				f(tc)
			}

			if f := tc.selfInit; f != nil {
				f(tc)
			}

			is := assert.New(t)

			expStr := tc.res.s
			if fields := tc.res.fields; fields != nil {
				if expStr != "" {
					panic("specifying fields and an expected string is invalid")
				}

				sep := tc.sep
				if sep == "" {
					sep = ","
				}

				var buf []byte
				for i, v := range fields {
					if i != 0 {
						buf = append(buf, sep...)
					}

					if strings.Contains(v, sep) {
						buf = append(buf, '"')
						buf = append(buf, v...)
						buf = append(buf, '"')

						continue
					}

					buf = append(buf, v...)
				}
				buf = append(buf, "\n"...)

				expStr = string(buf)
			}

			// var errWriterInIOErrState error
			var buf bytes.Buffer

			var cw *csv.Writer
			{
				writerOpt := csv.WriterOpts().Writer(&buf)

				opts := tc.newOpts
				if f := tc.newOptsF; f != nil {
					opts = append(append(slices.Clone(f()), writerOpt), opts...)
				} else {
					opts = append([]csv.WriterOption{writerOpt}, opts...)
				}

				if sep := tc.sep; sep != "" {
					r, n := utf8.DecodeRuneInString(sep)
					if (n == 1 && r == utf8.RuneError) || n != len(sep) {
						panic("field separator must be one valid utf8 rune")
					}

					opts = append(opts, csv.WriterOpts().FieldSeparator(r))
				}

				v, err := csv.NewWriter(opts...)
				is.Nil(err)
				is.NotNil(v)

				if v == nil {
					return
				}

				cw = v
			}

			var writeSuccess bool

			writeField := func(rw *csv.RecordWriter, v any) {
				switch v := v.(type) {
				case []byte:
					// if the string is empty, then randomly use either nil or empty slice
					// we should work with both in any case
					if len(v) == 0 && rand.IntN(2) == 0 {
						v = nil
					}

					rw.Bytes(v)
				case string:
					rw.String(v)
				case bool:
					rw.Bool(v)
				case time.Duration:
					rw.Duration(v)
				case float64:
					rw.Float64(v)
				case int:
					rw.Int(v)
				case int64:
					rw.Int64(v)
				case rune:
					rw.Rune(v)
				case time.Time:
					rw.Time(v)
				case uint64:
					rw.Uint64(v)
				case uncheckedUTF8Bytes:
					rw.UncheckedUTF8Bytes(v)
				case uncheckedUTF8String:
					rw.UncheckedUTF8String(string(v))
				default:
					panic("not a valid field type")
				}
			}

			rw := cw.NewRecord()

			for range tc.numRandomFields {
				candidates := fieldValuesByType[fieldType]
				writeField(rw, candidates[rand.IntN(len(candidates))])
				is.Nil(rw.Err())
			}

			for _, v := range tc.fields {
				writeField(rw, v.f)
				err := rw.Err()

				if !(v.hasErr || len(v.errAs) > 0 || len(v.errAsNot) > 0 || len(v.errIs) > 0 || len(v.errIsNot) > 0 || v.errStr != "") {
					is.Nil(err, v.errStrMsgAndArgs...)
					continue
				}

				is.NotNil(err, v.errStrMsgAndArgs...)

				for _, target := range v.errAs {
					is.ErrorAs(err, target, v.errStrMsgAndArgs...)
				}

				for _, target := range v.errAsNot {
					is.NotErrorAs(err, target, v.errStrMsgAndArgs...)
				}

				for _, target := range v.errIs {
					is.ErrorIs(err, target, v.errStrMsgAndArgs...)
				}

				for _, target := range v.errIsNot {
					is.NotErrorIs(err, target, v.errStrMsgAndArgs...)
				}

				if v.errStr != "" {
					is.Equal(v.errStr, err.Error(), v.errStrMsgAndArgs...)
				}
			}

			var expectWriteSuccess bool

			n, err := rw.Write()
			if err == nil {
				writeSuccess = true
			}
			if !(tc.res.hasErr || len(tc.res.errAs) > 0 || len(tc.res.errAsNot) > 0 || len(tc.res.errIs) > 0 || len(tc.res.errIsNot) > 0 || tc.res.errStr != "") {
				expectWriteSuccess = true
				is.Nil(err, tc.res.errStrMsgAndArgs...)
				if expStr == "" {
					if tc.res.n != -1 {
						is.Equal(tc.res.n, n, tc.res.errStrMsgAndArgs...)
					} else {
						r, _ := utf8.DecodeLastRuneInString(buf.String())
						// same as newlineRunesForWrite without \r (a.k.a. \x0D)
						is.Contains([]rune("\x0A\x0B\x0C\u0085\u2028"), r)
					}
				} else {
					is.Equal(expStr, buf.String())
				}
			} else {
				is.NotNil(err, tc.res.errStrMsgAndArgs...)
				is.Equal(tc.res.n, n, tc.res.errStrMsgAndArgs...)

				for _, target := range tc.res.errAs {
					is.ErrorAs(err, target, tc.res.errStrMsgAndArgs...)
				}

				for _, target := range tc.res.errAsNot {
					is.NotErrorAs(err, target, tc.res.errStrMsgAndArgs...)
				}

				for _, target := range tc.res.errIs {
					is.ErrorIs(err, target, tc.res.errStrMsgAndArgs...)
				}

				for _, target := range tc.res.errIsNot {
					is.NotErrorIs(err, target, tc.res.errStrMsgAndArgs...)
				}

				if tc.res.errStr != "" {
					is.Equal(tc.res.errStr, err.Error(), tc.res.errStrMsgAndArgs...)
				}

				if expStr == "" {
					is.Equal(tc.res.n, n, tc.res.errStrMsgAndArgs...)
				} else {
					is.Equal(expStr, buf.String())
				}
			}

			// once written or aborted, a record writer cannot be reused
			{
				err := rw.Err()
				if expectWriteSuccess {
					is.Equal(csv.ErrRecordWriterClosed, err)
				} else {
					// internal err state could be ErrRecordWriterClosed
					// or any other non-nil error state
					is.NotNil(err)
				}

				// attempting to write a field to a closed writer has no effect
				// and the Err() return value remains unchanged.
				rw.Bool(false)
				is.Equal(err, rw.Err())
			}

			// once closed, Writes should always return an error
			{
				{
					err := cw.Close()
					is.Nil(err)
				}

				// shuffling the order as it does not matter
				// any error here should be investigated very closely
				operations := [2]int{0, 1}
				rand.Shuffle(len(operations), func(x, y int) {
					operations[x], operations[y] = operations[y], operations[x]
				})

				for _, i := range operations {
					switch i {
					case 0:
						v, err := cw.WriteRow()
						is.Equal(v, 0)
						is.ErrorIs(err, csv.ErrWriterClosed)

					case 1:
						v, err := cw.WriteHeader()
						is.Equal(v, 0)
						is.ErrorIs(err, csv.ErrWriterClosed)
					}
				}

				// a second close should always return nil as well (be a nop)
				{
					err := cw.Close()
					is.Nil(err)
				}
			}

			// a checkpoint to ensure that if a write has taken place
			// with a custom writer that the afterTest callback
			// method is utilized
			//
			// if a write has happened, then there should never be zero
			// bytes written to the document, even if there is only one
			// column
			if writeSuccess && expStr == "" && buf.Len() == 0 {
				is.NotNil(tc.afterTest)
			}

			if tc.afterTest != nil {
				tc.afterTest(t)
			}
		}
	}

	if tc.numRandomFields > 0 {
		for _, v := range fieldTypes {
			f2 := f()
			f2(t, v)

			f2 = f(func(tc *functionalRecordWriterTestCase) {
				tc.newOpts = append(tc.newOpts, csv.WriterOpts().ClearFreedDataMemory(true))
			})
			f2(t, v)
		}
	} else {
		f2 := f()
		f2(t, nil)

		f2 = f(func(tc *functionalRecordWriterTestCase) {
			tc.newOpts = append(tc.newOpts, csv.WriterOpts().ClearFreedDataMemory(true))
		})
		f2(t, nil)
	}
}

func TestRecordWriterOKPaths(t *testing.T) {
	t.Parallel()

	tcs := []functionalRecordWriterTestCase{
		{
			when:            "writing a random field value",
			numRandomFields: 1,
			res:             resOnlyCheckNoErr,
		},
		{
			when:            "writing two random field values",
			numRandomFields: 2,
			res:             resOnlyCheckNoErr,
		},
		{
			when:            "writing three random field values",
			numRandomFields: 3,
			res:             resOnlyCheckNoErr,
		},
		{
			when: "writing fields without a custom field separator",
			fields: []rwf{
				{f: int(1)},
				{f: int64(1)},
				{f: uint64(1)},
				{f: time.Duration(1)},
				{f: math.NaN()},
				{f: true},
				{f: "s1"},
				{f: uncheckedUTF8String("us2")},
				{f: []byte("b3")},
				{f: uncheckedUTF8Bytes("ub4")},
				{f: rune('r')},
				{f: time.Time{}},
			},
			res: wrRes{
				fields: strings.Split("1,1,1,1,NaN,1,s1,us2,b3,ub4,r,0001-01-01T00:00:00Z", ","),
			},
		},
		{
			when: "writing fields with fieldSep=-",
			sep:  "-",
			fields: []rwf{
				{f: int(1)},
				{f: int64(1)},
				{f: uint64(1)},
				{f: time.Duration(1)},
				{f: math.NaN()},
				{f: true},
				{f: "s1"},
				{f: uncheckedUTF8String("us2")},
				{f: []byte("b3")},
				{f: uncheckedUTF8Bytes("ub4")},
				{f: rune('r')},
				{f: time.Time{}},
			},
			res: wrRes{
				fields: strings.Split("1,1,1,1,NaN,1,s1,us2,b3,ub4,r,0001-01-01T00:00:00Z", ","),
			},
		},
		{
			when: "writing a record that starts with a rune field containing a comment",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().CommentRune('#'),
			},
			fields: []rwf{
				{f: rune('#')},
			},
			res: wrRes{
				fields: strings.Split("\"#\"", ","),
			},
		},
		{
			when: "writing a record that starts with a rune field containing a MB field separator",
			sep:  "\U0001F600",
			fields: []rwf{
				{f: rune('\U0001F600')},
			},
			res: wrRes{
				fields: strings.Split("\U0001F600", ","),
			},
		},
		{
			when: "writing a record that starts with a MB rune field and fieldSep=-",
			sep:  "-",
			fields: []rwf{
				{f: rune('\U0001F600')},
			},
			res: wrRes{
				fields: strings.Split("\U0001F600", ","),
			},
		},
	}

	for i := range tcs {
		tc := &tcs[i]
		tc.when = "when " + tc.when

		if tc.then == "" {
			tc.then = "then no error should occur"
		} else {
			tc.then = "then " + tc.then
		}

		tc.Run(t)
	}
}
