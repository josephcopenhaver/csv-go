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

	cr := csv.NewReader(r)
	for cr.Scan() {
		row := cr.Row()
		println(strings.Join(row, ","))
	}
	if err := cr.Err(); err != nil {
		panic(err)
	}
}
