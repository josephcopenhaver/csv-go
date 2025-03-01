package csv_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/josephcopenhaver/csv-go"
)

func TestFunctionalReaderPrepareRowOKPaths(t *testing.T) {
	tcs := []functionalReaderTestCase{
		{
			when: "single non-utf8 byte",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(bytes.NewReader([]byte{0xC0})),
			},
			rows: [][]string{{string([]byte{0xC0})}},
		},
		{
			when: "single non-utf8 byte after a full ascii byte",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(bytes.NewReader([]byte{'A', 0xC0})),
			},
			rows: [][]string{{string([]byte{'A', 0xC0})}},
		},
		{
			when: "single non-utf8 byte after a comma",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(bytes.NewReader([]byte{',', 0xC0})),
			},
			rows: [][]string{{"", string([]byte{0xC0})}},
		},
		{
			when: "single non-utf8 byte in quotes with quote set",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(bytes.NewReader([]byte{'"', 0xC0, '"'})),
				csv.ReaderOpts().Quote('"'),
			},
			rows: [][]string{{string([]byte{0xC0})}},
		},
		{
			when: "single non-utf8 byte in comment with comment set",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(bytes.NewReader(append(append([]byte("# "), 0xC0), []byte("\n1,2")...))),
				csv.ReaderOpts().Comment('#'),
			},
			rows: [][]string{strings.Split("1,2", ",")},
		},
		{
			when: "configured to remove a byte order marker that exists",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(bytes.NewReader(append([]byte{0xEF, 0xBB, 0xBF}, []byte("1,2,3")...))),
				csv.ReaderOpts().RemoveByteOrderMarker(true),
			},
			rows: [][]string{strings.Split("1,2,3", ",")},
		},
		{
			when: "two non record sep characters separated by two record sep and ending in a record sep",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("a\n\nb\n")),
			},
			rows: [][]string{{"a"}, {""}, {"b"}},
		},
		{
			when: "two non record sep characters separated by two record sep and not ending in a record sep",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("a\n\nb")),
			},
			rows: [][]string{{"a"}, {""}, {"b"}},
		},
		{
			when: "two non record sep characters separated by two record sep and ending in a record sep with TerminalRecordSeparatorEmitsRecord=false",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("a\n\nb\n")),
				csv.ReaderOpts().TerminalRecordSeparatorEmitsRecord(false),
			},
			rows: [][]string{{"a"}, {""}, {"b"}},
		},
		{
			when: "two non record sep characters separated by two record sep and ending in a record sep with TerminalRecordSeparatorEmitsRecord=true",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("a\n\nb\n")),
				csv.ReaderOpts().TerminalRecordSeparatorEmitsRecord(true),
			},
			rows: [][]string{{"a"}, {""}, {"b"}, {""}},
		},
		{
			when: "discover record sep enabled and comment line ends file with CR",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("# neat\r")),
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().DiscoverRecordSeparator(true),
			},
		},
		{
			when: "EOF on first line with comment enabled",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("#neat")),
				csv.ReaderOpts().Comment('#'),
			},
		},
		{
			when: "just a comma",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader(",")),
			},
			rows: [][]string{{"", ""}},
		},
		{
			when: "quotes and escapes one column",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader(`"\"1\""`)),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			rows: [][]string{{`"1"`}},
		},
		{
			when: "quotes and escapes two columns with second empty",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader(`"\"1\"",`)),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			rows: [][]string{{`"1"`, ""}},
		},
		{
			when: "quotes and escapes two columns with second empty ending in newline",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader(`"\"1\"",` + "\n")),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			rows: [][]string{{`"1"`, ""}},
		},
		{
			when: "quotes and escapes two columns",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader(`"\"1\"","\"2\""`)),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			rows: [][]string{{`"1"`, `"2"`}},
		},
		{
			when: "quotes and escapes two columns ending in newline",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader(`"\"1\"","\"2\""` + "\n")),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			rows: [][]string{{`"1"`, `"2"`}},
		},
		{
			when: "quotes in field one column",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader(`"1""2"`)),
				csv.ReaderOpts().Quote('"'),
			},
			rows: [][]string{{`1"2`}},
		},
		{
			when: "quotes in field two columns with second empty",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader(`"1""2",`)),
				csv.ReaderOpts().Quote('"'),
			},
			rows: [][]string{{`1"2`, ""}},
		},
		{
			when: "quotes in field two columns with second empty ending in newline",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader(`"1""2",` + "\n")),
				csv.ReaderOpts().Quote('"'),
			},
			rows: [][]string{{`1"2`, ""}},
		},
		{
			when: "quotes in field two columns",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader(`"1""2","3""4"`)),
				csv.ReaderOpts().Quote('"'),
			},
			rows: [][]string{{`1"2`, `3"4`}},
		},
		{
			when: "quotes in field two columns ending in newline",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader(`"1""2","3""4"` + "\n")),
				csv.ReaderOpts().Quote('"'),
			},
			rows: [][]string{{`1"2`, `3"4`}},
		},
		{
			when: "comments after start of records and support enabled",
			then: "no error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("#neat1\na,b,c\n#neat2\n1,2,3")),
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().CommentsAllowedAfterStartOfRecords(true),
			},
			rows: [][]string{strings.Split("a,b,c", ","), strings.Split("1,2,3", ",")},
		},
		{
			when: "comments after start of records and support explicitly disabled",
			then: "no error processing second row because it is a comment interpreted as one field",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("#neat1\na\n#neat2\n3")),
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().CommentsAllowedAfterStartOfRecords(false),
			},
			rows: [][]string{{"a"}, {"#neat2"}, {"3"}},
		},
		{
			when: "comments after start of records and support explicitly disabled and numFields=1",
			then: "no error processing second row because it is a comment interpreted as one field",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("#neat1\na\n#neat2\n3")),
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().CommentsAllowedAfterStartOfRecords(false),
				csv.ReaderOpts().NumFields(1),
			},
			rows: [][]string{{"a"}, {"#neat2"}, {"3"}},
		},
		{
			when: "comments after start of records and support implicitly disabled",
			then: "no error processing second row because it is a comment interpreted as one field",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("#neat1\na\n#neat2\n3")),
				csv.ReaderOpts().Comment('#'),
			},
			rows: [][]string{{"a"}, {"#neat2"}, {"3"}},
		},
		{
			when: "comments after start of records and support implicitly disabled and numFields=1",
			then: "no error processing second row because it is a comment interpreted as one field",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("#neat1\na\n#neat2\n3")),
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().NumFields(1),
			},
			rows: [][]string{{"a"}, {"#neat2"}, {"3"}},
		},
		{
			when: "explicit error on newline in field is off where CR in middle of field",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("h\ri")),
				csv.ReaderOpts().RecordSeparator("\n"),
				csv.ReaderOpts().ErrorOnNewlineInUnquotedField(false),
			},
			rows: [][]string{{"h\ri"}},
		},
		{
			when: "explicit error on newline in field is off where LF in middle of field",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("h\ni")),
				csv.ReaderOpts().RecordSeparator(string(utf8LineSeparator)),
				csv.ReaderOpts().ErrorOnNewlineInUnquotedField(false),
			},
			rows: [][]string{{"h\ni"}},
		},
		{
			when: "explicit error on newline in field is off where CR at start of record",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("\r")),
				csv.ReaderOpts().RecordSeparator("\n"),
				csv.ReaderOpts().ErrorOnNewlineInUnquotedField(false),
			},
			rows: [][]string{{"\r"}},
		},
		{
			when: "explicit error on newline in field is off where LF at start of record",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader("\n")),
				csv.ReaderOpts().RecordSeparator(string(utf8LineSeparator)),
				csv.ReaderOpts().ErrorOnNewlineInUnquotedField(false),
			},
			rows: [][]string{{"\n"}},
		},
		{
			when: "explicit error on newline in field is off where CR at start of second field",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader(",\r")),
				csv.ReaderOpts().RecordSeparator("\n"),
				csv.ReaderOpts().ErrorOnNewlineInUnquotedField(false),
			},
			rows: [][]string{{"", "\r"}},
		},
		{
			when: "explicit error on newline in field is off where LF at start of second field",
			then: "error",
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Reader(strings.NewReader(",\n")),
				csv.ReaderOpts().RecordSeparator(string(utf8LineSeparator)),
				csv.ReaderOpts().ErrorOnNewlineInUnquotedField(false),
			},
			rows: [][]string{{"", "\n"}},
		},
	}

	for _, tc := range tcs {
		tc.Run(t)
	}
}
