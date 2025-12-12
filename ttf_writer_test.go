package csv_test

import (
	"bytes"
	"errors"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/josephcopenhaver/csv-go/v3"
	"github.com/stretchr/testify/assert"
)

type wr struct {
	r                []string
	fwr              []csv.FieldWriter
	errAs            []any
	errAsNot         []any
	errIs            []error
	errIsNot         []error
	errStr           string
	errStrMsgAndArgs []any
	n                int
	hasErr           bool
}

type functionalWriterTestCase struct {
	when, then                string
	selfInit                  func(*functionalWriterTestCase)
	afterInitWriter           func(*testing.T, *csv.Writer)
	newOpts                   []csv.WriterOption
	newOptsF                  func() []csv.WriterOption
	whOpts                    []csv.WriteHeaderOption
	whOptsF                   func() []csv.WriteHeaderOption
	newWriterErrAs            []any
	newWriterErrAsNot         []any
	newWriterErrIs            []error
	newWriterErrIsNot         []error
	newWriterErrStr           string
	newWriterErrStrMsgAndArgs []any
	whErrAs                   []any
	whErrAsNot                []any
	whErrIs                   []error
	whErrIsNot                []error
	whErrStr                  string
	whErrStrMsgAndArgs        []any
	wrs                       []wr
	res                       string
	afterTest                 func(*testing.T)
	whN                       int
	wh                        bool
	hasNewWriterErr           bool
	hasWHeaderErr             bool
	useRecordWriter           bool
	useRecordWriterBytes      bool
	useRecordWriterEmpty      bool
}

func (tc *functionalWriterTestCase) clone() *functionalWriterTestCase {
	ctc := *tc

	ctc.newOpts = slices.Clone(tc.newOpts)
	ctc.whOpts = slices.Clone(tc.whOpts)
	ctc.newWriterErrAs = slices.Clone(tc.newWriterErrAs)
	ctc.newWriterErrAsNot = slices.Clone(tc.newWriterErrAsNot)
	ctc.newWriterErrIs = slices.Clone(tc.newWriterErrIs)
	ctc.newWriterErrIsNot = slices.Clone(tc.newWriterErrIsNot)
	ctc.newWriterErrStrMsgAndArgs = slices.Clone(tc.newWriterErrStrMsgAndArgs)
	ctc.whErrAs = slices.Clone(tc.whErrAs)
	ctc.whErrAsNot = slices.Clone(tc.whErrAsNot)
	ctc.whErrIs = slices.Clone(tc.whErrIs)
	ctc.whErrIsNot = slices.Clone(tc.whErrIsNot)
	ctc.whErrStrMsgAndArgs = slices.Clone(tc.whErrStrMsgAndArgs)
	ctc.wrs = slices.Clone(tc.wrs)

	return &ctc
}

