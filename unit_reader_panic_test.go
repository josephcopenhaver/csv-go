package csv

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitReaderPanicOnValidate(t *testing.T) {
	//
	// when the config's record separator is in a corrupted / non possible value
	//
	// then a panic should occur when config is validated
	//

	is := assert.New(t)

	cfg := rCfg{
		reader:       strings.NewReader(""),
		recordSepLen: -2,
	}

	validate := func() (_ error, r any) {
		defer func() {
			r = recover()
		}()
		return cfg.validate(), nil
	}

	err, r := validate()
	is.NotNil(r)
	is.Nil(err)

	is.Equal(r, panicRecordSepLen)

	cfg = rCfg{
		reader:       strings.NewReader(""),
		recordSepLen: 3,
	}

	err, r = validate()
	is.NotNil(r)
	is.Nil(err)

	is.Equal(r, panicRecordSepLen)
}

func TestUnitReaderPanicOnHandleEOF(t *testing.T) {
	//
	// when reader is in a corrupted unknown state
	//
	// a panic should occur if handleEOF is called
	//

	is := assert.New(t)

	cr, err := NewReader(
		ReaderOpts().Reader(strings.NewReader("")),
	)
	is.Nil(err)
	is.NotNil(cr)
	cr.state = rState(rStateInLineComment + 1)

	handleEOF := func() (_ bool, r any) {
		defer func() {
			r = recover()
		}()
		return cr.handleEOF(), nil
	}

	resp, r := handleEOF()
	is.NotNil(r)
	is.False(resp)

	is.Equal(r, panicUnknownReaderStateDuringEOF)
}
