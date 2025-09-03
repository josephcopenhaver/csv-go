package csv_test

import (
	"bytes"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/josephcopenhaver/csv-go/v3"
	"github.com/stretchr/testify/assert"
)

type functionalReaderTestCase struct {
	when, then                string
	selfInit                  func(*functionalReaderTestCase)
	newOpts                   []csv.ReaderOption
	newOptsF                  func() []csv.ReaderOption
	newReaderErrAs            []any
	newReaderErrAsNot         []any
	newReaderErrIs            []error
	newReaderErrIsNot         []error
	newReaderErrStr           string
	newReaderErrStrMsgAndArgs []any
	iterErrAs                 []any
	iterErrAsNot              []any
	iterErrIs                 []error
	iterErrIsNot              []error
	iterErrStr                string
	iterErrStrMsgAndArgs      []any
	hadRowsMsgAndArgs         []any
	rows                      [][]string
	forEachRow                func(*testing.T, []string)
	afterTest                 func(*testing.T)
	numRows                   int
	hasNewReaderErr           bool
	hasIterErr                bool
	hadRows                   bool
	skipScan                  bool
}

func (tc *functionalReaderTestCase) Run(t *testing.T) {
	assert.NotEmpty(t, tc.when)
	t.Helper()

	f := func(options ...func(*functionalReaderTestCase)) func(t *testing.T) {
		tc_copy := *tc
		for _, f := range options {
			f(&tc_copy)
		}
		if f := tc_copy.selfInit; f != nil {
			f(&tc_copy)
		}

		return func(t *testing.T) {
			tc := &tc_copy
			is := assert.New(t)

			var cr csv.Reader
			{
				opts := tc.newOpts
				if f := tc.newOptsF; f != nil {
					opts = append(slices.Clone(f()), opts...)
				}

				v, err := csv.NewReader(opts...)
				if tc.hasNewReaderErr || len(tc.newReaderErrIs) > 0 || len(tc.newReaderErrAs) > 0 || len(tc.newReaderErrIsNot) > 0 || len(tc.newReaderErrAsNot) > 0 || tc.newReaderErrStr != "" {
					is.NotNil(t, err)
					is.Nil(t, v)

					for _, terr := range tc.newReaderErrIs {
						is.ErrorIs(err, terr)
					}

					for i := range tc.newReaderErrAs {
						v := tc.newReaderErrAs[i]
						is.True(errors.As(err, &v))
					}

					for _, target := range tc.newReaderErrIsNot {
						is.False(errors.Is(err, target))
					}

					for i := range tc.newReaderErrAsNot {
						v := tc.newReaderErrAsNot[i]
						is.False(errors.As(err, &v))
					}

					if tc.newReaderErrStr != "" && err != nil {
						is.Equal(tc.newReaderErrStr, err.Error(), tc.newReaderErrStrMsgAndArgs...)
					}

					if tc.afterTest != nil {
						tc.afterTest(t)
					}

					return
				}

				is.Nil(err)
				is.NotNil(v)

				if v == nil {
					return
				}

				cr = v
			}

			// until scan is called, should return nil
			// rows and nil error
			is.Nil(cr.Row())
			is.Nil(cr.Err())

			var hadRows bool
			numRows := -1
			if !tc.skipScan {
				for cr.Scan() {
					hadRows = true
					numRows += 1
					is.Less(numRows, len(tc.rows))
					if numRows >= len(tc.rows) {
						break
					}
					expRow := tc.rows[numRows]
					row := cr.Row()
					if expRow == nil {
						is.Nil(row)
					} else {
						is.True(slices.Equal(expRow, row), "row %d did not meet expectations", numRows)
					}
					if tc.forEachRow != nil {
						tc.forEachRow(t, row)
					}
				}
			}
			numRows += 1

			err := cr.Err()
			if tc.hasIterErr || len(tc.iterErrIs) > 0 || len(tc.iterErrIsNot) > 0 || len(tc.iterErrAs) > 0 || len(tc.iterErrAsNot) > 0 || tc.iterErrStr != "" {
				is.NotNil(err)

				for _, terr := range tc.iterErrIs {
					is.ErrorIs(err, terr)
				}

				for i := range tc.iterErrAs {
					v := tc.iterErrAs[i]
					is.True(errors.As(err, &v))
				}

				for _, target := range tc.iterErrIsNot {
					is.False(errors.Is(err, target))
				}

				for i := range tc.iterErrAsNot {
					v := tc.iterErrAsNot[i]
					is.False(errors.As(err, &v))
				}

				if tc.iterErrStr != "" && err != nil {
					is.Equal(tc.iterErrStr, err.Error(), tc.iterErrStrMsgAndArgs...)
				}
			} else {
				is.Nil(err)
			}

			// shorthand; rows were specified but counts and existence were not
			//
			// can infer what the count should be and they should exist
			if tc.numRows == 0 && !tc.hadRows && len(tc.rows) > 0 {
				tc.numRows = len(tc.rows)
				tc.hadRows = true
			}

			is.Equal(tc.hadRows, hadRows, tc.hadRowsMsgAndArgs...)
			is.Equal(tc.numRows, numRows)

			is.Nil(cr.Close())

			// once closed, Err should always return false
			is.Equal(csv.ErrReaderClosed, cr.Err())

			// once closed, Scan should always return false
			is.False(cr.Scan())

			// once closed, Row should always return nil
			is.Nil(cr.Row())

			if tc.afterTest != nil {
				tc.afterTest(t)
			}
		}
	}

	var name string
	if tc.then == "" {
		name = "then no error should occur"
	} else {
		name = "then " + tc.then
	}

	t.Run("when "+tc.when, func(t *testing.T) {
		t.Helper()

		t.Run(name, f())
	})

	t.Run("when clearmem+ and "+tc.when, func(t *testing.T) {
		t.Helper()

		t.Run(name, f(func(tc *functionalReaderTestCase) {
			v := slices.Clone(tc.newOpts)
			tc.newOpts = append(v, csv.ReaderOpts().ClearFreedDataMemory(true))
		}))
	})

	t.Run("when initRecBuffSize=4096 and "+tc.when, func(t *testing.T) {
		t.Helper()

		t.Run(name, f(func(tc *functionalReaderTestCase) {
			v := slices.Clone(tc.newOpts)
			tc.newOpts = append(v, csv.ReaderOpts().InitialRecordBufferSize(1024*4))
		}))
	})

	t.Run("when initRecBuff=[4096]byte and "+tc.when, func(t *testing.T) {
		t.Helper()

		t.Run(name, f(func(tc *functionalReaderTestCase) {
			v := slices.Clone(tc.newOpts)
			buf := make([]byte, 1024*4)
			tc.newOpts = append(v, csv.ReaderOpts().InitialRecordBuffer(buf))
		}))
	})
}

