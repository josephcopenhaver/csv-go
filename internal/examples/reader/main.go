package main

import (
	"bytes"
	_ "embed"
	"strings"

	"github.com/josephcopenhaver/csv-go"
)

//go:embed example.csv
var data []byte

func main() {
	r := bytes.NewReader(data)

	op := csv.ReaderOpts()
	cr, err := csv.NewReader(
		op.Reader(r),
		op.Quote('"'),
		op.ErrorOnQuotesInUnquotedField(false),
	)
	if err != nil {
		panic(err)
	}

	for row := range cr.IntoIter() {
		println(strings.Join(row, ","))
	}
	if err := cr.Err(); err != nil {
		panic(err)
	}
}
