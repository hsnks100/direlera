package main

import "errors"

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
func (c *CacheSystem) PutData(b []byte) uint8 {
	if v, ok := c.IncomingHitCache[string(b)]; ok {
		return v
	} else {
		c.IncomingData[c.Position] = b
		c.IncomingHitCache[string(b)] = c.Position
		c.Position++
		return c.Position - 1
	}
}

func (c *CacheSystem) GetData(pos uint8) ([]byte, error) {
	if v, ok := c.IncomingData[pos]; ok {
		return v, nil
	}
	return nil, errors.New("no cached")
}
