package main

import (
	"bytes"
	"encoding/binary"
)

/*
N
   seq1[2] length[2] msg_type [DATA]
   seq0[2] length[2] msg_type [DATA] ...

   length 는 msg_type 을 포함한 길이.
*/
type x086_ack struct {
	dummy0 uint8
	dummy1 uint32
	dummy2 uint32
	dummy3 uint32
	dummy4 uint32
}

type x086ServerStatus struct {
}

func MakeMergePacket(packets [][]byte) []byte {
	b := make([]byte, 0)
	for _, v := range packets {
		b = append(b, v...)
	}
	return b
}

func MakePacketHeaderBody(seq uint16, msgtype uint8, body []byte) []byte {
	return nil
}

func MakeServerAck() []byte {
	header := x086_ack{
		dummy0: 0,
		dummy1: 0,
		dummy2: 1,
		dummy3: 2,
		dummy4: 3,
	}
	buff := new(bytes.Buffer)
	binary.Write(buff, binary.LittleEndian, &header)
	return buff.Bytes()
}
