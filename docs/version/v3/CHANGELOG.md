# V3.* Changes[^1]

## v3.2.0 - 2025-10-19

The great long overdue Writer refactor.


Deprecated:
- (WriterOptions) InitialFieldBufferSize
- (WriterOptions) InitialFieldBuffer

The above methods are now deprecated and no longer have any effect on the writing strategy.


New Functions:
- (WriterOptions) CommentRune(r rune)
- (FieldWriterFactory) UncheckedUTF8Bytes(p []byte)
- (FieldWriterFactory) UncheckedUTF8String(p []byte)


It is now possible to specify the comment rune when constructing a writer regardless of whether or not
the WriteHeader method of the writer is called. This fixes a gap in deterministic parsing. The
CommentRune option used when calling WriteHeader has also been altered to behave in the exact same
fashion as this new option regardless of if the header writing operation has comment lines or not.
Previously if the comment rune was specified when writing a header and no comment lines existed then
the document could not be deterministically parsed by the reader with Comment set to the same rune.

This was confusing issue has been fixed.

Should it be specified during writer construction and when calling WriteHeader an error will now be
returned. It can only be specified when constructing the writer or calling WriteHeader - not both.


UncheckedUTF8* variants of the FieldWriter String and Byte methods are now offered to (in a very minor
capacity) speed up serialization operations should the author know for absolute certainty that the
values within them are already utf8 compliant and do not need to be checked when writing them out to
a utf8 csv document.


---

In this update the writer has been significantly refactored to increase the speed of writing documents
and ensure allocations are largely avoided.

## v3.1.1 - 2025-10-10

Internal allocation size bugfix. The FieldWriter types that are processed as signed integers
would for negative values allocate buffers exceeding the size of their values when writing via MarshalText.
The exceeding size would never be greater than 19 bytes.

This code path is not regularly traversed by csv processing so while no impact is expected - persons
who chose to use that code path for various tests or quick analysis will see less bytes allocated for their allocation actions.

## v3.1.0 - 2025-10-08

Changes in v3.0.1 and v3.0.2 are new additions and should have had the minor version bumped.
Going to fix forward by just crafting the new release.

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
