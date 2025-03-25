package csv

import (
	"bytes"
	"errors"
	"io"
	"unicode/utf8"
)

func (r *Reader) newPrepareRow(_ bool) func() bool {
	// TODO: make into generator based strategy selector and delete this whole file
	return r._prepareRow
}

func (r *Reader) handleEOF() bool {
	// r.done is always true when this function is called

	// check if we're in a terminal state otherwise error
	// there is no new character to process
	switch r.state {
	case rStateStartOfDoc:
		if (r.bitFlags & rFlagErrOnNoBOM) != 0 {
			r.parsingErr(errors.Join(ErrNoByteOrderMarker, io.ErrUnexpectedEOF))
			return false
		}

		r.state = rStateStartOfRecord // TODO: might be removable
		fallthrough
	case rStateStartOfRecord:
		if (r.bitFlags&rFlagTRSEmitsRecord) != 0 && r.numFields == 1 {
			r.fieldLengths = append(r.fieldLengths, 0)
			// field start is unchanged because the last one was zero length
			// r.fieldStart = len(r.recordBuf)
			return r.checkNumFields(io.ErrUnexpectedEOF)
		}
		return false
	case rStateInQuotedField, rStateInQuotedFieldAfterEscape:
		r.parsingErr(ErrIncompleteQuotedField)
		return false
	case rStateInLineComment:
		return false
	case rStateStartOfField:
		r.fieldLengths = append(r.fieldLengths, 0)
		// field start is unchanged because the last one was zero length
		// r.fieldStart = len(r.recordBuf)
		return r.checkNumFields(io.ErrUnexpectedEOF)
	case rStateEndOfQuotedField, rStateInField:
		r.fieldLengths = append(r.fieldLengths, len(r.recordBuf)-r.fieldStart)
		return r.checkNumFields(io.ErrUnexpectedEOF)
	}

	panic(panicUnknownReaderStateDuringEOF)
}

const (
	rMaxOverflowNumBytes = utf8.UTFMax - 1
	rMinRawBufSize       = utf8.UTFMax + rMaxOverflowNumBytes // TODO: ensure through various options the bounds are always greater than this min - also export so higher level layers can implement bounds validation if they need to.
)

