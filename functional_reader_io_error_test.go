package csv_test

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/josephcopenhaver/csv-go"
	"github.com/stretchr/testify/assert"
)

type bufferedReader interface {
	io.Reader
	ReadRune() (r rune, size int, err error)
	UnreadRune() error
	ReadByte() (byte, error)
}

type errReader struct {
	t        *testing.T
	reader   io.Reader
	numBytes int
	err      error
}

func (er *errReader) Read(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}

	if er.numBytes <= 0 {
		return 0, er.err
	}

	wantNum := er.numBytes
	if len(b) < er.numBytes {
		wantNum = len(b)
	}
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

type errBufferedReader struct {
	io.Reader
	numBytes        int
	readRune        func() (r rune, size int, err error)
	unreadRune      func() error
	readByte        func() (byte, error)
	hasReloadedByte bool
}

func (ebr *errBufferedReader) ReadRune() (r rune, size int, err error) {
	if f := ebr.readRune; f != nil {
		return f()
	}
	if v, ok := ebr.Reader.(interface{ ReadRune() (rune, int, error) }); ok {
		return v.ReadRune()
	}

	panic("underlying reader is not buffered")
}

func (ebr *errBufferedReader) UnreadRune() error {
	if f := ebr.unreadRune; f != nil {
		return f()
	}
	if v, ok := ebr.Reader.(interface{ UnreadRune() error }); ok {
		return v.UnreadRune()
	}

	panic("underlying reader is not buffered")
}

func (ebr *errBufferedReader) ReadByte() (byte, error) {
	if f := ebr.readByte; f != nil {
		return f()
	}
	if v, ok := ebr.Reader.(interface{ ReadByte() (byte, error) }); ok {
		return v.ReadByte()
	}

	panic("underlying reader is not buffered")
}

func newErrBufferedReader(er errReader) *errBufferedReader {
	ptrEr := &er
	br := bufio.NewReader(ptrEr)

	ebr := &errBufferedReader{
		Reader:   br,
		numBytes: ptrEr.numBytes,
	}
	if ptrEr.err != nil {
		ebr.readByte = func() (byte, error) {
			if ebr.numBytes <= 0 && !ebr.hasReloadedByte {
				return 0, ptrEr.err
			}
			ebr.hasReloadedByte = false
			ebr.numBytes -= 1
			b, err := br.ReadByte()
			if ebr.numBytes <= 0 {
				err = ptrEr.err
			}
			return b, err
		}
		ebr.readRune = func() (r rune, size int, err error) {
			if ebr.numBytes <= 0 && !ebr.hasReloadedByte {
				return 0, 0, ptrEr.err
			}
			ebr.hasReloadedByte = false
			ebr.numBytes -= 1
			r, s, err := br.ReadRune()
			if ebr.numBytes <= 0 {
				err = ptrEr.err
			}
			return r, s, err
		}
		ebr.unreadRune = func() error {
			if ebr.hasReloadedByte {
				panic("UnreadRune called twice in a row before a read operation")
			}
			err := br.UnreadRune()
			if err == nil {
				ebr.hasReloadedByte = true
				ebr.numBytes += 1
			}
			return err
		}
	}
	return ebr
}

var _ bufferedReader = &errBufferedReader{}

