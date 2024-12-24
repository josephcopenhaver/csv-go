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
	)
	if err != nil {
		panic(err)
	}

	if _, err := cw.WriteHeader(
		csv.WriteHeaderOpts().Headers([]string{"first", "second"}),
		csv.WriteHeaderOpts().CommentRune('#'),
		csv.WriteHeaderOpts().CommentLines("hello", "aloha"),
	); err != nil {
		panic(err)
	}

	if _, err := cw.WriteRow("a", "b"); err != nil {
		panic(err)
	}

	if _, err := cw.WriteRow("", ""); err != nil {
		panic(err)
	}

	fmt.Printf("%s", w.String())
}
