package protocol

import (
	"encoding/binary"
	"io"
)

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (e *Encoder) Encode(pkt *Packet) error {
	// Write header
	if err := binary.Write(e.w, binary.BigEndian, pkt.Header); err != nil {
		return err
	}

	// Write payload if any
	if len(pkt.Payload) > 0 {
		if _, err := e.w.Write(pkt.Payload); err != nil {
			return err
		}
	}

	return nil
}

type Decoder struct {
	r io.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

func (d *Decoder) Decode(pkt *Packet) error {
	// Read header
	if err := binary.Read(d.r, binary.BigEndian, &pkt.Header); err != nil {
		return err
	}

	// Read remaining bytes as payload
	payload, err := io.ReadAll(d.r)
	if err != nil {
		return err
	}
	pkt.Payload = payload

	return nil
}
