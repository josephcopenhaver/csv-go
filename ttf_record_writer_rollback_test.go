package csv_test

import (
	"io"
	"testing"

	"github.com/josephcopenhaver/csv-go/v3"
	"github.com/josephcopenhaver/tbdd-go"
	"github.com/stretchr/testify/assert"
)

func TestNewRecordWriterErrPaths(t *testing.T) {
	type TC struct {
		w *csv.Writer
	}

	type R struct {
		errBeforeRollback, errWClose, errAfterRollback error
	}

	type tci interface {
		RunI(*testing.T, int)
	}

	const given = "a writer with memclear+"
	givenF := func(t *testing.T, tc *TC) {
		is := assert.New(t)

		opt := csv.WriterOpts()

		var err error
		tc.w, err = csv.NewWriter(
			opt.Writer(io.Discard),
			opt.ClearFreedDataMemory(true),
		)
		is.Nil(err)
		is.NotNil(tc.w)
	}

	tcs := []tci{
		tbdd.GWT(
			TC{}, given, givenF,
			// when
			"creating a new record and rolling back",
			func(_ *testing.T, tc TC) R {
				rw := tc.w.MustNewRecord()
				err := rw.Err()
				rw.Rollback()

				return R{
					errBeforeRollback: err,
					errAfterRollback:  rw.Err(),
				}
			},
			// then
			"record writer should be closed after rollback",
			func(t *testing.T, tc TC, r R) {
				is := assert.New(t)

				is.Nil(r.errBeforeRollback)
				is.Nil(r.errWClose)
				is.NotNil(r.errAfterRollback)
				is.ErrorIs(r.errAfterRollback, csv.ErrRecordWriterClosed)
				is.Equal(csv.ErrRecordWriterClosed.Error(), r.errAfterRollback.Error())
			},
		),
		tbdd.GWT(
			TC{}, given, givenF,
			// when
			"creating a new record, closing the writer, and rolling back",
			func(_ *testing.T, tc TC) R {
				rw := tc.w.MustNewRecord()
				err := rw.Err()

				errWClose := tc.w.Close()

				rw.Rollback()

				return R{
					errBeforeRollback: err,
					errWClose:         errWClose,
					errAfterRollback:  rw.Err(),
				}
			},
			// then
			"record writer should be closed after rollback",
			func(t *testing.T, tc TC, r R) {
				is := assert.New(t)

				is.Nil(r.errBeforeRollback)
				is.Nil(r.errWClose)
				is.NotNil(r.errAfterRollback)
				is.ErrorIs(r.errAfterRollback, csv.ErrRecordWriterClosed)
				is.Equal(csv.ErrRecordWriterClosed.Error(), r.errAfterRollback.Error())
			},
		),
	}

	for i, tc := range tcs {
		tc.RunI(t, i)
	}
}
