package csv

import (
	"io"
	"testing"

	"github.com/josephcopenhaver/tbdd-go"
	"github.com/stretchr/testify/assert"
)

func TestNewRecordWriterErrPaths(t *testing.T) {
	type TC struct {
		w      *Writer
		expErr error
	}

	type R struct {
		err error
	}

	type tci interface {
		RunI(*testing.T, int)
	}

	tcs := []tci{
		tbdd.GWT(
			TC{},
			// given
			"a writer and a record is created but never rolled back or written",
			func(t *testing.T, tc *TC) {
				is := assert.New(t)

				opt := WriterOpts()

				var err error
				tc.w, err = NewWriter(
					opt.Writer(io.Discard),
				)
				is.Nil(err)
				is.NotNil(tc.w)

				_ = tc.w.NewRecord()
			},
			// when
			"creating a new record",
			func(t *testing.T, tc TC) R {
				is := assert.New(t)

				const exp = "invalid concurrent access detected on record creation"
				is.PanicsWithValue(exp, func() {
					_ = tc.w.NewRecord()
				})

				return R{}
			},
			// then
			"panics",
			func(*testing.T, TC, R) {},
		),
		tbdd.GWT(
			TC{
				expErr: ErrWriterClosed,
			},
			// given
			"a writer created and closed",
			func(t *testing.T, tc *TC) {
				is := assert.New(t)

				opt := WriterOpts()

				var err error
				tc.w, err = NewWriter(
					opt.Writer(io.Discard),
				)
				is.Nil(err)
				is.NotNil(tc.w)

				is.Nil(tc.w.Close())
			},
			// when
			"creating a new record",
			func(_ *testing.T, tc TC) R {
				rw := tc.w.NewRecord()

				return R{
					err: rw.Err(),
				}
			},
			// then
			"record writer should be in errored state ErrWriterClosed",
			func(t *testing.T, tc TC, r R) {
				is := assert.New(t)

				is.ErrorIs(r.err, tc.expErr)
				is.Equal(tc.expErr.Error(), r.err.Error())
			},
		),
		tbdd.GWT(
			TC{
				expErr: ErrWriterClosed,
			},
			// given
			"a writer created and closed but somehow has no error set",
			func(t *testing.T, tc *TC) {
				is := assert.New(t)

				opt := WriterOpts()

				var err error
				tc.w, err = NewWriter(
					opt.Writer(io.Discard),
				)
				is.Nil(err)
				is.NotNil(tc.w)

				err = tc.w.Close()
				is.Nil(err)

				// notice not a standard state:
				tc.w.err = nil
			},
			// when
			"creating a new record",
			func(_ *testing.T, tc TC) R {
				rw := tc.w.NewRecord()

				return R{
					err: rw.Err(),
				}
			},
			// then
			"record writer should be in errored state ErrWriterClosed",
			func(t *testing.T, tc TC, r R) {
				is := assert.New(t)
				is.ErrorIs(r.err, tc.expErr)
				is.Equal(tc.expErr.Error(), r.err.Error())
			},
		),
	}

	for i, v := range tcs {
		v.RunI(t, i)
	}
}
