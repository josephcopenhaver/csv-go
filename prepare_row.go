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

		r.state = rStateStartOfRecord
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

func (r *Reader) _prepareRow() bool {
	var c rune
	var size uint8

	for {
		if len(r.rawBuf) == r.rawIndex {
			if (r.bitFlags & stEOF) != 0 {
				r.setDone()
				return r.handleEOF()
			}
			if r.readErr != nil {
				r.setDone()
				r.ioErr(r.readErr)
				return false
			}

			for {
				n, err := r.reader.Read(r.rawBuf[0:cap(r.rawBuf)])
				r.rawIndex = 0
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

				if n != 0 {
					break
				}
			}
		}

		for {
			di := bytes.IndexAny(r.rawBuf[r.rawIndex:], r.controlRunes)
			if di == -1 {
				if (r.rawBuf[len(r.rawBuf)-1]&invalidControlRune) == 0 || endsInValidUTF8(r.rawBuf[r.rawIndex:]) {
					// ends in ascii byte or valid utf8, so a multi-byte character is not split awkwardly
					//
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

					r.byteIndex += uint64(di + 1)
					r.rawIndex = len(r.rawBuf)
					break
				}
				panic("chunk ends in multi-byte char") // TODO: implement
				// address this by making sure at least 4 bytes remain unprocessed at all times
				//
				// note if doing the above it might be best to use another "edge buffer" to avoid
				// scenarios where users provide the buffer-space they want to utilize in some fashion
				// (we support both the slice and the slice size)
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

					// r.state = rStateInQuotedField
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

					// r.state = rStateStartOfField
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
				panic("not implemented: handle escape") // TODO: implement
			case r.quote:
				switch r.state {
				case rStateStartOfDoc, rStateStartOfRecord:
					if di != 0 {
						panic("quote not escaped before start of first field of record") // TODO: handle
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
					panic("not implemented: handle after-escape state") // TODO: implement
				case rStateEndOfQuotedField:
					if di != 0 {
						r.setDone()
						r.parsingErr(ErrInvalidQuotedFieldEnding)
						return false
					}

					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx:idx+int(size)]...)
					r.byteIndex++
					r.rawIndex++

					r.state = rStateInQuotedField
				case rStateStartOfField:
					if di != 0 {
						r.byteIndex += uint64(di)
						r.setDone()
						r.parsingErr(ErrQuoteInUnquotedField)
						return false
					}

					r.byteIndex += uint64(size)
					r.rawIndex += int(size)

					r.state = rStateInQuotedField
				case rStateInField:
					r.byteIndex += uint64(di)
					r.setDone()
					r.parsingErr(ErrQuoteInUnquotedField)
					return false
				case rStateInLineComment:

					// TODO: zero out bytes

					r.byteIndex += uint64(di) + uint64(size)
					r.rawIndex = idx + int(size)
					continue
				}
			case r.recordSep[0]:
				if r.recordSepLen == 2 {
					// c is definitely CR
					//
					// if next exists and is LF then process as record sep
					//
					// otherwise process as just a data rune
					// continue
					panic("not implemented yet") // TODO: implement
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

					// r.state = rStateStartOfRecord
					return r.checkNumFields(nil)
				case rStateInQuotedField:
					// TODO: technically "skippable"

					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx+int(size)]...)
					r.byteIndex += uint64(di) + uint64(size)
					r.rawIndex = idx + int(size)
				case rStateInQuotedFieldAfterEscape:
					panic("not implemented: handle after-escape state") // TODO: implement
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
						r.state = rStateStartOfRecord
						fallthrough
					case rStateStartOfRecord, rStateStartOfField:
						if di > 0 {
							r.state = rStateInField
							r.byteIndex += uint64(di)
						}
						r.setDone()
						if c == '\n' {
							r.parsingErr(errNewlineInUnquotedFieldLineFeed)
						} else {
							r.parsingErr(errNewlineInUnquotedFieldCarriageReturn)
						}
						return false
					case rStateInField:
						r.byteIndex += uint64(di)
						r.setDone()
						if c == '\n' {
							r.parsingErr(errNewlineInUnquotedFieldLineFeed)
						} else {
							r.parsingErr(errNewlineInUnquotedFieldCarriageReturn)
						}
						return false
					case rStateInQuotedField:
						r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx+int(size)]...)
						r.byteIndex += uint64(di) + uint64(size)
						r.rawIndex = idx + int(size)
					case rStateInQuotedFieldAfterEscape:
						panic("not implemented: handle after-escape state") // TODO: implement
					case rStateEndOfQuotedField:
						r.setDone()
						r.parsingErr(ErrInvalidQuotedFieldEnding)
						return false
					case rStateInLineComment:
						r.byteIndex += uint64(di) + uint64(size)
						r.rawIndex = idx + int(size)
					}

					continue
				}

				// encountered a CR or LF character when record separator detection is automatically enabled

				switch c {
				case '\r':
					panic("CR or CRLF record sep discovery not implemented") // TODO: implement
				case '\n':
				default:
					panic("discovering record sep and found a non CR LF rune")
				}

				switch r.state {
				case rStateStartOfDoc, rStateStartOfRecord, rStateEndOfQuotedField, rStateInLineComment, rStateStartOfField, rStateInField:
					r.recordSepLen = 1
					r.recordSep[0] = c
				case rStateInQuotedField:
					// TODO: technically "skippable"

					r.recordBuf = append(r.recordBuf, r.rawBuf[r.rawIndex:idx+int(size)]...)
					r.byteIndex += uint64(di) + uint64(size)
					r.rawIndex = idx + int(size)
				case rStateInQuotedFieldAfterEscape:
					panic("not implemented: handle after-escape state") // TODO: implement
				}
			}
		}
	}
}

func endsInValidUTF8(p []byte) bool {
	r, s := utf8.DecodeLastRune(p)
	return (r != utf8.RuneError || s > 1)
}
