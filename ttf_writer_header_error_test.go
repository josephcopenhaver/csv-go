package csv_test

import (
	"testing"
	"unicode/utf8"

	"github.com/josephcopenhaver/csv-go/v2"
)

func TestFunctionalWriterHeaderErrorPaths(t *testing.T) {
	t.Parallel()

	tcs := []functionalWriterTestCase{
		{
			when: "zero header length",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().Headers([]string{}...),
			},
			whErrIs:  []error{csv.ErrBadConfig},
			whErrStr: csv.ErrBadConfig.Error() + "\nheaders length must be greater than zero",
		},
		{
			when: "nil headers",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().Headers([]string(nil)...),
			},
			whErrIs:  []error{csv.ErrBadConfig},
			whErrStr: csv.ErrBadConfig.Error() + "\nheaders length must be greater than zero",
		},
		{
			when: "headers len does not match numFields",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().NumFields(1),
			},
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().Headers("a", "b"),
			},
			whErrIs:  []error{csv.ErrBadConfig},
			whErrStr: csv.ErrBadConfig.Error() + "\nheaders length does not match number of fields",
		},
		{
			when: "comment rune is RuneError",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune(utf8.RuneError),
			},
			whErrIs:  []error{csv.ErrBadConfig},
			whErrStr: csv.ErrBadConfig.Error() + "\ninvalid comment rune",
		},
		{
			when: "comment rune is LF",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('\n'),
			},
			whErrIs:  []error{csv.ErrBadConfig},
			whErrStr: csv.ErrBadConfig.Error() + "\ninvalid comment rune",
		},
		{
			when: "comment rune is quote",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('"'),
			},
			whErrIs:  []error{csv.ErrBadConfig},
			whErrStr: csv.ErrBadConfig.Error() + "\ninvalid quote and comment rune combination",
		},
		{
			when: "comment rune is comma",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune(','),
			},
			whErrIs:  []error{csv.ErrBadConfig},
			whErrStr: csv.ErrBadConfig.Error() + "\ninvalid field separator and comment rune combination",
		},
		{
			when: "comment rune is escape",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\\'),
			},
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('\\'),
			},
			whErrIs:  []error{csv.ErrBadConfig},
			whErrStr: csv.ErrBadConfig.Error() + "\ninvalid escape and comment rune combination",
		},
		{
			when: "comment lines without rune",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentLines("hello world"),
			},
			whErrIs:  []error{csv.ErrBadConfig},
			whErrStr: csv.ErrBadConfig.Error() + "\ncomment lines require a comment rune",
		},
		{
			when: "comment lines contains an invalid utf8 character",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines(string([]byte{0xC0})),
			},
			whErrIs:  []error{csv.ErrNonUTF8InComment},
			whErrStr: csv.ErrNonUTF8InComment.Error(),
		},
	}

	for _, tc := range tcs {
		if tc.then == "" {
			tc.then = "a coupled error should occur"
		}
		tc.Run(t)
	}
}
