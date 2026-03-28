package sound

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"io"

	"gopkg.in/hraban/opus.v2"
)

const channels = 2
const sampleRate = 48000
const frameDuration = 20

func ReadOpusFrame(reader io.Reader) ([]byte, error) {
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

func WriteOpusFrames(reader io.Reader, writer io.Writer) (string, int64, error) {
	hash := sha1.New()
	size := int64(0)
	for {
		// read next frame
		frame, err := ReadOpusFrame(reader)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", 0, err
		}

		// write frame to dest
		hash.Write(frame)
		frameLenBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(frameLenBytes, uint16(len(frame)))
		n1, err := writer.Write(frameLenBytes)
		if err != nil {
			return "", int64(0), err
		}
		n2, err := writer.Write(frame)
		if err != nil {
			return "", int64(0), err
		}
		size += int64(n1 + n2)
	}
	return hex.EncodeToString(hash.Sum(nil)), size, nil
}

/* Decode opus frame data to raw PCM. */
func DecodeOpusFrames(reader io.Reader) (io.Reader, error) {
	// TODO: add wav container?
	frameSize := channels * frameDuration * sampleRate / 1000

	dec, err := opus.NewDecoder(sampleRate, 2)
	if err != nil {
		return nil, err
	}

	out := make([]byte, 0)
	for {
		frameData, err := ReadOpusFrame(reader)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		pcm := make([]int16, int(frameSize))
		if _, err := dec.Decode(frameData, pcm); err != nil {
			return nil, err
		}
		for _, d := range pcm {
			out = binary.LittleEndian.AppendUint16(out, uint16(d))
		}
	}

	return bytes.NewReader(out), nil
}
