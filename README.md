# csv-go

csv-go

[![Go Report Card](https://goreportcard.com/badge/github.com/josephcopenhaver/csv-go)](https://goreportcard.com/report/github.com/josephcopenhaver/csv-go/v3)
![tests](https://github.com/josephcopenhaver/csv-go/actions/workflows/tests.yaml/badge.svg)
![code-coverage](https://img.shields.io/badge/code_coverage-100%25-rgb%2852%2C208%2C88%29)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

This package is a highly flexible and performant single threaded csv stream reader and writer. It opts for strictness with nearly all options off by default. Using the option functions pattern on Reader and Writer creation ensures extreme flexibility can be offered while configuration can be validated up-front in cold paths. This creates an immutable, clear execution of the csv file/stream parsing strategy. It has been battle tested thoroughly in production contexts for both correctness and speed so feel free to use in any way you like.

The reader is also [more performant than the standard go csv package](docs/BENCHMARKS.md) when compared in an apples-to-apples configuration between the two. The writer also has several optimizations for non-string type serialization via FieldWriters(). I expect mileage here to vary over time. My primary goal with this lib was to solve my own edge case problems like suspect-encodings/loose-rules and offer something back more aligned with others that think like myself with regard to reducing allocations, GC pause, and increasing efficiency.

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
	defer w.Close()

	cw, err := csv.NewWriter(
		csv.WriterOpts().Writer(w),
	)
	if err != nil {
		panic(err)
	}
	defer cw.Close()

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

---

[CHANGELOG](docs/version/v3/CHANGELOG.md)

---

[![Go Reference](https://pkg.go.dev/badge/github.com/josephcopenhaver/csv-go/v3.svg)](https://pkg.go.dev/github.com/josephcopenhaver/csv-go/v3)
