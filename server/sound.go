package main

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
	"path"
)

type SoundReader struct {
	reader io.ReadCloser
}

func NewSoundReader(reader io.ReadCloser) *SoundReader {
	return &SoundReader{reader: reader}
}

func NewSoundReaderFromStorage(hash string) (*SoundReader, error) {
	log.Printf("> Load sound %s.", hash)
	storePath := path.Join(storagePath, hash+".dat")
	file, err := os.Open(storePath)
	if err != nil {
		return nil, err
	}
	return NewSoundReader(file), nil
}

func (s *SoundReader) NextFrame() ([]byte, error) {
	frameSizeBytes := make([]byte, 2)
	_, err := io.ReadAtLeast(s.reader, frameSizeBytes, 2)
	if err != nil {
		if err == io.EOF {
			s.reader.Close()
		}
		return nil, err
	}

	frameSize := uint16(0)
	frameSize = binary.LittleEndian.Uint16(frameSizeBytes)

	frameBuffer := make([]byte, frameSize)
	_, err = io.ReadAtLeast(s.reader, frameBuffer, int(frameSize))
	if err != nil {
		if err == io.EOF {
			s.reader.Close()
		}
		return nil, err
	}
	return frameBuffer, nil
}

func (s *SoundReader) Save() (string, error) {
	file, err := os.CreateTemp(storagePath, "_*")
	if err != nil {
		return "", err
	}
	defer func() {
		file.Close()
		os.Remove(file.Name())
	}()

	hash := sha1.New()
	size := 0

	for {
		frame, err := s.NextFrame()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		hash.Write(frame)

		frameLenBytes := make([]byte, 2)
		binary.LittleEndian.PutUint16(frameLenBytes, uint16(len(frame)))

		file.Write(frameLenBytes)
		file.Write(frame)
		size += len(frameLenBytes) + len(frame)
	}

	file.Close()
	hashStr := hex.EncodeToString(hash.Sum(nil))
	storePath := path.Join(storagePath, hashStr+".dat")

	if _, err := os.Stat(storePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("> Saved sound %s (%d bytes).", hashStr, size)
			return hashStr, os.Rename(file.Name(), storePath)
		}
		return "", err
	}

	log.Printf("> Sound %s already exists, skipped ", hashStr)

	return hashStr, nil
}
