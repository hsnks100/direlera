package main

import (
	"bytes"
	"encoding/binary"
)

type x086_ack struct {
	dummy0 uint8
	dummy1 uint32
	dummy2 uint32
	dummy3 uint32
	dummy4 uint32
}

func MakeMergePacket(packets [][]byte) []byte {
	b := make([]byte, 0)
	for _, v := range packets {
		b = append(b, v...)
	}
	return b
}

func MakePacketHeaderBody(seq uint16, msgtype uint8, body []byte) []byte {
	header := ProtocolHeader{}
	header.N = 1
	header.Seq = seq
	header.Length = uint16(len(body) + 1)
	header.MessageType = msgtype
	ret := make([]byte, 0)

	buff := new(bytes.Buffer)
	binary.Write(buff, binary.LittleEndian, &header)
	ret = append(ret, buff.Bytes()...)
	ret = append(ret, body...)
	return ret
}

func MakeServerAck(seq uint16) []byte {
	header := x086_ack{
		dummy0: 0,
		dummy1: 0,
		dummy2: 1,
		dummy3: 2,
		dummy4: 3,
	}
	buff := new(bytes.Buffer)
	binary.Write(buff, binary.LittleEndian, &header)
	return MakePacketHeaderBody(seq, 0x05, buff.Bytes())
}
