package csv_test

import (
	"testing"
	"unicode/utf8"

	"github.com/josephcopenhaver/csv-go/v3"
)

func TestFunctionalWriterInitializationErrorPaths(t *testing.T) {
	t.Parallel()

	tcs := []functionalWriterTestCase{
		{
			when: "record separator is not a newline character",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().RecordSeparator(","),
			},
			newWriterErrIs:  []error{csv.ErrBadConfig},
			newWriterErrStr: csv.ErrBadConfig.Error() + "\nrecord separator can only be one valid utf8 newline rune long or \"\\r\\n\"",
		},
		{
			when: "record separator is newline characters but 2 CRLF",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().RecordSeparator("\r\n\r\n"),
			},
			newWriterErrIs:  []error{csv.ErrBadConfig},
			newWriterErrStr: csv.ErrBadConfig.Error() + "\nrecord separator can only be one valid utf8 newline rune long or \"\\r\\n\"",
		},
		{
			when: "record separator is newline characters but 2 and not CRLF",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().RecordSeparator("\n\n"),
			},
			newWriterErrIs:  []error{csv.ErrBadConfig},
			newWriterErrStr: csv.ErrBadConfig.Error() + "\nrecord separator can only be one valid utf8 newline rune long or \"\\r\\n\"",
		},
		{
			when: "record separator is empty string",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().RecordSeparator(""),
			},
			newWriterErrIs:  []error{csv.ErrBadConfig},
			newWriterErrStr: csv.ErrBadConfig.Error() + "\nrecord separator can only be one valid utf8 newline rune long or \"\\r\\n\"",
		},
		{
			when: "record separator is RuneError",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().RecordSeparator(string(utf8.RuneError)),
			},
			newWriterErrIs:  []error{csv.ErrBadConfig},
			newWriterErrStr: csv.ErrBadConfig.Error() + "\nrecord separator can only be one valid utf8 newline rune long or \"\\r\\n\"",
		},
		{
			when: "record separator is CR RuneError",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().RecordSeparator("\r" + string(utf8.RuneError)),
			},
			newWriterErrIs:  []error{csv.ErrBadConfig},
			newWriterErrStr: csv.ErrBadConfig.Error() + "\nrecord separator can only be one valid utf8 newline rune long or \"\\r\\n\"",
		},
		{
			when: "nil writer",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Writer(nil),
			},
			newWriterErrIs:  []error{csv.ErrBadConfig},
			newWriterErrStr: csv.ErrBadConfig.Error() + "\nnil writer",
		},
		{
			when: "zero num fields",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().NumFields(0),
			},
			newWriterErrIs:  []error{csv.ErrBadConfig},
			newWriterErrStr: csv.ErrBadConfig.Error() + "\nnum fields must be greater than zero",
		},
		{
			when: "negative num fields",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().NumFields(-1),
			},
			newWriterErrIs:  []error{csv.ErrBadConfig},
			newWriterErrStr: csv.ErrBadConfig.Error() + "\nnum fields must be greater than zero",
		},
		{
			when: "LF for field separator",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().FieldSeparator('\n'),
			},
			newWriterErrIs:  []error{csv.ErrBadConfig},
			newWriterErrStr: csv.ErrBadConfig.Error() + "\ninvalid field separator value",
		},
		{
			when: "RuneError for field separator",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().FieldSeparator(utf8.RuneError),
			},
			newWriterErrIs:  []error{csv.ErrBadConfig},
			newWriterErrStr: csv.ErrBadConfig.Error() + "\ninvalid field separator value",
		},
		{
			when: "LF for quote",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Quote('\n'),
			},
			newWriterErrIs:  []error{csv.ErrBadConfig},
			newWriterErrStr: csv.ErrBadConfig.Error() + "\ninvalid quote value",
		},
		{
			when: "RuneError for quote",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Quote(utf8.RuneError),
			},
			newWriterErrIs:  []error{csv.ErrBadConfig},
			newWriterErrStr: csv.ErrBadConfig.Error() + "\ninvalid quote value",
		},
		{
			when: "LF for escape",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape('\n'),
			},
			newWriterErrIs:  []error{csv.ErrBadConfig},
			newWriterErrStr: csv.ErrBadConfig.Error() + "\ninvalid escape value",
		},
		{
			when: "RuneError for escape",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape(utf8.RuneError),
			},
			newWriterErrIs:  []error{csv.ErrBadConfig},
			newWriterErrStr: csv.ErrBadConfig.Error() + "\ninvalid escape value",
		},
		{
			when: "comma for quote",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Quote(','),
			},
			newWriterErrIs:  []error{csv.ErrBadConfig},
			newWriterErrStr: csv.ErrBadConfig.Error() + "\ninvalid field separator and quote combination",
		},
		{
			when: "comma for escape",
			newOpts: []csv.WriterOption{
				csv.WriterOpts().Escape(','),
			},
			newWriterErrIs:  []error{csv.ErrBadConfig},
			newWriterErrStr: csv.ErrBadConfig.Error() + "\ninvalid field separator and escape combination",
		},
	}

	for _, tc := range tcs {
		if tc.then == "" {
			tc.then = "a coupled error should occur"
		}
		tc.Run(t)
	}
}
