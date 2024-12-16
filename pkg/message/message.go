package message

import (
	"encoding/binary"
	"fmt"
	"io"
)

type msgID uint16

const MsgChoke msgID = 0
const MsgUnchoke msgID = 1
const MsgInterested msgID = 2
const MsgUninterested msgID = 3
const MsgHave msgID = 4
const MsgBitfield msgID = 5
const MsgRequest msgID = 6
const MsgPiece msgID = 7
const MsgCancel msgID = 8

type Msg struct {
	ID msgID
	Payload []byte
}

func FormatRequest(ind int, begin int, length int) *Msg {
	payload := make([]byte, 12)
	binary.BigEndian.PutUint32(payload[0:4], uint32(ind))
	binary.BigEndian.PutUint32(payload[4:8], uint32(begin))
	binary.BigEndian.PutUint32(payload[8:12], uint32(length))
	return &Msg {
		ID: MsgRequest, 
		Payload: payload,
	}
}

func FormatHave(ind int) *Msg {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, uint32(ind))
	return &Msg {
		ID: MsgHave,
		Payload: payload,
	}
}

func ParsePiece(ind int, buffer []byte, msg *Msg) (int, error) {
	if msg.ID != MsgPiece {
		return 0, fmt.Errorf("expected piece with ID %d but got ID %d", MsgPiece, msg.ID)
	}
	if len(msg.Payload) < 8 {
		return 0, fmt.Errorf("payload size too small")
	}

	parsedInd := int(binary.BigEndian.Uint32(msg.Payload[0:4]))
	if parsedInd != ind {
		return 0, fmt.Errorf("expected index %d but got %d", ind, parsedInd)
	}

	begin := int(binary.BigEndian.Uint32(msg.Payload[4:8]))
	if begin >= len(buffer) {
		return 0, fmt.Errorf("begin offset too high")
	}

	data := msg.Payload[8:]
	if begin + len(data) > len(buffer) {
		return 0, fmt.Errorf("data too long")
	}
	copy(buffer[begin:], data)
	return len(data), nil
}

func ParseHave(msg *Msg) (int, error) {
	if msg.ID != MsgHave {
		return 0, fmt.Errorf("expected have with ID %d but got ID %d", MsgHave, msg.ID)
	}
	if len(msg.Payload) != 4 {
		return 0, fmt.Errorf("did not get payload of expected length 4")
	}
	index := int(binary.BigEndian.Uint32(msg.Payload))
	return index, nil
}

func (m *Msg) Serialize() []byte {
	if m == nil {
		return make([]byte, 4)
	}

	len := uint32(len(m.Payload) + 1)
	buffer := make([]byte, len + 4)
	binary.BigEndian.PutUint32(buffer[0:4], len)
	buffer[4] = byte(m.ID)
	copy(buffer[5:], m.Payload)

	return buffer
}

func Read(r io.Reader) (*Msg, error) {
	lengthBuffer := make([]byte, 4)
	_, err := io.ReadFull(r, lengthBuffer)
	if err != nil {
		return nil, err
	}
	len := binary.BigEndian.Uint32(lengthBuffer)

	if len == 0 {
		return nil, nil
	}

	msgBuffer := make([]byte, len)
	_, err = io.ReadFull(r, msgBuffer)
	if err != nil {
		return nil, err
	}

	m := Msg {
		ID: msgID(msgBuffer[0]),
		Payload: msgBuffer[1:],
	}

	return &m, nil
}

/*
func (m *Msg) name() string {
	if m == nil {
		return "KeepAlive"
	}

	switch m.ID {
	case MsgChoke:
		return "Choke"
	
	case MsgUnchoke:
		return "Unchoke"

	case MsgInterested:
		return "Interested"
	
	case MsgUninterested:
		return "Uninterested"

	case MsgHave:
		return "Have"

	case MsgBitfield:
		return "Bitfield"

	case MsgRequest:
		return "Request"

	case MsgCancel:
		return "Cancel"

	default:
		return fmt.Sprintf("Unknown#%d", m.ID)
	}
}

func (m *Msg) String() string {
	if m == nil {
		return m.name()
	}

	return fmt.Sprintf("%s [%d]", m.name(), len(m.Payload))
}
*/