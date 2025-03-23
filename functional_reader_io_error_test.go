package csv_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/josephcopenhaver/csv-go"
	"github.com/stretchr/testify/assert"
)

type errReader struct {
	t        *testing.T
	reader   io.Reader
	numBytes int
	err      error
}

func (er *errReader) Read(b []byte) (int, error) {
	if len(b) == 0 {
		if er.numBytes > 0 {
			return 0, nil
		}
		return 0, er.err
	}

	if er.numBytes <= 0 {
		if er.err == nil {
			return 0, io.EOF
		}
		return 0, er.err
	}

	wantNum := min(len(b), er.numBytes)
	n, err := er.reader.Read(b[:wantNum])
	assert.LessOrEqual(er.t, n, wantNum)
	if err != nil {
		assert.ErrorIs(er.t, err, io.EOF)
	}

	er.numBytes -= n

	if er.numBytes <= 0 {
		if er.err == nil {
			return n, err
		}
		return n, er.err
	}

	return n, nil
}

func TestFunctionalReaderIOErrorPaths(t *testing.T) {

	readErr := errors.New("some test read error")

	tcs := []functionalReaderTestCase{
		{
			when: "reader raises an error",
			then: "error should be raised to the .Err method unless EOF",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(&errReader{
						err: readErr,
					}),
				}
			},
			iterErrIs:  []error{csv.ErrIO, readErr},
			iterErrStr: csv.ErrIO.Error() + " at byte 0, record 0, field 0: " + readErr.Error(),
		},
		{
			when: "reader has record sep of CRLF and a non EOF error is thrown before LF",
			then: "error should be raised to the .Err method",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(&errReader{
						t:        t,
						reader:   strings.NewReader("a,b\r"),
						numBytes: 4,
						err:      readErr,
					}),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r\n"),
			},
			iterErrIs:  []error{csv.ErrIO, readErr},
			iterErrStr: csv.ErrIO.Error() + " at byte 4, record 1, field 2: " + readErr.Error(),
		},
	}

	for _, tc := range tcs {
		tc.Run(t)
	}
}
