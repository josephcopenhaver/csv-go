package main

import (
	"bytes"
	_ "embed"

	"github.com/josephcopenhaver/csv-go/v2"
)

//go:embed pizza.csv
var data []byte

var expHeaders = []string{
	"Pizza Type",
	"Origin Locality",
	"Is Round",
}

func main() {
	r := bytes.NewReader(data)

	op := csv.ReaderOpts()
	cr, err := csv.NewReader(
		op.Reader(r),                 // Reader must always be specified
		op.ExpectHeaders(expHeaders), // ExpectHeaders is an optional validation check
		op.RemoveHeaderRow(true),     // stops header record from emitting with the data records
		op.Quote('"'),                // by default quotes have no meaning, so must be specified to match RFC 4180
		op.ErrorOnNoRows(true),       //
		// op.NumFields(3),           // not required: will be auto-discovered
		// op.FieldSeparator(','),    // not required: matches default value
		// op.RecordSeparator("\n"),  // not required: matches default value
	)
	if err != nil {
		panic(err)
	}
	defer cr.Close()

	for row := range cr.IntoIter() {
		println(row[0], row[1], row[2])
	}
	if err := cr.Err(); err != nil {
		// for a given document stream:
		// should the number of fields per line ever change,
		// the reader error unexpectedly, or the contents
		// of the document violate expectations specified
		// via config options, then there will be an error
		// returned here

		// There is no guarantee that the stream has been fully
		// read when an error is encountered nor should authors
		// use the errors to infer such a low level of detail.
		panic(err)
	}
}
