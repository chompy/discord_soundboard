package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/jonas747/ogg"
)

// loadSound attempts to load an encoded sound file from disk.
func loadSound(path string) ([][]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	oggDecoder := ogg.NewDecoder(file)
	oggPacketDecoder := ogg.NewPacketDecoder(oggDecoder)
	skipPackets := 2
	out := make([][]byte, 0)
	for {
		packet, _, err := oggPacketDecoder.Decode()
		if skipPackets > 0 {
			skipPackets--
			continue
		}
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}
		out = append(out, packet)
	}
	return out, nil
}

// loadAllSounds iterates sound directory and attempts to load all opus sound files in to memory.
func (a *App) loadAllSounds() error {
	dirRead, err := os.ReadDir(a.Config.SoundPath)
	if err != nil {
		return err
	}
	a.sounds = make(map[string][][]byte)
	for _, file := range dirRead {
		fullPath := filepath.Join(a.Config.SoundPath, file.Name())
		if filepath.Ext(fullPath) == ".opus" {
			log.Printf("> Load sound '%s.'", fullPath)
			data, err := loadSound(fullPath)
			if err != nil {
				log.Println("> WARNING: Load sound error: ", err)
				continue
			}
			a.sounds[strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))] = data
		}
	}
	return nil
}