func (tc *functionalWriterTestCase) Run(t *testing.T) {
	assert.NotEmpty(t, tc.when)
	t.Helper()

	f := func(options ...func(*functionalWriterTestCase)) func(t *testing.T) {
		tc := tc.clone()

		for _, f := range options {
			f(tc)
		}
		if f := tc.selfInit; f != nil {
			f(tc)
		}

		return func(t *testing.T) {
			t.Helper()

			is := assert.New(t)
			var errWriterInIOErrState error
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

				v, err := csv.NewWriter(opts...)
				if tc.hasNewWriterErr || len(tc.newWriterErrIs) > 0 || len(tc.newWriterErrAs) > 0 || len(tc.newWriterErrIsNot) > 0 || len(tc.newWriterErrAsNot) > 0 || tc.newWriterErrStr != "" {
					is.NotNil(err)
					is.Nil(v)

					for _, terr := range tc.newWriterErrIs {
						is.ErrorIs(err, terr)
					}

					for i := range tc.newWriterErrAs {
						v := tc.newWriterErrAs[i]
						is.True(errors.As(err, &v))
					}

					for _, target := range tc.newWriterErrIsNot {
						is.False(errors.Is(err, target))
					}

					for i := range tc.newWriterErrAsNot {
						v := tc.newWriterErrAsNot[i]
						is.False(errors.As(err, &v))
					}

					if tc.newWriterErrStr != "" && err != nil {
						is.Equal(tc.newWriterErrStr, err.Error(), tc.newWriterErrStrMsgAndArgs...)
					}

					if tc.afterTest != nil {
						tc.afterTest(t)
					}

					return
				}

				is.Nil(err)
				is.NotNil(v)

				if f := tc.afterInitWriter; f != nil {
					f(t, v)
				}

				if v == nil {
					return
				}

				cw = v
			}

			var writeSuccess bool

			if tc.wh || len(tc.whOpts) > 0 || tc.whOptsF != nil {
				opts := tc.whOpts
				if f := tc.whOptsF; f != nil {
					opts = append(slices.Clone(f()), opts...)
				}

				n, err := cw.WriteHeader(opts...)
				if err == nil && n > 0 {
					writeSuccess = true
				}
				if errors.Is(err, csv.ErrIO) {
					errWriterInIOErrState = err
					is.ErrorIs(err, csv.ErrWriteHeaderFailed)
				}
				if tc.hasWHeaderErr || len(tc.whErrIs) > 0 || len(tc.whErrAs) > 0 || len(tc.whErrIsNot) > 0 || len(tc.whErrAsNot) > 0 || tc.whErrStr != "" {
					is.NotNil(err)
					if len(tc.wrs) > 0 {
						is.Equal(tc.whN, n)
					} else {
						is.Equal(0, tc.whN)
					}

					for _, terr := range tc.whErrIs {
						is.ErrorIs(err, terr)
					}

					for i := range tc.whErrAs {
						v := tc.whErrAs[i]
						is.True(errors.As(err, &v))
					}

					for _, target := range tc.whErrIsNot {
						is.False(errors.Is(err, target))
					}

					for i := range tc.whErrAsNot {
						v := tc.whErrAsNot[i]
						is.False(errors.As(err, &v))
					}

					if tc.whErrStr != "" && err != nil {
						is.Equal(tc.whErrStr, err.Error(), tc.whErrStrMsgAndArgs...)
					}
				} else {
					is.Nil(err)
					if len(tc.wrs) > 0 {
						is.Equal(tc.whN, n)
					} else {
						// if this assert fails, you have a redundant assert m8
						//
						// remove the whN value and just use the res value since
						// there were no rows written
						//
						// that or your test case should have rows...
						//
						//
						// keep it simple, keep it clean
						is.Equal(0, tc.whN)
					}
				}

				// attempting to call write headers a second time should always
				// result in a header already written error
				//
				// unless there was an io level error previously or the writer was closed
				//
				// note that another category of errors exists: preflight validation
				// failed given call inputs and no state was altered but since this is a
				// second attempt call it cannot be returned
				{
					v, err := cw.WriteHeader()
					is.Zero(v)
					is.NotNil(err)
					if errWriterInIOErrState != nil {
						is.True(err == errWriterInIOErrState)
					} else if writeSuccess {
						is.ErrorIs(err, csv.ErrHeaderWritten)
					} else {
						// could have been already written OR the write failed such that ErrWriterNotReady was returned
						//
						// all other error types are io level errors.
						is.True(errors.Is(err, csv.ErrHeaderWritten) || errors.Is(err, csv.ErrWriterNotReady))
					}
				}
			}

			for _, v := range tc.wrs {
				var n int
				var err error
				if tc.useRecordWriter {
					if v.fwr != nil {
						t.Fatal("record writer requires record slices to write be specified using r => []string")
					}

					rw := cw.MustNewRecord()

					if rw.Err() == nil {
						if tc.useRecordWriterBytes {
							for _, v := range v.r {
								v := []byte(v)

								// if the string is empty, then randomly use either nil or empty slice
								// we should work with both in any case
								if len(v) == 0 && rand.IntN(2) == 0 {
									v = nil
								}

								rw.Bytes(v)
								if rw.Err() != nil {
									break
								}
							}
						} else {
							for _, v := range v.r {
								if tc.useRecordWriterEmpty && v == "" {
									rw.Empty()
								} else {
									rw.String(v)
								}
								if rw.Err() != nil {
									break
								}
							}
						}
					}

					n, err = rw.Write()
				} else if v.fwr == nil {
					n, err = cw.WriteRow(v.r...)
				} else {
					if v.r != nil {
						t.Fatal("record slices to write can only be specified using one of r => []string or fwr => []csv.FieldWriter")
					}

					n, err = cw.WriteFieldRow(v.fwr...)
				}

				if errors.Is(err, csv.ErrIO) {
					errWriterInIOErrState = err
				}
				if v.hasErr || len(v.errIs) > 0 || len(v.errAs) > 0 || len(v.errIsNot) > 0 || len(v.errAsNot) > 0 || v.errStr != "" {
					is.NotNil(err)
					is.Equal(v.n, n)

					for _, v := range v.errIs {
						is.ErrorIs(err, v)
					}

					for i := range v.errAs {
						v := v.errAs[i]
						is.True(errors.As(err, &v))
					}

					for _, target := range v.errIsNot {
						is.False(errors.Is(err, target))
					}

					for i := range v.errAsNot {
						v := v.errAsNot[i]
						is.False(errors.As(err, &v))
					}

					if v.errStr != "" && err != nil {
						is.Equal(v.errStr, err.Error(), v.errStrMsgAndArgs...)
					}

					continue
				}

				if errWriterInIOErrState != nil {
					is.NotNil(err)
					is.Equal(0, n)
					is.True(err == errWriterInIOErrState)

					continue
				}

				if err == nil {
					writeSuccess = true
				}

				is.Nil(err)
				is.Equal(v.n, n)
			}

			is.Equal(tc.res, buf.String())

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
			if writeSuccess && tc.res == "" && buf.Len() == 0 {
				is.NotNil(tc.afterTest)
			}

			if tc.afterTest != nil {
				tc.afterTest(t)
			}
		}
	}

	var name string
	if tc.then == "" {
		name = "then no error should occur"
	} else {
		name = "then " + tc.then
	}

	t.Run("when "+tc.when, func(t *testing.T) {
		t.Helper()

		t.Run(name, f())
	})

	t.Run("when clearmem+ and "+tc.when, func(t *testing.T) {
		t.Helper()

		t.Run(name, f(func(tc *functionalWriterTestCase) {
			tc.newOpts = append(tc.newOpts, csv.WriterOpts().ClearFreedDataMemory(true))
		}))
	})

	if canConvertStrWriteToFieldWrite(tc) {

		//
		// convert calls from string based APIs to fieldWriter API calls
		//

		t.Run("when str2field+ and "+tc.when, func(t *testing.T) {
			t.Helper()

			t.Run(name, f(func(tc *functionalWriterTestCase) {
				for i := range tc.wrs {
					src := tc.wrs[i].r
					if src == nil {
						continue
					}
					tc.wrs[i].r = nil

					dst := make([]csv.FieldWriter, len(src))
					tc.wrs[i].fwr = dst
					for i, s := range src {
						dst[i] = csv.FieldWriters().String(s)
					}
				}
			}))
		})

		t.Run("when str2field+ clearmem+ and "+tc.when, func(t *testing.T) {
			t.Helper()

			t.Run(name, f(func(tc *functionalWriterTestCase) {
				for i := range tc.wrs {
					src := tc.wrs[i].r
					if src == nil {
						continue
					}
					tc.wrs[i].r = nil

					dst := make([]csv.FieldWriter, len(src))
					tc.wrs[i].fwr = dst
					for i, s := range src {
						dst[i] = csv.FieldWriters().String(s)
					}
				}

				tc.newOpts = append(tc.newOpts, csv.WriterOpts().ClearFreedDataMemory(true))
			}))
		})
	}

	if canConvertStrWriteToFluentWrite(tc) {

		t.Run("when str2fluent+ and "+tc.when, func(t *testing.T) {
			t.Helper()

			t.Run(name, f(func(tc *functionalWriterTestCase) {
				tc.useRecordWriter = true
			}))
		})

		t.Run("when str2fluent+ clearmem+ and "+tc.when, func(t *testing.T) {
			t.Helper()

			t.Run(name, f(func(tc *functionalWriterTestCase) {
				tc.useRecordWriter = true
				tc.newOpts = append(tc.newOpts, csv.WriterOpts().ClearFreedDataMemory(true))
			}))
		})

		t.Run("when str2fluentBytes+ and "+tc.when, func(t *testing.T) {
			t.Helper()

			t.Run(name, f(func(tc *functionalWriterTestCase) {
				tc.useRecordWriter = true
				tc.useRecordWriterBytes = true
			}))
		})

		t.Run("when str2fluentBytes+ clearmem+ and "+tc.when, func(t *testing.T) {
			t.Helper()

			t.Run(name, f(func(tc *functionalWriterTestCase) {
				tc.useRecordWriter = true
				tc.useRecordWriterBytes = true
				tc.newOpts = append(tc.newOpts, csv.WriterOpts().ClearFreedDataMemory(true))
			}))
		})

		t.Run("when useRWEmpty+ and "+tc.when, func(t *testing.T) {
			t.Helper()

			t.Run(name, f(func(tc *functionalWriterTestCase) {
				tc.useRecordWriter = true
				tc.useRecordWriterEmpty = true
			}))
		})

		t.Run("when clearmem+ useRWEmpty+ and "+tc.when, func(t *testing.T) {
			t.Helper()

			t.Run(name, f(func(tc *functionalWriterTestCase) {
				tc.useRecordWriter = true
				tc.newOpts = append(tc.newOpts, csv.WriterOpts().ClearFreedDataMemory(true))
				tc.useRecordWriterEmpty = true
			}))
		})
	}
}

