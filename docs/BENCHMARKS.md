## 📊 Benchmark Results

#### Metric Details
| Metric | Details |
|-|-|
| ns/op | less is better |
| B/op | (usually) less is better |
| allocs/op | (usually) less is better |
| Samples | (usually) more is better |

| Host Detail | Value |
|-|-|
| goos | darwin |
| goarch | arm64 |
| cpu | Apple M1 Max |


#### Types of Tests

- Post-Init - focuses on the details of the parsing method after initialization of the resources involved all the way to their deallocation

### ✅ Standard lib's Post-Init Benchmarks

| Link | Benchmark | ns/op | B/op | allocs/op | Samples |
|---|:---|---:|---:|---:|---:|
| [SOURCE](../bench_reader_test.go) | BenchmarkSTDReadPostInit256Rows | 26,573 | 16,208 | 522 | 42,980 |
| [SOURCE](../bench_reader_test.go) | BenchmarkSTDReadPostInit256RowsBorrowRow | 21,175 | 3,920 | 266 | 56,851 |
| [SOURCE](../bench_writer_test.go) | BenchmarkSTDWritePostInitSmallNegInts | 50.08 | 4 | 2 | 23,710,572 |
| [SOURCE](../bench_writer_test.go) | BenchmarkSTDWritePostInitLargeNegInts | 103.1 | 48 | 2 | 11,735,142 |
| [SOURCE](../bench_writer_test.go) | BenchmarkSTDWritePostInitStrings | 24.10 | 0 | 0 | 49,049,661 |

### 🚀 This lib's Post-Init Benchmarks

| Link | Benchmark | ns/op | B/op | allocs/op | Samples |
|---|:---|---:|---:|---:|---:|
| [SOURCE](../bench_reader_test.go) | BenchmarkReadPostInit256Rows | 21,047 | 16,128 | 521 | 56,418 |
| [SOURCE](../bench_reader_test.go) | BenchmarkReadPostInit256RowsBorrowRow | 15,415 | 3,824 | 264 | 78,158 |
| [SOURCE](../bench_reader_test.go) | BenchmarkReadPostInit256RowsBorrowRowBorrowFields | 12,639 | 144 | 7 | 92,061 |
| [SOURCE](../bench_reader_test.go) | BenchmarkReadPostInit256RowsBorrowRowBorrowFieldsReadBuf | 12,422 | 144 | 7 | 94,563 |
| [SOURCE](../bench_reader_test.go) | BenchmarkReadPostInit256RowsBorrowRowBorrowFieldsRecBuf | 12,645 | 128 | 5 | 95,444 |
| [SOURCE](../bench_reader_test.go) | BenchmarkReadPostInit256RowsBorrowRowBorrowFieldsReadBufRecBuf | 12,371 | 128 | 5 | 94,317 |
| [SOURCE](../bench_reader_test.go) | BenchmarkReadPostInit256RowsBorrowRowBorrowFieldsRecBufNumFields | 11,999 | 0 | 0 | 97,297 |
| [SOURCE](../bench_reader_test.go) | BenchmarkReadPostInit256RowsBorrowRowBorrowFieldsReadBufRecBufNumFields | 12,167 | 0 | 0 | 98,736 |
| [SOURCE](../bench_writer_test.go) | BenchmarkWritePostInitSmallNegInts | 35.97 | 0 | 0 | 32,497,794 |
| [SOURCE](../bench_writer_test.go) | BenchmarkWritePostInitLargeNegInts | 54.78 | 0 | 0 | 21,899,774 |
| [SOURCE](../bench_writer_test.go) | BenchmarkWritePostInitStrings | 22.24 | 0 | 0 | 53,430,211 |
