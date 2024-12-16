package peerconnection

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/lokeshllkumar/bittorent-client/pkg/connect"
	"github.com/lokeshllkumar/bittorent-client/pkg/message"
	"github.com/lokeshllkumar/bittorent-client/pkg/peers"
)

const MaxReqSize = 16384

const MaxBacklog = 5

type Torrent struct {
	Peers       []peers.Peer
	PeerID      [20]byte
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

type pieceWork struct {
	ind  int
	hash [20]byte
	len  int
}

type pieceRes struct {
	ind    int
	buffer []byte
}

type pieceProgress struct {
	ind        int
	client     *connect.Client
	buffer     []byte
	downloaded int
	requested  int
	backlog    int
}

func (state *pieceProgress) readMessage() error {
	msg, err := state.client.Read() // blocking call
	if err != nil {
		return err
	}

	if msg == nil {
		return nil
	}

	switch msg.ID {
	case message.MsgUnchoke:
		state.client.Choked = false

	case message.MsgChoke:
		state.client.Choked = true

	case message.MsgHave:
		ind, err := message.ParseHave(msg)
		if err != nil {
			return err
		}
		state.client.Bitfield.SetPiece(ind)

	case message.MsgPiece:
		n, err := message.ParsePiece(state.ind, state.buffer, msg)
		if err != nil {
			return err
		}
		state.downloaded += n
		state.backlog--
	}
	return nil
}

func attemptDownload(c *connect.Client, pw *pieceWork) ([]byte, error) {
	state := pieceProgress{
		ind:    pw.ind,
		client: c,
		buffer: make([]byte, pw.len),
	}

	c.Conn.SetDeadline(time.Now().Add(30 * time.Second))
	defer c.Conn.SetDeadline(time.Time{})

	for state.downloaded < pw.len {
		if !state.client.Choked {
			for state.backlog < MaxBacklog && state.requested < pw.len {
				reqSize := MaxReqSize
				if pw.len-state.requested < reqSize {
					reqSize = pw.len - state.requested
				}

				err := c.SendRequest(pw.ind, state.requested, reqSize)
				if err != nil {
					return nil, err
				}
				state.backlog++
				state.requested += reqSize
			}
		}

		err := state.readMessage()
		if err != nil {
			return nil, err
		}
	}

	return state.buffer, nil
}

func checkIntegrity(pw *pieceWork, buffer []byte) error {
	hash := sha1.Sum(buffer)
	if !bytes.Equal(hash[:], pw.hash[:]) {
		return fmt.Errorf("index %d failed integrity check", pw.ind)
	}
	return nil
}

func (t *Torrent) startDownloadWorker(peer peers.Peer, workQueue chan *pieceWork, res chan *pieceRes) {
	c, err := connect.NewConnection(peer, t.PeerID, t.InfoHash)
	if err != nil {
		log.Printf("Could not handshake with %s. Disconnecting with peer...\n", peer.IP)
		return
	}
	defer c.Conn.Close()
	log.Printf("Completed handshake with %s\n", peer.IP)

	c.SendUnchoke()
	c.SendInterested()

	for pw := range workQueue {
		if !c.Bitfield.HasPiece(pw.ind) {
			workQueue <- pw
			continue
		}

		buffer, err := attemptDownload(c, pw)
		if err != nil {
			log.Println("Exiting", err)
			workQueue <- pw
			return
		}

		err = checkIntegrity(pw, buffer)
		if err != nil {
			log.Printf("Piece #%d fails integrity check\n", pw.ind)
			workQueue <- pw
			continue
		}

		c.SendHave(pw.ind)
		res <- &pieceRes{
			pw.ind,
			buffer,
		}
	}
}

func (t *Torrent) calcBounds(ind int) (begin int, end int) {
	begin = ind * t.PieceLength
	end = begin + t.PieceLength
	if end > t.Length {
		end = t.Length
	}

	return begin, end
}

func (t *Torrent) calcPieceSize(ind int) int {
	begin, end := t.calcBounds(ind)
	size := end - begin

	return size
}

func (t *Torrent) Download() ([]byte, error) {
	log.Println("Starting download for", t.Name)

	workQueue := make(chan *pieceWork, len(t.PieceHashes))
	results := make(chan *pieceRes)
	for ind, hash := range t.PieceHashes {
		len := t.calcPieceSize(ind)
		workQueue <- &pieceWork{
			ind,
			hash,
			len,
		}
	}

	for _, peer := range t.Peers {
		go t.startDownloadWorker(peer, workQueue, results)
	}

	buffer := make([]byte, t.Length)
	donePieces := 0
	for donePieces < len(t.PieceHashes) {
		res := <-results
		begin, end := t.calcBounds(res.ind)
		copy(buffer[begin:end], res.buffer)
		donePieces++

		percent := float64(donePieces) / float64(len(t.PieceHashes)) * 100
		numWorkers := runtime.NumGoroutine() - 1
		log.Printf("(%0.2f%%) Downloaded piece #%d from %d peers\n", percent, res.ind, numWorkers)
	}
	close(workQueue)

	return buffer, nil
}
