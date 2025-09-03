package csv_test

import (
	"errors"
	"strconv"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/josephcopenhaver/csv-go/v3"
	"github.com/stretchr/testify/assert"
)

func TestFunctionalReaderInitializationErrorPaths(t *testing.T) {
	t.Parallel()

	t.Run("when creating a CSV reader without a reader", func(t *testing.T) {
		t.Run("should return an error indicating the reader option is nil and a nil csv reader instance", func(t *testing.T) {
			cr, err := csv.NewReader()
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.ErrorIs(t, err, csv.ErrNilReader)
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader and expecting an non-nil yet empty set of headers", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().ExpectHeaders([]string{}...),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New("empty set of headers expected")).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader and expecting a nil set of headers", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().ExpectHeaders([]string(nil)...),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New("empty set of headers expected")).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader and expecting a header and field length that do not match", func(t *testing.T) {
		t.Run("should return an error indicating option value combinations are invalid", func(t *testing.T) {
			{
				cr, err := csv.NewReader(
					csv.ReaderOpts().Reader(strings.NewReader("")),
					csv.ReaderOpts().ExpectHeaders("a"),
					csv.ReaderOpts().NumFields(2),
				)
				assert.NotNil(t, err)
				assert.ErrorIs(t, err, csv.ErrBadConfig)
				assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New("explicitly specified NumFields does not match length of ExpectHeaders list")).Error(), err.Error())
				assert.Nil(t, cr)
			}
			{
				cr, err := csv.NewReader(
					csv.ReaderOpts().Reader(strings.NewReader("")),
					csv.ReaderOpts().ExpectHeaders("a", "b"),
					csv.ReaderOpts().NumFields(1),
				)
				assert.NotNil(t, err)
				assert.ErrorIs(t, err, csv.ErrBadConfig)
				assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New("explicitly specified NumFields does not match length of ExpectHeaders list")).Error(), err.Error())
				assert.Nil(t, cr)
			}
		})
	})

	t.Run("when creating a CSV reader and enabling record separator discovery and specifying one explicitly", func(t *testing.T) {
		t.Run("should return an error indicating option value combinations are invalid", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().DiscoverRecordSeparator(true),
				csv.ReaderOpts().RecordSeparator("\n"),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New("must specify one and only one of automatic record separator discovery or a specific recordSeparator")).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader and specifying an empty record separator", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().RecordSeparator(""),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`record separator can only be one valid utf8 rune long or "\r\n"`)).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader with no explicit record separator", func(t *testing.T) {
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
			t.Run("when specifying "+p.OptionName+" of \\n", func(t *testing.T) {
				t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
					cr, err := csv.NewReader(
						csv.ReaderOpts().Reader(strings.NewReader("")),
						p.Option('\n'),
					)
					assert.NotNil(t, err)
					assert.ErrorIs(t, err, csv.ErrBadConfig)
					assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`invalid record separator and `+p.OptionName+` combination`)).Error(), err.Error())
					assert.Nil(t, cr)
				})
			})
		}
	})

	t.Run("when creating a CSV reader with an explicit record separator of '\r\n'", func(t *testing.T) {
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
			t.Run("when specifying "+p.OptionName+" of "+p.ValName, func(t *testing.T) {
				t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
					cr, err := csv.NewReader(
						csv.ReaderOpts().Reader(strings.NewReader("")),
						csv.ReaderOpts().RecordSeparator("\r\n"),
						p.Option(p.Val),
					)
					assert.NotNil(t, err)
					assert.ErrorIs(t, err, csv.ErrBadConfig)
					assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`invalid record separator and `+p.OptionName+` combination`)).Error(), err.Error())
					assert.Nil(t, cr)
				})
			})
		}
	})

	t.Run("when creating a CSV reader with record separator discovery enabled", func(t *testing.T) {
		type config struct {
			Val        rune
			ValName    string
			OptionName string
			Option     func(rune) csv.ReaderOption
		}
		permutations := []config{
			{Val: '\r', ValName: "\\r", OptionName: "quote", Option: csv.ReaderOpts().Quote},
			{Val: '\r', ValName: "\\r", OptionName: "comment", Option: csv.ReaderOpts().Comment},
			{Val: '\r', ValName: "\\r", OptionName: "field separator", Option: csv.ReaderOpts().FieldSeparator},
			{Val: '\r', ValName: "\\r", OptionName: "escape", Option: csv.ReaderOpts().Escape},
		}
		for _, p := range permutations {
			t.Run("when specifying "+p.OptionName+" of "+p.ValName, func(t *testing.T) {
				t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
					/*
						This is desired behavior because eof signaled-one processing implementation is coupled to this expectation at this time.

						Also how on earth would discovery work should this not be the case? Should we assume that the runes assigned to the other concerns explicitly should be reserved for those singular other purposes? How about the default field separator value of ','?

						Seems like the only way to win this game is not to play.
					*/
					cr, err := csv.NewReader(
						csv.ReaderOpts().Reader(strings.NewReader("")),
						csv.ReaderOpts().DiscoverRecordSeparator(true),
						p.Option(p.Val),
					)
					assert.NotNil(t, err)
					assert.ErrorIs(t, err, csv.ErrBadConfig)
					assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(p.OptionName+` cannot be a discoverable newline character when record separator discovery is enabled`)).Error(), err.Error())
					assert.Nil(t, cr)
				})
			})
		}
	})

	t.Run("when creating a CSV reader with only a reader", func(t *testing.T) {
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
			t.Run("when specifying a \\uFFFD rune "+p.OptionName, func(t *testing.T) {
				t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
					cr, err := csv.NewReader(
						csv.ReaderOpts().Reader(strings.NewReader("")),
						p.Option,
					)
					assert.NotNil(t, err)
					assert.ErrorIs(t, err, csv.ErrBadConfig)
					errMsg := p.ErrMsg
					if errMsg == "" {
						errMsg = `invalid ` + p.OptionName + ` value`
					}
					assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(errMsg)).Error(), err.Error())
					assert.Nil(t, cr)
				})
			})
		}
	})

	t.Run("when creating a CSV reader and specifying the same valid rune for comment and quote", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().Quote('#'),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`invalid comment and quote combination`)).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader and specifying the same valid rune for field separator and quote", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().FieldSeparator('#'),
				csv.ReaderOpts().Quote('#'),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`invalid field separator and quote combination`)).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader and specifying the same valid rune for comment and escape", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().Escape('#'),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`invalid comment and escape combination`)).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader and specifying escape but no quote", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().Escape('#'),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`escape can only be used when quoting is enabled`)).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader and specifying a zero NumFields value", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().NumFields(0),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`num fields must be greater than zero or not specified`)).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader and specifying a negative NumFields value", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().NumFields(-1),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`num fields must be greater than zero or not specified`)).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader and specifying a buffer of negative length", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().InitialRecordBufferSize(-1),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`initial record buffer size must be greater than or equal to zero`)).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader and specifying both a buffer length and a buffer", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().InitialRecordBufferSize(1024*4),
				csv.ReaderOpts().InitialRecordBuffer(nil),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`initial record buffer size cannot be specified when also setting the initial record buffer`)).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader and specifying a reader buffer length and a reader buffer", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			buf := [csv.ReaderMinBufferSize]byte{}
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().ReaderBufferSize(len(buf)),
				csv.ReaderOpts().ReaderBuffer(buf[:]),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`cannot specify both ReaderBuffer and ReaderBufferSize`)).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader and specifying a reader buffer length that is too small", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			bufLen := 6
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().ReaderBufferSize(bufLen),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`ReaderBufferSize must be greater than or equal to 7`)).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader and specifying a reader buffer that is too small", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			buf := [csv.ReaderMinBufferSize - 1]byte{}
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().ReaderBuffer(buf[:]),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`ReaderBuffer must have a length greater than or equal to `+strconv.Itoa(csv.ReaderMinBufferSize))).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader and specifying field borrowing is enabled with row borrowing implicitly disabled", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().BorrowFields(true),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`field borrowing cannot be enabled without enabling row borrowing`)).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader and specifying field borrowing is enabled with row borrowing explicitly disabled", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().BorrowFields(true),
				csv.ReaderOpts().BorrowRow(false),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`field borrowing cannot be enabled without enabling row borrowing`)).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader with maxFields=2 and headersLength=3", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().MaxFields(2),
				csv.ReaderOpts().ExpectHeaders("1", "2", "3"),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`max fields should not be specified or should be larger: max fields was specified with a value less than the specified number of fields per record`)).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader with maxFields=2 and NumFields=3", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().MaxFields(2),
				csv.ReaderOpts().NumFields(3),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`max fields should not be specified or should be larger: max fields was specified with a value less than the specified number of fields per record`)).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader with maxFields=1", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().MaxFields(1),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`max fields cannot be set to a value less than or equal to one`)).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader with MaxRecordBytes=0", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().MaxRecordBytes(0),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`max record bytes cannot be less than or equal to zero`)).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader with MaxRecords=0", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().MaxRecords(0),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`max records cannot be equal to zero`)).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader with MaxComments=-1", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().MaxComments(-1),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`max comments cannot be less than zero`)).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})

	t.Run("when creating a CSV reader with MaxCommentBytes=-1", func(t *testing.T) {
		t.Run("should return an error indicating option value is invalid", func(t *testing.T) {
			cr, err := csv.NewReader(
				csv.ReaderOpts().Reader(strings.NewReader("")),
				csv.ReaderOpts().MaxCommentBytes(-1),
			)
			assert.NotNil(t, err)
			assert.ErrorIs(t, err, csv.ErrBadConfig)
			assert.Equal(t, errors.Join(csv.ErrBadConfig, errors.New(`max comment bytes cannot be less than zero`)).Error(), err.Error())
			assert.Nil(t, cr)
		})
	})
}
