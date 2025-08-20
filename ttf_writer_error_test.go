package csv_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/josephcopenhaver/csv-go/v2"
	"github.com/stretchr/testify/assert"
)

type errWriter struct {
	writer    io.Writer
	numWrites int
	err       error
}

func (ew *errWriter) Write(p []byte) (int, error) {
	if ew.numWrites <= 0 {
		if ew.err != nil {
			return 0, ew.err
		}

		return ew.writer.Write(p)
	}
	ew.numWrites -= 1

	return ew.writer.Write(p)
}

func TestFunctionalWriterErrorPaths(t *testing.T) {
	t.Parallel()

	tcs := []functionalWriterTestCase{
		{
			when: "a row with no fields - nil",
			wrs: []wr{
				{r: nil, errIs: []error{csv.ErrRowNilOrEmpty}, errStr: csv.ErrRowNilOrEmpty.Error()},
			},
		},
		{
			when: "a row with no fields - empty",
			wrs: []wr{
				{r: []string{}, errIs: []error{csv.ErrRowNilOrEmpty}, errStr: csv.ErrRowNilOrEmpty.Error()},
			},
		},
		{
			when: "row length mismatch",
			wrs: []wr{
				{r: []string{"hello"}, n: 6},
				{r: strings.Split("how,are", ","), errIs: []error{csv.ErrInvalidFieldCountInRecord}, errStr: csv.ErrInvalidFieldCountInRecord.Error()},
			},
			res: "hello\n",
		},
		{
			when: "configured row length mismatch with data row",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().NumFields(2),
			},
			wrs: []wr{
				{r: []string{"hello"}, errIs: []error{csv.ErrInvalidFieldCountInRecord}, errStr: csv.ErrInvalidFieldCountInRecord.Error()},
			},
		},
		{
			when: "configured row length mismatch with header row",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().NumFields(2),
			},
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().Headers("a", "b", "c"),
			},
			whErrIs:  []error{csv.ErrBadConfig},
			whErrStr: csv.ErrBadConfig.Error() + "\nheaders length does not match number of fields",
		},
		{
			when: "header row and data row length mismatch",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().Headers("a", "b", "c"),
			},
			whN: 6,
			wrs: []wr{
				{r: strings.Split("1,2", ","), errIs: []error{csv.ErrInvalidFieldCountInRecord}, errStr: csv.ErrInvalidFieldCountInRecord.Error()},
			},
			res: "a,b,c\n",
		},
		{
			when: "first field has an invalid utf8 sequence",
			wrs: []wr{
				{r: []string{string([]byte{0xC0}), "b"}, errIs: []error{csv.ErrNonUTF8InRecord}, errStr: csv.ErrNonUTF8InRecord.Error()},
			},
		},
		{
			when: "second field has an invalid utf8 sequence",
			wrs: []wr{
				{r: []string{"a", string([]byte{0xC0})}, errIs: []error{csv.ErrNonUTF8InRecord}, errStr: csv.ErrNonUTF8InRecord.Error()},
			},
		},
		{
			when: "io error encountered when writing a BOM",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().IncludeByteOrderMarker(true),
			},
			selfInit: func(tc *functionalWriterTestCase) {
				var buf bytes.Buffer
				w := &errWriter{writer: &buf, err: io.ErrClosedPipe}
				tc.newOpts = append(tc.newOpts, csv.WriterOpts().Writer(w))
				tc.afterTest = func(t *testing.T) {
					assert.Empty(t, buf.String())
				}
			},
			whErrStr: csv.ErrIO.Error() + ": " + io.ErrClosedPipe.Error(),
		},
		{
			when: "io error encountered when writing a comment",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines("hello"),
			},
			selfInit: func(tc *functionalWriterTestCase) {
				var buf bytes.Buffer
				w := &errWriter{writer: &buf, err: io.ErrClosedPipe}
				tc.newOpts = append(tc.newOpts, csv.WriterOpts().Writer(w))
				tc.afterTest = func(t *testing.T) {
					assert.Empty(t, buf.String())
				}
			},
			whErrStr: csv.ErrIO.Error() + ": " + io.ErrClosedPipe.Error(),
		},
		{
			when: "io error encountered when writing a comment after successful BOM",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().IncludeByteOrderMarker(true),
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines("hello"),
			},
			selfInit: func(tc *functionalWriterTestCase) {
				var buf bytes.Buffer
				w := &errWriter{writer: &buf, numWrites: 1, err: io.ErrClosedPipe}
				tc.newOpts = append(tc.newOpts, csv.WriterOpts().Writer(w))
				tc.afterTest = func(t *testing.T) {
					assert.Equal(t, string(bomBytes()), buf.String())
				}
			},
			whErrStr: csv.ErrIO.Error() + ": " + io.ErrClosedPipe.Error(),
		},
		{
			when: "io error encountered when writing a header row",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines("hello"),
				csv.WriteHeaderOpts().Headers(strings.Split("a,b,c", ",")...),
			},
			selfInit: func(tc *functionalWriterTestCase) {
				var buf bytes.Buffer
				w := &errWriter{writer: &buf, numWrites: 1, err: io.ErrClosedPipe}
				tc.newOpts = append(tc.newOpts, csv.WriterOpts().Writer(w))
				tc.afterTest = func(t *testing.T) {
					assert.Equal(t, "# hello\n", buf.String())
				}
			},
			whErrStr: csv.ErrIO.Error() + ": " + io.ErrClosedPipe.Error(),
		},
	}

	for _, tc := range tcs {
		if tc.then == "" {
			tc.then = "a coupled error should occur"
		}
		tc.Run(t)
	}
}
