package main

import (
	"bytes"
	_ "embed"
	"fmt"

	"github.com/josephcopenhaver/csv-go"
)

func main() {
	var w bytes.Buffer

	op := csv.WriterOpts()
	cw, err := csv.NewWriter(
		op.Writer(&w),
		op.DiscoverFieldCount(true),
	)
	if err != nil {
		panic(err)
	}

	if _, err := cw.WriteRow([]string{"a", "b"}); err != nil {
		panic(err)
	}

	if _, err := cw.WriteRow([]string{"", ""}); err != nil {
		panic(err)
	}

	fmt.Printf("%s", w.String())
}
