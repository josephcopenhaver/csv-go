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
| [SOURCE](../bench_reader_test.go) | BenchmarkSTDReadPostInit256Rows | 26,518 | 16,208 | 522 | 45,469 |
| [SOURCE](../bench_reader_test.go) | BenchmarkSTDReadPostInit256RowsBorrowRow | 21,104 | 3,920 | 266 | 57,159 |
| [SOURCE](../bench_writer_test.go) | BenchmarkSTDWritePostInitSmallNegInts | 50.93 | 4 | 2 | 20,604,350 |
| [SOURCE](../bench_writer_test.go) | BenchmarkSTDWritePostInitLargeNegInts | 102.7 | 48 | 2 | 11,465,218 |
| [SOURCE](../bench_writer_test.go) | ⚠️ BenchmarkSTDWritePostInitStrings ⚠️ | 24.03 | 0 | 0 | 49,673,070 |

### 🚀 This lib's Post-Init Benchmarks

| Link | Benchmark | ns/op | B/op | allocs/op | Samples |
|---|:---|---:|---:|---:|---:|
| [SOURCE](../bench_reader_test.go) | BenchmarkReadPostInit256Rows | 23,544 | 16,128 | 521 | 50,815 |
| [SOURCE](../bench_reader_test.go) | BenchmarkReadPostInit256RowsBorrowRow | 17,741 | 3,840 | 265 | 67,639 |
| [SOURCE](../bench_reader_test.go) | BenchmarkReadPostInit256RowsBorrowRowBorrowFields | 14,518 | 160 | 8 | 82,717 |
| [SOURCE](../bench_reader_test.go) | BenchmarkReadPostInit256RowsBorrowRowBorrowFieldsReadBuf | 14,400 | 160 | 8 | 82,689 |
| [SOURCE](../bench_reader_test.go) | BenchmarkReadPostInit256RowsBorrowRowBorrowFieldsRecBuf | 14,457 | 144 | 6 | 80,601 |
| [SOURCE](../bench_reader_test.go) | BenchmarkReadPostInit256RowsBorrowRowBorrowFieldsReadBufRecBuf | 14,396 | 144 | 6 | 83,990 |
| [SOURCE](../bench_reader_test.go) | BenchmarkReadPostInit256RowsBorrowRowBorrowFieldsRecBufNumFields | 14,084 | 0 | 0 | 85,561 |
| [SOURCE](../bench_reader_test.go) | BenchmarkReadPostInit256RowsBorrowRowBorrowFieldsReadBufRecBufNumFields | 14,006 | 0 | 0 | 85,872 |
| [SOURCE](../bench_writer_test.go) | BenchmarkWritePostInitSmallNegInts | 36.57 | 0 | 0 | 32,581,101 |
| [SOURCE](../bench_writer_test.go) | BenchmarkWritePostInitLargeNegInts | 57.14 | 0 | 0 | 19,322,439 |
| [SOURCE](../bench_writer_test.go) | ⚠️ BenchmarkWritePostInitStrings ⚠️ | 38.72 | 0 | 0 | 30,218,833 |

⚠️ - Turns out BenchmarkWritePostInitStrings shows string handling is less efficient than standard csv writing at the moment. So if you are only using strings and you know the content is non-overlapping with record separator, quotes, etc - you may better be served by using only the standard SDK. This lib excels at writing documents made of mixed type content and may be optimized further in the future.
