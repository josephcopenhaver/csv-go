# V3.* Changes

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

### Significantly altered functions
- NewReader() (Reader, error)

---

## Moving to V3

- NewReader now returns an interface rather than a pointer to a concrete exported type to match the spirit of hiding the internals as much as possible. To ease transition the *Reader struct that was returned is now a Reader interface with the same exported function signature.
- ExpectHeaders now errors when a nil value is passed in. In addition it now takes a variadic slice of strings rather than a slice of strings.
