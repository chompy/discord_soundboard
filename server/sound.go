package main

import (
	"encoding/binary"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/jonas747/ogg"
)

// SoundData contains raw sound bytes as well as name and size.
type SoundData struct {
	name string
	data [][]byte
	size int
}

// loadOpusFramesFromReader loads Opus frames from reader
func loadOpusFramesFromReader(sound io.Reader, name string) (SoundData, error) {
	log.Printf("> Load sound '%s' from Opus frame", name)
	out := SoundData{name: name, data: make([][]byte, 0)}
	frameSizeBytes := make([]byte, 2)
	frameSize := uint16(0)
	for {
		_, err := io.ReadAtLeast(sound, frameSizeBytes, 2)
		if err != nil {
			if err == io.EOF {
				break
			}
			return SoundData{}, err
		}
		frameSize = binary.LittleEndian.Uint16(frameSizeBytes)
		frameBuffer := make([]byte, frameSize)
		_, err = io.ReadAtLeast(sound, frameBuffer, int(frameSize))
		if err != nil {
			if err == io.EOF {
				break
			}
			return SoundData{}, err
		}
		out.data = append(out.data, frameBuffer)
		out.size += int(frameSize) + 2
	}
	log.Printf("> Loaded sound '%s' (%d bytes).", out.name, out.size)
	return out, nil
}

// loadOpusFromReader loads Opus frames from OGG Opus file
func loadOpusFromReader(sound io.Reader, name string) (SoundData, error) {
	log.Printf("> Load sound '%s'", name)
	oggDecoder := ogg.NewDecoder(sound)
	oggPacketDecoder := ogg.NewPacketDecoder(oggDecoder)
	skipPackets := 2
	out := SoundData{name: name, data: make([][]byte, 0), size: 0}
	for {
		packet, _, err := oggPacketDecoder.Decode()
		if err != nil {
			if err != io.EOF {
				return out, err
			}
			break
		}
		if skipPackets > 0 {
			skipPackets--
			continue
		}
		out.size += len(packet)
		out.data = append(out.data, packet)
	}
	log.Printf("> Loaded sound '%s' (%d bytes).", out.name, out.size)
	return out, nil
}

// LoadSound loads sound from sound path in to memory.
func (a *App) LoadSound(name string) (*SoundData, error) {
	cachedSound := a.fetchSoundCache(name)
	if cachedSound != nil {
		log.Printf("> Load sound '%s' from cache.", name)
		return cachedSound, nil
	}
	fullPath := filepath.Join(a.Config.SoundPath, name+".opus")
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	sound, err := loadOpusFromReader(file, name)
	if err != nil {
		return nil, err
	}
	a.pushSoundCache(sound)
	return a.fetchSoundCache(name), nil
}

// ListSounds lists all available sounds in sound path.
func (a *App) ListSounds() ([]string, error) {
	dirRead, err := os.ReadDir(a.Config.SoundPath)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0)
	for _, file := range dirRead {
		fullPath := filepath.Join(a.Config.SoundPath, file.Name())
		if filepath.Ext(fullPath) == ".opus" {
			name := filepath.Base(fullPath)
			out = append(out, name[0:len(name)-5])
		}
	}
	return out, nil
}

func (a *App) clearSoundCache() {
	a.soundCache = make([]SoundData, 0)
}

func (a *App) soundCacheSize() int {
	out := 0
	for _, sound := range a.soundCache {
		out += sound.size
	}
	return out
}

func (a *App) cleanSoundCache() {
	currentSize := a.soundCacheSize()
	if currentSize < a.Config.MaxMemory {
		return
	}
	log.Printf("> Sound cache exceeds %d bytes, run clean up.", a.Config.MaxMemory)
	for i, sound := range a.soundCache {
		log.Printf("  - Clear '%s' (%d bytes).", sound.name, sound.size)
		currentSize -= sound.size
		if currentSize < a.Config.MaxMemory {
			a.soundCache = a.soundCache[i+1 : len(a.soundCache)]
			return
		}
	}
}

func (a *App) pushSoundCache(sound SoundData) {
	a.cleanSoundCache()
	a.soundCache = append(a.soundCache, sound)
}

func (a *App) fetchSoundCache(name string) *SoundData {
	for i := range a.soundCache {
		if a.soundCache[i].name == name {
			return &a.soundCache[i]
		}
	}
	return nil
}
