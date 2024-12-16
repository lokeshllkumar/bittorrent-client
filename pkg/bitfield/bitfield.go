package bitfield

type Bitfield []byte

func (bf Bitfield) HasPiece(ind int) bool {
	byteInd := ind / 8
	offset := ind % 8
	if byteInd < 0 || byteInd >= len(bf) {
		return false
	}

	return bf[byteInd]>>uint(7-offset)&1 != 0
}

func (bf Bitfield) SetPiece(ind int) {
	byteInd := ind / 8
	offset := ind % 8

	if byteInd < 0 || byteInd >= len(bf) {
		return
	}

	bf[byteInd] |= 1 << uint(7-offset)
}
