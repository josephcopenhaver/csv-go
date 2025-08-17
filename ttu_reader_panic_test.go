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

	validate := func() (r any, _ error) {
		defer func() {
			r = recover()
		}()
		return nil, cfg.validate()
	}

	r, err := validate()
	is.NotNil(r)
	is.Nil(err)

	is.Equal(r, panicRecordSepLen)

	cfg = rCfg{
		reader:       strings.NewReader(""),
		recordSepLen: 3,
	}

	r, err = validate()
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

	cr, crv, err := internalNewReader(
		ReaderOpts().Reader(strings.NewReader("")),
	)
	is.Nil(err)
	is.NotNil(cr)
	is.NotNil(crv)

	cri, ok := crv.(*fastReader)
	if !ok {
		is.Fail("expected fastReader type, got %T", cr)
	}

	cri.state = rState(rStateInLineComment + 1)

	handleEOF := func() (_ bool, r any) {
		defer func() {
			r = recover()
		}()
		return cri.handleEOF(), nil
	}

	resp, r := handleEOF()
	is.NotNil(r)
	is.False(resp)

	is.Equal(r, panicUnknownReaderStateDuringEOF)
}
