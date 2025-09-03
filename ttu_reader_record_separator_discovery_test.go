package csv

import (
	"bytes"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitReaderDiscoverRecordSeparator(t *testing.T) {
	t.Parallel()

	tcs := [][]string{
		{
			"start of record",
			`[[""],["a"]]`,
			"", "a",
		},
		{
			"end of quoted field",
			`[["a"]]`,
			"\"a\"",
		},
		{
			"line comment",
			`[["a"]]`,
			"# a", "a",
		},
		{
			"start of field",
			`[["","b"]]`,
			",b",
		},
		{
			"in field",
			`[["","234567"],["a","b"]]`,
			",234567", "a,b",
		},
	}

	recSepSeqs := [][]string{
		{"utf8Newline", string(rune(utf8NextLine))},
		{"CR", string(rune('\r'))},
		{"LF", string(rune('\n'))},
		{"CRLF", string("\r\n")},
	}
	type optsCollection struct {
		name string
		opts []ReaderOption
	}
	iterOpts := []optsCollection{
		{"with clearMem=false", []ReaderOption{ReaderOpts().ClearFreedDataMemory(false)}},
		{"with clearMem=true", []ReaderOption{ReaderOpts().ClearFreedDataMemory(true)}},
	}

	for _, iOpts := range iterOpts {
		for _, recSep := range recSepSeqs {
			for _, tc := range tcs {
				tc := slices.Clone(tc)
				name := tc[0]
				expJSON := tc[1]

				tc = tc[2:]
				for i := range tc {
					tc[i] = tc[i] + recSep[1]
				}

				t.Run(name+" encounter "+recSep[0]+" with RecSepDiscovery=true withReadBufSize=ReaderMinBufferSize "+iOpts.name, func(t *testing.T) {
					is := assert.New(t)

					cr, crv, err := internalNewReader(append([]ReaderOption{
						ReaderOpts().Reader(strings.NewReader(strings.Join(tc, ""))),
						ReaderOpts().Quote('"'),
						ReaderOpts().Escape('\\'),
						ReaderOpts().Comment('#'),
						ReaderOpts().DiscoverRecordSeparator(true),
						ReaderOpts().ReaderBufferSize(ReaderMinBufferSize),
					}, iOpts.opts...)...)
					is.Nil(err)
					is.NotNil(cr)
					is.NotNil(crv)

					var cri *fastReader
					if v, ok := crv.(*secOpReader); ok {
						cri = v.fastReader
					} else if v, ok := crv.(*fastReader); ok {
						cri = v
					} else {
						is.FailNow("unexpected reader type")
					}

					// verify control runes initialized as expected
					{
						allPossibleNLRunes := []rune{asciiCarriageReturn, asciiLineFeed, asciiVerticalTab, asciiFormFeed, utf8NextLine, utf8LineSeparator}

						expControlRunes := string(cri.fieldSeparator) + string(cri.quote) + string(cri.escape) + string(cri.comment)
						for _, r := range allPossibleNLRunes {
							if !bytes.ContainsRune([]byte(expControlRunes), r) {
								expControlRunes += string(r)
							}
						}

						is.Equal(expControlRunes, cri.controlRunes)
					}

					is.Nil(cr.Row())

					rows := [][]string{}
					for row := range cr.IntoIter() {
						rows = append(rows, row)
					}
					is.Nil(cr.Err())

					b, err := json.Marshal(rows)
					is.Nil(err)
					is.NotNil(b)

					encodedVal := strings.ReplaceAll(strings.ReplaceAll(string(b), " ", ""), "\n", "")
					is.Equal(expJSON, encodedVal)

					// verify control runes changed to a subset as expected
					{
						expControlRunes := string(cri.fieldSeparator)
						if strings.HasPrefix(recSep[1], "\r") {
							expControlRunes += "\r"
						} else {
							expControlRunes += string(recSep[1])
						}
						expControlRunes += string(cri.quote) + string(cri.escape) + string(cri.comment)
						if !bytes.ContainsRune([]byte(expControlRunes), asciiCarriageReturn) {
							expControlRunes += string(rune(asciiCarriageReturn))
						}
						if !bytes.ContainsRune([]byte(expControlRunes), asciiLineFeed) {
							expControlRunes += string(rune(asciiLineFeed))
						}

						is.Equal(expControlRunes, cri.controlRunes)
					}

					is.Nil(cr.Close())
					is.Nil(cr.Row())
				})
			}
		}
	}
}