func (r *Reader) _prepareRow() bool {
	var c rune
	var size uint8

	for {
		if len(r.rawBuf)+int(r.rawNumHiddenBytes)-r.rawIndex < rMinRawBufSize {
			var lastProcessedByte byte
			if r.rawIndex > 0 {
				lastProcessedByte = r.rawBuf[r.rawIndex-1]
			}

			copy(r.rawBuf[0:cap(r.rawBuf)], r.rawBuf[r.rawIndex:len(r.rawBuf)+int(r.rawNumHiddenBytes)])
			r.rawBuf = r.rawBuf[0 : len(r.rawBuf)+int(r.rawNumHiddenBytes)-r.rawIndex]
			r.rawIndex = 0
			r.rawNumHiddenBytes = 0

			if (r.bitFlags & stEOF) != 0 {
				if len(r.rawBuf) == 0 {
					r.setDone()

					// if CRLF is the record sep, no error has been thrown before now
					// and we've reached EOF with the last byte being a CR
					//
					// It's unsafe to assume that the field has ended correctly and that
					// the file has been generated reliably.
					//
					// In such cases where strict RFC compliance is enabled and CRLF is supported
					// this character along with LF should be encased in quotes and an error should
					// be raised.
					//
					// An argument could be made that this should be allowed when rFlagErrOnNLInUF
					// is off that this should also be off, but I will not be making that decision
					// without a stronger opinion. A pull request with strong justification or a new
					// option would be welcome here should you have a strong opinion dear reader.
					if lastProcessedByte == asciiCarriageReturn && r.recordSepLen == 2 {
						r.parsingErr(ErrUnsafeCRFileEnd)
						return false
					}

					return r.handleEOF()
				}
			} else if r.readErr != nil {
				if len(r.rawBuf) != 0 {
					r.setDone()
					r.ioErr(r.readErr)
					return false
				}
			} else {

				for {
					n, err := r.reader.Read(r.rawBuf[len(r.rawBuf):cap(r.rawBuf)])
					n += len(r.rawBuf)
					r.rawBuf = r.rawBuf[0:n]
					if err != nil {
						if errors.Is(err, io.EOF) {
							r.bitFlags |= stEOF
							if n == 0 {
								r.setDone()
								return r.handleEOF()
							}
							break
						}

						if n == 0 {
							r.setDone()
							r.ioErr(err)
							return false
						}

						r.readErr = err
						break
					}

					if n >= rMinRawBufSize {
						if r.rawBuf[n-1] == asciiCarriageReturn {
							if r.recordSepLen != 1 {
								r.rawBuf = r.rawBuf[0 : len(r.rawBuf)-1]
								r.rawNumHiddenBytes = 1
							}
							break
						}

						for i := 0; i <= int(rMaxOverflowNumBytes); i++ {
							if (r.rawBuf[n-i-1]&invalidControlRune) == 0 || endsInValidUTF8(r.rawBuf[n-rMinRawBufSize:n-i]) {
								r.rawBuf = r.rawBuf[0 : len(r.rawBuf)-i]
								r.rawNumHiddenBytes = uint8(i)
								break
							}
						}

						break
					}
				}
			}
		}

		for {
			di := bytes.IndexAny(r.rawBuf[r.rawIndex:], r.controlRunes)
			if di == -1 {
				// consume it all without adjustment

				switch r.state {
				case rStateStartOfDoc, rStateStartOfRecord, rStateStartOfField:
					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:]...)

					r.state = rStateInField
				case rStateInQuotedField, rStateInField:
					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:]...)

					// r.state = ... (unchanged)
				case rStateInQuotedFieldAfterEscape:
					r.setDone()
					r.parsingErr(ErrInvalidEscSeqInQuotedField)
					return false
				case rStateEndOfQuotedField:
					r.setDone()
					r.parsingErr(ErrInvalidQuotedFieldEnding)
					return false
				case rStateInLineComment:
					// TODO: zero out bytes

					// r.state = ... (unchanged)
				}

				r.byteIndex += uint64(len(r.rawBuf) - r.rawIndex)
				r.rawIndex = len(r.rawBuf)
				break
			}
			idx := r.rawIndex + di

			c = rune(r.rawBuf[idx])
			if (c & invalidControlRune) == 0 {
				size = 1
			} else {
				v, s := utf8.DecodeRune(r.rawBuf[idx:])
				if s == 1 {
					panic("DecodeRune failed")
				}
				c = v
				size = uint8(s)
			}

			// TODO: benchmark if skipping intermediate copies for signals not valid for a state saves time
			//
			// if it does then use multiple sets of runes for IndexAny operation

			switch c {
			case r.fieldSeparator:
				switch r.state {
				case rStateStartOfDoc, rStateStartOfRecord:
					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx]...)
					r.byteIndex += uint64(di) + uint64(size)

					if r.fieldNumOverflow() {
						return false
					}

					r.rawIndex = idx + int(size)
					r.fieldLengths = append(r.fieldLengths, len(r.recordBuf)-r.fieldStart)
					r.fieldStart = len(r.recordBuf)
					r.fieldIndex++

					r.state = rStateStartOfField
				case rStateInQuotedField:
					// TODO: technically "skippable"

					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx+int(size)]...)
					r.byteIndex += uint64(di) + uint64(size)
					r.rawIndex = idx + int(size)

					// r.state = ... (unchanged)
				case rStateInQuotedFieldAfterEscape:
					r.setDone()
					r.parsingErr(ErrInvalidEscSeqInQuotedField)
					return false
				case rStateEndOfQuotedField:
					if di != 0 {
						r.setDone()
						r.parsingErr(ErrInvalidQuotedFieldEnding)
						return false
					}

					r.byteIndex += uint64(size)
					r.rawIndex += int(size)

					if r.fieldNumOverflow() {
						return false
					}

					r.fieldLengths = append(r.fieldLengths, len(r.recordBuf)-r.fieldStart)
					r.fieldStart = len(r.recordBuf)
					r.fieldIndex++

					r.state = rStateStartOfField
				case rStateStartOfField:
					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx]...)
					r.byteIndex += uint64(di) + uint64(size)
					r.rawIndex = idx + int(size)

					if r.fieldNumOverflow() {
						return false
					}

					r.fieldLengths = append(r.fieldLengths, len(r.recordBuf)-r.fieldStart)
					r.fieldStart = len(r.recordBuf)
					r.fieldIndex++

					// r.state = ... (unchanged)
				case rStateInField:
					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx]...)
					r.byteIndex += uint64(di) + uint64(size)
					r.rawIndex = idx + int(size)

					if r.fieldNumOverflow() {
						return false
					}

					r.fieldLengths = append(r.fieldLengths, len(r.recordBuf)-r.fieldStart)
					r.fieldStart = len(r.recordBuf)
					r.fieldIndex++

					r.state = rStateStartOfField
				case rStateInLineComment:
					// TODO: zero out bytes
					r.byteIndex += uint64(di) + uint64(size)
					r.rawIndex = idx + int(size)
					continue
				}
			case r.escape:
				switch r.state {
				case rStateStartOfDoc, rStateStartOfRecord, rStateStartOfField:
					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx+int(size)]...)
					r.byteIndex += uint64(di) + uint64(size)
					r.rawIndex = idx + int(size)

					r.state = rStateInField
				case rStateInQuotedField:
					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx]...)
					r.byteIndex += uint64(di) + uint64(size)
					r.rawIndex = idx + int(size)

					r.state = rStateInQuotedFieldAfterEscape
				case rStateInQuotedFieldAfterEscape:
					if di != 0 {
						r.setDone()
						r.parsingErr(ErrInvalidEscSeqInQuotedField)
						return false
					}

					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx+int(size)]...)
					r.byteIndex += uint64(size)
					r.rawIndex += int(size)

					r.state = rStateInQuotedField
				case rStateEndOfQuotedField:
					r.setDone()
					r.parsingErr(ErrInvalidQuotedFieldEnding)
					return false
				case rStateInField:
					// TODO: technically "skippable"

					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx+int(size)]...)
					r.byteIndex += uint64(di) + uint64(size)
					r.rawIndex = idx + int(size)

					// r.state = ... (unchanged)
				case rStateInLineComment:
					// TODO: zero out bytes
					r.byteIndex += uint64(di) + uint64(size)
					r.rawIndex = idx + int(size)
					continue
				}
			case r.quote:
				switch r.state {
				case rStateStartOfDoc, rStateStartOfRecord:
					if di != 0 {
						if (r.bitFlags & rFlagErrOnQInUF) != 0 {
							r.byteIndex += uint64(di)

							r.state = rStateInField // TODO: might be removable

							r.setDone()
							r.parsingErr(ErrQuoteInUnquotedField)
							return false
						}

						r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx+int(size)]...)
						r.byteIndex += uint64(di) + uint64(size)
						r.rawIndex = idx + int(size)

						r.state = rStateInField
						continue
					}

					r.byteIndex += uint64(size)
					r.rawIndex += int(size)

					r.state = rStateInQuotedField
				case rStateInQuotedField:
					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx]...)
					r.byteIndex += uint64(di) + uint64(size)
					r.rawIndex = idx + int(size)

					r.state = rStateEndOfQuotedField
				case rStateInQuotedFieldAfterEscape:
					if di != 0 {
						r.setDone()
						r.parsingErr(ErrInvalidEscSeqInQuotedField)
						return false
					}

					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx+int(size)]...)
					r.byteIndex += uint64(size)
					r.rawIndex += int(size)

					r.state = rStateInQuotedField
				case rStateEndOfQuotedField:
					if di != 0 {
						r.setDone()
						r.parsingErr(ErrInvalidQuotedFieldEnding)
						return false
					}

					if (r.bitFlags & rFlagEscape) != 0 {
						r.setDone()
						r.parsingErr(ErrUnexpectedQuoteAfterField)
						return false
					}

					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx:idx+int(size)]...)
					r.byteIndex += uint64(size)
					r.rawIndex += int(size)

					r.state = rStateInQuotedField
				case rStateStartOfField:
					if di != 0 {
						if (r.bitFlags & rFlagErrOnQInUF) != 0 {
							r.byteIndex += uint64(di)
							r.setDone()
							r.parsingErr(ErrQuoteInUnquotedField)
							return false
						}

						r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx+int(size)]...)
						r.byteIndex += uint64(di) + uint64(size)
						r.rawIndex = idx + int(size)

						r.state = rStateInField
						continue
					}

					r.byteIndex += uint64(size)
					r.rawIndex += int(size)

					r.state = rStateInQuotedField
				case rStateInField:
					if (r.bitFlags & rFlagErrOnQInUF) != 0 {
						r.byteIndex += uint64(di)
						r.setDone()
						r.parsingErr(ErrQuoteInUnquotedField)
						return false
					}

					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx+int(size)]...)
					r.byteIndex += uint64(di) + uint64(size)
					r.rawIndex = idx + int(size)

					// r.state = ... (unchanged)
					continue
				case rStateInLineComment:
					// TODO: zero out bytes
					r.byteIndex += uint64(di) + uint64(size)
					r.rawIndex = idx + int(size)
					continue
				}
			case r.recordSep[0]:
				if r.recordSepLen == 2 {
					// checking for a full CRLF
					//
					// if not a CRLF sequence then just process the CR as field data

					if idx+int(size) >= len(r.rawBuf) || r.rawBuf[idx+int(size)] != asciiLineFeed {
						// definitely not a CRLF sequence, just an isolated CR byte
						// not followed by LF
						//
						// so treat as a field data byte

						switch r.state {
						case rStateStartOfDoc, rStateStartOfRecord, rStateStartOfField:
							r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:]...)

							r.state = rStateInField
						case rStateInQuotedField, rStateInField:
							r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:]...)

							// r.state = ... (unchanged)
						case rStateInQuotedFieldAfterEscape:
							r.setDone()
							r.parsingErr(ErrInvalidEscSeqInQuotedField)
							return false
						case rStateEndOfQuotedField:
							r.setDone()
							r.parsingErr(ErrInvalidQuotedFieldEnding)
							return false
						case rStateInLineComment:
							// TODO: zero out bytes

							// r.state = ... (unchanged)
						}

						r.byteIndex += uint64(di) + uint64(size)
						r.rawIndex = idx + int(size)

						continue
					}

					// we are handling a CRLF sequence
					// so increase size by the length of LF
					// and continue with record separator processing
					size++
				}
				switch r.state {
				case rStateStartOfDoc:
					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx]...)
					r.byteIndex += uint64(di) + uint64(size)
					r.rawIndex = idx + int(size)
					r.fieldLengths = append(r.fieldLengths, len(r.recordBuf)-r.fieldStart)

					r.state = rStateStartOfRecord
					return r.checkNumFields(nil)
				case rStateStartOfRecord:
					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx]...)
					r.byteIndex += uint64(di) + uint64(size)
					r.rawIndex = idx + int(size)
					r.fieldLengths = append(r.fieldLengths, len(r.recordBuf)-r.fieldStart)

					// r.state = ... (unchanged)
					return r.checkNumFields(nil)
				case rStateInQuotedField:
					// TODO: technically "skippable"

					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx+int(size)]...)
					r.byteIndex += uint64(di) + uint64(size)
					r.rawIndex = idx + int(size)

					// r.state = ... (unchanged)
				case rStateInQuotedFieldAfterEscape:
					r.setDone()
					r.parsingErr(ErrInvalidEscSeqInQuotedField)
					return false
				case rStateEndOfQuotedField:
					if di != 0 {
						r.setDone()
						r.parsingErr(ErrInvalidQuotedFieldEnding)
						return false
					}

					r.byteIndex += uint64(size)
					r.rawIndex += int(size)
					r.fieldLengths = append(r.fieldLengths, len(r.recordBuf)-r.fieldStart)

					r.state = rStateStartOfRecord
					return r.checkNumFields(nil)
				case rStateStartOfField, rStateInField:
					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx]...)
					r.byteIndex += uint64(di) + uint64(size)
					r.rawIndex = idx + int(size)
					r.fieldLengths = append(r.fieldLengths, len(r.recordBuf)-r.fieldStart)

					r.state = rStateStartOfRecord
					return r.checkNumFields(nil)
				case rStateInLineComment:
					// TODO: zero out bytes
					r.byteIndex += uint64(di) + uint64(size)
					r.rawIndex = idx + int(size)

					r.state = rStateStartOfRecord
				}
			default:
				if r.recordSepLen != 0 {
					// record separator detection is disabled or already hardened
					//
					// must have found a CR or LF character under circumstances where we're aiming to error
					// if discovered outside of a quoted state
					switch r.state {
					case rStateStartOfDoc:
						r.state = rStateStartOfRecord // TODO: might be removable
						fallthrough
					case rStateStartOfRecord, rStateStartOfField:
						if di > 0 {
							r.state = rStateInField // TODO: might be removable
							r.byteIndex += uint64(di)
						}
						r.setDone()
						if c == asciiLineFeed {
							r.parsingErr(errNewlineInUnquotedFieldLineFeed)
						} else {
							r.parsingErr(errNewlineInUnquotedFieldCarriageReturn)
						}
						return false
					case rStateInField:
						r.byteIndex += uint64(di)
						r.setDone()
						if c == asciiLineFeed {
							r.parsingErr(errNewlineInUnquotedFieldLineFeed)
						} else {
							r.parsingErr(errNewlineInUnquotedFieldCarriageReturn)
						}
						return false
					case rStateInQuotedField:
						r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx+int(size)]...)
						r.byteIndex += uint64(di) + uint64(size)
						r.rawIndex = idx + int(size)

						// r.state = ... (unchanged)
					case rStateInQuotedFieldAfterEscape:
						r.setDone()
						r.parsingErr(ErrInvalidEscSeqInQuotedField)
						return false
					case rStateEndOfQuotedField:
						r.setDone()
						r.parsingErr(ErrInvalidQuotedFieldEnding)
						return false
					case rStateInLineComment:
						// TODO: zero out bytes
						r.byteIndex += uint64(di) + uint64(size)
						r.rawIndex = idx + int(size)
						// continue // can skip, very next instruction after the switch is this
					}

					continue
				}

				//
				// record separator discovery handling block
				//
				// c contains the first rune of the record separator sequence
				//
				// only CRLF is a valid two-rune sequence, all others are one rune
				//

				switch r.state {
				case rStateStartOfDoc, rStateStartOfRecord, rStateEndOfQuotedField, rStateInLineComment, rStateStartOfField, rStateInField:
					if c == asciiCarriageReturn && idx+1 < len(r.rawBuf) && r.rawBuf[idx+1] == asciiLineFeed {
						r.recordSepLen = 2
						r.recordSep[0] = asciiCarriageReturn
						r.recordSep[1] = asciiLineFeed
					} else {
						r.recordSepLen = 1
						r.recordSep[0] = c
					}

					// preserve field separator
					var buf [7]rune
					controlRunes := append(buf[:0], r.fieldSeparator)
					controlRunes = append(controlRunes, c)

					if (r.bitFlags & rFlagQuote) != 0 {
						controlRunes = append(controlRunes, r.quote)
					}
					if (r.bitFlags & rFlagEscape) != 0 {
						controlRunes = append(controlRunes, r.escape)
					}
					if (r.bitFlags & rFlagComment) != 0 {
						controlRunes = append(controlRunes, r.comment)
					}

					if (r.bitFlags & rFlagErrOnNLInUF) != 0 {
						crs := []byte(string(controlRunes))

						if !bytes.Contains(crs, []byte{asciiCarriageReturn}) {
							controlRunes = append(controlRunes, asciiCarriageReturn)
						}

						if !bytes.Contains(crs, []byte{asciiLineFeed}) {
							controlRunes = append(controlRunes, asciiLineFeed)
						}
					}

					r.controlRunes = string(controlRunes)

					// r.state = ... (unchanged)
				case rStateInQuotedField:
					// TODO: technically "skippable"

					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx+int(size)]...)
					r.byteIndex += uint64(di) + uint64(size)
					r.rawIndex = idx + int(size)

					// r.state = ... (unchanged)
				case rStateInQuotedFieldAfterEscape:
					r.setDone()
					r.parsingErr(ErrInvalidEscSeqInQuotedField)
					return false
				}
			}
		}
	}
}

func endsInValidUTF8(p []byte) bool {
	r, s := utf8.DecodeLastRune(p)
	return (r != utf8.RuneError || s > 1)
}
