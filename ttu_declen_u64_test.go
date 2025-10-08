package csv

import (
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
