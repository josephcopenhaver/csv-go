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
		reader:           strings.NewReader(""),
		recordSepRuneLen: -2,
	}

	is.PanicsWithValue(panicRecordSepRuneLen, func() {
		_ = cfg.validate()
	})

	cfg = rCfg{
		reader:           strings.NewReader(""),
		recordSepRuneLen: 3,
	}

	is.PanicsWithValue(panicRecordSepRuneLen, func() {
		_ = cfg.validate()
	})

	is.Equal(panicRecordSepRuneLen.String(), panicRecordSepRuneLen.Error())
	is.Equal(panicRecordSepRuneLen.String(), "invalid record separator rune length")
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

	is.PanicsWithValue(panicUnknownReaderStateDuringEOF, func() {
		_ = cri.handleEOF()
	})

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

	is.PanicsWithValue(panicMissedHandlingMaxRecordIndex, cri.incRecordIndex)

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

	is.PanicsWithValue(panicMissedHandlingMaxExpectedFieldIndex, func() {
		cri.checkNumFields(nil)
	})

	is.Equal(panicMissedHandlingMaxExpectedFieldIndex.String(), panicMissedHandlingMaxExpectedFieldIndex.Error())
	is.Equal(panicMissedHandlingMaxExpectedFieldIndex.String(), "missed handling field index at expected max value")
}