func canConvertStrWriteToFieldWrite(tc *functionalWriterTestCase) bool {
	if tc.afterInitWriter != nil {
		return false
	}

	if len(tc.wrs) == 0 {
		return false
	}

	for i := range tc.wrs {

		if tc.wrs[i].fwr != nil {
			return false
		}
	}

	return true
}

func canConvertStrWriteToFluentWrite(tc *functionalWriterTestCase) bool {
	if tc.afterInitWriter != nil {
		return false
	}

	if len(tc.wrs) == 0 {
		return false
	}

	for i := range tc.wrs {

		if tc.wrs[i].fwr != nil {
			return false
		}
	}

	return true
}

func bomBytes() []byte {
	return []byte{0xEF, 0xBB, 0xBF}
}

func TestFunctionalWriterOKPaths(t *testing.T) {
	t.Parallel()

	tcs := []functionalWriterTestCase{
		{
			when: "rendering a comment header",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines("hello world"),
			},
			res: "# hello world\n",
		},
		{
			when: "rendering a comment header with a BOM",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines("hello world"),
				csv.WriteHeaderOpts().IncludeByteOrderMarker(true),
			},
			res: string(bomBytes()) + "# hello world\n",
		},
		{
			when: "rendering a comment header with a 2 col csv header trimmed",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines("hello world"),
				csv.WriteHeaderOpts().Headers(" a ", " b "),
				csv.WriteHeaderOpts().TrimHeaders(true),
			},
			res: "# hello world\na,b\n",
		},
		{
			when: "rendering a comment header with a 2 col csv header",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines("hello world"),
				csv.WriteHeaderOpts().Headers(" a ", " b "),
			},
			res: "# hello world\n a , b \n",
		},
		{
			when: "rendering a comment header with a 1 col csv header",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines("hello world"),
				csv.WriteHeaderOpts().Headers(" a "),
			},
			res: "# hello world\n a \n",
		},
		{
			when: "rendering a comment header with a 1 empty col csv header",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines("hello world"),
				csv.WriteHeaderOpts().Headers(""),
			},
			res: "# hello world\n\"\"\n",
		},
		{
			when: "rendering a comment header with a 2 col csv header where first col is comment rune",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines("hello world"),
				csv.WriteHeaderOpts().Headers("#", "b"),
			},
			res: "# hello world\n\"#\",b\n",
		},
		{
			when: "rendering a comment header with a 2 col csv header where first col is comment rune + word",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines("hello world"),
				csv.WriteHeaderOpts().Headers("#rawr", "b"),
			},
			res: "# hello world\n\"#rawr\",b\n",
		},
		{
			when: "CRLF record separator",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().RecordSeparator("\r\n"),
			},
			wrs: []wr{
				{r: strings.Split("a,b", ","), n: 5},
			},
			res: "a,b\r\n",
		},
		{
			when: "CR record separator",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().RecordSeparator("\r"),
			},
			wrs: []wr{
				{r: strings.Split("a,b", ","), n: 4},
			},
			res: "a,b\r",
		},
		{
			when: "err on non-utf8 enabled implicitly, record has literal replacement utf8 char",
			then: "RuneError should be in output because it was in the source data and is a valid utf8 character sequence",
			wrs: []wr{
				{r: strings.Split(string(utf8.RuneError)+",b", ","), n: 6},
			},
			res: string(utf8.RuneError) + ",b\n",
		},
		{
			when: "err on non-utf8 enabled explicitly, record has literal replacement utf8 char",
			then: "RuneError should be in output because it was in the source data and is a valid utf8 character sequence",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().ErrorOnNonUTF8(true),
			},
			wrs: []wr{
				{r: strings.Split(string(utf8.RuneError)+",b", ","), n: 6},
			},
			res: string(utf8.RuneError) + ",b\n",
		},
		{
			when: "escape set to implicit quote value",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('"'),
			},
			wrs: []wr{
				{r: strings.Split("a,b", ","), n: 4},
			},
			res: "a,b\n",
		},
		{
			when: "escape set to explicit quote value",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('"'),
				csv.WriterOpts().Quote('"'),
			},
			wrs: []wr{
				{r: strings.Split("a,b", ","), n: 4},
			},
			res: "a,b\n",
		},
		{
			when: "two rows and two columns with second column empty",
			wrs: []wr{
				{r: strings.Split("a,", ","), n: 3},
				{r: strings.Split("b,", ","), n: 3},
			},
			res: "a,\nb,\n",
		},
		{
			when: "field contains LF character",
			then: "LF should be rendered within quotes",
			wrs: []wr{
				{r: []string{"\n"}, n: 4},
			},
			res: "\"\n\"\n",
		},
		{
			when: "field contains quote character",
			then: "field should have quote doubled and then wrapped in quotes",
			wrs: []wr{
				{r: []string{"\""}, n: 5},
			},
			res: "\"\"\"\"\n",
		},
		{
			when: "short field and long field with quotes",
			then: "should cover field buffer reallocation code path",
			wrs: []wr{
				{r: strings.Split("a\"b,zzzzzzzzzzzzzzz\"zzzzzzzzzzzzzzz", ","), n: 42},
			},
			res: "\"a\"\"b\",\"zzzzzzzzzzzzzzz\"\"zzzzzzzzzzzzzzz\"\n",
		},
		{
			when: "comment rune on writer, comment header, escape set, and two records that start with comment runes",
			then: "first row, first column should be quoted",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\\'),
				csv.WriterOpts().CommentRune('#'),
			},
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentLines("hello world"),
			},
			whN: 14,
			wrs: []wr{
				{r: strings.Split("#,b", ","), n: 6},
				{r: strings.Split("#,2", ","), n: 4},
			},
			res: "# hello world\n\"#\",b\n#,2\n",
		},
		{
			when: "comment rune on writer, no comment header, escape set, and two records that start with comment runes",
			then: "first row, first column should be quoted",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\\'),
				csv.WriterOpts().CommentRune('#'),
			},
			whN: 14,
			wrs: []wr{
				{r: strings.Split("#,b", ","), n: 6},
				{r: strings.Split("#,2", ","), n: 4},
			},
			res: "\"#\",b\n#,2\n",
		},
		{
			when: "writing a record using supposedly already UTF8 encoded string and []byte field writers which are not properly utf8 encoded",
			then: "no error should be thrown on write",
			wrs: []wr{
				{fwr: []csv.FieldWriter{csv.FieldWriters().UncheckedUTF8Bytes([]byte("\xF8")), csv.FieldWriters().UncheckedUTF8String("\xFF")}, n: 4},
			},
			res: "\xF8,\xFF\n",
		},
		{
			when: "writing a record using string and []byte field writers which are not properly utf8 encoded and the writer has ErrOnNonUTF8=false",
			then: "no error should be thrown on write",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().ErrorOnNonUTF8(false),
			},
			wrs: []wr{
				{fwr: []csv.FieldWriter{csv.FieldWriters().Bytes([]byte("\xF8")), csv.FieldWriters().String("\xFF")}, n: 4},
			},
			res: "\xF8,\xFF\n",
		},
		{
			when: "data to write has 2 columns and a field separator value that is numeric",
			then: "wFlagControlRuneOverlap paths are used to format the output",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().FieldSeparator('0'),
			},
			wrs: []wr{
				{fwr: []csv.FieldWriter{csv.FieldWriters().Int(1), csv.FieldWriters().Int(9)}, n: 4},
			},
			res: "109\n",
		},
		{
			when: "data to write has 2 columns and a field separator value that is numeric and fields are the numeric sep value",
			then: "wFlagControlRuneOverlap paths are used to format the output and quotes are used",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().FieldSeparator('0'),
			},
			wrs: []wr{
				{fwr: []csv.FieldWriter{csv.FieldWriters().Int(0), csv.FieldWriters().Int(0)}, n: 8},
			},
			res: "\"0\"0\"0\"\n",
		},
		{
			when: "writing a record of one empty byte slice column",
			then: "the output is quoted",
			wrs: []wr{
				{fwr: []csv.FieldWriter{csv.FieldWriters().Bytes(nil)}, n: 3},
			},
			res: "\"\"\n",
		},
		{
			when: "writing a record of two empty byte slice columns",
			then: "the output is not quoted",
			wrs: []wr{
				{fwr: []csv.FieldWriter{csv.FieldWriters().Bytes(nil), csv.FieldWriters().Bytes(nil)}, n: 2},
			},
			res: ",\n",
		},
		{
			when: "writing a record of one empty string slice column",
			then: "the output is quoted",
			wrs: []wr{
				{r: []string{""}, n: 3},
			},
			res: "\"\"\n",
		},
		{
			when: "writing a record of two empty string slice columns",
			then: "the output is not quoted",
			wrs: []wr{
				{r: []string{"", ""}, n: 2},
			},
			res: ",\n",
		},
		{
			when: "writing a record of one rune column",
			then: "the output is not quoted",
			wrs: []wr{
				{fwr: []csv.FieldWriter{csv.FieldWriters().Rune('A')}, n: 2},
			},
			res: "A\n",
		},
		{
			when: "writing a record of one UncheckedUTF8String column with value set to escape rune",
			then: "escape rune should be doubly escaped and quoted",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{fwr: []csv.FieldWriter{csv.FieldWriters().UncheckedUTF8String(`\`)}, n: 5},
			},
			res: "\"\\\\\"\n",
		},
		{
			when: "writing a record of one UncheckedUTF8String column with value set to two escape runes",
			then: "escape rune should be doubly escaped and quoted",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{fwr: []csv.FieldWriter{csv.FieldWriters().UncheckedUTF8String(`\\`)}, n: 7},
			},
			res: "\"\\\\\\\\\"\n",
		},
		{
			when: "writing a record of two UncheckedUTF8String columns with second value set to escape rune",
			then: "escape rune should be doubly escaped and quoted",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{fwr: []csv.FieldWriter{csv.FieldWriters().UncheckedUTF8String(""), csv.FieldWriters().UncheckedUTF8String(`\`)}, n: 6},
			},
			res: ",\"\\\\\"\n",
		},
		{
			when: "writing a record of two UncheckedUTF8String column with second value set to two escape runes",
			then: "escape rune should be doubly escaped and quoted",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{fwr: []csv.FieldWriter{csv.FieldWriters().UncheckedUTF8String(""), csv.FieldWriters().UncheckedUTF8String(`\\`)}, n: 8},
			},
			res: ",\"\\\\\\\\\"\n",
		},
		{
			when: "writing a record of two UncheckedUTF8String columns with second value set to quote rune",
			then: "quote rune should be escaped and quoted",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{fwr: []csv.FieldWriter{csv.FieldWriters().UncheckedUTF8String(""), csv.FieldWriters().UncheckedUTF8String(`"`)}, n: 6},
			},
			res: ",\"\\\"\"\n",
		},
		{
			when: "writing a record of two UncheckedUTF8String column with second value set to two quote runes",
			then: "quote rune should be escaped and quoted",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{fwr: []csv.FieldWriter{csv.FieldWriters().UncheckedUTF8String(""), csv.FieldWriters().UncheckedUTF8String(`""`)}, n: 8},
			},
			res: ",\"\\\"\\\"\"\n",
		},
		{
			when: "writing a record of two String columns with second value set to escape rune",
			then: "escape rune should be doubly escaped and quoted",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{r: []string{"", `\`}, n: 6},
			},
			res: ",\"\\\\\"\n",
		},
		{
			when: "writing a record of two String columns with second value set to quote rune",
			then: "quote rune should be escaped and quoted",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{r: []string{"", `"`}, n: 6},
			},
			res: ",\"\\\"\"\n",
		},
		{
			when: "writing a record of two Rune column with second value set to comma",
			then: "comma rune should be quoted",
			wrs: []wr{
				{fwr: []csv.FieldWriter{csv.FieldWriters().Rune('0'), csv.FieldWriters().Rune(',')}, n: 6},
			},
			res: "0,\",\"\n",
		},
		{
			when: "writing a record of two String column with second value set to comma",
			then: "comma rune should be quoted",
			wrs: []wr{
				{r: []string{"", ","}, n: 5},
			},
			res: ",\",\"\n",
		},
		{
			when: "writing a record of two String column with second value set to comma",
			then: "comma rune should be quoted",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().RecordSeparator("\u2028"),
			},
			wrs: []wr{
				{r: []string{"abc", "123"}, n: 10},
			},
			res: "abc,123\u2028",
		},
		{
			when: "quote character is a 4 byte grinning face icon and field contains another multibyte character",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Quote('\U0001F600'),
			},
			wrs: []wr{
				{r: []string{"a\u20ACc", "d\u20ACf"}, n: 12},
			},
			res: "a\u20ACc,d\u20ACf\n",
		},
		{
			when: "quote character is a 4 byte grinning face and field contains another multibyte character",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Quote('\U0001F600'),
			},
			wrs: []wr{
				{r: []string{"a\u20ACc", "d\u20ACf"}, n: 12},
			},
			res: "a\u20ACc,d\u20ACf\n",
		},
		{
			when: "quote character is a 4 byte grinning face and field contains quote and another multibyte character",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Quote('\U0001F600'),
			},
			wrs: []wr{
				{r: []string{"a\u20ACc", "d\U0001F600f\u20ACh"}, n: 29},
			},
			res: "a\u20ACc,\U0001F600d\U0001F600\U0001F600f\u20ACh\U0001F600\n",
		},
		{
			when: "ErrorOnNonUTF8=false second field has no control character overlap",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().ErrorOnNonUTF8(false),
			},
			wrs: []wr{
				{r: []string{"abc", "123"}, n: 8},
			},
			res: "abc,123\n",
		},
		{
			when: "ErrorOnNonUTF8=false second field has quote",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().ErrorOnNonUTF8(false),
			},
			wrs: []wr{
				{r: []string{"abc", `1"3`}, n: 11},
			},
			res: `abc,"1""3"` + "\n",
		},
		{
			when: "ErrorOnNonUTF8=false second field has record separator",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().ErrorOnNonUTF8(false),
			},
			wrs: []wr{
				{r: []string{"abc", "1\n3"}, n: 10},
			},
			res: "abc,\"1\n3\"" + "\n",
		},
		{
			when: "ErrorOnNonUTF8=false second field has escape",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().ErrorOnNonUTF8(false),
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{r: []string{"abc", `1\3`}, n: 11},
			},
			res: `abc,"1\\3"` + "\n",
		},
		{
			when: "ErrorOnNonUTF8=false first field has two escapes",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().ErrorOnNonUTF8(false),
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{r: []string{`1\3\5`, "abc"}, n: 14},
			},
			res: `"1\\3\\5",abc` + "\n",
		},
		{
			when: "ErrorOnNonUTF8=false second field has two escapes",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().ErrorOnNonUTF8(false),
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{r: []string{"abc", `1\3\5`}, n: 14},
			},
			res: `abc,"1\\3\\5"` + "\n",
		},
		{
			when: "ErrorOnNonUTF8=true first field has two escapes",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().ErrorOnNonUTF8(true),
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{r: []string{`1\3\5`, "abc"}, n: 14},
			},
			res: `"1\\3\\5",abc` + "\n",
		},
		{
			when: "ErrorOnNonUTF8=true second field has two escapes",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().ErrorOnNonUTF8(true),
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{r: []string{"abc", `1\3\5`}, n: 14},
			},
			res: `abc,"1\\3\\5"` + "\n",
		},
	}

	for _, tc := range tcs {
		tc.Run(t)
	}
}

