package csv_test

import (
	"strings"
	"testing"

	"github.com/josephcopenhaver/csv-go/v2"
)

func TestFunctionalWriterHeaderOKPaths(t *testing.T) {
	tcs := []functionalWriterTestCase{
		{
			when: "header comment contains invalid utf8 sequence and non-utf8 errors are disabled",
			then: "the byte is still written to the writer correctly as-is",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().ErrorOnNonUTF8(false),
			},
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines(string([]byte{0xC0})),
			},
			res: "# " + string([]byte{0xC0}) + "\n",
		},
		{
			when: "header comment contains a CR newline character",
			then: "the newline character is discovered, replaced with the record separator and followed by the comment rune",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines("hello\rworld"),
			},
			res: "# hello\n# world\n",
		},
		{
			when: "header comment contains a CRLF newline sequence",
			then: "the newline sequence is discovered, replaced with the record separator and followed by the comment rune",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines("hello\r\nworld"),
			},
			res: "# hello\n# world\n",
		},
		{
			when: "comment header and two records that start with comment runes",
			then: "first data record first column should be quoted",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
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
			when: "comment header and two records with one empty column",
			then: "data row should be rendered as a pair of empty quotes",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines("hello world"),
			},
			whN: 14,
			wrs: []wr{
				{r: []string{""}, n: 3},
				{r: []string{""}, n: 3},
			},
			res: "# hello world\n\"\"\n\"\"\n",
		},
		{
			when: "comment header, escape set, and two records that start with comment runes",
			then: "first row, first column should be quoted",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\\'),
			},
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
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
			when: "comment header, escape set, and two records that start with escape runes",
			then: "data row fields with escapes should be doubly escaped",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\\'),
			},
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines("hello world"),
			},
			whN: 14,
			wrs: []wr{
				{r: strings.Split("\\,b", ","), n: 7},
				{r: strings.Split("\\,2", ","), n: 7},
			},
			res: "# hello world\n\"\\\\\",b\n\"\\\\\",2\n",
		},
		{
			when: "comment lines contains an empty string",
			then: "document should just have the comment rune line sequence",
			whOpts: []csv.WriteHeaderOption{
				csv.WriteHeaderOpts().CommentRune('#'),
				csv.WriteHeaderOpts().CommentLines(""),
			},
			res: "# \n",
		},
	}

	for _, tc := range tcs {
		tc.Run(t)
	}
}
