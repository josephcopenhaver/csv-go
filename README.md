# csv-go

csv-go

[![Go Report Card](https://goreportcard.com/badge/github.com/josephcopenhaver/csv-go)](https://goreportcard.com/report/github.com/josephcopenhaver/csv-go/v2)
![tests](https://github.com/josephcopenhaver/csv-go/actions/workflows/tests.yaml/badge.svg)
![code-coverage](https://img.shields.io/badge/code_coverage-100%25-rgb%2852%2C208%2C88%29)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

This package is a highly flexible and performant csv stream reader and writer that opts for strictness and nearly all options off by default. By using the option functions pattern on Reader and Writer creation extreme flexibility can be offered and configuration can be validated up-front. This creates an immutable, clear execution of the strategy.

That being said, the reader is also more performant at the moment than the standard go csv package when compared in an apples-to-apples configuration between the two.

```go
package main

import (
	"os"

	"github.com/josephcopenhaver/csv-go/v2"
)

func main() {
	r, err := os.Open("input.csv")
	if err != nil {
		panic(err)
	}
	defer r.Close()

	cr, err := csv.NewReader(
		csv.ReaderOpts().Reader(r),
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

See the [Reader](./internal/examples/reader/main.go) and [Writer](./internal/examples/writer/main.go) examples for more in-depth usages.

## Reader Features

| Name | option(s) |
| - | - |
| Zero allocations | BorrowRow + BorrowFields + InitialRecordBuffer + InitialRecordBufferSize |
| Format Specification | Comment + CommentsAllowedAfterStartOfRecords + Escape + FieldSeparator + Quote + RecordSeparator + NumFields |
| Format Discovery | DiscoverRecordSeparator |
| Data Loss Prevention | ClearFreedDataMemory |
| Byte Order Marker Support | RemoveByteOrderMarker + ErrorOnNoByteOrderMarker
| Headers Support | ExpectHeaders + RemoveHeaderRow + TrimHeaders |
| Reader Buffer tuning | ReaderBuffer + ReaderBufferSize |
| Format Validation | ErrorOnNoRows + ErrorOnNewlineInUnquotedField + ErrorOnQuotesInUnquotedField |
| Security Limits | *planned* |

## Writer Features

| Name | option(s) |
| - | - |
| Zero allocations | *planned* |
| Header and Comment Specification | CommentRune + CommentLines + IncludeByteOrderMarker + Headers + TrimHeaders|
| Format Specification | Escape + FieldSeparator + Quote + RecordSeparator + NumFields |
| Data Loss Prevention | ClearFreedDataMemory |
| Encoding Validation | ErrorOnNonUTF8 |
| Security Limits | *planned* |

---

[CHANGELOG](./docs/version/v2/CHANGELOG.md)

---

[![Go Reference](https://pkg.go.dev/badge/github.com/josephcopenhaver/csv-go/v2.svg)](https://pkg.go.dev/github.com/josephcopenhaver/csv-go/v2)
