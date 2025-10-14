package csv_test

import (
	"testing"

	"github.com/josephcopenhaver/csv-go/v3"
)

func TestFunctionalWriterProcessFieldOKPaths(t *testing.T) {
	t.Parallel()

	tcs := []functionalWriterTestCase{
		{
			when: "err on non-utf8 is disabled and non-utf8 present in field and escape disabled",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().ErrorOnNonUTF8(false),
			},
			wrs: []wr{
				{r: []string{string([]byte{0xC0})}, n: 2},
			},
			res: "\xc0\n",
		},
		{
			when: "err on non-utf8 is disabled and non-utf8 present in field and escape enabled",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().ErrorOnNonUTF8(false),
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{r: []string{string([]byte{0xC0})}, n: 2},
			},
			res: "\xc0\n",
		},
		{
			when: "escape enabled and quote in field",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{r: []string{"\""}, n: 5},
			},
			res: "\"\\\"\"\n",
		},
		{
			when: "two quotes in field",
			wrs: []wr{
				{r: []string{"\"\""}, n: 7},
			},
			res: "\"\"\"\"\"\"\n",
		},
		{
			when: "field separator in field",
			wrs: []wr{
				{r: []string{","}, n: 4},
			},
			res: "\",\"\n",
		},
		{
			when: "escape enabled and field separator in field",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{r: []string{","}, n: 4},
			},
			res: "\",\"\n",
		},
		{
			when: "escape enabled and escape in field",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{r: []string{"\\"}, n: 5},
			},
			res: "\"\\\\\"\n",
		},
		{
			when: "escape enabled and two escapes in field",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{r: []string{"\\\\"}, n: 7},
			},
			res: "\"\\\\\\\\\"\n",
		},
		{
			when: "escape enabled and escape followed by quote in field",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{r: []string{"\\\""}, n: 7},
			},
			res: "\"\\\\\\\"\"\n",
		},
		{
			when: "quote in field followed by non-utf8 and error on non-utf8 is disabled",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().ErrorOnNonUTF8(false),
			},
			wrs: []wr{
				{r: []string{"\"" + string([]byte{0xC0})}, n: 6},
			},
			res: "\"\"\"" + string([]byte{0xC0}) + "\"\n",
		},
		{
			when: "quote in field followed by non-utf8 and error on non-utf8 is disabled and escape enabled",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().ErrorOnNonUTF8(false),
				csv.WriterOpts().Escape('\\'),
			},
			wrs: []wr{
				{r: []string{"\"" + string([]byte{0xC0})}, n: 6},
			},
			res: "\"\\\"" + string([]byte{0xC0}) + "\"\n",
		},
	}

	for _, tc := range tcs {
		tc.Run(t)
	}
}
