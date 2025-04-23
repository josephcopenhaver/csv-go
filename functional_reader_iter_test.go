package csv_test

import (
	"strings"
	"testing"

	"github.com/josephcopenhaver/csv-go/v2"
	"github.com/stretchr/testify/assert"
)

func TestFunctionalReaderIteratorPaths(t *testing.T) {

	t.Run("given a CSV reader with two records", func(t *testing.T) {
		newReader := func() (*csv.Reader, error) {
			return csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("a,b,c\n1,2,3")),
			)
		}

		t.Run("should support breaking the iterator for loop without losing state", func(t *testing.T) {
			cr, err := newReader()
			assert.Nil(t, err)
			assert.NotNil(t, cr)

			var row []string
			for v := range cr.IntoIter() {
				row = v
				break
			}

			x := cr.Row()
			assert.NotEqual(t, []string{"a", "b", ""}, row)
			assert.Equal(t, "a,b,c", strings.Join(row, ","))
			assert.Equal(t, x, row)

			assert.True(t, cr.Scan(), "there should be one more row to read")
			assert.Equal(t, 3, len(cr.Row()))
			assert.Equal(t, "1,2,3", strings.Join(cr.Row(), ","))

			assert.False(t, cr.Scan(), "there should be no more rows to read")
			assert.Nil(t, cr.Err())
			// close has not been called so rows should be non-nil
			assert.Equal(t, 3, len(cr.Row()))
			assert.Equal(t, "1,2,3", strings.Join(cr.Row(), ","))

			// and now it should be nil after Close is called
			assert.Nil(t, cr.Close())
			assert.Nil(t, cr.Row())
		})

		t.Run("should support breaking the iterator for loop without losing state and converting into an iterator again", func(t *testing.T) {
			cr, err := newReader()
			assert.Nil(t, err)
			assert.NotNil(t, cr)

			var row []string
			for v := range cr.IntoIter() {
				row = v
				break
			}

			x := cr.Row()
			assert.NotEqual(t, []string{"a", "b", ""}, row)
			assert.Equal(t, "a,b,c", strings.Join(row, ","))
			assert.Equal(t, x, row)

			var numIters int
			for v := range cr.IntoIter() {
				numIters += 1
				row = v
			}
			assert.Equal(t, 1, numIters)

			assert.False(t, cr.Scan(), "there should not be any more rows to load")
			assert.Equal(t, 3, len(row))
			assert.Equal(t, "1,2,3", strings.Join(cr.Row(), ","))

			assert.False(t, cr.Scan(), "there should not be any more rows to read")
			assert.Nil(t, cr.Err())
			// close has not been called so rows should be non-nil
			assert.Equal(t, 3, len(cr.Row()))
			assert.Equal(t, "1,2,3", strings.Join(cr.Row(), ","))

			// and now it should be nil after Close is called
			assert.Nil(t, cr.Close())
			assert.Nil(t, cr.Row())
		})
	})
}
