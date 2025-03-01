package csv_test

import (
	"strings"
	"testing"

	"github.com/josephcopenhaver/csv-go"
	"github.com/stretchr/testify/assert"
)

func TestFunctionalReaderInitializationPaths(t *testing.T) {

	t.Run("when creating a csv reader and using the same value for quote and escape", func(t *testing.T) {
		t.Run("should not error", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('"'),
			)
			assert.Nil(t, err)
			assert.NotNil(t, cr)

			for row := range cr.IntoIter() {
				_ = row
			}

			assert.Nil(t, cr.Err())
			assert.Nil(t, cr.Close())
		})
	})

	t.Run("when creating a csv reader and specifying headers", func(t *testing.T) {
		t.Run("should not error", func(t *testing.T) {
			header := "a,b,c"
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader(header)),
				csv.ReaderOpts().ExpectHeaders(strings.Split(header, ",")),
			)
			assert.Nil(t, err)
			assert.NotNil(t, cr)

			for row := range cr.IntoIter() {
				_ = row
			}

			assert.Nil(t, cr.Err())
			assert.Nil(t, cr.Close())
		})
	})
}
