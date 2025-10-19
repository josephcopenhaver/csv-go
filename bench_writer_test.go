package csv_test

import (
	"io"
	"math"
	"runtime"
	"strconv"
	"testing"
	"time"

	std_csv "encoding/csv"

	"github.com/josephcopenhaver/csv-go/v3"
)

func BenchmarkSTDWrite(b *testing.B) {
	b.ReportAllocs()

	cw := std_csv.NewWriter(io.Discard)

	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := cw.Write([]string{strconv.Itoa(-1), strconv.Itoa(-1)})
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkWrite(b *testing.B) {
	b.ReportAllocs()

	cw, err := csv.NewWriter(
		csv.WriterOpts().Writer(io.Discard),
		csv.WriterOpts().ErrorOnNonUTF8(false),
	)
	if err != nil {
		panic(err)
	}
	// defer cw.Close() // for the sake of the benchmark, calling explicitly and the end of the loop

	fw := csv.FieldWriters()

	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = cw.WriteFieldRow(
			fw.String("-1"),
			fw.String("-1"),
		)
		if err != nil {
			panic(err)
		}
	}

	_ = cw.Close()
}

func BenchmarkWriteWithSliceExpansion(b *testing.B) {
	//
	// To Reader/Author: if you find yourself using this reused row pattern, just use WriteFieldRowBorrowed instead
	//

	b.ReportAllocs()

	cw, err := csv.NewWriter(
		csv.WriterOpts().Writer(io.Discard),
	)
	if err != nil {
		panic(err)
	}
	// defer cw.Close() // for the sake of the benchmark, calling explicitly and the end of the loop

	fw := csv.FieldWriters()

	row := []csv.FieldWriter{
		fw.Int(-1),
		fw.Int(-1),
	}

	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = cw.WriteFieldRow(row...)
		if err != nil {
			panic(err)
		}
	}

	_ = cw.Close()
}

func BenchmarkWriteWithSliceBorrowed(b *testing.B) {
	b.ReportAllocs()

	cw, err := csv.NewWriter(
		csv.WriterOpts().Writer(io.Discard),
	)
	if err != nil {
		panic(err)
	}
	// defer cw.Close() // for the sake of the benchmark, calling explicitly and the end of the loop

	fw := csv.FieldWriters()
	row := []csv.FieldWriter{
		fw.Int(-1),
		fw.Int(-1),
	}

	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = cw.WriteFieldRowBorrowed(row)
		if err != nil {
			panic(err)
		}
	}

	_ = cw.Close()
}

func BenchmarkWriteWithQuotes(b *testing.B) {
	b.ReportAllocs()

	cw, err := csv.NewWriter(
		csv.WriterOpts().Writer(io.Discard),
		csv.WriterOpts().Quote('"'),
	)
	if err != nil {
		panic(err)
	}
	// defer cw.Close() // for the sake of the benchmark, calling explicitly and the end of the loop

	fw := csv.FieldWriters()

	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = cw.WriteFieldRow(
			fw.Int(-1),
			fw.Rune('"'),
		)
		if err != nil {
			panic(err)
		}
	}

	_ = cw.Close()
}

func BenchmarkFieldWriterAppendMinInt(b *testing.B) {
	buf := make([]byte, 0, 20)
	b.ReportAllocs()

	fw := csv.FieldWriters()

	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := fw.Int(math.MinInt)
		_, _ = f.AppendText(buf)
	}
}

func BenchmarkFieldWriterAppendMaxInt(b *testing.B) {
	buf := make([]byte, 0, 19)
	b.ReportAllocs()

	fw := csv.FieldWriters()

	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := fw.Int(math.MaxInt)
		_, _ = f.AppendText(buf)
	}
}

func BenchmarkFieldWriterAppendMinInt64(b *testing.B) {
	buf := make([]byte, 0, 20)
	b.ReportAllocs()

	fw := csv.FieldWriters()

	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := fw.Int64(math.MinInt64)
		_, _ = f.AppendText(buf)
	}
}

