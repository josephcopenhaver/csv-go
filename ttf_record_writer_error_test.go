package csv_test

import (
	"io"
	"testing"

	"github.com/josephcopenhaver/csv-go/v3"
	"github.com/josephcopenhaver/tbdd-go"
	"github.com/stretchr/testify/assert"
)

func TestRecordWriterErrPaths(t *testing.T) {
	type TC struct {
		w *csv.Writer
	}

	type R struct {
		errAfterRuneWrite, errAfterClosedFieldWrite, errWrite error
		n                                                     int
	}

	const given = "a normal csv writer"
	givenF := func(t *testing.T, tc *TC) {
		is := assert.New(t)

		opt := csv.WriterOpts()

		var err error
		tc.w, err = csv.NewWriter(
			opt.Writer(io.Discard),
		)
		is.Nil(err)
		is.NotNil(tc.w)
	}

	tcs := []tbdd.Lifecycle[TC, R]{
		tbdd.GWT(
			TC{}, given, givenF,
			// when
			"creating a new record writing an invalid utf8 rune",
			func(_ *testing.T, tc TC) R {
				rw := tc.w.NewRecord()
				errAfterRuneWrite := rw.Rune(rune(0x7FFFFFFF)).Err()
				n, err := rw.Write()

				return R{
					errAfterRuneWrite: errAfterRuneWrite,
					n:                 n,
					errWrite:          err,
				}
			},
			// then
			"should error",
			func(t *testing.T, tc TC, r R) {
				is := assert.New(t)

				is.NotNil(r.errAfterRuneWrite)
				is.ErrorIs(r.errAfterRuneWrite, csv.ErrInvalidRune)
				is.Equal(csv.ErrInvalidRune.Error(), r.errAfterRuneWrite.Error())
				is.NotNil(r.errWrite)
				is.ErrorIs(r.errWrite, csv.ErrInvalidRune)
				is.Equal(csv.ErrInvalidRune.Error(), r.errWrite.Error())
				is.Equal(int(0), r.n)
			},
		),
		tbdd.GWT(
			TC{}, given, givenF,
			// when
			"creating a record writer, writing a rune, closing the parent writer, and writing another string field",
			func(_ *testing.T, tc TC) R {
				rw := tc.w.NewRecord()
				errAfterRuneWrite := rw.Rune('R').Err()
				tc.w.Close()
				errAfterClosedFieldWrite := rw.String("").Err()
				n, err := rw.Write()

				return R{
					errAfterRuneWrite:        errAfterRuneWrite,
					errAfterClosedFieldWrite: errAfterClosedFieldWrite,
					n:                        n,
					errWrite:                 err,
				}
			},
			// then
			"should error",
			func(t *testing.T, tc TC, r R) {
				is := assert.New(t)

				is.Nil(r.errAfterRuneWrite)
				is.Nil(r.errAfterClosedFieldWrite)
				is.NotNil(r.errWrite)
				is.ErrorIs(r.errWrite, csv.ErrWriterClosed)
				is.Equal(csv.ErrWriterClosed.Error(), r.errWrite.Error())
				is.Equal(int(0), r.n)
			},
		),
		tbdd.GWT(
			TC{}, given, givenF,
			// when
			"creating a record writer, closing the parent writer, writing a rune",
			func(_ *testing.T, tc TC) R {
				rw := tc.w.NewRecord()
				tc.w.Close()
				errAfterRuneWrite := rw.Rune('R').Err()
				n, err := rw.Write()

				return R{
					errAfterRuneWrite: errAfterRuneWrite,
					n:                 n,
					errWrite:          err,
				}
			},
			// then
			"should error",
			func(t *testing.T, tc TC, r R) {
				is := assert.New(t)

				is.NotNil(r.errAfterRuneWrite)
				is.ErrorIs(r.errAfterRuneWrite, csv.ErrWriterClosed)
				is.Equal(csv.ErrWriterClosed.Error(), r.errAfterRuneWrite.Error())
				is.NotNil(r.errWrite)
				is.ErrorIs(r.errWrite, csv.ErrWriterClosed)
				is.Equal(csv.ErrWriterClosed.Error(), r.errWrite.Error())
				is.Equal(int(0), r.n)
			},
		),
	}

	for i, tc := range tcs {
		tc.RunI(t, i)

		{
			tc.Arrange = func(_ *testing.T, cfg tbdd.Arrange[TC, R]) (string, func(*testing.T)) {
				tc := cfg.TC

				return "a csv writer with clearmem+", func(t *testing.T) {
					is := assert.New(t)

					w, err := csv.NewWriter(
						csv.WriterOpts().Writer(io.Discard),
						csv.WriterOpts().ClearFreedDataMemory(true),
					)
					is.Nil(err)
					is.NotNil(w)

					tc.w = w
				}
			}

			tc.RunI(t, i)
		}
	}
}
