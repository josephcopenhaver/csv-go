package csv

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// statically validates panicErr type satisfies interface implementations
// that are normally required by panic handlers
var (
	// verify panicErr implements error interface
	_ error = panicErr(0)

	// verify panicErr implements Stringer interface
	_ fmt.Stringer = panicErr(0)
)

func TestUnitReaderPanicOnValidate(t *testing.T) {
	t.Parallel()

	//
	// when the config's record separator is in a corrupted / impossible value
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
	is.Equal(panicRecordSepLen.String(), panicRecordSepLen.Error())
	is.Equal(panicRecordSepLen.String(), "invalid record separator length")
}

func TestUnitReaderPanicOnHandleEOF(t *testing.T) {
	t.Parallel()

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
		t.Fatalf("expected fastReader type, got %T", cr)
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
	is.Equal(panicUnknownReaderStateDuringEOF.String(), panicUnknownReaderStateDuringEOF.Error())
	is.Equal(panicUnknownReaderStateDuringEOF.String(), "reader in unknown state when EOF encountered")
}

func TestUnitSecOpReaderPanicOnHandleEOF(t *testing.T) {
	t.Parallel()

	//
	// when secOpReader's recordIndex is in a corrupted state
	//
	// a panic should occur if incRecordIndexWithMax is called
	//

	is := assert.New(t)

	cr, crv, err := internalNewReader(
		ReaderOpts().Reader(strings.NewReader("")),
		ReaderOpts().MaxRecords(1),
	)
	is.Nil(err)
	is.NotNil(cr)
	is.NotNil(crv)

	cri, ok := crv.(*secOpReader)
	if !ok {
		is.Fail("expected secOpReader type, got %T", cr)
	}

	cri.recordIndex = 1

	handlePanic := func() (r any) {
		defer func() {
			r = recover()
		}()
		cri.incRecordIndex()
		return nil
	}

	r := handlePanic()
	is.NotNil(r)

	is.Equal(r, panicMissedHandlingMaxRecordIndex)
	is.Equal(panicMissedHandlingMaxRecordIndex.String(), panicMissedHandlingMaxRecordIndex.Error())
	is.Equal(panicMissedHandlingMaxRecordIndex.String(), "missed handling record index at max value")
}

func TestUnitReaderPanicOnCorruptedFieldLengthsEOF(t *testing.T) {
	t.Parallel()

	//
	// when secOpReader's recordIndex is in a corrupted state
	//
	// a panic should occur if incRecordIndexWithMax is called
	//

	is := assert.New(t)

	cr, crv, err := internalNewReader(
		ReaderOpts().Reader(strings.NewReader("")),
		ReaderOpts().NumFields(1),
	)
	is.Nil(err)
	is.NotNil(cr)
	is.NotNil(crv)

	cri, ok := crv.(*fastReader)
	if !ok {
		t.Fatalf("expected fastReader type, got %T", cr)
	}

	cri.fieldLengths = append(cri.fieldLengths, 0)
	cri.fieldIndex++

	cri.fieldLengths = append(cri.fieldLengths, 0)
	cri.fieldIndex++

	handlePanic := func() (r any) {
		defer func() {
			r = recover()
		}()
		cri.checkNumFields(nil)
		return nil
	}

	r := handlePanic()
	is.NotNil(r)

	is.Equal(r, panicMissedHandlingMaxExpectedFieldIndex)
	is.Equal(panicMissedHandlingMaxExpectedFieldIndex.String(), panicMissedHandlingMaxExpectedFieldIndex.Error())
	is.Equal(panicMissedHandlingMaxExpectedFieldIndex.String(), "missed handling field index at expected max value")
}
