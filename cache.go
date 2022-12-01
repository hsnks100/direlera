package main

import (
	"errors"
	"fmt"
)

// 인풋 []byte 에 대해 uint8 로 접근 할 수 있게 하는 시스템

type CacheSystem struct {
	Position         uint8
	IncomingData     map[uint8][]byte
	IncomingHitCache map[string]uint8
}

func NewCacheSystem() *CacheSystem {
	return &CacheSystem{IncomingData: map[uint8][]byte{},
		IncomingHitCache: map[string]uint8{}}
}

func (c *CacheSystem) Reset() {
	c.Position = 0
	c.IncomingData = map[uint8][]byte{}
	c.IncomingHitCache = map[string]uint8{}
}
func (c *CacheSystem) GetCachePosition(b []byte) (uint8, error) {
	if v, ok := c.IncomingHitCache[string(b)]; ok {
		return v, nil
	} else {
		return 0, errors.New("no cache")
	}
}
func (c *CacheSystem) PutData(b []byte) uint8 {
	if v, ok := c.IncomingHitCache[string(b)]; ok {
		return v
	} else {
		c.IncomingData[c.Position] = b
		c.IncomingHitCache[string(b)] = c.Position
		c.Position++
		if c.Position >= 250 {
			fmt.Printf("cache 위험~~: %d", c.Position)
		}
		return c.Position - 1
	}
}

func (c *CacheSystem) GetData(pos uint8) ([]byte, error) {
	if v, ok := c.IncomingData[pos]; ok {
		return v, nil
	}
	return nil, errors.New("no cached")
}