func TestFunctionalReaderOKPaths(t *testing.T) {
	t.Parallel()

	tcs := []functionalReaderTestCase{
		func() functionalReaderTestCase {
			selfInit := func(tc *functionalReaderTestCase) {
				var expAddress string
				forEachRow := func() func(*testing.T, []string) {
					var expLen int
					var expCap int
					var handler func(*testing.T, []string)
					handler = func(t *testing.T, row []string) {
						t.Helper()

						handler = func(t *testing.T, row []string) {
							t.Helper()

							actual := fmt.Sprintf("%p", row)
							assert.Equal(t, expAddress, actual)
							assert.Equal(t, expLen, len(row))
							assert.Equal(t, expCap, cap(row))
						}
						expAddress = fmt.Sprintf("%p", row)
						expLen = len(row)
						expCap = cap(row)
						assert.Equal(t, expCap, expLen)
					}
					return func(t *testing.T, row []string) {
						t.Helper()

						handler(t, row)
					}
				}()

				tc.forEachRow = forEachRow

				oldAfterTest := tc.afterTest
				tc.afterTest = func(t *testing.T) {
					t.Helper()

					if oldAfterTest != nil {
						oldAfterTest(t)
					}

					assert.NotEmpty(t, expAddress)
				}
			}
			return functionalReaderTestCase{
				when: "borrowing rows",
				then: "the same slice reference should be returned each iteration",
				newOptsF: func() []csv.ReaderOption {
					return []csv.ReaderOption{
						csv.ReaderOpts().Reader(strings.NewReader("a,b,c\n,,3")),
					}
				},
				newOpts: []csv.ReaderOption{
					csv.ReaderOpts().BorrowRow(true),
					csv.ReaderOpts().BorrowFields(true),
				},
				rows:     [][]string{strings.Split("a,b,c", ","), strings.Split(",,3", ",")},
				selfInit: selfInit,
			}
		}(),
		{
			when: "there are empty fields in a row",
			then: "rows should have empty fields",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(",,3")),
				}
			},
			rows: [][]string{strings.Split(",,3", ",")},
		},
		{
			when: "when record sep is CRLF",
			then: "rows should parse fine if document uses CRLF",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("a,b,c\r\n1,2,3")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r\n"),
			},
			rows: [][]string{strings.Split("a,b,c", ","), strings.Split("1,2,3", ",")},
		},
		{
			when: "when record sep is being discovered and first row ends in CR",
			then: "row should parse without error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("a,b\r")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().DiscoverRecordSeparator(true),
			},
			rows: [][]string{strings.Split("a,b", ",")},
		},
		{
			when: "reader is discovering the record sep and first row has one column and ends in CR+short-multibyte+EOF",
			then: "now one row should be returned and error should be raised to the .Err method",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(bytes.NewReader(append([]byte("a\r"), 0xC0))),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().DiscoverRecordSeparator(true),
			},
			rows: [][]string{{"a"}, {string([]byte{0xC0})}},
		},
		{
			when: "reader is discovering the record sep and it is CRLF",
			then: "now rows should be returned and error should be raised to the .Err method",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("a,b\r\nc,d")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().DiscoverRecordSeparator(true),
			},
			rows: [][]string{strings.Split("a,b", ","), strings.Split("c,d", ",")},
		},
		{
			when: "reader is discovering the record sep and it is LF",
			then: "now rows should be returned and error should be raised to the .Err method",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("a,b\nc,d")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().DiscoverRecordSeparator(true),
			},
			rows: [][]string{strings.Split("a,b", ","), strings.Split("c,d", ",")},
		},
		{
			when: "erroring on no rows, expecting headers, and document is not empty",
			then: "header row should be returned with data row",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("a,b,c\n1,2,3")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ErrorOnNoRows(true),
				csv.ReaderOpts().ExpectHeaders("a", "b", "c"),
			},
			rows: [][]string{strings.Split("a,b,c", ","), strings.Split("1,2,3", ",")},
		},
		{
			when: "erroring on no rows, removing the first header row, and document is not empty",
			then: "header row should not be returned",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("a,b,c\n1,2,3")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ErrorOnNoRows(true),
				csv.ReaderOpts().RemoveHeaderRow(true),
			},
			rows: [][]string{strings.Split("1,2,3", ",")},
		},
		{
			when: "erroring on no rows, whitespace-trimming the header row values, and document is not empty",
			then: "trimmed header row should be returned with data row",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(" a , b , c \n1,2,3")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ErrorOnNoRows(true),
				csv.ReaderOpts().TrimHeaders(true),
			},
			rows: [][]string{strings.Split("a,b,c", ","), strings.Split("1,2,3", ",")},
		},
		{
			when: "whitespace-trimming the header row values, one header is all whitespace, but expecting headers to begin and end with whitespace",
			then: "should return trimmed header row and data row",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(" a ,  , c \n1,2,3")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ExpectHeaders(" a ", "  ", " c "),
				csv.ReaderOpts().TrimHeaders(true),
			},
			rows: [][]string{strings.Split("a,,c", ","), strings.Split("1,2,3", ",")},
		},
		{
			when: "header expected where one value is empty",
			then: "should return header row and data row",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("a,,c\n1,2,3")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ExpectHeaders("a", "", "c"),
			},
			rows: [][]string{strings.Split("a,,c", ","), strings.Split("1,2,3", ",")},
		},
		{
			when: "terminal record separator emits record but there is not one with one column",
			then: "returns one row",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("a")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().TerminalRecordSeparatorEmitsRecord(true),
			},
			rows: [][]string{{"a"}},
		},
		{
			when: "terminal record separator emits record and there is one with one column",
			then: "returns two rows",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("a\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().TerminalRecordSeparatorEmitsRecord(true),
			},
			rows: [][]string{{"a"}, {""}},
		},
		{
			when: "terminal record separator emits record but there is not one with two columns",
			then: "returns one row",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("a,b")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().TerminalRecordSeparatorEmitsRecord(true),
			},
			rows: [][]string{{"a", "b"}},
		},
		{
			when: "terminal record separator emits record and there is one with two columns",
			then: "returns one row as it is only applicable to one column datasets",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("a,b\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().TerminalRecordSeparatorEmitsRecord(true),
			},
			rows: [][]string{{"a", "b"}},
		},
	}

	for _, tc := range tcs {
		tc.Run(t)
	}
}
