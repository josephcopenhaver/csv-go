// fast_csv_rune_set.go
//
// This file defined compact rune sets (runeSet4 / runeSet6) and extremely
// fast scanning routines used by the CSV Reader and Writer to locate
// structural runes in input data.
//
// To ensure implementation between runeSet4 and runeSet6 stay in sync they are
// now generated via the template in ./internal/cmd/generate/fast_csv_rune_set.go.tmpl
//
// Key ideas:
//
//   - singleBytes is a 256-bit table of "which single-byte runes are interesting"
//     so ASCII membership checks are just bit tests.
//
//   - mbByteEnds is a 64-bit table of "which final bytes of multi-byte UTF-8
//     sequences are interesting". We use that to cheaply guess (when we've just
//     reached the end of a multi-byte rune) that we care about it before decoding it.
//
//   - We track lastMBStartIdx as well as the rune byte-encoded length and only decode
//     when (a) the byte index is the end byte for the last valid start byte, and
//     (b) the current byte could be an interesting end byte.
//
// Invariants:
//
//   - runeSet4 is only ever populated with <=4 multi-byte runes. Writer config
//     enforces this. runeSet6 is <=6 for Reader config.
//     Code assumes these caps and will panic on misuse instead of bounds checking.
//
// =====
//
// internal* functions
// -------------------
// internal* funcs are only for use inside this file (and its white-box tests).
// They rely on preconditions enforced by this scanning logic and skip checks
// for speed that would normally be there to guard against bad usage by external
// callers.
//
// UTF-8 scanning strategy
// -----------------------
// When multi-byte runes are in play, indexAny* will:
//   1. Check if the current byte could be the final byte of a tracked rune
//      (fast bitset match on mbByteEnds).
//   2. validate and extract the rune encoded byte length from the start byte.
//   3. Check that continuation bytes exist contiguously
//   4. Check that the end byte is an interesting end-byte value to decode fully.
//   5. Only decode the rune once.
//   6. Confirm that:
//        - decoding succeeded,
//        - and that the decoded rune is in the set.

package csv

import (
	"unicode/utf8"
)

const (
	// contMBMask is the mask applied to a byte to see if it is a continuation byte in a utf8 encoded rune
	contMBMask = 0xC0
	// contMBVal is the value that indicates a byte is a continuation byte after the bit mask contMBMask has been applied
	// should it equal this constant.
	contMBVal = 0x80

	// startMBMin marks the start of a multi-byte utf8 encoded rune should a byte be greater than or equal to it.
	startMBMin = 0xC0

	startMB2ByteMin = 0xC2
	startMBMax      = 0xF4

	// invalidMBStartIdx is a safe value to use that indicates that there is no known valid start of utf8 multi-byte encoded rune
	//
	// It is safe because we set the lastMBStartIdx value to it and check that the distance between the current index and that value are
	// less than utf8.UTFMax before attempting a full rune decode. (we switch on distances 1-3)
	invalidMBStartIdx = -utf8.UTFMax
)

//
// NOTICE: runeSet4 and runeSet6 implementation is now generated and inflated via a template into gen_strategies.go
//
// ---
//
// A bit more about the above. Yes it was possible to implement the logic using generics and deeply reduce the duplications
// - it was also possible to avoid generics and implement using different data access strategies with reused top-level logic
// - even further it was also possible to just implement one runeSet6 and use that everywhere rather than make a tighter runeSet4.
//
// each code reduction tactic came at the expense of speed which makes complete sense given golang is NOT a zero-cost
// abstraction language and the compiler is VERY optimized to work with SIMPLE code at the moment rather than all the
// other more DRY tactics. Code-gen is currently still king in golang (2025-11-05) when aiming for speed.
//
// With a little bit of testing on top of the duplications we can create not just full coverage - but coverage that
// ensures parity between all implementations. It is also kinda moot to prove since we're using code-gen as the
// strategy between the specific methods CANNOT vary unless the code-gen fails or contains a bug. The latter is
// the main reason why there are additional tests over the entire runeSet. Just because it is generated does not mean
// we can skip tests - in fact it should be very much the opposite.
