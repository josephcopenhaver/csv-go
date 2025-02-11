package csv_test

import (
	"strings"

	"github.com/josephcopenhaver/csv-go"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Functional CSV Reader iterator Paths", func() {

	Context("when creating a CSV reader with two records", func() {
		It("should support breaking the iterator for loop without losing state", func() {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("a,b,c\n1,2,3")),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(cr).ToNot(BeNil())

			var row []string
			for v := range cr.IntoIter() {
				row = v
				break
			}

			x := cr.Row()
			Expect(row).ToNot(Equal([]string{"a", "b", "d"}))
			Expect(strings.Join(row, ",")).To(Equal("a,b,c"))
			Expect(row).To(Equal(x))

			Expect(cr.Scan()).To(BeTrueBecause("there should be one more row to read"))
			Expect(len(cr.Row())).To(Equal(3))
			Expect(strings.Join(cr.Row(), ",")).To(Equal("1,2,3"))

			Expect(cr.Scan()).To(BeFalseBecause("there should be no more rows to read"))
			Expect(cr.Err()).To(BeNil())
			// close has not been called so rows should be non-nil
			Expect(len(cr.Row())).To(Equal(3))
			Expect(strings.Join(cr.Row(), ",")).To(Equal("1,2,3"))

			// and now it should be nil after Close is called
			Expect(cr.Close()).To(BeNil())
			Expect(cr.Row()).To(BeNil())
		})
	})

	Context("when creating a CSV reader with two records", func() {
		It("should support breaking the iterator for loop without losing state and converting into an iterator again", func() {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("a,b,c\n1,2,3")),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(cr).ToNot(BeNil())

			var row []string
			for v := range cr.IntoIter() {
				row = v
				break
			}

			x := cr.Row()
			Expect(row).ToNot(Equal([]string{"a", "b", "d"}))
			Expect(len(row)).To(Equal(3))
			Expect(strings.Join(row, ",")).To(Equal("a,b,c"))
			Expect(row).To(Equal(x))

			var numIters int
			for v := range cr.IntoIter() {
				numIters += 1
				row = v
			}
			Expect(numIters).To(Equal(1))

			Expect(cr.Scan()).To(BeFalseBecause("there should not be any more rows to load"))
			Expect(len(row)).To(Equal(3))
			Expect(strings.Join(row, ",")).To(Equal("1,2,3"))

			Expect(cr.Scan()).To(BeFalseBecause("there should be no more rows to read"))
			Expect(cr.Err()).To(BeNil())
			// close has not been called so rows should be non-nil
			Expect(len(cr.Row())).To(Equal(3))
			Expect(strings.Join(cr.Row(), ",")).To(Equal("1,2,3"))

			// and now it should be nil after Close is called
			Expect(cr.Close()).To(BeNil())
			Expect(cr.Row()).To(BeNil())
		})
	})
})
