package csv_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/josephcopenhaver/csv-go"
	"github.com/stretchr/testify/assert"
)

func TestFunctionalReaderEOFPaths(t *testing.T) {

	type TC struct {
		Name        string
		reader      func() (*csv.Reader, error)
		err         error
		validateErr func(*testing.T, error)
	}

	tcs := []TC{
		{"doc start, errOnNoByteOrderMarker=default", func() (*csv.Reader, error) {
			return csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
			)
		}, nil, nil},
		{"doc start, errOnNoByteOrderMarker=false", func() (*csv.Reader, error) {
			return csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().ErrorOnNoByteOrderMarker(false),
			)
		}, nil, nil},
		{"doc start, errOnNoByteOrderMarker=true", func() (*csv.Reader, error) {
			return csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().ErrorOnNoByteOrderMarker(true),
			)
		}, nil, func(t *testing.T, err error) {
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrNoByteOrderMarker)
			assert.ErrorIs(t, err, io.ErrUnexpectedEOF)
			assert.EqualError(t, err, "parsing error at byte 0, record 0, field 0: "+errors.Join(csv.ErrNoByteOrderMarker, io.ErrUnexpectedEOF).Error())
		}},
	}

	for _, tc := range tcs {
		t.Run(tc.Name, func(t *testing.T) {
			cr, err := tc.reader()
			assert.Nil(t, err)
			assert.NotNil(t, cr)
			defer func() {
				assert.Nil(t, cr.Close())
			}()

			for range cr.IntoIter() {
			}

			err = cr.Err()
			var checked bool

			if tc.err != nil {
				checked = true
				assert.ErrorIs(t, err, tc.err)
			}

			if f := tc.validateErr; f != nil {
				checked = true
				f(t, err)
			}

			if !checked {
				assert.Nil(t, err)
			}
		})
	}
}
