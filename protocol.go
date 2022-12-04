package main

import (
	"bytes"
	"encoding/binary"
)

type ProtocolPackets struct {
	N uint8
}

const (
	ProtocolPacketsSize = 1
	ProtocolBodySize    = 5
)

type ProtocolHeader struct {
	// TODO: Seq 세팅할 때 서비스 코드단에서 이걸 설정 안할 수 있게 수정 필요함.
	Seq         uint16
	Length      uint16 // msgtype 포함한 길이
	MessageType uint8
}
type Protocol struct {
	header ProtocolHeader
	data   []byte
}

func (t *Protocol) MakePacket() []byte {
	prob := ProtocolHeader{}
	prob.Seq = t.header.Seq
	prob.Length = uint16(len(t.data) + 1)
	prob.MessageType = t.header.MessageType
	ret := make([]byte, 0)

	buff := new(bytes.Buffer)
	binary.Write(buff, binary.LittleEndian, &prob)
	ret = append(ret, buff.Bytes()...)
	ret = append(ret, t.data...)
	return ret
}

func GetProtocolFromBytes(data []byte) []Protocol {
	curPos := 0
	ret := make([]Protocol, 0)
	for {
		// log.Infof("%d <= %d", curPos+ProtocolBodySize, len(data))
		if curPos+ProtocolBodySize <= len(data) {
			msgtype := Protocol{}
			// log.Infof("range1 [%d, %d)", curPos, curPos+ProtocolBodySize)
			buf2 := bytes.NewBuffer(data[curPos : curPos+ProtocolBodySize])

			err := binary.Read(buf2, binary.LittleEndian, &msgtype.header)
			// log.Infof("range2 [%d, %d)", curPos+ProtocolBodySize, curPos+ProtocolBodySize+int(msgtype.header.Length)-1)
			msgtype.data = data[curPos+ProtocolBodySize : curPos+ProtocolBodySize+int(msgtype.header.Length)-1]
			curPos += ProtocolBodySize + int(msgtype.header.Length) - 1
			if err != nil {
				break
			}
			ret = append(ret, msgtype)
		} else {
			break
		}
	}
	return ret
}
