package csv_test

import (
	std_csv "encoding/csv"
	"io"
	"strings"
	"testing"

	"github.com/josephcopenhaver/csv-go/v2"
)

const fileContents3c256Rows = `a,b,c
1,odd,r1
2,even,r2
3,odd,r3
4,even,r4
5,odd,r5
6,even,r6
7,odd,r7
8,even,r8
9,odd,r9
10,even,r10
11,odd,r11
12,even,r12
13,odd,r13
14,even,r14
15,odd,r15
16,even,r16
17,odd,r17
18,even,r18
19,odd,r19
20,even,r20
21,odd,r21
22,even,r22
23,odd,r23
24,even,r24
25,odd,r25
26,even,r26
27,odd,r27
28,even,r28
29,odd,r29
30,even,r30
31,odd,r31
32,even,r32
33,odd,r33
34,even,r34
35,odd,r35
36,even,r36
37,odd,r37
38,even,r38
39,odd,r39
40,even,r40
41,odd,r41
42,even,r42
43,odd,r43
44,even,r44
45,odd,r45
46,even,r46
47,odd,r47
48,even,r48
49,odd,r49
50,even,r50
51,odd,r51
52,even,r52
53,odd,r53
54,even,r54
55,odd,r55
56,even,r56
57,odd,r57
58,even,r58
59,odd,r59
60,even,r60
61,odd,r61
62,even,r62
63,odd,r63
64,even,r64
65,odd,r65
66,even,r66
67,odd,r67
68,even,r68
69,odd,r69
70,even,r70
71,odd,r71
72,even,r72
73,odd,r73
74,even,r74
75,odd,r75
76,even,r76
77,odd,r77
78,even,r78
79,odd,r79
80,even,r80
81,odd,r81
82,even,r82
83,odd,r83
84,even,r84
85,odd,r85
86,even,r86
87,odd,r87
88,even,r88
89,odd,r89
90,even,r90
91,odd,r91
92,even,r92
93,odd,r93
94,even,r94
95,odd,r95
96,even,r96
97,odd,r97
98,even,r98
99,odd,r99
100,even,r100
101,odd,r101
102,even,r102
103,odd,r103
104,even,r104
105,odd,r105
106,even,r106
107,odd,r107
108,even,r108
109,odd,r109
110,even,r110
111,odd,r111
112,even,r112
113,odd,r113
114,even,r114
115,odd,r115
116,even,r116
117,odd,r117
118,even,r118
119,odd,r119
120,even,r120
121,odd,r121
122,even,r122
123,odd,r123
124,even,r124
125,odd,r125
126,even,r126
127,odd,r127
128,even,r128
129,odd,r129
130,even,r130
131,odd,r131
132,even,r132
133,odd,r133
134,even,r134
135,odd,r135
136,even,r136
137,odd,r137
138,even,r138
139,odd,r139
140,even,r140
141,odd,r141
142,even,r142
143,odd,r143
144,even,r144
145,odd,r145
146,even,r146
147,odd,r147
148,even,r148
149,odd,r149
150,even,r150
151,odd,r151
152,even,r152
153,odd,r153
154,even,r154
155,odd,r155
156,even,r156
157,odd,r157
158,even,r158
159,odd,r159
160,even,r160
161,odd,r161
162,even,r162
163,odd,r163
164,even,r164
165,odd,r165
166,even,r166
167,odd,r167
168,even,r168
169,odd,r169
170,even,r170
171,odd,r171
172,even,r172
173,odd,r173
174,even,r174
175,odd,r175
176,even,r176
177,odd,r177
178,even,r178
179,odd,r179
180,even,r180
181,odd,r181
182,even,r182
183,odd,r183
184,even,r184
185,odd,r185
186,even,r186
187,odd,r187
188,even,r188
189,odd,r189
190,even,r190
191,odd,r191
192,even,r192
193,odd,r193
194,even,r194
195,odd,r195
196,even,r196
197,odd,r197
198,even,r198
199,odd,r199
200,even,r200
201,odd,r201
202,even,r202
203,odd,r203
204,even,r204
205,odd,r205
206,even,r206
207,odd,r207
208,even,r208
209,odd,r209
210,even,r210
211,odd,r211
212,even,r212
213,odd,r213
214,even,r214
215,odd,r215
216,even,r216
217,odd,r217
218,even,r218
219,odd,r219
220,even,r220
221,odd,r221
222,even,r222
223,odd,r223
224,even,r224
225,odd,r225
226,even,r226
227,odd,r227
228,even,r228
229,odd,r229
230,even,r230
231,odd,r231
232,even,r232
233,odd,r233
234,even,r234
235,odd,r235
236,even,r236
237,odd,r237
238,even,r238
239,odd,r239
240,even,r240
241,odd,r241
242,even,r242
243,odd,r243
244,even,r244
245,odd,r245
246,even,r246
247,odd,r247
248,even,r248
249,odd,r249
250,even,r250
251,odd,r251
252,even,r252
253,odd,r253
254,even,r254
255,odd,r255
256,even,r256
`

