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
