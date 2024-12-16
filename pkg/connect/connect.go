package connect

import (
	"bytes"
	"fmt"
	"net"
	"time"

	"github.com/lokeshllkumar/bittorent-client/pkg/bitfield"
	"github.com/lokeshllkumar/bittorent-client/pkg/handshake"
	"github.com/lokeshllkumar/bittorent-client/pkg/message"
	"github.com/lokeshllkumar/bittorent-client/pkg/peers"
)

type Client struct {
	Conn     net.Conn
	Choked   bool
	Bitfield bitfield.Bitfield
	peer     peers.Peer
	infoHash [20]byte
	peerID   [20]byte
}

func completeHandshake(conn net.Conn, infohash [20]byte, peerID [20]byte) (*handshake.Handshake, error) {
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	defer conn.SetDeadline(time.Time{})

	req := handshake.NewHandshake(infohash, peerID)
	_, err := conn.Write(req.Serialize())
	if err != nil {
		return nil, err
	}

	res, err := handshake.Read(conn)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(res.InfoHash[:], infohash[:]) {
		return nil, fmt.Errorf("expected infohash %x but got %x", res.InfoHash, infohash)
	}

	return res, nil
}

func receiveBitfield(conn net.Conn) (bitfield.Bitfield, error) {
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetDeadline(time.Time{})

	msg, err := message.Read(conn)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		err := fmt.Errorf("did not get bitfield")
		return nil, err
	}
	if msg.ID != message.MsgBitfield {
		err := fmt.Errorf("expected bitfield but received wrong ID")
		return nil, err
	}

	return msg.Payload, nil
}

func NewConnection(peer peers.Peer, peerID [20]byte, infoHash [20]byte) (*Client, error) {
	conn, err := net.DialTimeout("tcp", peer.ToString(), 10 * time.Second)
	if err != nil {
		return nil, err
	}
	
	_, err = completeHandshake(conn, infoHash, peerID)
	if err != nil {
		conn.Close()
		return nil, err
	}

	bitfield, err := receiveBitfield(conn)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Client {
		Conn: conn,
		Choked: true,
		Bitfield: bitfield,
		peer: peer,
		infoHash: infoHash,
		peerID: peerID,
	}, nil
}

func (c *Client) Read() (*message.Msg, error) {
	msg, err := message.Read(c.Conn)
	return msg, err
}

func (c *Client) SendRequest(ind int, begin int, len int) error {
	req := message.FormatRequest(ind, begin, len)

	_, err := c.Conn.Write(req.Serialize())
	return err
}

func (c *Client) SendInterested() error {
	msg := message.Msg {
		ID: message.MsgInterested,
	}

	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendNotInterested() error {
	msg := message.Msg {
		ID: message.MsgUninterested,
	}

	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendUnchoke() error {
	msg := message.Msg {
		ID: message.MsgUnchoke,
	}

	_, err := c.Conn.Write(msg.Serialize())
	return err
}

func (c *Client) SendHave(ind int) error {
	msg := message.FormatHave(ind)

	_, err := c.Conn.Write(msg.Serialize())
	return err
}