func TestWriteFieldRow(t *testing.T) {
	t.Parallel()

	is := assert.New(t)

	var buf bytes.Buffer

	func() {
		fw := csv.FieldWriters()

		cw, err := csv.NewWriter(
			csv.WriterOpts().Writer(&buf),
		)
		is.Nil(err)
		defer func() {
			is.Nil(cw.Close())

			n, err := cw.WriteFieldRow()
			is.Equal(csv.ErrWriterClosed, err)
			is.Zero(n)
		}()

		n, err := cw.WriteFieldRow(
			fw.Int(-1),
			fw.String(""),
			fw.String("a"),
			fw.Bytes(nil),
			fw.Bytes([]byte("b")),
			fw.Bool(true),
		)
		is.Nil(err)
		is.Equal(11, n)
	}()

	is.Equal(buf.String(), `-1,,a,,b,1`+"\n")
}

func TestWriteFieldRowWithQuote(t *testing.T) {
	t.Parallel()

	is := assert.New(t)

	var buf bytes.Buffer

	func() {
		fw := csv.FieldWriters()

		cw, err := csv.NewWriter(
			csv.WriterOpts().Writer(&buf),
			csv.WriterOpts().Quote('"'),
		)
		is.Nil(err)
		defer func() {
			is.Nil(cw.Close())

			n, err := cw.WriteFieldRow()
			is.Equal(csv.ErrWriterClosed, err)
			is.Zero(n)
		}()

		n, err := cw.WriteFieldRow(
			fw.Int(-1),
			fw.String(""),
			fw.String("a"),
			fw.Bytes(nil),
			fw.Bytes([]byte("b")),
			fw.Bool(true),
			fw.Rune('"'),
		)
		is.Nil(err)
		is.Equal(16, n)
	}()

	is.Equal(buf.String(), `-1,,a,,b,1,""""`+"\n")
}

