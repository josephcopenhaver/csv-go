package csv

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_pow10u64(t *testing.T) {
	t.Parallel()

	is := assert.New(t)

	v := uint64(1)
	is.Equal(v, pow10u64[0])
	is.Equal(1, decLenU64(v-1))
	is.Equal(1, decLenU64(v))

	for i := 1; i < len(pow10u64); i++ {
		v *= 10
		is.Equal(v, pow10u64[i])

		is.Equal(i, decLenU64(v-1))
		is.Equal(i+1, decLenU64(v))
	}
}

func TestUsageOf_decLenU64(t *testing.T) {
	t.Parallel()

	is := assert.New(t)

	is.Equal(int(19), decLenU64(math.MaxInt64))
	is.Equal(int(20), decLenU64(math.MaxUint64))

	// prove the process for negatives
	{
		minInt64 := int64(math.MinInt64)
		overflowedMinInt64AsUint64 := uint64(minInt64)

		// this process is a false positive, do not replicate this block
		// instead use the next block style!
		//
		// is.Equal(overflowedMinInt64AsUint64, uint64(1<<63))
		// is.Equal(int(19), decLenU64(overflowedMinInt64AsUint64))

		// The above is a false positive, we need to use this process below
		// for negatives then add one to the result to account for the sign

		is.Equal(int(1), decLenU64(uint64(-int64(-1))))
		is.Equal(int(9), decLenU64(uint64(-int64(-111111111))))
		is.Equal(int(19), decLenU64(uint64(-int64(-1111111111111111111))))
		is.Equal(int(19), decLenU64(uint64(-int64(overflowedMinInt64AsUint64))))
		is.Equal(int(18), decLenU64(uint64(-int64(-999999999999999999))))
	}
}
