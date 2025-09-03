package csv_test

import (
	"testing"

	"github.com/josephcopenhaver/csv-go/v3"
)

func TestFunctionalWriterProcessFieldErrorPaths(t *testing.T) {
	t.Parallel()

	tcs := []functionalWriterTestCase{
		{
			when: "field contains quote followed by non-UTF8 sequence",
			wrs: []wr{
				{r: []string{"\"" + string([]byte{0xC0})}, errIs: []error{csv.ErrNonUTF8InRecord}, errStr: csv.ErrNonUTF8InRecord.Error()},
			},
		},
		{
			when: "field contains non-UTF8 sequence and escape set",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{r: []string{string([]byte{0xC0})}, errIs: []error{csv.ErrNonUTF8InRecord}, errStr: csv.ErrNonUTF8InRecord.Error()},
			},
		},
		{
			when: "field contains quote followed by non-UTF8 sequence and escape set",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{r: []string{"\"" + string([]byte{0xC0})}, errIs: []error{csv.ErrNonUTF8InRecord}, errStr: csv.ErrNonUTF8InRecord.Error()},
			},
		},
		{
			when: "non-UTF8 sequence after writing a comment with the first field starting with a comment rune",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines("hello"),
			},
			whN: 8,
			wrs: []wr{
				{r: []string{"#" + string([]byte{0xC0})}, errIs: []error{csv.ErrNonUTF8InRecord}, errStr: csv.ErrNonUTF8InRecord.Error()},
			},
			res: "# hello\n",
		},
		{
			when: "escape enabled and non-UTF8 sequence after writing a comment with the first field starting with a comment rune",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\\'),
			},
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines("hello"),
			},
			whN: 8,
			wrs: []wr{
				{r: []string{"#" + string([]byte{0xC0})}, errIs: []error{csv.ErrNonUTF8InRecord}, errStr: csv.ErrNonUTF8InRecord.Error()},
			},
			res: "# hello\n",
		},
	}

	for _, tc := range tcs {
		if tc.then == "" {
			tc.then = "a coupled error should occur"
		}
		tc.Run(t)
	}
}
