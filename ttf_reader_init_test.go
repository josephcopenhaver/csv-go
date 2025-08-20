package csv_test

import (
	"strings"
	"testing"

	"github.com/josephcopenhaver/csv-go/v2"
	"github.com/stretchr/testify/assert"
)

func TestFunctionalReaderInitializationPaths(t *testing.T) {
	t.Parallel()

	t.Run("when creating a csv reader and using the same value for quote and escape", func(t *testing.T) {
		t.Run("should not error", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('"'),
			)
			assert.Nil(t, err)
			assert.NotNil(t, cr)
			assert.Nil(t, cr.Row())

			for row := range cr.IntoIter() {
				_ = row
			}

			assert.Nil(t, cr.Err())
			assert.Nil(t, cr.Close())
			assert.Nil(t, cr.Row())
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
			assert.Nil(t, cr.Row())

			for row := range cr.IntoIter() {
				_ = row
			}

			assert.Nil(t, cr.Err())
			assert.Nil(t, cr.Close())
			assert.Nil(t, cr.Row())
		})
	})

	t.Run("when creating a csv reader and specifying reader buffer", func(t *testing.T) {
		t.Run("should not error", func(t *testing.T) {
			buf := [csv.ReaderMinBufferSize]byte{}
			header := "a,b,c"
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader(header)),
				csv.ReaderOpts().ReaderBuffer(buf[:]),
			)
			assert.Nil(t, err)
			assert.NotNil(t, cr)
			assert.Nil(t, cr.Row())

			for row := range cr.IntoIter() {
				_ = row
			}

			assert.Nil(t, cr.Err())
			assert.Nil(t, cr.Close())
			assert.Nil(t, cr.Row())
		})
	})

	t.Run("when creating a csv reader and specifying reader buffer size", func(t *testing.T) {
		t.Run("should not error", func(t *testing.T) {
			header := "a,b,c"
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader(header)),
				csv.ReaderOpts().ReaderBufferSize(csv.ReaderMinBufferSize),
			)
			assert.Nil(t, err)
			assert.NotNil(t, cr)
			assert.Nil(t, cr.Row())

			for row := range cr.IntoIter() {
				_ = row
			}

			assert.Nil(t, cr.Err())
			assert.Nil(t, cr.Close())
			assert.Nil(t, cr.Row())
		})
	})

	t.Run("when creating a csv reader and row borrowing with no field borrowing implicitly", func(t *testing.T) {
		t.Run("should not error and un-cloned extracted first row values should remain unchanged", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("a,b\n1,2")),
				csv.ReaderOpts().BorrowRow(true),
			)
			assert.Nil(t, err)
			assert.NotNil(t, cr)
			assert.Nil(t, cr.Row())

			firstRow := true
			var f1, f2 string
			for row := range cr.IntoIter() {
				if firstRow {
					firstRow = false
					f1 = row[0]
					f2 = row[1]
				}
			}

			assert.Nil(t, cr.Err())
			assert.Nil(t, cr.Close())
			assert.Nil(t, cr.Row())

			assert.Equal(t, "a", f1)
			assert.Equal(t, "b", f2)
		})
	})

	t.Run("when creating a csv reader and row borrowing with no field borrowing implicitly and NumFields set", func(t *testing.T) {
		t.Run("should not error and un-cloned extracted first row values should remain unchanged", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("a,b\n1,2")),
				csv.ReaderOpts().BorrowRow(true),
				csv.ReaderOpts().NumFields(2),
			)
			assert.Nil(t, err)
			assert.NotNil(t, cr)
			assert.Nil(t, cr.Row())

			firstRow := true
			var f1, f2 string
			for row := range cr.IntoIter() {
				if firstRow {
					firstRow = false
					f1 = row[0]
					f2 = row[1]
				}
			}

			assert.Nil(t, cr.Err())
			assert.Nil(t, cr.Close())
			assert.Nil(t, cr.Row())

			assert.Equal(t, "a", f1)
			assert.Equal(t, "b", f2)
		})
	})

	t.Run("when creating a csv reader and row borrowing with no field borrowing explicitly", func(t *testing.T) {
		t.Run("should not error and un-cloned extracted first row values should remain unchanged", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("a,b\n1,2")),
				csv.ReaderOpts().BorrowRow(true),
				csv.ReaderOpts().BorrowFields(false),
			)
			assert.Nil(t, err)
			assert.NotNil(t, cr)
			assert.Nil(t, cr.Row())

			firstRow := true
			var f1, f2 string
			for row := range cr.IntoIter() {
				if firstRow {
					firstRow = false
					f1 = row[0]
					f2 = row[1]
				}
			}

			assert.Nil(t, cr.Err())
			assert.Nil(t, cr.Close())
			assert.Nil(t, cr.Row())

			assert.Equal(t, "a", f1)
			assert.Equal(t, "b", f2)
		})
	})

	t.Run("when creating a csv reader and row borrowing with no field borrowing explicitly and NumFields set", func(t *testing.T) {
		t.Run("should not error and un-cloned extracted first row values should remain unchanged", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("a,b\n1,2")),
				csv.ReaderOpts().BorrowRow(true),
				csv.ReaderOpts().BorrowFields(false),
				csv.ReaderOpts().NumFields(2),
			)
			assert.Nil(t, err)
			assert.NotNil(t, cr)
			assert.Nil(t, cr.Row())

			firstRow := true
			var f1, f2 string
			for row := range cr.IntoIter() {
				if firstRow {
					firstRow = false
					f1 = row[0]
					f2 = row[1]
				}
			}

			assert.Nil(t, cr.Err())
			assert.Nil(t, cr.Close())
			assert.Nil(t, cr.Row())

			assert.Equal(t, "a", f1)
			assert.Equal(t, "b", f2)
		})
	})

	t.Run("when creating a csv reader and row borrowing with field borrowing", func(t *testing.T) {
		t.Run("should not error and un-cloned extracted first row values should change", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("a,b\n1,2")),
				csv.ReaderOpts().BorrowRow(true),
				csv.ReaderOpts().BorrowFields(true),
			)
			assert.Nil(t, err)
			assert.NotNil(t, cr)
			assert.Nil(t, cr.Row())

			firstRow := true
			var f1, f2 string
			for row := range cr.IntoIter() {
				if firstRow {
					firstRow = false
					f1 = row[0]
					f2 = row[1]
				}
			}

			assert.Nil(t, cr.Err())
			assert.Nil(t, cr.Close())
			assert.Nil(t, cr.Row())

			assert.Equal(t, "1", f1)
			assert.Equal(t, "2", f2)
		})
	})

	t.Run("when creating a csv reader and row borrowing with field borrowing and NumFields set", func(t *testing.T) {
		t.Run("should not error and un-cloned extracted first row values should change", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("a,b\n1,2")),
				csv.ReaderOpts().BorrowRow(true),
				csv.ReaderOpts().BorrowFields(true),
				csv.ReaderOpts().NumFields(2),
			)
			assert.Nil(t, err)
			assert.NotNil(t, cr)
			assert.Nil(t, cr.Row())

			firstRow := true
			var f1, f2 string
			for row := range cr.IntoIter() {
				if firstRow {
					firstRow = false
					f1 = row[0]
					f2 = row[1]
				}
			}

			assert.Nil(t, cr.Err())
			assert.Nil(t, cr.Close())
			assert.Nil(t, cr.Row())

			assert.Equal(t, "1", f1)
			assert.Equal(t, "2", f2)
		})
	})

	t.Run("when creating a csv reader and clearing freed data mem and row borrowing with field borrowing", func(t *testing.T) {
		t.Run("should not error and un-cloned extracted first row values should zero out", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("a,b\n1,2")),
				csv.ReaderOpts().BorrowRow(true),
				csv.ReaderOpts().BorrowFields(true),
				csv.ReaderOpts().ClearFreedDataMemory(true),
			)
			assert.Nil(t, err)
			assert.NotNil(t, cr)
			assert.Nil(t, cr.Row())

			firstRow := true
			var f1, f2 string
			for row := range cr.IntoIter() {
				if firstRow {
					firstRow = false
					f1 = row[0]
					f2 = row[1]
				}
			}

			assert.Nil(t, cr.Err())
			assert.Nil(t, cr.Close())
			assert.Nil(t, cr.Row())

			assert.Equal(t, "\x00", f1)
			assert.Equal(t, "\x00", f2)
		})
	})

	t.Run("when creating a csv reader and clearing freed data mem and row borrowing with field borrowing and NumFields set", func(t *testing.T) {
		t.Run("should not error and un-cloned extracted first row values should zero out", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("a,b\n1,2")),
				csv.ReaderOpts().BorrowRow(true),
				csv.ReaderOpts().BorrowFields(true),
				csv.ReaderOpts().ClearFreedDataMemory(true),
				csv.ReaderOpts().NumFields(2),
			)
			assert.Nil(t, err)
			assert.NotNil(t, cr)
			assert.Nil(t, cr.Row())

			firstRow := true
			var f1, f2 string
			for row := range cr.IntoIter() {
				if firstRow {
					firstRow = false
					f1 = row[0]
					f2 = row[1]
				}
			}

			assert.Nil(t, cr.Err())
			assert.Nil(t, cr.Close())
			assert.Nil(t, cr.Row())

			assert.Equal(t, "\x00", f1)
			assert.Equal(t, "\x00", f2)
		})
	})

	t.Run("when creating a csv reader and clearing freed data mem and row borrowing with field borrowing and NumFields set and record row col 1 is empty", func(t *testing.T) {
		t.Run("should not error and un-cloned extracted first row values should zero out", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("a,b\n,2")),
				csv.ReaderOpts().BorrowRow(true),
				csv.ReaderOpts().BorrowFields(true),
				csv.ReaderOpts().ClearFreedDataMemory(true),
				csv.ReaderOpts().NumFields(2),
			)
			assert.Nil(t, err)
			assert.NotNil(t, cr)
			assert.Nil(t, cr.Row())

			firstRow := true
			var f1, f2 string
			for row := range cr.IntoIter() {
				if firstRow {
					firstRow = false
					f1 = row[0]
					f2 = row[1]
				}
			}

			assert.Nil(t, cr.Err())
			assert.Nil(t, cr.Close())
			assert.Nil(t, cr.Row())

			assert.Equal(t, "\x00", f1)
			assert.Equal(t, "\x00", f2)
		})
	})

	t.Run("when creating a csv reader with MaxField=2 NumField=1", func(t *testing.T) {
		t.Run("should not error on init", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().MaxFields(2),
				csv.ReaderOpts().NumFields(1),
			)
			assert.Nil(t, err)
			assert.NotNil(t, cr)
			assert.Nil(t, cr.Row())
		})
	})
}