func TestWriteFieldRowWithMemclearOn(t *testing.T) {
	t.Parallel()

	is := assert.New(t)

	var buf bytes.Buffer

	func() {
		fw := csv.FieldWriters()

		cw, err := csv.NewWriter(
			csv.WriterOpts().Writer(&buf),
			csv.WriterOpts().ClearFreedDataMemory(true),
		)
		is.Nil(err)
		defer func() {
			is.Nil(cw.Close())

			n, err := cw.WriteFieldRow()
			is.Equal(csv.ErrWriterClosed, err)
			is.Zero(n)
		}()

		n, err := cw.WriteFieldRow(
			fw.Int(-1),
			fw.String(""),
			fw.String("a"),
			fw.Bytes(nil),
			fw.Bytes([]byte("b")),
			fw.Bool(true),
		)
		is.Nil(err)
		is.Equal(11, n)
	}()

	is.Equal(buf.String(), `-1,,a,,b,1`+"\n")
}

func TestWriteFieldRowBorrowed(t *testing.T) {
	t.Parallel()

	is := assert.New(t)

	var buf bytes.Buffer

	func() {
		fw := csv.FieldWriters()

		cw, err := csv.NewWriter(
			csv.WriterOpts().Writer(&buf),
		)
		is.Nil(err)
		defer func() {
			is.Nil(cw.Close())

			n, err := cw.WriteFieldRowBorrowed(nil)
			is.Equal(csv.ErrWriterClosed, err)
			is.Zero(n)
		}()

		fields := []csv.FieldWriter{
			fw.Int(-1),
			fw.String(""),
			fw.String("a"),
			fw.Bytes(nil),
			fw.Bytes([]byte("b")),
			fw.Bool(true),
		}

		n, err := cw.WriteFieldRowBorrowed(fields)
		is.Nil(err)
		is.Equal(11, n)
	}()

	is.Equal(buf.String(), `-1,,a,,b,1`+"\n")
}