func BenchmarkRead256Rows(b *testing.B) {

	strReader := strings.NewReader(fileContents3c256Rows)
	opts := csv.ReaderOpts()

	for b.Loop() {
		if _, err := strReader.Seek(0, io.SeekStart); err != nil {
			panic(err)
		}
		cr, err := csv.NewReader(
			opts.Reader(strReader),
		)
		if err != nil {
			panic(err)
		}
		defer cr.Close()

		for row := range cr.IntoIter() {
			_ = row
		}
		if err := cr.Err(); err != nil {
			panic(err)
		}
	}
}

func BenchmarkSTDRead256Rows(b *testing.B) {

	strReader := strings.NewReader(fileContents3c256Rows)

	for b.Loop() {
		if _, err := strReader.Seek(0, io.SeekStart); err != nil {
			panic(err)
		}
		cr := std_csv.NewReader(strReader)

		for {
			row, err := cr.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				panic(err)
			}
			_ = row
		}
	}
}

func BenchmarkRead256RowsBorrowRow(b *testing.B) {

	strReader := strings.NewReader(fileContents3c256Rows)
	opts := csv.ReaderOpts()

	for b.Loop() {
		if _, err := strReader.Seek(0, io.SeekStart); err != nil {
			panic(err)
		}
		cr, err := csv.NewReader(
			opts.Reader(strReader),
			opts.BorrowRow(true),
		)
		if err != nil {
			panic(err)
		}
		defer cr.Close()

		for row := range cr.IntoIter() {
			_ = row
		}
		if err := cr.Err(); err != nil {
			panic(err)
		}
	}
}

func BenchmarkSTDRead256RowsBorrowRow(b *testing.B) {

	strReader := strings.NewReader(fileContents3c256Rows)

	for b.Loop() {
		if _, err := strReader.Seek(0, io.SeekStart); err != nil {
			panic(err)
		}
		cr := std_csv.NewReader(strReader)
		cr.ReuseRecord = true

		for {
			row, err := cr.Read()
			if err != nil {
				if err == io.EOF {
					break
				}
				panic(err)
			}
			_ = row
		}
	}
}

func BenchmarkRead256RowsBorrowRowBorrowFields(b *testing.B) {

	strReader := strings.NewReader(fileContents3c256Rows)
	opts := csv.ReaderOpts()

	for b.Loop() {
		if _, err := strReader.Seek(0, io.SeekStart); err != nil {
			panic(err)
		}
		cr, err := csv.NewReader(
			opts.Reader(strReader),
			opts.BorrowRow(true),
			opts.BorrowFields(true),
		)
		if err != nil {
			panic(err)
		}
		defer cr.Close()

		for row := range cr.IntoIter() {
			_ = row
		}
		if err := cr.Err(); err != nil {
			panic(err)
		}
	}
}

func BenchmarkRead256RowsBorrowRowBorrowFieldsRecBuf(b *testing.B) {

	strReader := strings.NewReader(fileContents3c256Rows)
	opts := csv.ReaderOpts()
	var recBuf [32]byte

	for b.Loop() {
		if _, err := strReader.Seek(0, io.SeekStart); err != nil {
			panic(err)
		}
		cr, err := csv.NewReader(
			opts.Reader(strReader),
			opts.BorrowRow(true),
			opts.BorrowFields(true),
			opts.InitialRecordBuffer(recBuf[:]),
		)
		if err != nil {
			panic(err)
		}
		defer cr.Close()

		for row := range cr.IntoIter() {
			_ = row
		}
		if err := cr.Err(); err != nil {
			panic(err)
		}
	}
}

func BenchmarkRead256RowsBorrowRowBorrowFieldsRecBufReadBuf(b *testing.B) {

	strReader := strings.NewReader(fileContents3c256Rows)
	opts := csv.ReaderOpts()
	var recBuf [32]byte
	var readerBuf [32]byte

	for b.Loop() {
		if _, err := strReader.Seek(0, io.SeekStart); err != nil {
			panic(err)
		}
		cr, err := csv.NewReader(
			opts.Reader(strReader),
			opts.BorrowRow(true),
			opts.BorrowFields(true),
			opts.InitialRecordBuffer(recBuf[:]),
			opts.ReaderBuffer(readerBuf[:]),
		)
		if err != nil {
			panic(err)
		}
		defer cr.Close()

		for row := range cr.IntoIter() {
			_ = row
		}
		if err := cr.Err(); err != nil {
			panic(err)
		}
	}
}

func BenchmarkRead256RowsBorrowRowBorrowFieldsRecBufReadBufRecSepLF(b *testing.B) {

	strReader := strings.NewReader(fileContents3c256Rows)
	opts := csv.ReaderOpts()
	var recBuf [32]byte
	var readerBuf [32]byte

	for b.Loop() {
		if _, err := strReader.Seek(0, io.SeekStart); err != nil {
			panic(err)
		}
		cr, err := csv.NewReader(
			opts.Reader(strReader),
			opts.BorrowRow(true),
			opts.BorrowFields(true),
			opts.InitialRecordBuffer(recBuf[:]),
			opts.ReaderBuffer(readerBuf[:]),
			opts.RecordSeparator("\n"),
		)
		if err != nil {
			panic(err)
		}
		defer cr.Close()

		for row := range cr.IntoIter() {
			_ = row
		}
		if err := cr.Err(); err != nil {
			panic(err)
		}
	}
}
