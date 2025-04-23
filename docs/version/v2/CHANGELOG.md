# V2.* Changes

## v2.0.0 - 2025-04-22

### Removed Interfaces
- BufferedReader

### Removed Constants
- ErrBadReadRuneImpl
- ErrInvalidEscapeInQuotedField
- ErrBadUnreadRuneImpl
- ErrBadReadByteImpl

### New Functions
- ReaderOptions.ReaderBufferSize()
- ReaderOptions.ReaderBuffer()

### New Constants
- ReaderMinBufferSize
- ErrInvalidEscSeqInQuotedField

---

# Limitations of V1

V1 was based on an experiment that started off focusing on utf8 compliance by leaning off the back of bufio and its reader implementation that supported reading one utf8 rune at a time, going back a byte when decoding failed, and falling back to reading a byte at a time. Because the implementation did not have assurances that utf8 would be a strong requirement, it did not offer support for erroring when character decoding failed, and ultimately the context switching to and from the reader and the state machine byte by byte was a clear bottleneck - it did not make sense to continue to maintain an implementation orders of magnitude less efficient than the stock csv parser. Luckily this same end result can be achieved by maintaining a read buffer slice and operating on sub-slices at a time. This did require that the more simple state machine expand from an operation for a byte encountered in a given reader state to an operation for a slice before a critical sequence encountered in a given reader state where that critical sequence and slice size may require different strategies of handling. This allows for hot paths to stay in one function context, copy operations to span much more than one byte, configuration to alter the sequences being matched in one highly efficient call at reader initialization time / record separator discovery time, increase branchless executions and reduce performance impacts related to execution strategy variation over time. It's much more consistent regardless of configuration options, state, and input stream.

## Moving to V2

- In v1 InitialRecordBuffer had a counter-to-best-practice behavior: the supplied slice would be used up to its cap rather than its length.
  - InitialRecordBuffer and ReaderBuffer use now will never exceed the len of the slice passed to them.
  - If you wish for the full capacity of the slice to be used and are not certain of the len alignment to the capacity of the slice then increase the slice length accordingly before passing to these functions.
- ReaderBufferSize and ReaderBuffer options now exist
  - They operate similarly to InitialRecordBufferSize and InitialRecordBuffer. But are specifically used to buffer the results of reading the specified Reader.
  - If using ReaderBuffer/ReaderBufferSize the slice passed in/created will be owned by the reader instance and never be reallocated / increased in size. So if this value is too small for the average record/field size throughput could be negatively impacted. Size this up as the performance tests you author for your case would suggest you should.
- Creating a Reader will result in a 4096 byte slice being allocated by default.
  - Counter this behavior by specifying a specific buffer to use or a smaller/larger reader buffer via ReaderBuffer or ReaderBufferSize respectively
  - Note that the reader buffer size/len must be at least csv.ReaderMinBufferSize (7) bytes long otherwise an error will be thrown on reader initialization
  - The above limitation exists to ensure correct utf8 character alignment and CRLF sequence validation.
- The functionality of csv.ReaderOpts().BorrowRow() has been split into BorrowRow and BorrowFields.
  - BorrowRow no longer borrows both string fields and the row when true. However BorrowRow still enables reuse of the returned row slice between calls to Row() of a Reader.
  - When setting BorrowFields(false) or not specifying it then the strings in the slice are cloned before the Row() call returns, granting ownership of the strings to the calling context to do with them as they wish. As a micro optimization, to avoid this behavior and get unsafe strings that must be cloned (before saved or sub-sliced and saved with a lifetime beyond the next call to Scan) then set BorrowFields(true).
  - Avoiding clones reduces allocations and should be considered an unsafe micro-optimization for the case of Fields. For Row it is a more socially acceptable optimization as long as the slice does not persist in any form by the next call to Scan.