func TestFunctionalReaderIOErrorPaths(t *testing.T) {

	readErr := errors.New("some test read error")
	unreadErr := errors.New("some test unread error")

	tcs := []functionalReaderTestCase{
		{
			when: "reader raises an error",
			then: "error should be raised to the .Err method unless EOF",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(&errReader{
					err: readErr,
				}),
			},
			iterErrIs:  []error{csv.ErrIO, readErr},
			iterErrStr: csv.ErrIO.Error() + " at byte 0, record 0, field 0: " + readErr.Error(),
		},
		{
			when: "reader has record sep of CRLF and a non EOF error is thrown before LF",
			then: "error should be raised to the .Err method",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(&errReader{
					t:        t,
					reader:   strings.NewReader("a,b\r"),
					numBytes: 4,
					err:      readErr,
				}),
				csv.ReaderOpts().RecordSeparator("\r\n"),
			},
			iterErrIs:  []error{csv.ErrIO, readErr},
			iterErrStr: csv.ErrIO.Error() + " at byte 4, record 1, field 2: " + readErr.Error(),
		},
		{
			when: "reader has record sep of CRLF and a non EOF error is thrown as LF is read",
			then: "now rows should be returned and error should be raised to the .Err method",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(newErrBufferedReader(errReader{
					t:        t,
					reader:   strings.NewReader("a,b\r\n"),
					numBytes: 5,
					err:      readErr,
				})),
				csv.ReaderOpts().RecordSeparator("\r\n"),
			},
			iterErrIs: []error{csv.ErrIO, csv.ErrBadReadRuneImpl, readErr},
			// 4 is correct here because it was the last successfully read byte before the bad implementation detail
			// was uncovered
			iterErrStr: csv.ErrIO.Error() + " at byte 4, record 1, field 2: " + errors.Join(csv.ErrBadReadRuneImpl, readErr).Error(),
		},
		{
			when: "reader has record sep of CRLF and an EOF error is thrown as LF is read",
			then: "now rows should be returned and error should be raised to the .Err method",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(newErrBufferedReader(errReader{
					t:        t,
					reader:   strings.NewReader("a,b\r\n"),
					numBytes: 5,
					err:      io.EOF,
				})),
				csv.ReaderOpts().RecordSeparator("\r\n"),
			},
			iterErrIs: []error{csv.ErrIO, csv.ErrBadReadRuneImpl, io.EOF},
			// 4 is correct here because it was the last successfully read byte before the bad implementation detail
			// was uncovered
			iterErrStr: csv.ErrIO.Error() + " at byte 4, record 1, field 2: " + errors.Join(csv.ErrBadReadRuneImpl, io.EOF).Error(),
		},
		{
			when: "reader is discovering the record sep and first row ends in CR+short-multibyte and UnreadRune implementation is bad",
			then: "now rows should be returned and error should be raised to the .Err method",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(func() *errBufferedReader {
					ebr := newErrBufferedReader(errReader{
						t:        t,
						reader:   bytes.NewReader(append([]byte("a,b\r"), 0xC0)),
						numBytes: 5,
					})

					ebr.unreadRune = func() error { return unreadErr }

					return ebr
				}()),
				csv.ReaderOpts().DiscoverRecordSeparator(true),
			},
			iterErrIs: []error{csv.ErrIO, csv.ErrBadUnreadRuneImpl, unreadErr},
			// 4 is correct here because it was the last successfully read byte before the bad implementation detail
			// was uncovered
			iterErrStr: csv.ErrIO.Error() + " at byte 4, record 1, field 2: " + errors.Join(csv.ErrBadUnreadRuneImpl, unreadErr).Error(),
		},

		{
			when: "ReadRune returns an error and a byte on the first read",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(func() *errBufferedReader {
					return newErrBufferedReader(errReader{
						t:        t,
						reader:   strings.NewReader("1"),
						numBytes: 1,
						err:      io.EOF,
					})
				}()),
			},
			iterErrIs:  []error{csv.ErrIO, csv.ErrBadReadRuneImpl, io.EOF},
			iterErrStr: csv.ErrIO.Error() + " at byte 0, record 0, field 0: " + errors.Join(csv.ErrBadReadRuneImpl, io.EOF).Error(),
		},
		{
			when: "incomplete utf8 rune and UnreadRune returns an error",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(func() *errBufferedReader {
					ebr := newErrBufferedReader(errReader{
						t:        t,
						reader:   bytes.NewReader([]byte{0xC0}),
						numBytes: 2,
						err:      io.EOF,
					})

					ebr.unreadRune = func() error { return io.ErrClosedPipe }

					return ebr
				}()),
			},
			iterErrIs:  []error{csv.ErrIO, csv.ErrBadUnreadRuneImpl, io.ErrClosedPipe},
			iterErrStr: csv.ErrIO.Error() + " at byte 1, record 1, field 1: " + errors.Join(csv.ErrBadUnreadRuneImpl, io.ErrClosedPipe).Error(),
		},
		{
			when: "incomplete utf8 rune and ReadByte returns an error after UnreadRune",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(func() *errBufferedReader {
					ebr := newErrBufferedReader(errReader{
						t:        t,
						reader:   bytes.NewReader([]byte{0xC0}),
						numBytes: 2,
						err:      io.EOF,
					})

					ebr.readByte = func() (byte, error) { return 0, io.ErrClosedPipe }

					return ebr
				}()),
			},
			iterErrIs:  []error{csv.ErrIO, csv.ErrBadReadByteImpl, io.ErrClosedPipe},
			iterErrStr: csv.ErrIO.Error() + " at byte 1, record 1, field 1: " + errors.Join(csv.ErrBadReadByteImpl, io.ErrClosedPipe).Error(),
		},
		{
			when: "record sep is CRLF and record start has CR then reader errors",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(func() *errBufferedReader {
					return newErrBufferedReader(errReader{
						t:        t,
						reader:   bytes.NewReader([]byte{'\r'}),
						numBytes: 2,
						err:      io.ErrClosedPipe,
					})
				}()),
				csv.ReaderOpts().RecordSeparator("\r\n"),
			},
			iterErrIs:  []error{csv.ErrIO, io.ErrClosedPipe},
			iterErrStr: csv.ErrIO.Error() + " at byte 1, record 1, field 1: " + io.ErrClosedPipe.Error(),
		},
		{
			when: "record sep is CRLF, two quote chars, CR and ReadRune errors",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(func() *errBufferedReader {
					return newErrBufferedReader(errReader{
						t:        t,
						reader:   strings.NewReader("\"\"\r"),
						numBytes: 4,
						err:      io.ErrClosedPipe,
					})
				}()),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().RecordSeparator("\r\n"),
			},
			iterErrIs:  []error{csv.ErrIO, io.ErrClosedPipe},
			iterErrStr: csv.ErrIO.Error() + " at byte 3, record 1, field 1: " + io.ErrClosedPipe.Error(),
		},
		{
			when: "record sep is CRLF, reader errors after comma+CR",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(func() *errBufferedReader {
					return newErrBufferedReader(errReader{
						t:        t,
						reader:   strings.NewReader(",\r"),
						numBytes: 3,
						err:      io.ErrClosedPipe,
					})
				}()),
				csv.ReaderOpts().RecordSeparator("\r\n"),
			},
			iterErrIs:  []error{csv.ErrIO, io.ErrClosedPipe},
			iterErrStr: csv.ErrIO.Error() + " at byte 2, record 1, field 2: " + io.ErrClosedPipe.Error(),
		},
	}

	for _, tc := range tcs {
		tc.Run(t)
	}
}