func BenchmarkFieldWriterAppendMaxInt64(b *testing.B) {
	buf := make([]byte, 0, 19)
	b.ReportAllocs()

	fw := csv.FieldWriters()

	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := fw.Int64(math.MaxInt64)
		_, _ = f.AppendText(buf)
	}
}

func BenchmarkFieldWriterAppendMaxUInt(b *testing.B) {
	buf := make([]byte, 0, 20)
	b.ReportAllocs()

	fw := csv.FieldWriters()

	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := fw.Uint64(math.MaxUint64)
		_, _ = f.AppendText(buf)
	}
}

func BenchmarkFieldWriterAppendMinUInt(b *testing.B) {
	buf := make([]byte, 0, 20)
	b.ReportAllocs()

	fw := csv.FieldWriters()

	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := fw.Uint64(0)
		_, _ = f.AppendText(buf)
	}
}

func BenchmarkFieldWriterAppendBytes(b *testing.B) {
	buf := make([]byte, 0, 5)
	b.ReportAllocs()

	fw := csv.FieldWriters()

	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := fw.Bytes([]byte(`12345`))
		_, _ = f.AppendText(buf)
	}
}

func BenchmarkFieldWriterAppendString(b *testing.B) {
	buf := make([]byte, 0, 5)
	b.ReportAllocs()

	fw := csv.FieldWriters()

	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := fw.String(`12345`)
		_, _ = f.AppendText(buf)
	}
}

func BenchmarkFieldWriterAppendTime(b *testing.B) {
	buf := make([]byte, 0, 35)
	b.ReportAllocs()

	fw := csv.FieldWriters()
	longestFormTime := time.Time{}.Add(1).In(time.FixedZone("unused", int(time.Hour/time.Second)))

	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := fw.Time(longestFormTime)
		_, _ = f.AppendText(buf)
	}
}

func BenchmarkFieldWriterAppendRune(b *testing.B) {
	buf := make([]byte, 0, 1)
	b.ReportAllocs()

	fw := csv.FieldWriters()

	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := fw.Rune('T')
		_, _ = f.AppendText(buf)
	}
}

func BenchmarkFieldWriterAppendBoolTrue(b *testing.B) {
	buf := make([]byte, 0, 1)
	b.ReportAllocs()

	fw := csv.FieldWriters()

	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := fw.Bool(true)
		_, _ = f.AppendText(buf)
	}
}

func BenchmarkFieldWriterAppendBoolFalse(b *testing.B) {
	buf := make([]byte, 0, 1)
	b.ReportAllocs()

	fw := csv.FieldWriters()

	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := fw.Bool(false)
		_, _ = f.AppendText(buf)
	}
}

func BenchmarkFieldWriterAppendDurationMax(b *testing.B) {
	buf := make([]byte, 0, 19)
	b.ReportAllocs()

	fw := csv.FieldWriters()

	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := fw.Duration(math.MaxInt64)
		_, _ = f.AppendText(buf)
	}
}

func BenchmarkFieldWriterAppendDurationMin(b *testing.B) {
	buf := make([]byte, 0, 20)
	b.ReportAllocs()

	fw := csv.FieldWriters()

	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := fw.Duration(math.MinInt64)
		_, _ = f.AppendText(buf)
	}
}

func BenchmarkFieldWriterAppendFloat64Min(b *testing.B) {
	buf := make([]byte, 0, 24)
	b.ReportAllocs()

	fw := csv.FieldWriters()

	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := fw.Float64(-math.MaxFloat64)
		_, _ = f.AppendText(buf)
	}
}

func BenchmarkFieldWriterAppendFloat64Max(b *testing.B) {
	buf := make([]byte, 0, 23)
	b.ReportAllocs()

	fw := csv.FieldWriters()

	runtime.GC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := fw.Float64(math.MaxFloat64)
		_, _ = f.AppendText(buf)
	}
}