func TestWriteFieldRowBorrowedWithMemclearOn(t *testing.T) {
	t.Parallel()

	is := assert.New(t)

	var buf bytes.Buffer

	func() {
		fw := csv.FieldWriters()

		cw, err := csv.NewWriter(
			csv.WriterOpts().Writer(&buf),
			csv.WriterOpts().ClearFreedDataMemory(true),
		)
		is.Nil(err)
		defer func() {
			is.Nil(cw.Close())

			n, err := cw.WriteFieldRowBorrowed(nil)
			is.Equal(csv.ErrWriterClosed, err)
			is.Zero(n)
		}()

		fields := []csv.FieldWriter{
			fw.Int(-1),
			fw.String(""),
			fw.String("a"),
			fw.Bytes(nil),
			fw.Bytes([]byte("b")),
			fw.Bool(true),
		}

		n, err := cw.WriteFieldRowBorrowed(fields)
		is.Nil(err)
		is.Equal(11, n)
	}()

	is.Equal(buf.String(), `-1,,a,,b,1`+"\n")
}

func TestWriteFieldRowWithInvalidField(t *testing.T) {
	t.Parallel()

	is := assert.New(t)

	var buf bytes.Buffer

	func() {
		cw, err := csv.NewWriter(
			csv.WriterOpts().Writer(&buf),
		)
		is.Nil(err)
		defer func() {
			is.Nil(cw.Close())

			n, err := cw.WriteFieldRow()
			is.Equal(csv.ErrWriterClosed, err)
			is.Zero(n)
		}()

		invalidField := csv.FieldWriter{}
		n, err := cw.WriteFieldRow(
			invalidField,
		)
		is.NotNil(err)
		is.ErrorIs(err, csv.ErrInvalidFieldWriter)
		is.Zero(n)
	}()

	is.Equal(buf.String(), ``)
}

func TestWriteFieldRowWithInvalidFieldWithMemclearOn(t *testing.T) {
	t.Parallel()

	is := assert.New(t)

	var buf bytes.Buffer

	func() {
		cw, err := csv.NewWriter(
			csv.WriterOpts().Writer(&buf),
			csv.WriterOpts().ClearFreedDataMemory(true),
		)
		is.Nil(err)
		defer func() {
			is.Nil(cw.Close())

			n, err := cw.WriteFieldRow()
			is.Equal(csv.ErrWriterClosed, err)
			is.Zero(n)
		}()

		invalidField := csv.FieldWriter{}
		n, err := cw.WriteFieldRow(
			invalidField,
		)
		is.NotNil(err)
		is.ErrorIs(err, csv.ErrInvalidFieldWriter)
		is.Zero(n)
	}()

	is.Equal(buf.String(), ``)
}
