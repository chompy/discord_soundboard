package main

import "log"

const maxSoundCacheSize = 10240

type SoundCacheData struct {
	hash string
	data [][]byte
	size int
}

type SoundCache struct {
	data []SoundCacheData
}

func (c *SoundCache) Size() int {
	out := 0
	for _, sound := range c.data {
		out += sound.size
	}
	return out
}

func (c *SoundCache) Clean() {
	currentSize := c.Size()
	if currentSize < maxSoundCacheSize {
		return
	}
	log.Printf("> Sound cache exceeds %d bytes, run clean up.", maxSoundCacheSize)
	for i, sound := range c.data {
		log.Printf("  - Clear '%s' (%d bytes).", sound.hash, sound.size)
		currentSize -= sound.size
		if currentSize < maxSoundCacheSize {
			c.data = c.data[i+1 : len(c.data)]
			return
		}
	}
}

func (c *SoundCache) Push(hash string, data [][]byte) {
	size := 0
	for _, frame := range data {
		size += len(frame)
	}
	c.data = append(c.data, SoundCacheData{hash: hash, data: data, size: size})
}

func (c *SoundCache) Get(hash string) *SoundCacheData {
	for i := range c.data {
		if c.data[i].hash == hash {
			return &c.data[i]
		}
	}
	return nil
}
