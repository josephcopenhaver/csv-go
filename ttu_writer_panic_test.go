package csv

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_intSumOverflowCheck(t *testing.T) {
	t.Parallel()

	is := assert.New(t)

	//
	// no panic paths
	//

	{
		assertNoPanic := func(a, b int) {
			is.NotPanics(func() { intSumOverflowCheck(a, b) })
		}

		{
			sum := int(0)
			term := 1
			sum += term

			assertNoPanic(sum, term)
		}

		{
			sum := int(0)
			term := math.MaxInt
			sum += term

			assertNoPanic(sum, term)
		}
	}

	//
	// panic paths
	//

	{
		assertPanics := func(a, b int) {
			is.Panics(func() {
				defer func() {
					r := recover()
					is.NotNil(r)
					is.Equal(panicIntOverflow, r)
					if r != nil {
						panic(r)
					}
				}()

				intSumOverflowCheck(a, b)
			})
		}

		{
			sum := math.MaxInt
			term := 1
			sum += term

			assertPanics(sum, term)
		}

		{
			sum := math.MaxInt
			term := math.MaxInt
			sum += term

			assertPanics(sum, term)
		}
	}
}
