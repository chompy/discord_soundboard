package sound

import (
	"bytes"
	"io"

	"github.com/jonas747/ogg"
)

type OggReader struct {
	baseReader     io.Reader
	buffer         bytes.Buffer
	packetsRead    int
	packetDecorder *ogg.PacketDecoder
}

func (r *OggReader) bufferNextFrame() error {
	if r.packetDecorder == nil {
		oggDecoder := ogg.NewDecoder(r.baseReader)
		r.packetDecorder = ogg.NewPacketDecoder(oggDecoder)
	}
	packet, _, err := r.packetDecorder.Decode()
	if err != nil {
		return err
	}
	r.packetsRead++
	if r.packetsRead < 2 {
		return r.bufferNextFrame()
	}
	r.buffer = bytes.Buffer{}
	if _, err := writeOpusFrame(packet, &r.buffer, nil); err != nil {
		return err
	}
	return nil
}

func (r *OggReader) Read(p []byte) (n int, err error) {
	if r.buffer.Len() == 0 {
		if err := r.bufferNextFrame(); err != nil {
			return 0, err
		}
	}
	return r.buffer.Read(p)
}
