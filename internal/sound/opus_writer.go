package sound

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"hash"
	"io"
)

func WriteOpusFrames(reader io.Reader, writer io.Writer) (string, int64, error) {
	hash := sha1.New()
	size := int64(0)
	for {
		frame, err := NextFrame(reader)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", 0, err
		}
		n, err := writeOpusFrame(frame, writer, hash)
		if err != nil {
			return "", 0, err
		}
		size += int64(n)
	}
	return hex.EncodeToString(hash.Sum(nil)), size, nil
}

func writeOpusFrame(frame []byte, writer io.Writer, hasher hash.Hash) (int64, error) {
	if hasher != nil {
		if _, err := hasher.Write(frame); err != nil {
			return 0, err
		}
	}
	frameLenBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(frameLenBytes, uint16(len(frame)))
	n1, err := writer.Write(frameLenBytes)
	if err != nil {
		return int64(n1), err
	}
	n2, err := writer.Write(frame)
	return int64(n1 + n2), err
}
