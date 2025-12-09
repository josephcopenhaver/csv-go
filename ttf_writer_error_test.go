package csv_test

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/josephcopenhaver/csv-go/v3"
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
			whErrStr: csv.ErrWriteHeaderFailed.Error() + "\n" + csv.ErrIO.Error() + ": " + io.ErrClosedPipe.Error(),
		},
		{
			when: "io error encountered when writing a BOM and a comment",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().IncludeByteOrderMarker(true),
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines("hia"),
			},
			selfInit: func(tc *functionalWriterTestCase) {
				var buf bytes.Buffer
				w := &errWriter{writer: &buf, err: io.ErrClosedPipe}
				tc.newOpts = append(tc.newOpts, csv.WriterOpts().Writer(w))
				tc.afterTest = func(t *testing.T) {
					assert.Empty(t, buf.String())
				}
			},
			whErrStr: csv.ErrWriteHeaderFailed.Error() + "\n" + csv.ErrIO.Error() + ": " + io.ErrClosedPipe.Error(),
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
			whErrStr: csv.ErrWriteHeaderFailed.Error() + "\n" + csv.ErrIO.Error() + ": " + io.ErrClosedPipe.Error(),
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
			whErrStr: csv.ErrWriteHeaderFailed.Error() + "\n" + csv.ErrIO.Error() + ": " + io.ErrClosedPipe.Error(),
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
				w := &errWriter{writer: &buf, numWrites: 3, err: io.ErrClosedPipe}
				tc.newOpts = append(tc.newOpts, csv.WriterOpts().Writer(w))
				tc.afterTest = func(t *testing.T) {
					assert.Equal(t, "# hello\n", buf.String())
				}
			},
			whErrIs: []error{
				csv.ErrWriteHeaderFailed,
				csv.ErrIO,
				io.ErrClosedPipe,
			},
			whErrStr: csv.ErrWriteHeaderFailed.Error() + "\n" + csv.ErrIO.Error() + ": " + io.ErrClosedPipe.Error(),
		},
		{
			when: "an io error is encountered when writing a single column using an empty bytes writer",
			selfInit: func(tc *functionalWriterTestCase) {
				w := &errWriter{err: errors.New("test-error: 0199fa41b62a2d0d6c4df872eaa9c367")}
				tc.newOpts = append(tc.newOpts, csv.WriterOpts().Writer(w))
			},
			wrs: []wr{
				{fwr: []csv.FieldWriter{csv.FieldWriters().Bytes(nil)}, errIs: []error{csv.ErrIO}, errStr: csv.ErrIO.Error() + `: test-error: 0199fa41b62a2d0d6c4df872eaa9c367`},
			},
		},
		{
			when: "an io error is encountered when writing a single column using an empty string writer",
			selfInit: func(tc *functionalWriterTestCase) {
				w := &errWriter{err: errors.New("test-error: 0199fa565f45ca1407f90d3baa484ce7")}
				tc.newOpts = append(tc.newOpts, csv.WriterOpts().Writer(w))
			},
			wrs: []wr{
				{r: []string{""}, errIs: []error{csv.ErrIO}, errStr: csv.ErrIO.Error() + `: test-error: 0199fa565f45ca1407f90d3baa484ce7`},
			},
		},
		{
			when: "writing a single column using a rune writer and an invalid rune value",
			wrs: []wr{
				{fwr: []csv.FieldWriter{csv.FieldWriters().Rune(0x7FFFFFFF)}, errIs: []error{csv.ErrInvalidRune}, errStr: csv.ErrInvalidRune.Error()},
			},
		},
		{
			when: "writing a single column using an incorrectly initialized FieldWriter with field sep set to a colon rune overlapping with field writer chars",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().FieldSeparator(':'),
			},
			wrs: []wr{
				{fwr: []csv.FieldWriter{{}}, errIs: []error{csv.ErrInvalidFieldWriter}, errStr: csv.ErrInvalidFieldWriter.Error()},
			},
		},
		{
			when: "writing two columns using an incorrectly initialized FieldWriter for the second field",
			wrs: []wr{
				{fwr: []csv.FieldWriter{csv.FieldWriters().Bool(false), {}}, errIs: []error{csv.ErrInvalidFieldWriter}, errStr: csv.ErrInvalidFieldWriter.Error()},
			},
		},
		{
			when: "writing two columns using a RuneWriter with an invalid rune for the second field",
			wrs: []wr{
				{fwr: []csv.FieldWriter{csv.FieldWriters().Bool(false), csv.FieldWriters().Rune(0x7FFFFFFF)}, errIs: []error{csv.ErrInvalidRune}, errStr: csv.ErrInvalidRune.Error()},
			},
		},
		{
			when: "writing two columns using an incorrectly initialized FieldWriter for the second field with field sep set to a colon rune overlapping with field writer chars",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().FieldSeparator(':'),
			},
			wrs: []wr{
				{fwr: []csv.FieldWriter{csv.FieldWriters().Bool(false), {}}, errIs: []error{csv.ErrInvalidFieldWriter}, errStr: csv.ErrInvalidFieldWriter.Error()},
			},
		},
		{
			when: "writing two string columns where the second column as a comma followed by an incorrect uf8 byte sequence",
			wrs: []wr{
				{r: []string{"", ",\x7F\xFF\xFF\xFF"}, errIs: []error{csv.ErrNonUTF8InRecord}, errStr: csv.ErrNonUTF8InRecord.Error()},
			},
		},
		{
			when: "io error encountered when writing a terminating comment record separator",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines("hello"),
				csv.WriteHeaderOpts().Headers(strings.Split("a,b,c", ",")...),
			},
			selfInit: func(tc *functionalWriterTestCase) {
				var buf bytes.Buffer
				w := &errWriter{writer: &buf, numWrites: 2, err: io.ErrClosedPipe}
				tc.newOpts = append(tc.newOpts, csv.WriterOpts().Writer(w))
				tc.afterTest = func(t *testing.T) {
					assert.Equal(t, "# hello", buf.String())
				}
			},
			whErrIs: []error{
				csv.ErrWriteHeaderFailed,
				csv.ErrIO,
				io.ErrClosedPipe,
			},
			whErrStr: csv.ErrWriteHeaderFailed.Error() + "\n" + csv.ErrIO.Error() + ": " + io.ErrClosedPipe.Error(),
		},
		{
			when: "io error encountered when writing a comment recSepAndLinePrefix",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines("hell\no"),
				csv.WriteHeaderOpts().Headers(strings.Split("a,b,c", ",")...),
			},
			selfInit: func(tc *functionalWriterTestCase) {
				var buf bytes.Buffer
				w := &errWriter{writer: &buf, numWrites: 2, err: io.ErrClosedPipe}
				tc.newOpts = append(tc.newOpts, csv.WriterOpts().Writer(w))
				tc.afterTest = func(t *testing.T) {
					assert.Equal(t, "# hell", buf.String())
				}
			},
			whErrIs: []error{
				csv.ErrWriteHeaderFailed,
				csv.ErrIO,
				io.ErrClosedPipe,
			},
			whErrStr: csv.ErrWriteHeaderFailed.Error() + "\n" + csv.ErrIO.Error() + ": " + io.ErrClosedPipe.Error(),
		},
		{
			when: "io error encountered when writing a comment segment before recSepAndLinePrefix",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines("hell\no"),
				csv.WriteHeaderOpts().Headers(strings.Split("a,b,c", ",")...),
			},
			selfInit: func(tc *functionalWriterTestCase) {
				var buf bytes.Buffer
				w := &errWriter{writer: &buf, numWrites: 1, err: io.ErrClosedPipe}
				tc.newOpts = append(tc.newOpts, csv.WriterOpts().Writer(w))
				tc.afterTest = func(t *testing.T) {
					assert.Equal(t, "# ", buf.String())
				}
			},
			whErrIs: []error{
				csv.ErrWriteHeaderFailed,
				csv.ErrIO,
				io.ErrClosedPipe,
			},
			whErrStr: csv.ErrWriteHeaderFailed.Error() + "\n" + csv.ErrIO.Error() + ": " + io.ErrClosedPipe.Error(),
		},
		{
			when: "io error encountered when writing a comment line without writer newline runes",
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
					assert.Equal(t, "# ", buf.String())
				}
			},
			whErrIs: []error{
				csv.ErrWriteHeaderFailed,
				csv.ErrIO,
				io.ErrClosedPipe,
			},
			whErrStr: csv.ErrWriteHeaderFailed.Error() + "\n" + csv.ErrIO.Error() + ": " + io.ErrClosedPipe.Error(),
		},
		{
			when: "writer errors on all writes and attempting to write one record",
			selfInit: func(tc *functionalWriterTestCase) {
				w := &errWriter{writer: io.Discard, err: io.ErrClosedPipe}
				tc.newOpts = append(tc.newOpts, csv.WriterOpts().Writer(w))
			},
			wrs: []wr{
				{r: []string{"hello", "dave"}, errIs: []error{csv.ErrIO, io.ErrClosedPipe}, errStr: csv.ErrIO.Error() + ": " + io.ErrClosedPipe.Error()},
			},
		},
		{
			when: "writer is locked by a RecordWriter and attempting to use non-fluent Write",
			afterInitWriter: func(t *testing.T, w *csv.Writer) {
				t.Helper()

				// leaves the writer locked by the external RecordWriter
				_ = w.NewRecord()
			},
			wrs: []wr{
				{r: []string{"hello", "dave"}, errIs: []error{csv.ErrWriterNotReady}, errStr: csv.ErrWriterNotReady.Error()},
			},
		},
	}

	for _, tc := range tcs {
		if tc.then == "" {
			tc.then = "a coupled error should occur"
		}
		tc.Run(t)
	}
}
