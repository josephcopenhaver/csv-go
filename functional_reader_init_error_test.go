package csv_test

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/josephcopenhaver/csv-go"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Functional CSV Reader Initialization Error Paths", func() {

	Context("when creating a CSV reader without a reader", func() {
		It("should return an error indicating the reader option is nil and a nil csv reader instance", func() {
			reader, err := csv.NewReader()
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(csv.ErrBadConfig))
			Expect(err).To(MatchError(csv.ErrNilReader))
			Expect(reader).To(BeNil())
		})
	})

	Context("when creating a CSV reader and expecting an non-nil yet empty set of headers", func() {
		It("should return an error indicating option value is invalid", func() {
			reader, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().ExpectHeaders([]string{}),
			)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(csv.ErrBadConfig))
			Expect(err.Error()).To(Equal(errors.Join(csv.ErrBadConfig, errors.New("empty set of headers expected")).Error()))
			Expect(reader).To(BeNil())
		})
	})

	Context("when creating a CSV reader and expecting a header and field length that do not match", func() {
		It("should return an error indicating option value combinations are invalid", func() {
			{
				reader, err := csv.NewReader(
					csv.ReaderOpts().Reader(strings.NewReader("")),
					csv.ReaderOpts().ExpectHeaders([]string{"a"}),
					csv.ReaderOpts().NumFields(2),
				)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(csv.ErrBadConfig))
				Expect(err.Error()).To(Equal(errors.Join(csv.ErrBadConfig, errors.New("explicitly specified NumFields does not match length of ExpectHeaders list")).Error()))
				Expect(reader).To(BeNil())
			}
			{
				reader, err := csv.NewReader(
					csv.ReaderOpts().Reader(strings.NewReader("")),
					csv.ReaderOpts().ExpectHeaders([]string{"a", "b"}),
					csv.ReaderOpts().NumFields(1),
				)
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(csv.ErrBadConfig))
				Expect(err.Error()).To(Equal(errors.Join(csv.ErrBadConfig, errors.New("explicitly specified NumFields does not match length of ExpectHeaders list")).Error()))
				Expect(reader).To(BeNil())
			}
		})
	})

	Context("when creating a CSV reader and enabling record separator discovery and specifying one explicitly", func() {
		It("should return an error indicating option value combinations are invalid", func() {
			reader, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().DiscoverRecordSeparator(true),
				csv.ReaderOpts().RecordSeparator("\n"),
			)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(csv.ErrBadConfig))
			Expect(err.Error()).To(Equal(errors.Join(csv.ErrBadConfig, errors.New("must specify one and only one of automatic record separator discovery or a specific recordSeparator")).Error()))
			Expect(reader).To(BeNil())
		})
	})

	Context("when creating a CSV reader and specifying an empty record separator", func() {
		It("should return an error indicating option value is invalid", func() {
			reader, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().RecordSeparator(""),
			)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(csv.ErrBadConfig))
			Expect(err.Error()).To(Equal(errors.Join(csv.ErrBadConfig, errors.New(`record separator can only be one valid utf8 rune long or "\r\n"`)).Error()))
			Expect(reader).To(BeNil())
		})
	})

	Context("when creating a CSV reader with no explicit record separator", func() {
		type config struct {
			OptionName string
			Option     func(rune) csv.ReaderOption
		}
		permutations := []config{
			{OptionName: "quote", Option: csv.ReaderOpts().Quote},
			{OptionName: "field separator", Option: csv.ReaderOpts().FieldSeparator},
			{OptionName: "comment", Option: csv.ReaderOpts().Comment},
			{OptionName: "escape", Option: csv.ReaderOpts().Escape},
		}
		for _, p := range permutations {
			Context("when specifying "+p.OptionName+" of \\n", func() {
				It("should return an error indicating option value is invalid", func() {
					reader, err := csv.NewReader(
						csv.ReaderOpts().Reader(strings.NewReader("")),
						p.Option('\n'),
					)
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(csv.ErrBadConfig))
					Expect(err.Error()).To(Equal(errors.Join(csv.ErrBadConfig, errors.New(`invalid record separator and `+p.OptionName+` combination`)).Error()))
					Expect(reader).To(BeNil())
				})
			})
		}
	})

	Context("when creating a CSV reader with an explicit record separator of '\r\n'", func() {
		type config struct {
			Val        rune
			ValName    string
			OptionName string
			Option     func(rune) csv.ReaderOption
		}
		permutations := []config{
			{Val: '\r', ValName: "\\r", OptionName: "quote", Option: csv.ReaderOpts().Quote},
			{Val: '\n', ValName: "\\n", OptionName: "quote", Option: csv.ReaderOpts().Quote},
			{Val: '\r', ValName: "\\r", OptionName: "field separator", Option: csv.ReaderOpts().FieldSeparator},
			{Val: '\n', ValName: "\\n", OptionName: "field separator", Option: csv.ReaderOpts().FieldSeparator},
			{Val: '\r', ValName: "\\r", OptionName: "comment", Option: csv.ReaderOpts().Comment},
			{Val: '\n', ValName: "\\n", OptionName: "comment", Option: csv.ReaderOpts().Comment},
			{Val: '\r', ValName: "\\r", OptionName: "escape", Option: csv.ReaderOpts().Escape},
			{Val: '\n', ValName: "\\n", OptionName: "escape", Option: csv.ReaderOpts().Escape},
		}
		for _, p := range permutations {
			Context("when specifying "+p.OptionName+" of "+p.ValName, func() {
				It("should return an error indicating option value is invalid", func() {
					reader, err := csv.NewReader(
						csv.ReaderOpts().Reader(strings.NewReader("")),
						csv.ReaderOpts().RecordSeparator("\r\n"),
						p.Option(p.Val),
					)
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(csv.ErrBadConfig))
					Expect(err.Error()).To(Equal(errors.Join(csv.ErrBadConfig, errors.New(`invalid record separator and `+p.OptionName+` combination`)).Error()))
					Expect(reader).To(BeNil())
				})
			})
		}
	})

	Context("when creating a CSV reader with only a reader", func() {
		type config struct {
			OptionName string
			Option     csv.ReaderOption
			ErrMsg     string
		}
		permutations := []config{
			{OptionName: "quote", Option: csv.ReaderOpts().Quote(utf8.RuneError)},
			{OptionName: "field separator", Option: csv.ReaderOpts().FieldSeparator(utf8.RuneError)},
			{OptionName: "comment", Option: csv.ReaderOpts().Comment(utf8.RuneError)},
			{OptionName: "escape", Option: csv.ReaderOpts().Escape(utf8.RuneError)},
			{OptionName: "record separator", Option: csv.ReaderOpts().RecordSeparator(string(utf8.RuneError)), ErrMsg: "record separator can only be one valid utf8 rune long or \"\\r\\n\""},
		}
		for _, p := range permutations {
			Context("when specifying a \\uFFFD rune "+p.OptionName, func() {
				It("should return an error indicating option value is invalid", func() {
					reader, err := csv.NewReader(
						csv.ReaderOpts().Reader(strings.NewReader("")),
						p.Option,
					)
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError(csv.ErrBadConfig))
					errMsg := p.ErrMsg
					if errMsg == "" {
						errMsg = `invalid ` + p.OptionName + ` value`
					}
					Expect(err.Error()).To(Equal(errors.Join(csv.ErrBadConfig, errors.New(errMsg)).Error()))
					Expect(reader).To(BeNil())
				})
			})
		}
	})

	Context("when creating a CSV reader and specifying the same valid rune for comment and quote", func() {
		It("should return an error indicating option value is invalid", func() {
			reader, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().Quote('#'),
			)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(csv.ErrBadConfig))
			Expect(err.Error()).To(Equal(errors.Join(csv.ErrBadConfig, errors.New(`invalid comment and quote combination`)).Error()))
			Expect(reader).To(BeNil())
		})
	})

	Context("when creating a CSV reader and specifying the same valid rune for field separator and quote", func() {
		It("should return an error indicating option value is invalid", func() {
			reader, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().FieldSeparator('#'),
				csv.ReaderOpts().Quote('#'),
			)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(csv.ErrBadConfig))
			Expect(err.Error()).To(Equal(errors.Join(csv.ErrBadConfig, errors.New(`invalid field separator and quote combination`)).Error()))
			Expect(reader).To(BeNil())
		})
	})

	Context("when creating a CSV reader and specifying the same valid rune for comment and escape", func() {
		It("should return an error indicating option value is invalid", func() {
			reader, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().Escape('#'),
			)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(csv.ErrBadConfig))
			Expect(err.Error()).To(Equal(errors.Join(csv.ErrBadConfig, errors.New(`invalid comment and escape combination`)).Error()))
			Expect(reader).To(BeNil())
		})
	})

	Context("when creating a CSV reader and specifying escape but no quote", func() {
		It("should return an error indicating option value is invalid", func() {
			reader, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().Escape('#'),
			)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(csv.ErrBadConfig))
			Expect(err.Error()).To(Equal(errors.Join(csv.ErrBadConfig, errors.New(`escape can only be used when quoting is enabled`)).Error()))
			Expect(reader).To(BeNil())
		})
	})

	Context("when creating a CSV reader and specifying a zero NumFields value", func() {
		It("should return an error indicating option value is invalid", func() {
			reader, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().NumFields(0),
			)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(csv.ErrBadConfig))
			Expect(err.Error()).To(Equal(errors.Join(csv.ErrBadConfig, errors.New(`num fields must be greater than zero or not specified`)).Error()))
			Expect(reader).To(BeNil())
		})
	})

	Context("when creating a CSV reader and specifying a negative NumFields value", func() {
		It("should return an error indicating option value is invalid", func() {
			reader, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().NumFields(-1),
			)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(csv.ErrBadConfig))
			Expect(err.Error()).To(Equal(errors.Join(csv.ErrBadConfig, errors.New(`num fields must be greater than zero or not specified`)).Error()))
			Expect(reader).To(BeNil())
		})
	})
})

var _ = Describe("CSV Reader Initialization success Paths", func() {

	Context("when creating a csv reader and using the same value for quote and escape", func() {
		It("should not error", func() {
			reader, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('"'),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(reader).ToNot(BeNil())

			for row := range reader.IntoIter() {
				_ = row
			}

			err = reader.Err()
			Expect(err).ToNot(HaveOccurred())

			err = reader.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("when creating a csv reader and specifying headers", func() {
		It("should not error", func() {
			header := "a,b,c"
			reader, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader(header)),
				csv.ReaderOpts().ExpectHeaders(strings.Split(header, ",")),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(reader).ToNot(BeNil())

			for row := range reader.IntoIter() {
				_ = row
			}

			err = reader.Err()
			Expect(err).ToNot(HaveOccurred())

			err = reader.Close()
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
