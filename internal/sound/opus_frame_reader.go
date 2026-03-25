package sound

import (
	"encoding/binary"
	"io"
)

func NextFrame(reader io.Reader) ([]byte, error) {
	frameSizeBytes := make([]byte, 2)
	_, err := io.ReadAtLeast(reader, frameSizeBytes, 2)
	if err != nil {
		return nil, err
	}

	frameSize := uint16(0)
	frameSize = binary.LittleEndian.Uint16(frameSizeBytes)

	frameBuffer := make([]byte, frameSize)
	_, err = io.ReadAtLeast(reader, frameBuffer, int(frameSize))
	if err != nil {
		return nil, err
	}
	return frameBuffer, nil
}
