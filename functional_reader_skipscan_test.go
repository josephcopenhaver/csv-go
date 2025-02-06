package csv_test

import (
	"strings"
	"testing"

	"github.com/josephcopenhaver/csv-go"
)

func TestFunctionalReaderSkipscanPaths(t *testing.T) {

	tcs := []functionalReaderTestCase{
		{
			when: "reading from a one-row row-borrow-unspecified reader",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("1,2,3")),
			},
		},
		{
			when: "reading from a one-row row-borrow-disabled reader",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("1,2,3")),
				csv.ReaderOpts().BorrowRow(false),
			},
		},
		{
			when: "reading from a one-row row-borrow-enabled reader",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("1,2,3")),
				csv.ReaderOpts().BorrowRow(true),
			},
		},
	}

	for _, tc := range tcs {
		tc.skipScan = true
		tc.when += " closed before calling Scan"
		tc.Run(t)
	}
}
