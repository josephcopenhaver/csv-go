package csv_test

import (
	"bytes"
	"errors"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/josephcopenhaver/csv-go"
	"github.com/stretchr/testify/assert"
)

type wr struct {
	r                []string
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
}

func (tc *functionalWriterTestCase) Run(t *testing.T) {
	assert.NotEmpty(t, tc.when)
	t.Helper()

	f := func(options ...func(*functionalWriterTestCase)) func(t *testing.T) {
		tc_copy := *tc
		for _, f := range options {
			f(&tc_copy)
		}
		if f := tc_copy.selfInit; f != nil {
			f(&tc_copy)
		}

		return func(t *testing.T) {
			tc := &tc_copy
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
						is.Equal(0, tc.whN)
					}
				}

				// attempting to call write headers a second time should always
				// result in a header already written error
				//
				// unless there was an io level error previously or the writer was closed
				{
					v, err := cw.WriteHeader()
					is.Zero(v)
					is.NotNil(err)
					if errWriterInIOErrState != nil {
						is.True(err == errWriterInIOErrState)
					} else {
						is.ErrorIs(err, csv.ErrHeaderWritten)
					}
				}
			}

			for _, v := range tc.wrs {
				n, err := cw.WriteRow(v.r...)
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
}

func TestFunctionalWriterOKPaths(t *testing.T) {
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
			res: string([]byte{0xEF, 0xBB, 0xBF}) + "# hello world\n",
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
	}

	for _, tc := range tcs {
		tc.Run(t)
	}
}
