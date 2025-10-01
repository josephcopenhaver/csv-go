package csv

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnitWriterInitializationBufferPaths(t *testing.T) {
	t.Parallel()

	is := assert.New(t)

	mockBuf := &bytes.Buffer{}

	// InitialFieldBufferSize(254)
	{
		cw, err := NewWriter(
			WriterOpts().Writer(mockBuf),
			WriterOpts().InitialFieldBufferSize(254),
		)
		is.Nil(err)
		is.Equal(0, len(cw.fieldBuf))
		is.Equal(254, cap(cw.fieldBuf))
	}

	// InitialFieldBufferSize(255)
	{
		cw, err := NewWriter(
			WriterOpts().Writer(mockBuf),
			WriterOpts().InitialFieldBufferSize(255),
		)
		is.Nil(err)
		is.Equal(0, len(cw.fieldBuf))
		is.Equal(255, cap(cw.fieldBuf))
	}

	// InitialFieldBuffer(254)
	{
		buf := [254]byte{}

		cw, err := NewWriter(
			WriterOpts().Writer(mockBuf),
			WriterOpts().InitialFieldBuffer(buf[:]),
		)
		is.Nil(err)
		is.Equal(0, len(cw.fieldBuf))
		is.Equal(254, cap(cw.fieldBuf))
	}

	// InitialFieldBuffer(255)
	{
		buf := [255]byte{}

		cw, err := NewWriter(
			WriterOpts().Writer(mockBuf),
			WriterOpts().InitialFieldBuffer(buf[:]),
		)
		is.Nil(err)
		is.Equal(0, len(cw.fieldBuf))
		is.Equal(255, cap(cw.fieldBuf))
	}
	//
	//
	//

	// InitialRecordBufferSize(254)
	{
		cw, err := NewWriter(
			WriterOpts().Writer(mockBuf),
			WriterOpts().InitialRecordBufferSize(254),
		)
		is.Nil(err)
		is.Equal(0, len(cw.recordBuf))
		is.Equal(254, cap(cw.recordBuf))
	}

	// InitialRecordBufferSize(255)
	{
		cw, err := NewWriter(
			WriterOpts().Writer(mockBuf),
			WriterOpts().InitialRecordBufferSize(255),
		)
		is.Nil(err)
		is.Equal(0, len(cw.recordBuf))
		is.Equal(255, cap(cw.recordBuf))
	}

	// InitialRecordBuffer(254)
	{
		buf := [254]byte{}

		cw, err := NewWriter(
			WriterOpts().Writer(mockBuf),
			WriterOpts().InitialRecordBuffer(buf[:]),
		)
		is.Nil(err)
		is.Equal(0, len(cw.recordBuf))
		is.Equal(254, cap(cw.recordBuf))
	}

	// InitialRecordBuffer(255)
	{
		buf := [255]byte{}

		cw, err := NewWriter(
			WriterOpts().Writer(mockBuf),
			WriterOpts().InitialRecordBuffer(buf[:]),
		)
		is.Nil(err)
		is.Equal(0, len(cw.recordBuf))
		is.Equal(255, cap(cw.recordBuf))
	}
}
