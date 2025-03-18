package csv

type rFlag uint16

const (
	stDone rFlag = 1 << iota
	stAfterSOR

	rFlagDropBOM
	rFlagErrOnNoBOM

	// rFlagStartOfDoc = either of rFlagDropBOM, rFlagErrOnNoBOM

	rFlagClearMemAfterUser
	rFlagErrOnNLInUF
	rFlagErrOnQInUF
	rFlagOneRuneRecSep
	rFlagTwoRuneRecSep
	rFlagComment
	rFlagQuote
	rFlagEscape
	rFlagCommentAfterSOR
	rFlagTRSEmitsRecord
)

const (
	cf_next_step    = 0
	cf_return_true  = 1
	cf_return_false = 2
	cf_fallthrough  = 3
	cf_continue     = 4
)
