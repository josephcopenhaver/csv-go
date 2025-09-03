package csv_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/josephcopenhaver/csv-go/v3"
)

func TestFunctionalReaderPrepareRowOKPaths(t *testing.T) {
	t.Parallel()

	tcs := []functionalReaderTestCase{
		{
			when: "in quoted field encounter CRLF with RecSepDiscovery=true",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("\"234567", "\r\n\"")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().DiscoverRecordSeparator(true),
				csv.ReaderOpts().Quote('"'),
			},
			rows: [][]string{{"234567\r\n"}},
		},
		{
			when: "in quoted field encounter LF with RecSepDiscovery=true",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("\"234567", "\n\"")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().DiscoverRecordSeparator(true),
				csv.ReaderOpts().Quote('"'),
			},
			rows: [][]string{{"234567\n"}},
		},
		{
			when: "start of doc found CRLF with DropBOM=true with RecSepDiscovery=true",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(string(bomBytes()) + "\r\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().DiscoverRecordSeparator(true),
				csv.ReaderOpts().RemoveByteOrderMarker(true),
			},
			rows: [][]string{{""}},
		},
		{
			when: "start of doc found LF with DropBOM=true with RecSepDiscovery=true",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(string(bomBytes()) + "\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().DiscoverRecordSeparator(true),
				csv.ReaderOpts().RemoveByteOrderMarker(true),
			},
			rows: [][]string{{""}},
		},
		{
			when: "within line comment encounter LF then EOF with RecSep=CR",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("#234567", "\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r"),
				csv.ReaderOpts().Comment('#'),
			},
		},
		{
			when: "within line comment encounter CR then EOF with RecSep=LF",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("#234567", "\r")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
			},
		},
		{
			when: "within line comment encounter LF with RecSep=CR",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("#234567", "\n#3456\ra,b")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r"),
				csv.ReaderOpts().Comment('#'),
			},
			rows: [][]string{strings.Split("a,b", ",")},
		},
		{
			when: "within line comment encounter CR with RecSep=LF",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("#234567", "\r#3456\na,b")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
			},
			rows: [][]string{strings.Split("a,b", ",")},
		},
		{
			when: "quoted field encounter LF ending buf with RecSep=CR",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("\"234567", "123456\n", "\"")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r"),
				csv.ReaderOpts().Quote('"'),
			},
			rows: [][]string{{"234567123456\n"}},
		},
		{
			when: "quoted field encounter CR ending buf with RecSep=LF",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("\"234567", "123456\r", "\"")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
			},
			rows: [][]string{{"234567123456\r"}},
		},
		{
			when: "quoted field encounter LF with RecSep=CR",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("\"234567", "\n\"")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r"),
				csv.ReaderOpts().Quote('"'),
			},
			rows: [][]string{{"234567\n"}},
		},
		{
			when: "quoted field encounter CR with RecSep=LF",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("\"234567", "\r\"")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
			},
			rows: [][]string{{"234567\r"}},
		},
		{
			when: "in line comment encounter comment",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("#234567", "#\na,b")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
			},
			rows: [][]string{strings.Split("a,b", ",")},
		},
		{
			when: "in field encounter comment",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("1234567", "#")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
			},
			rows: [][]string{{"1234567#"}},
		},
		{
			when: "in quoted field encounter comment",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("\"234567", "#\"")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Comment('#'),
			},
			rows: [][]string{{"234567#"}},
		},
		{
			when: "start of doc encounter comment",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("#2345\n#", "\na,b")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
			},
			rows: [][]string{strings.Split("a,b", ",")},
		},
		{
			when: "start of doc encounter comment with DropBOM=true",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(string(bomBytes()) + "# neat\na,b")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RemoveByteOrderMarker(true),
				csv.ReaderOpts().Comment('#'),
			},
			rows: [][]string{strings.Split("a,b", ",")},
		},
		{
			when: "in quoted field encounter 1 byte RecSep",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("\"234567", "\n\"")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
			},
			rows: [][]string{{"234567\n"}},
		},
		{
			when: "in quoted field encounter CRLF with RecSep=CRLF",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("\"234567", "\r\n\"")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().RecordSeparator("\r\n"),
			},
			rows: [][]string{{"234567\r\n"}},
		},
		{
			when: "start of doc encounter 1 byte RecSep with DropBOM=true",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(string(bomBytes()) + "456\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RemoveByteOrderMarker(true),
			},
			rows: [][]string{{"456"}},
		},
		{
			when: "start of doc, encounter CRLF with DropBOM=true with RecSep=CRLF",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(string(bomBytes()) + "45\r\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r\n"),
				csv.ReaderOpts().RemoveByteOrderMarker(true),
			},
			rows: [][]string{{"45"}},
		},
		{
			when: "In unquoted field encounter CR without LF with RecSep=CRLF",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("a234567", "\rb")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r\n"),
			},
			rows: [][]string{{"a234567\rb"}},
		},
		{
			when: "In quoted field encounter CR without LF with RecSep=CRLF",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("\"234567", "\r\"")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().RecordSeparator("\r\n"),
			},
			rows: [][]string{{"234567\r"}},
		},
		{
			when: "BOM followed by CR without LF at start of doc with data after it while DropBOM=true and RecSep=CRLF",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(string(bomBytes()) + "\rb")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RemoveByteOrderMarker(true),
				csv.ReaderOpts().RecordSeparator("\r\n"),
			},
			rows: [][]string{{"\rb"}},
		},
		{
			when: "CR without LF at start of doc with data after it while DropBOM=true and RecSep=CRLF",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("\rb")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RemoveByteOrderMarker(true),
				csv.ReaderOpts().RecordSeparator("\r\n"),
			},
			rows: [][]string{{"\rb"}},
		},
		{
			when: "quote char found in line comment state",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`#234567`, "\"\na,b")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Comment('#'),
			},
			rows: [][]string{strings.Split(`a,b`, ",")},
		},
		{
			when: "quote char found after the start of a field, but it is not the first character and ErrOnQInUF=false and it is the last character",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`a23456,`, `b"`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().ErrorOnQuotesInUnquotedField(false),
			},
			rows: [][]string{strings.Split(`a23456,b"`, ",")},
		},
		{
			when: "quote char found after the start of a field, but it is not the first character and ErrOnQInUF=false and it is not the last character",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader(`a23456,`, `b"c`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().ErrorOnQuotesInUnquotedField(false),
			},
			rows: [][]string{strings.Split(`a23456,b"c`, ",")},
		},
		{
			when: "quote char found at start of record but not in first char and rFlagErrOnQInUF=false and there is no closing quote",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`a"`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().ErrorOnQuotesInUnquotedField(false),
			},
			rows: [][]string{{`a"`}},
		},
		{
			when: "quote char found at start of record but not in first char and rFlagErrOnQInUF=false",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`a"b"`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().ErrorOnQuotesInUnquotedField(false),
			},
			rows: [][]string{{`a"b"`}},
		},
		{
			when: "quote char found at start of doc while DropBOM=true",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(string(bomBytes()) + `"a",b,c`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().RemoveByteOrderMarker(true),
			},
			rows: [][]string{strings.Split("a,b,c", ",")},
		},
		{
			when: "escape char found while in a line comment",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("#234567", "\\\na,b")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
				csv.ReaderOpts().Comment('#'),
			},
			rows: [][]string{strings.Split("a,b", ",")},
		},
		{
			when: "escape char found while in a field",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("a2345,7", "\\nice")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			rows: [][]string{strings.Split("a2345,7\\nice", ",")},
		},
		{
			when: "escape char found right after the first escape in a quoted field",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("a,b4,\"\\", "\\nice\"")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			rows: [][]string{strings.Split("a,b4,\\nice", ",")},
		},
		{
			when: "escape char found while in start of field state",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("a,b456,", "\\")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			rows: [][]string{strings.Split("a,b456,\\", ",")},
		},
		{
			when: "require BOM, drop it, and starts with escape char",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(string(bomBytes()) + "\\,b,c")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ErrorOnNoByteOrderMarker(true),
				csv.ReaderOpts().RemoveByteOrderMarker(true),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			rows: [][]string{strings.Split("\\,b,c", ",")},
		},
		{
			when: "field separator in comment",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(newPartReader("#234,567\na,b,c")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
			},
			rows: [][]string{strings.Split("a,b,c", ",")},
		},
		{
			when: "field separator in quoted field",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`a,b,"c,d"`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
			},
			rows: [][]string{append(strings.Split("a,b", ","), "c,d")},
		},
		{
			when: "multibyte field separator",
			newOptsF: func() []csv.ReaderOption {
				mindBlownEmoji := string([]byte{0xF0, 0x9F, 0xA4, 0xAF})
				csvContents := strings.ReplaceAll("a,b,c", ",", mindBlownEmoji)
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(csvContents)),
					csv.ReaderOpts().FieldSeparator([]rune(mindBlownEmoji)[0]),
				}
			},
			rows: [][]string{strings.Split("a,b,c", ",")},
		},
		{
			when: "starts with BOM, has minimum buffer size, and removing BOM",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(string(bomBytes()) + "1234,b,c")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RemoveByteOrderMarker(true),
				csv.ReaderOpts().ReaderBufferSize(csv.ReaderMinBufferSize),
			},
			rows: [][]string{strings.Split("1234,b,c", ",")},
		},
		{
			when: "starts with BOM, has minimum buffer size, has no headers or data, and removing BOM",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(bytes.NewReader(bomBytes())),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RemoveByteOrderMarker(true),
				csv.ReaderOpts().ReaderBufferSize(csv.ReaderMinBufferSize),
			},
		},
		{
			when: "single non-utf8 byte",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(bytes.NewReader([]byte{0xC0})),
				}
			},
			rows: [][]string{{string([]byte{0xC0})}},
		},
		{
			when: "single non-utf8 byte after a full ascii byte",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(bytes.NewReader([]byte{'A', 0xC0})),
				}
			},
			rows: [][]string{{string([]byte{'A', 0xC0})}},
		},
		{
			when: "single non-utf8 byte after a comma",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(bytes.NewReader([]byte{',', 0xC0})),
				}
			},
			rows: [][]string{{"", string([]byte{0xC0})}},
		},
		{
			when: "single non-utf8 byte in quotes with quote set",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(bytes.NewReader([]byte{'"', 0xC0, '"'})),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
			},
			rows: [][]string{{string([]byte{0xC0})}},
		},
		{
			when: "single non-utf8 byte in comment with comment set",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(bytes.NewReader(append(append([]byte("# "), 0xC0), []byte("\n1,2")...))),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
			},
			rows: [][]string{strings.Split("1,2", ",")},
		},
		{
			when: "configured to remove a byte order marker that exists",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(bytes.NewReader(append(bomBytes(), []byte("1,2,3")...))),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RemoveByteOrderMarker(true),
			},
			rows: [][]string{strings.Split("1,2,3", ",")},
		},
		{
			when: "two non record sep characters separated by two record sep and ending in a record sep",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("a\n\nb\n")),
				}
			},
			rows: [][]string{{"a"}, {""}, {"b"}},
		},
		{
			when: "two non record sep characters separated by two record sep and not ending in a record sep",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("a\n\nb")),
				}
			},
			rows: [][]string{{"a"}, {""}, {"b"}},
		},
		{
			when: "two non record sep characters separated by two record sep and ending in a record sep with TerminalRecordSeparatorEmitsRecord=false",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("a\n\nb\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().TerminalRecordSeparatorEmitsRecord(false),
			},
			rows: [][]string{{"a"}, {""}, {"b"}},
		},
		{
			when: "two non record sep characters separated by two record sep and ending in a record sep with TerminalRecordSeparatorEmitsRecord=true",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("a\n\nb\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().TerminalRecordSeparatorEmitsRecord(true),
			},
			rows: [][]string{{"a"}, {""}, {"b"}, {""}},
		},
		{
			when: "discover record sep enabled and comment line ends file with CR",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("# neat\r")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().DiscoverRecordSeparator(true),
			},
		},
		{
			when: "EOF on first line with comment enabled",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("#neat")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
			},
		},
		{
			when: "just a comma",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(",")),
				}
			},
			rows: [][]string{{"", ""}},
		},
		{
			when: "quotes and escapes one column",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`"\"1\""`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			rows: [][]string{{`"1"`}},
		},
		{
			when: "quotes and escapes two columns with second empty",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`"\"1\"",`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			rows: [][]string{{`"1"`, ""}},
		},
		{
			when: "quotes and escapes two columns with second empty ending in newline",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`"\"1\"",` + "\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			rows: [][]string{{`"1"`, ""}},
		},
		{
			when: "quotes and escapes two columns",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`"\"1\"","\"2\""`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			rows: [][]string{{`"1"`, `"2"`}},
		},
		{
			when: "quotes and escapes two columns ending in newline",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`"\"1\"","\"2\""` + "\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			rows: [][]string{{`"1"`, `"2"`}},
		},
		{
			when: "quotes in field one column",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`"1""2"`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
			},
			rows: [][]string{{`1"2`}},
		},
		{
			when: "quotes in field two columns with second empty",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`"1""2",`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
			},
			rows: [][]string{{`1"2`, ""}},
		},
		{
			when: "quotes in field two columns with second empty ending in newline",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`"1""2",` + "\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
			},
			rows: [][]string{{`1"2`, ""}},
		},
		{
			when: "quotes in field two columns",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`"1""2","3""4"`)),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
			},
			rows: [][]string{{`1"2`, `3"4`}},
		},
		{
			when: "quotes in field two columns ending in newline",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(`"1""2","3""4"` + "\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
			},
			rows: [][]string{{`1"2`, `3"4`}},
		},
		{
			when: "comments after start of records and support enabled",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("#neat1\na,b,c\n#neat2\n1,2,3")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().CommentsAllowedAfterStartOfRecords(true),
			},
			rows: [][]string{strings.Split("a,b,c", ","), strings.Split("1,2,3", ",")},
		},
		{
			when: "comments after start of records and support explicitly disabled",
			then: "no error processing second row because it is a comment interpreted as one field",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("#neat1\na\n#neat2\n3")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().CommentsAllowedAfterStartOfRecords(false),
			},
			rows: [][]string{{"a"}, {"#neat2"}, {"3"}},
		},
		{
			when: "comments after start of records and support explicitly disabled and numFields=1",
			then: "no error processing second row because it is a comment interpreted as one field",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("#neat1\na\n#neat2\n3")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().CommentsAllowedAfterStartOfRecords(false),
				csv.ReaderOpts().NumFields(1),
			},
			rows: [][]string{{"a"}, {"#neat2"}, {"3"}},
		},
		{
			when: "comments after start of records and support implicitly disabled",
			then: "no error processing second row because it is a comment interpreted as one field",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("#neat1\na\n#neat2\n3")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
			},
			rows: [][]string{{"a"}, {"#neat2"}, {"3"}},
		},
		{
			when: "comments after start of records and support implicitly disabled and numFields=1",
			then: "no error processing second row because it is a comment interpreted as one field",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("#neat1\na\n#neat2\n3")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Comment('#'),
				csv.ReaderOpts().NumFields(1),
			},
			rows: [][]string{{"a"}, {"#neat2"}, {"3"}},
		},
		{
			when: "explicit error on newline in field is off where CR in middle of field",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("h\ri")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\n"),
				csv.ReaderOpts().ErrorOnNewlineInUnquotedField(false),
			},
			rows: [][]string{{"h\ri"}},
		},
		{
			when: "explicit error on newline in field is off where LF in middle of field",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("h\ni")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator(string(utf8LineSeparator)),
				csv.ReaderOpts().ErrorOnNewlineInUnquotedField(false),
			},
			rows: [][]string{{"h\ni"}},
		},
		{
			when: "explicit error on newline in field is off where CR at start of record",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("\r")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\n"),
				csv.ReaderOpts().ErrorOnNewlineInUnquotedField(false),
			},
			rows: [][]string{{"\r"}},
		},
		{
			when: "explicit error on newline in field is off where LF at start of record",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator(string(utf8LineSeparator)),
				csv.ReaderOpts().ErrorOnNewlineInUnquotedField(false),
			},
			rows: [][]string{{"\n"}},
		},
		{
			when: "explicit error on newline in field is off where CR at start of second field",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(",\r")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\n"),
				csv.ReaderOpts().ErrorOnNewlineInUnquotedField(false),
			},
			rows: [][]string{{"", "\r"}},
		},
		{
			when: "explicit error on newline in field is off where LF at start of second field",
			then: "error",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader(",\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator(string(utf8LineSeparator)),
				csv.ReaderOpts().ErrorOnNewlineInUnquotedField(false),
			},
			rows: [][]string{{"", "\n"}},
		},
		{
			when: "BOM removal enabled and EOF is the first event encountered",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RemoveByteOrderMarker(true),
			},
		},
		{
			when: "BOM removal enabled then normal-rune+EOF",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("A")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RemoveByteOrderMarker(true),
			},
			rows: [][]string{{"A"}},
		},
		{
			when: "BOM removal enabled then normal-rune+EOF",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(bytes.NewReader([]byte{0xC0})),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RemoveByteOrderMarker(true),
			},
			rows: [][]string{{string([]byte{0xC0})}},
		},
		{
			when: "BOM removal and error on no-BOM enabled",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(bytes.NewReader(append(bomBytes(), []byte("a,b,c")...))),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RemoveByteOrderMarker(true),
				csv.ReaderOpts().ErrorOnNoByteOrderMarker(true),
			},
			rows: [][]string{strings.Split("a,b,c", ",")},
		},
		{
			when: "error on no-BOM enabled",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(bytes.NewReader(append(bomBytes(), []byte("a,b,c")...))),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().ErrorOnNoByteOrderMarker(true),
			},
			rows: [][]string{append([]string{string(bomBytes()) + "a"}, strings.Split("b,c", ",")...)},
		},
		{
			when: "quotes enabled and err on quote in unquoted field enabled but no quotes present",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("a1,b2,c3")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().ErrorOnQuotesInUnquotedField(true),
			},
			rows: [][]string{strings.Split("a1,b2,c3", ",")},
		},
		{
			when: "escape set and record sep CRLF after closing quote",
			newOptsF: func() []csv.ReaderOption {
				return []csv.ReaderOption{
					csv.ReaderOpts().Reader(strings.NewReader("\"\"\r\n")),
				}
			},
			newOpts: []csv.ReaderOption{
				csv.ReaderOpts().RecordSeparator("\r\n"),
				csv.ReaderOpts().Quote('"'),
				csv.ReaderOpts().Escape('\\'),
			},
			rows: [][]string{{""}},
		},
	}

	for _, tc := range tcs {
		tc.Run(t)
	}
}
