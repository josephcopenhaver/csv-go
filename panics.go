package csv

// panicErr ensures the source of a panic comes from this module and not some other one
// when values are checked in white-box unit tests
//
// there is now no chance of confusion
type panicErr uint8

const (
	_ panicErr = iota

	panicRecordSepRuneLen                    // "invalid record separator rune length"
	panicUnknownReaderStateDuringEOF         // "reader in unknown state when EOF encountered"
	panicMissedHandlingMaxRecordIndex        // "missed handling record index at max value"
	panicMissedHandlingMaxSecOpFieldIndex    // "missed handling field index at max SecOp value"
	panicMissedHandlingMaxExpectedFieldIndex // "missed handling field index at expected max configured value"
	panicInvalidByteSequenceLength           // "invalid byte sequence length"
)

func (p panicErr) String() string {
	return []string{
		"invalid record separator rune length",              // panicRecordSepRuneLen
		"reader in unknown state when EOF encountered",      // panicUnknownReaderStateDuringEOF
		"missed handling record index at max value",         // panicMissedHandlingMaxRecordIndex
		"missed handling field index at SecOp max value",    // panicMissedHandlingMaxSecOpFieldIndex
		"missed handling field index at expected max value", // panicMissedHandlingMaxExpectedFieldIndex
		"invalid byte sequence length",                      // panicInvalidByteSequenceLength
	}[p-1]
}

func (p panicErr) Error() string {
	return p.String()
}
