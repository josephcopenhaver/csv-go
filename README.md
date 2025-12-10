# csv-go

csv-go

[![Go Report Card](https://goreportcard.com/badge/github.com/josephcopenhaver/csv-go)](https://goreportcard.com/report/github.com/josephcopenhaver/csv-go/v3)
![tests](https://github.com/josephcopenhaver/csv-go/actions/workflows/tests.yaml/badge.svg)
![code-coverage](https://img.shields.io/badge/code_coverage-100%25-rgb%2852%2C208%2C88%29)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

This package is a highly flexible and performant single threaded csv stream reader and writer. It opts for strictness with nearly all options off by default. Using the option functions pattern on Reader and Writer creation ensures extreme flexibility can be offered while configuration can be validated up-front in cold paths. This creates an immutable, clear execution of the csv file/stream parsing strategy. It has been battle tested thoroughly in production contexts for both correctness and speed so feel free to use in any way you like.

Both the reader and writer are [more performant than the standard go csv package](docs/BENCHMARKS.md) when compared in an apples-to-apples configuration between the two. The writer also has several optimizations for non-string type serialization via the fluent api returned by csv.Writer.NewRecord() and FieldWriters(). I expect mileage here to vary over time. My primary goal with this lib was to solve my own edge case problems like suspect-encodings/loose-rules and offer something back more aligned with others that think like myself regarding reducing allocations, GC pause, and increasing efficiency.

```go
package main

// this is a toy example that reads a csv file and writes to another

import (
	"os"

	"github.com/josephcopenhaver/csv-go/v3"
)

func main() {
	r, err := os.Open("input.csv")
	if err != nil {
		panic(err)
	}
	defer r.Close()

	cr, err := csv.NewReader(
		csv.ReaderOpts().Reader(r),
		// by default quotes have no meaning
		// so must be specified to match RFC 4180
		// csv.ReaderOpts().Quote('"'),
	)
	if err != nil {
		panic(err)
	}
	defer cr.Close()

	w, err := os.Create("output.csv")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := w.Close(); err != nil {
			panic(err)
		}
	}()

	cw, err := csv.NewWriter(
		csv.WriterOpts().Writer(w),
	)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := cw.Close(); err != nil {
			panic(err)
		}
	}()

	for row := range cr.IntoIter() {
		if _, err := cw.WriteRow(row...); err != nil {
			panic(err)
		}
	}
	if err := cr.Err(); err != nil {
		panic(err)
	}
}
```

See the [Reader](internal/examples/reader/main.go) and [Writer](internal/examples/writer/main.go) examples for more in-depth usages.

## Reader Features

| Name | option(s) |
| - | - |
| Zero allocations during processing | BorrowRow + BorrowFields + InitialRecordBuffer + InitialRecordBufferSize + NumFields |
| Format Specification | Comment + CommentsAllowedAfterStartOfRecords + Escape + FieldSeparator + Quote + RecordSeparator + NumFields |
| Format Discovery | DiscoverRecordSeparator |
| Data Loss Prevention | ClearFreedDataMemory |
| Byte Order Marker Support | RemoveByteOrderMarker + ErrorOnNoByteOrderMarker
| Headers Support | ExpectHeaders + RemoveHeaderRow + TrimHeaders |
| Reader Buffer tuning | ReaderBuffer + ReaderBufferSize |
| Format Validation | ErrorOnNoRows + ErrorOnNewlineInUnquotedField + ErrorOnQuotesInUnquotedField |
| Security Limits | MaxFields + MaxRecordBytes + MaxRecords + MaxComments + MaxCommentBytes |

## Writer Features

| Name | option(s) |
| - | - |
| Zero allocations | InitialRecordBufferSize + InitialRecordBuffer |
| Header and Comment Specification | CommentRune + CommentLines + IncludeByteOrderMarker + Headers + TrimHeaders|
| Format Specification | CommentRune + Escape + FieldSeparator + Quote + RecordSeparator + NumFields |
| Data Loss Prevention | ClearFreedDataMemory |
| Encoding Validation | ErrorOnNonUTF8 |
| Security Limits | *planned* |

Note that the writer also has WriteFieldRow*() functions (WriteFieldRow, WriteFieldRowBorrowed) to reduce allocations when converting non‑string types to human‑readable CSV field values via the FieldWriter generating functions under csv.FieldWriters().

Note that after a number of columns, the WriteFieldRow*() calls flush less efficiently given they can leak to the heap and the cost of staging the non-serialized forms in a slice of wide-structs can add up quickly. To address this case, a fluent API has been added to the csv.Writer instance which can be utilized per some record to write via .NewRecord() which returns a RecordWriter instance. In a single-threaded fashion it locks the writer until Write() or Rollback() is called. Each field can be buffered for writing via the "FieldTypeName()" functions on the RecordWriter instance. Only one RecordWriter instance can be alive at a time for any given Writer.

Performance testing should be utilized to choose which writing methodology is ideal for your case.
In general choose the method most sympathetic to your hardware and data formats. For most cases, csv.Writer.NewRecord() should achieve a nice balance that scales very high in terms of both utility and efficiency.

---

[CHANGELOG](docs/version/v3/CHANGELOG.md)

---

[![Go Reference](https://pkg.go.dev/badge/github.com/josephcopenhaver/csv-go/v3.svg)](https://pkg.go.dev/github.com/josephcopenhaver/csv-go/v3)

---

Here's the same example as above adjusted to optimize throughput via additional configurations.

```go
package main

// this is a toy example that reads a csv file and writes to another without making allocations while processing

import (
	"bufio"
	"os"

	"github.com/josephcopenhaver/csv-go/v3"
)

func main() {
	r, err := os.Open("input.csv")
	if err != nil {
		panic(err)
	}
	defer r.Close()

	// using a buffered reader to avoid hot io pipes / writing less than the system storage device block size or ideal network protocol packet payload size
	// could instead use something async powered to get concurrent behaviors
	br := bufio.NewReader(r)

	var cr csv.Reader
	{
		op := csv.ReaderOpts()
		cr, err = csv.NewReader(
			op.Reader(br),
			op.RecordSeparator("\n"), // simplifies the execution plan ever so slightly and ensures consistent parsing rather than depending on automatic discovery
			op.InitialRecordBufferSize(4*1024*1024), // seeds the reading record buffer to a particular initial capacity
			op.BorrowRow(true),                      // evades allocations BUT makes it unsafe to store/use the resulting slice past the next call to Scan
			op.BorrowFields(true),                   // evades allocations BUT makes it unsafe to store/use the resulting character content of each slice element result anywhere past the next call to Scan
			op.NumFields(2),                         // simplifies the execution plan ever so slightly
			// by default quotes have no meaning
			// so must be specified to match RFC 4180
			// op.Quote('"'),
		)
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := cr.Close(); err != nil {
				panic(err)
			}
		}()
	}

	w, err := os.Create("output.csv")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := w.Close(); err != nil {
			panic(err)
		}
	}()

	// using a buffered writer to avoid hot io pipes / writing less than the system storage device block size or ideal network protocol packet payload size
	// could instead use something async powered to get concurrent behaviors
	bw := bufio.NewWriterSize(w, 4*1024*1024)
	defer func() {
		if err := bw.Flush(); err != nil {
			panic(err)
		}
	}()

	var cw *csv.Writer
	{
		op := csv.WriterOpts()
		cw, err = csv.NewWriter(
			op.Writer(bw),
			op.InitialRecordBufferSize(4*1024*1024), // seeds the writing record buffer to a particular initial capacity
		)
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := cw.Close(); err != nil {
				panic(err)
			}
		}()
	}

	// using Scan instead of the iterator sugar to avoid allocation of the iterator closures
	for cr.Scan() {
		// if BorrowRow=true or BorrowFields=true then implementation reading rows from the Reader MUST NOT keep the rows or byte sub-slices alive beyond the next call to cr.Scan()
		rw, err := cw.NewRecord()
		if err != nil {
			// note if you are just going to panic, consider calling MustNewRecord()
			// instead of the NewRecord() function
			panic(err)
		}
		for _, s := range cr.Row() {
			rw.String(s)
		}
		if _, err := rw.Write(); err != nil {
			panic(err)
		}
	}
	if err := cr.Err(); err != nil {
		panic(err)
	}
}
```
