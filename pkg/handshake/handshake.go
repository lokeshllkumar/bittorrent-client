package handshake

import (
	"fmt"
	"io"
)

type Handshake struct {
	Pstr string
	InfoHash [20]byte
	PeerID   [20]byte
}

func NewHandshake(infoHash [20]byte, peerID [20]byte) *Handshake {
	return &Handshake {
		Pstr: "BitTorrent Protocol",
		InfoHash: infoHash,
		PeerID: peerID,
	}
}

func (h *Handshake) Serialize() []byte {
	buffer := make([]byte, len(h.Pstr) + 49)
	buffer[0] = byte(len(h.Pstr))
	cur := 1
	cur += copy(buffer[cur:], []byte(h.Pstr))
	cur += copy(buffer[cur:], make([]byte, 8))
	cur += copy(buffer[cur:], h.InfoHash[:])
	cur += copy(buffer[cur:], h.PeerID[:])

	return buffer
}

func Read(r io.Reader) (*Handshake, error) {
	lengthBuffer := make([]byte, 1)
	_, err := io.ReadFull(r, lengthBuffer)
	if err != nil {
		return nil, err
	}

	pstrlen := int(lengthBuffer[0])

	if pstrlen == 0 {
		err := fmt.Errorf("pstrlen cannot be 0")
		return nil, err
	}

	handshakeBuffer := make([]byte, pstrlen + 48)
	_, err = io.ReadFull(r, handshakeBuffer)
	if err != nil {
		return nil, err
	}

	var infoHash [20]byte
	var peerID [20]byte

	copy(infoHash[:], handshakeBuffer[pstrlen + 8: pstrlen + 28])
	copy(peerID[:], handshakeBuffer[pstrlen + 28:])

	h := Handshake {
		Pstr: string(handshakeBuffer[0: pstrlen]),
		InfoHash: infoHash,
		PeerID: peerID,
	}

	return &h, nil
}