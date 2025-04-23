package main

import (
	"bytes"
	"fmt"

	"github.com/josephcopenhaver/csv-go/v2"
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
	defer func() {
		if err := cw.Close(); err != nil {
			panic(err)
		}
	}()

	if _, err := cw.WriteHeader(
		csv.WriteHeaderOpts().CommentRune('#'),
		csv.WriteHeaderOpts().CommentLines(
			"hello",
			"aloha",
		),
		csv.WriteHeaderOpts().Headers("first", "second"),
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
