package csv

// updatePrepareRow acts as glue from generated code for a unique configuration set
// to a specific strategy of preparing a row.
func (r *Reader) updatePrepareRow(clearMemAfterUse bool) {
	if !clearMemAfterUse {
		if (r.bitFlags & rFlagErrOnNLInUF) == 0 {
			if (r.bitFlags & rFlagErrOnQInUF) == 0 {
				if (r.bitFlags & rFlagComment) == 0 {
					if (r.bitFlags & rFlagDropBOM) == 0 {
						if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
							if (r.bitFlags & rFlagQuote) == 0 {
								if r.recordSepLen < 1 {
									r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
								} else if r.recordSepLen == 1 {
									r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
								} else {
									r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
								}
							} else if (r.bitFlags & rFlagEscape) == 0 {
								if r.recordSepLen < 1 {
									r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
								} else if r.recordSepLen == 1 {
									r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
								} else {
									r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
								}
							} else if r.recordSepLen < 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
							} else if r.recordSepLen == 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
							} else {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
							}
						} else if (r.bitFlags & rFlagQuote) == 0 {
							if r.recordSepLen < 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
							} else if r.recordSepLen == 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
							} else {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
							}
						} else if (r.bitFlags & rFlagEscape) == 0 {
							if r.recordSepLen < 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
							} else if r.recordSepLen == 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
							} else {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
							}
						} else if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
						if (r.bitFlags & rFlagQuote) == 0 {
							if r.recordSepLen < 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
							} else if r.recordSepLen == 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
							} else {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
							}
						} else if (r.bitFlags & rFlagEscape) == 0 {
							if r.recordSepLen < 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
							} else if r.recordSepLen == 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
							} else {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
							}
						} else if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if (r.bitFlags & rFlagQuote) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if (r.bitFlags & rFlagEscape) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagDropBOM) == 0 {
					if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
						if (r.bitFlags & rFlagQuote) == 0 {
							if r.recordSepLen < 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
							} else if r.recordSepLen == 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
							} else {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
							}
						} else if (r.bitFlags & rFlagEscape) == 0 {
							if r.recordSepLen < 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
							} else if r.recordSepLen == 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
							} else {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
							}
						} else if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if (r.bitFlags & rFlagQuote) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if (r.bitFlags & rFlagEscape) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
					if (r.bitFlags & rFlagQuote) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if (r.bitFlags & rFlagEscape) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagQuote) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagEscape) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagComment) == 0 {
				if (r.bitFlags & rFlagDropBOM) == 0 {
					if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
						if (r.bitFlags & rFlagQuote) == 0 {
							if r.recordSepLen < 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
							} else if r.recordSepLen == 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
							} else {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
							}
						} else if (r.bitFlags & rFlagEscape) == 0 {
							if r.recordSepLen < 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
							} else if r.recordSepLen == 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
							} else {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
							}
						} else if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if (r.bitFlags & rFlagQuote) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if (r.bitFlags & rFlagEscape) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
					if (r.bitFlags & rFlagQuote) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if (r.bitFlags & rFlagEscape) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagQuote) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagEscape) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagDropBOM) == 0 {
				if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
					if (r.bitFlags & rFlagQuote) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if (r.bitFlags & rFlagEscape) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagQuote) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagEscape) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
				if (r.bitFlags & rFlagQuote) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagEscape) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagQuote) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagEscape) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if (r.bitFlags & rFlagErrOnQInUF) == 0 {
			if (r.bitFlags & rFlagComment) == 0 {
				if (r.bitFlags & rFlagDropBOM) == 0 {
					if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
						if (r.bitFlags & rFlagQuote) == 0 {
							if r.recordSepLen < 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
							} else if r.recordSepLen == 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
							} else {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
							}
						} else if (r.bitFlags & rFlagEscape) == 0 {
							if r.recordSepLen < 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
							} else if r.recordSepLen == 1 {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
							} else {
								r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
							}
						} else if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if (r.bitFlags & rFlagQuote) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if (r.bitFlags & rFlagEscape) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
					if (r.bitFlags & rFlagQuote) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if (r.bitFlags & rFlagEscape) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagQuote) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagEscape) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagDropBOM) == 0 {
				if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
					if (r.bitFlags & rFlagQuote) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if (r.bitFlags & rFlagEscape) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagQuote) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagEscape) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
				if (r.bitFlags & rFlagQuote) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagEscape) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagQuote) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagEscape) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if (r.bitFlags & rFlagComment) == 0 {
			if (r.bitFlags & rFlagDropBOM) == 0 {
				if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
					if (r.bitFlags & rFlagQuote) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if (r.bitFlags & rFlagEscape) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagQuote) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagEscape) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
				if (r.bitFlags & rFlagQuote) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagEscape) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagQuote) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagEscape) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if (r.bitFlags & rFlagDropBOM) == 0 {
			if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
				if (r.bitFlags & rFlagQuote) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagEscape) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagQuote) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagEscape) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
			if (r.bitFlags & rFlagQuote) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagEscape) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if (r.bitFlags & rFlagQuote) == 0 {
			if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if (r.bitFlags & rFlagEscape) == 0 {
			if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if r.recordSepLen < 1 {
			r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
		} else if r.recordSepLen == 1 {
			r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
		} else {
			r.prepareRow = r.prepareRow_memclearOff_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
		}
	} else if (r.bitFlags & rFlagErrOnNLInUF) == 0 {
		if (r.bitFlags & rFlagErrOnQInUF) == 0 {
			if (r.bitFlags & rFlagComment) == 0 {
				if (r.bitFlags & rFlagDropBOM) == 0 {
					if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
						if (r.bitFlags & rFlagQuote) == 0 {
							if r.recordSepLen < 1 {
								r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
							} else if r.recordSepLen == 1 {
								r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
							} else {
								r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
							}
						} else if (r.bitFlags & rFlagEscape) == 0 {
							if r.recordSepLen < 1 {
								r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
							} else if r.recordSepLen == 1 {
								r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
							} else {
								r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
							}
						} else if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if (r.bitFlags & rFlagQuote) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if (r.bitFlags & rFlagEscape) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
					if (r.bitFlags & rFlagQuote) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if (r.bitFlags & rFlagEscape) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagQuote) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagEscape) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagDropBOM) == 0 {
				if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
					if (r.bitFlags & rFlagQuote) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if (r.bitFlags & rFlagEscape) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagQuote) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagEscape) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
				if (r.bitFlags & rFlagQuote) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagEscape) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagQuote) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagEscape) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if (r.bitFlags & rFlagComment) == 0 {
			if (r.bitFlags & rFlagDropBOM) == 0 {
				if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
					if (r.bitFlags & rFlagQuote) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if (r.bitFlags & rFlagEscape) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagQuote) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagEscape) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
				if (r.bitFlags & rFlagQuote) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagEscape) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagQuote) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagEscape) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if (r.bitFlags & rFlagDropBOM) == 0 {
			if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
				if (r.bitFlags & rFlagQuote) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagEscape) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagQuote) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagEscape) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
			if (r.bitFlags & rFlagQuote) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagEscape) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if (r.bitFlags & rFlagQuote) == 0 {
			if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if (r.bitFlags & rFlagEscape) == 0 {
			if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if r.recordSepLen < 1 {
			r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
		} else if r.recordSepLen == 1 {
			r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
		} else {
			r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOff_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
		}
	} else if (r.bitFlags & rFlagErrOnQInUF) == 0 {
		if (r.bitFlags & rFlagComment) == 0 {
			if (r.bitFlags & rFlagDropBOM) == 0 {
				if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
					if (r.bitFlags & rFlagQuote) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if (r.bitFlags & rFlagEscape) == 0 {
						if r.recordSepLen < 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
						} else if r.recordSepLen == 1 {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
						} else {
							r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
						}
					} else if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagQuote) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagEscape) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
				if (r.bitFlags & rFlagQuote) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagEscape) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagQuote) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagEscape) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if (r.bitFlags & rFlagDropBOM) == 0 {
			if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
				if (r.bitFlags & rFlagQuote) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagEscape) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagQuote) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagEscape) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
			if (r.bitFlags & rFlagQuote) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagEscape) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if (r.bitFlags & rFlagQuote) == 0 {
			if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if (r.bitFlags & rFlagEscape) == 0 {
			if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if r.recordSepLen < 1 {
			r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
		} else if r.recordSepLen == 1 {
			r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
		} else {
			r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOff_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
		}
	} else if (r.bitFlags & rFlagComment) == 0 {
		if (r.bitFlags & rFlagDropBOM) == 0 {
			if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
				if (r.bitFlags & rFlagQuote) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if (r.bitFlags & rFlagEscape) == 0 {
					if r.recordSepLen < 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
					} else if r.recordSepLen == 1 {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
					} else {
						r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
					}
				} else if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagQuote) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagEscape) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
			if (r.bitFlags & rFlagQuote) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagEscape) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if (r.bitFlags & rFlagQuote) == 0 {
			if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if (r.bitFlags & rFlagEscape) == 0 {
			if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if r.recordSepLen < 1 {
			r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
		} else if r.recordSepLen == 1 {
			r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
		} else {
			r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOff_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
		}
	} else if (r.bitFlags & rFlagDropBOM) == 0 {
		if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
			if (r.bitFlags & rFlagQuote) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if (r.bitFlags & rFlagEscape) == 0 {
				if r.recordSepLen < 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
				} else if r.recordSepLen == 1 {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
				} else {
					r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
				}
			} else if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if (r.bitFlags & rFlagQuote) == 0 {
			if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if (r.bitFlags & rFlagEscape) == 0 {
			if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if r.recordSepLen < 1 {
			r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
		} else if r.recordSepLen == 1 {
			r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
		} else {
			r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOff_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
		}
	} else if (r.bitFlags & rFlagErrOnNoBOM) == 0 {
		if (r.bitFlags & rFlagQuote) == 0 {
			if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if (r.bitFlags & rFlagEscape) == 0 {
			if r.recordSepLen < 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
			} else if r.recordSepLen == 1 {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
			} else {
				r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
			}
		} else if r.recordSepLen < 1 {
			r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
		} else if r.recordSepLen == 1 {
			r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
		} else {
			r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOff_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
		}
	} else if (r.bitFlags & rFlagQuote) == 0 {
		if r.recordSepLen < 1 {
			r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOff
		} else if r.recordSepLen == 1 {
			r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOn_2RuneRecSepOff
		} else {
			r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOff_escapeOff_1RuneRecSepOff_2RuneRecSepOn
		}
	} else if (r.bitFlags & rFlagEscape) == 0 {
		if r.recordSepLen < 1 {
			r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOff
		} else if r.recordSepLen == 1 {
			r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOn_2RuneRecSepOff
		} else {
			r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOff_1RuneRecSepOff_2RuneRecSepOn
		}
	} else if r.recordSepLen < 1 {
		r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOff
	} else if r.recordSepLen == 1 {
		r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOn_2RuneRecSepOff
	} else {
		r.prepareRow = r.prepareRow_memclearOn_errOnNLInUFOn_errOnQInUFOn_commentOn_dropBOMOn_errOnNoBOMOn_quoteOn_escapeOn_1RuneRecSepOff_2RuneRecSepOn
	}
}
