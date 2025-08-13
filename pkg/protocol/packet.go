package protocol

import (
	"bytes"
	"errors"
	"time"
)

/*
Binary Protocol Format

The protocol uses a fixed-length header with variable-length payload:

 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|        Version (2 bytes)      |      Operation (2 bytes)      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                       SequenceID (4 bytes)                    |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                           Payload                             |
|                          (variable)                           |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

Header Fields:
- Version: Protocol version, currently 1 (2 bytes)
- Operation: Operation type (2 bytes)
- SequenceID: Sequence ID for message tracking (4 bytes)

Total header size: 8 bytes
All fields are in network byte order
*/

const (
	Version uint16 = 1
)

var (
	ErrInvalidLength  = errors.New("invalid packet length")
	ErrInvalidPayload = errors.New("invalid payload")
	ErrInvalidVersion = errors.New("invalid packet version")
)

// FixedLengthHeader represents the fixed-length header of a packet
type FixedLengthHeader struct {
	Version    uint16 // Protocol version, currently 1
	Operation  uint16 // Operation type
	SequenceID uint32 // Sequence ID for message tracking
}

// Packet represents a complete data packet
type Packet struct {
	Header  FixedLengthHeader
	Payload []byte
}

func (p *Packet) validate() error {
	if p.Header.Version != Version {
		return ErrInvalidVersion
	}
	return nil
}

type PacketOption func(*Packet)

func WithSequenceID(sequenceID uint32) PacketOption {
	return func(p *Packet) {
		p.Header.SequenceID = sequenceID
	}
}

func NewPacket(operation uint16, payload []byte, opts ...PacketOption) *Packet {
	pkt := &Packet{
		Header: FixedLengthHeader{
			Version:   Version,
			Operation: operation,
		},
		Payload: payload,
	}

	for _, opt := range opts {
		opt(pkt)
	}

	if pkt.Header.SequenceID == 0 {
		pkt.Header.SequenceID = uint32(time.Now().UnixNano())
	}

	return pkt
}

// Pack encodes packet to binary data
func Pack(pkt *Packet) ([]byte, error) {
	if err := pkt.validate(); err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(make([]byte, 0, 4096))
	encoder := NewEncoder(buf)
	if err := encoder.Encode(pkt); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Unpack decodes binary data to packet
func Unpack(data []byte) (*Packet, error) {
	buf := bytes.NewReader(data)
	decoder := NewDecoder(buf)

	pkt := &Packet{}
	if err := decoder.Decode(pkt); err != nil {
		return nil, err
	}

	if err := pkt.validate(); err != nil {
		return nil, err
	}

	return pkt, nil
}
