package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var log = logrus.New()

func init() {
	// logrus.For
	log.Formatter = new(prefixed.TextFormatter)
	log.Level = logrus.DebugLevel
}

type ProtocolPackets struct {
	N uint8
}

const (
	ProtocolPacketsSize = 1
	ProtocolBodySize    = 5
)

type ProtocolHeader struct {
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

func MakeProcServer() net.Addr {
	server, err := net.ListenPacket("udp", "0.0.0.0:27999")
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println("222server address: ", server.LocalAddr().String())
	go func() {
		for {
			buf := make([]byte, 1024)
			n, clientAddress, err := server.ReadFrom(buf)
			if err != nil {
				continue
			}
			fmt.Printf("<== received: %+v ( %s ) / %s / %s\n", buf[:n], string(buf[:n]), "from", clientAddress)
			ph := ProtocolPackets{}
			buf2 := bytes.NewBuffer(buf[:ProtocolPacketsSize])
			err = binary.Read(buf2, binary.LittleEndian, &ph)
			// fmt.Printf("ph: %+v\n", ph)
			// fmt.Printf("header size: %d\n", ProtocolPacketsSize)
			msgtype := Protocol{}
			buf2 = bytes.NewBuffer(buf[ProtocolPacketsSize : ProtocolPacketsSize+ProtocolBodySize])
			err = binary.Read(buf2, binary.LittleEndian, &msgtype.header)
			msgtype.data = buf[ProtocolPacketsSize+ProtocolBodySize : ProtocolPacketsSize+ProtocolBodySize+msgtype.header.Length-1]
			// fmt.Printf("%+v\n", msgtype)

			if msgtype.header.MessageType == 0x03 {
				GetUC().AddUser(clientAddress, NewUserStruct())
				GetUC().Users[clientAddress.String()].Id = "temp"
				SvcUserLogin(server, clientAddress, msgtype) // , rawdata[:])
			} else if msgtype.header.MessageType == 0x06 {
				SvcAck(server, clientAddress, msgtype) // , rawdata[:])
			} else if msgtype.header.MessageType == 0x07 {
				SvcChatMesg(server, clientAddress, msgtype) // , rawdata[:])
			} else if msgtype.header.MessageType == 0x01 {
				SvcUserQuit(server, clientAddress, msgtype)
			} else {
				log.Infof("unknown Message[%2x]: %+v", msgtype.header.MessageType, msgtype)
			}
		}
	}()
	return server.LocalAddr()
}

func MakeUDPServer() net.Addr {
	server, err := net.ListenPacket("udp", "0.0.0.0:27888")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("server address: ", server.LocalAddr().String())
	go func() {
		for {
			buf := make([]byte, 1024)
			n, clientAddress, err := server.ReadFrom(buf)
			if err != nil {
				continue
			}
			fmt.Printf("<- received: %+v ( %s ) / %s / %s\n", buf[:n], string(buf[:n]), "from", clientAddress)
			if string(buf[:n]) == "PING\x00" {
				_, err = server.WriteTo([]byte("PONG\x00"), clientAddress)
			}
			if n >= 5 && string(buf[:5]) == "HELLO" {
				_, err = server.WriteTo([]byte("HELLOD00D27999\x00"), clientAddress)
			}
		}
	}()
	return server.LocalAddr()
}
func main() {

	log.SetReportCaller(true)

	for i, j := range GetUC().Users {
		log.Infof("%s: %+v", i, j)
	}
	MakeUDPServer()
	MakeProcServer()

	time.Sleep(1000 * time.Second)

}
