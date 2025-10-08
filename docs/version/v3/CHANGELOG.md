# V3.* Changes[^1]

## v3.0.2 - 2025-10-08

Rune writing now behaves in a simpler/faster fashion while keeping all other behaviors the same.

## v3.0.1 - 2025-10-07

### New Functions
- ReaderOptions.InitialRecordBufferSize()
- ReaderOptions.InitialRecordBuffer()
- ReaderOptions.InitialFieldBufferSize()
- ReaderOptions.InitialFieldBuffer()
- (*Writer).WriteFieldRow(row ...FieldWriter) (int, error)
- (*Writer).WriteFieldRowBorrowed(row []FieldWriter) (int, error)
- FieldWriters()
- FieldWriterFactory.Bytes()
- FieldWriterFactory.String()
- FieldWriterFactory.Int()
- FieldWriterFactory.Int64()
- FieldWriterFactory.Uint64()
- FieldWriterFactory.Time()
- FieldWriterFactory.Rune()
- FieldWriterFactory.Bool()
- FieldWriterFactory.Duration()
- FieldWriterFactory.Float64()

### New Constants
- ErrWriteHeaderFailed
- ErrInvalidFieldWriter

### NewStructs
- FieldWriterFactory
- FieldWriter

This update adds allocation functions and structures which authors can
use to speed up write operations. Specifically two new functions have
been added to the csv Writer type: WriteFieldRow and WriteFieldRowBorrowed.
Both these functions take a slice of FieldWriter instances which can be
obtained via FieldWriters().SomeType(valueOfTypeToSerialize). The Borrowed
variant can be used to reuse a slice from the parent context over and over
to avoid some internal book-keeping should that be ideal for the developer.

ErrWriteHeaderFailed is a new error type joined with errors that occur
during the phase of writing a header should the writer state be negatively
impacted from the attempt and prevents any other attempts to write a header
or record.

---

## v3.0.0 - 2025-09-03

### Removed Structs
- Reader

### New Interfaces
- Reader

### New Functions
- ReaderOptions.MaxFields()
- ReaderOptions.MaxRecordBytes()
- ReaderOptions.MaxRecords()
- ReaderOptions.MaxComments()
- ReaderOptions.MaxCommentBytes()

### New Constants
- ErrSecOp
- ErrSecOpRecordByteCountAboveMax
- ErrSecOpFieldCountAboveMax
- ErrSecOpRecordCountAboveMax
- ErrSecOpCommentBytesAboveMax
- ErrSecOpCommentsAboveMax

### Breaking API changes
- NewReader now returns the Reader interface: `Reader (interface)` (previously `*Reader (struct pointer)`).
- ReaderOptions.ExpectHeaders() has changed signatures from `func([]string)` to `func(...string)` and no longer accepts a nil value.
- Minimum supported Go version: 1.25.

---

## Moving to V3

- NewReader now returns an interface rather than a pointer to a concrete exported type to match the spirit of hiding the internals as much as possible. To ease transition the *Reader struct that was returned is now a Reader interface with the same exported function signature.
- ExpectHeaders now errors when a nil value is passed in. In addition it now takes a variadic slice of strings rather than a slice of strings to better match other option styles.

[^1]: For V2.* changes [see here](/docs/version/v2/CHANGELOG.md)
