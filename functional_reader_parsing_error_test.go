package csv_test

import (
	"errors"
	"io"
	"strings"

	"github.com/josephcopenhaver/csv-go"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Functional CSV Reader Parsing Error Paths", func() {

	Context("when reader errors on no rows and file is zero length", func() {
		var cr *csv.Reader

		It("should initialize without error", func() {
			var err error
			cr, err = csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().ErrorOnNoRows(true),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(err).To(BeNil())
			Expect(cr).ToNot(BeNil())
		})

		It("should return a specific error when .Err() is called and contain no rows", func() {

			var hadRows bool
			for cr.Scan() {
				hadRows = true
			}

			err := cr.Err()
			Expect(err).ToNot(BeNil())
			Expect(err).To(MatchError(io.ErrUnexpectedEOF))
			Expect(err.Error()).To(Equal("parsing error at byte 0, record 1, field 1: no rows: unexpected EOF"))

			Expect(hadRows).To(BeFalseBecause("%s", "zero length file"))
		})

		It("should not error when .Close() is called", func() {
			Expect(cr.Close()).To(BeNil())
		})
	})

	Context("when reader errors on no rows and file is zero length", func() {
		var cr *csv.Reader

		It("should initialize without error", func() {
			var err error
			cr, err = csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().ErrorOnNoByteOrderMarker(true),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(err).To(BeNil())
			Expect(cr).ToNot(BeNil())
		})

		It("should return a specific error when .Err() is called and contain no rows", func() {

			var hadRows bool
			for cr.Scan() {
				hadRows = true
			}

			err := cr.Err()
			Expect(err).ToNot(BeNil())
			Expect(err).To(MatchError(io.ErrUnexpectedEOF))
			Expect(err.Error()).To(Equal("io error at byte 0, record 1, field 1: no byte order marker\nunexpected EOF"))

			Expect(hadRows).To(BeFalseBecause("%s", "zero length file"))
		})

		It("should not error when .Close() is called", func() {
			Expect(cr.Close()).To(BeNil())
		})
	})

	Context("when read file contains quotes in unquoted field", func() {
		var cr *csv.Reader

		It("should initialize without error", func() {
			var err error
			cr, err = csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("1,2\",3")),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().ErrorOnQuotesInUnquotedField(true),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(err).To(BeNil())
			Expect(cr).ToNot(BeNil())
		})

		It("should return a specific error when .Err() is called and contain no rows", func() {

			var hadRows bool
			for cr.Scan() {
				hadRows = true
			}

			err := cr.Err()
			Expect(err).ToNot(BeNil())
			Expect(err).To(MatchError(csv.ErrQuoteInUnquotedField))
			Expect(err.Error()).To(Equal("parsing error at byte 4, record 1, field 2: quote found in unquoted field"))

			Expect(hadRows).To(BeFalseBecause("%s", "errors before first row"))
		})

		It("should not error when .Close() is called", func() {
			Expect(cr.Close()).To(BeNil())
		})
	})

	Context("when reading from a row-borrow-unspecified reader before calling scan", func() {
		var cr *csv.Reader

		It("should initialize without error", func() {
			var err error
			cr, err = csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(err).To(BeNil())
			Expect(cr).ToNot(BeNil())
		})

		It("should return nil when calling .Row()", func() {

			row := cr.Row()
			Expect(row).To(BeNil())
		})

		It("should not error when .Close() is called", func() {
			Expect(cr.Close()).To(BeNil())
		})
	})

	Context("when reading from row-borrow-disabled reader before calling scan", func() {
		var cr *csv.Reader

		It("should initialize without error", func() {
			var err error
			cr, err = csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().BorrowRow(false),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(err).To(BeNil())
			Expect(cr).ToNot(BeNil())
		})

		It("should return nil when calling .Row()", func() {

			row := cr.Row()
			Expect(row).To(BeNil())
		})

		It("should not error when .Close() is called", func() {
			Expect(cr.Close()).To(BeNil())
		})
	})

	Context("when reading from row-borrow-enabled reader before calling scan", func() {
		var cr *csv.Reader

		It("should initialize without error", func() {
			var err error
			cr, err = csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().BorrowRow(true),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(err).To(BeNil())
			Expect(cr).ToNot(BeNil())
		})

		It("should return nil when calling .Row()", func() {

			row := cr.Row()
			Expect(row).To(BeNil())
		})

		It("should not error when .Close() is called", func() {
			Expect(cr.Close()).To(BeNil())
		})
	})

	Context("when reading from a one-row row-borrow-unspecified reader closed before calling Scan", func() {
		var cr *csv.Reader

		It("should initialize without error", func() {
			var err error
			cr, err = csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("1,2,3")),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(err).To(BeNil())
			Expect(cr).ToNot(BeNil())
		})

		It("should not error when .Close() is called and subsequent calls to Scan() return false", func() {
			Expect(cr.Close()).To(BeNil())

			Expect(cr.Scan()).To(BeFalseBecause("scan operation should short circuit: reader is closed"))

			Expect(cr.Row()).To(BeNil())
		})
	})

	Context("when reading from a one-row row-borrow-disabled reader closed before calling Scan", func() {
		var cr *csv.Reader

		It("should initialize without error", func() {
			var err error
			cr, err = csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("1,2,3")),
				csv.ReaderOpts().BorrowRow(false),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(err).To(BeNil())
			Expect(cr).ToNot(BeNil())
		})

		It("should not error when .Close() is called and subsequent calls to Scan() return false", func() {
			Expect(cr.Close()).To(BeNil())

			Expect(cr.Scan()).To(BeFalseBecause("scan operation should short circuit: reader is closed"))

			Expect(cr.Row()).To(BeNil())
		})
	})

	Context("when reading from a one-row row-borrow-enabled reader closed before calling Scan", func() {
		var cr *csv.Reader

		It("should initialize without error", func() {
			var err error
			cr, err = csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("1,2,3")),
				csv.ReaderOpts().BorrowRow(true),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(err).To(BeNil())
			Expect(cr).ToNot(BeNil())
		})

		It("should not error when .Close() is called and subsequent calls to Scan() return false", func() {
			Expect(cr.Close()).To(BeNil())

			Expect(cr.Scan()).To(BeFalseBecause("scan operation should short circuit: reader is closed"))

			Expect(cr.Row()).To(BeNil())
		})
	})

	Context("when reading from a two-row misaligned reader", func() {
		var cr *csv.Reader

		It("should initialize without error", func() {
			var err error
			cr, err = csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("1\n2,3")),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(err).To(BeNil())
			Expect(cr).ToNot(BeNil())
		})

		It("reads first row fine", func() {
			Expect(cr.Scan()).To(BeTrue())
			Expect(cr.Err()).To(BeNil())
			row := cr.Row()
			Expect(len(row)).To(Equal(1))
			Expect(row[0]).To(Equal("1"))
			Expect(cr.Err()).To(BeNil())
		})

		It("errors on second row", func() {
			Expect(cr.Scan()).To(BeFalse())
			err := cr.Err()
			Expect(err).ToNot(BeNil())
			Expect(errors.As(err, &csv.ErrTooManyFields{})).To(BeTrueBecause("should be too many fields on second row"))
			Expect(err).To(MatchError("parsing error at byte 4, record 2, field 1: more than 1 field(s) found in record"))
		})

		It("should not error when .Close() is called", func() {
			Expect(cr.Close()).To(BeNil())
		})
	})

	Context("when reading from a two-row misaligned reader with \"early\" eof", func() {
		var cr *csv.Reader

		It("should initialize without error", func() {
			var err error
			cr, err = csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("1,2\n3")),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(err).To(BeNil())
			Expect(cr).ToNot(BeNil())
		})

		It("reads first row fine", func() {
			Expect(cr.Scan()).To(BeTrue())
			Expect(cr.Err()).To(BeNil())
			row := cr.Row()
			Expect(len(row)).To(Equal(2))
			Expect(row[0]).To(Equal("1"))
			Expect(row[1]).To(Equal("2"))
			Expect(cr.Err()).To(BeNil())
		})

		It("errors on second row", func() {
			Expect(cr.Scan()).To(BeFalse())
			err := cr.Err()
			Expect(err).ToNot(BeNil())
			Expect(errors.As(err, &csv.ErrFieldCountMismatch{})).To(BeTrueBecause("should be too little fields on second row"))
			Expect(errors.Is(err, io.ErrUnexpectedEOF)).To(BeTrueBecause("should have more fields"))
			Expect(err).To(MatchError("parsing error at byte 5, record 2, field 1: expected 2 fields but found 1 instead\nunexpected EOF"))
		})

		It("should not error when .Close() is called", func() {
			Expect(cr.Close()).To(BeNil())
		})
	})

	Context("when reading from a one-row 2 field reader with NumFields(1) option", func() {
		var cr *csv.Reader

		It("should initialize without error", func() {
			var err error
			cr, err = csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("1,2")),
				csv.ReaderOpts().NumFields(1),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(err).To(BeNil())
			Expect(cr).ToNot(BeNil())
		})

		It("errors on first row", func() {
			Expect(cr.Scan()).To(BeFalse())
			err := cr.Err()
			Expect(err).ToNot(BeNil())
			Expect(errors.As(err, &csv.ErrTooManyFields{})).To(BeTrueBecause("too many fields on first row"))
			Expect(errors.Is(err, io.ErrUnexpectedEOF)).To(BeFalseBecause("EOF should not have been observed"))
			Expect(err).To(MatchError("parsing error at byte 2, record 1, field 1: more than 1 field(s) found in record"))
		})

		It("should not error when .Close() is called", func() {
			Expect(cr.Close()).To(BeNil())
		})
	})

	Context("when reading from a two-row 1,2 field reader with NumFields(2) option", func() {
		var cr *csv.Reader

		It("should initialize without error", func() {
			var err error
			cr, err = csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("1\n2,3")),
				csv.ReaderOpts().NumFields(2),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(err).To(BeNil())
			Expect(cr).ToNot(BeNil())
		})

		It("errors on first row", func() {
			Expect(cr.Scan()).To(BeFalse())
			err := cr.Err()
			Expect(err).ToNot(BeNil())
			Expect(errors.As(err, &csv.ErrFieldCountMismatch{})).To(BeTrueBecause("too few fields on first row"))
			Expect(errors.Is(err, io.ErrUnexpectedEOF)).To(BeFalseBecause("EOF should not have been observed"))
			Expect(err).To(MatchError("parsing error at byte 2, record 1, field 1: expected 2 fields but found 1 instead"))
		})

		It("should not error when .Close() is called", func() {
			Expect(cr.Close()).To(BeNil())
		})
	})
})
