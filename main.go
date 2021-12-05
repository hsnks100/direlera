package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

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
			fmt.Printf("ph: %+v\n", ph)
			fmt.Printf("header size: %d\n", ProtocolPacketsSize)
			msgtype := Protocol{}
			buf2 = bytes.NewBuffer(buf[ProtocolPacketsSize : ProtocolPacketsSize+ProtocolBodySize])
			err = binary.Read(buf2, binary.LittleEndian, &msgtype.header)
			msgtype.data = buf[ProtocolPacketsSize+ProtocolBodySize : ProtocolPacketsSize+ProtocolBodySize+msgtype.header.Length-1]
			fmt.Printf("%+v\n", msgtype)

			if msgtype.header.MessageType == 0x03 {
				GetUC().AddUser(clientAddress.String(), NewUserStruct())
				GetUC().Users[clientAddress.String()].Id = "temp"
				SvcUserLogin(server, clientAddress, msgtype) // , rawdata[:])
			} else if msgtype.header.MessageType == 0x06 {
				SvcAck(server, clientAddress, msgtype) // , rawdata[:])
			}
		}
	}()
	return server.LocalAddr()
}

func SvcUserLogin(server net.PacketConn, addr net.Addr, ph Protocol) { // , body []byte) {
	log.Infof("================ SvcUserLogin ===============")
	/*
		real:		01 00 00 12 00 05 00 00 00 00 00 01 00 00 00 02 00 00 00 03 00 00 00                                                .
		me:         01 00 00 12 00 05 00 00 00 00 00 01 00 00 00 02 00 00 00 03 00 00 00
	*/
	v3 := MakeServerAck()
	send := Protocol{}
	send.header.MessageType = 0x05
	send.header.Seq = 0
	send.data = v3
	packet := make([]byte, 0)
	packet = append(packet, 1) // N = 1
	packet = append(packet, send.MakePacket()...)
	fmt.Printf("%s\n", hex.Dump(packet))
	user := GetUC().Users[addr.String()]
	user.SendCount += 1
	user.Packets = append(user.Packets, send)

	server.WriteTo(packet, addr)
}

func SvcAck(server net.PacketConn, addr net.Addr, ph Protocol) { // Header, body []byte) {
	log.Infof("================ SvcAck ===============")
	user := GetUC().Users[addr.String()]
	log.Infof("user: %+v", user)
	if user.SendCount <= 4 {
		v3 := MakeServerAck()
		send := Protocol{}
		send.header.MessageType = 0x05
		send.header.Seq = uint16(user.SendCount)
		send.data = v3
		packet := make([]byte, 0)
		packet = append(packet, 1) // N = 1
		packet = append(packet, send.MakePacket()...)
		fmt.Printf("%s\n", hex.Dump(packet))
		user.SendCount += 1
		user.Packets = append(user.Packets, send)

		server.WriteTo(packet, addr)

	} else {
		p := GetUC().MakeServerStatus(uint16(user.SendCount))
		user.Packets = append(user.Packets, p)

		packet := make([]byte, 0)
		packet = append(packet, 1) // N = 1
		packet = append(packet, p.MakePacket()...)
		server.WriteTo(packet, addr)
		user.SendCount += 1
	}
	fmt.Printf("%s\n", hex.Dump(ph.data))
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
	GetUC().AddUser("key1", NewUserStruct())
	GetUC().AddUser("key2", NewUserStruct())
	GetUC().Users["key1"] = &UserStruct{
		Ip:           "key1",
		Id:           "i1",
		Name:         "name1",
		Ping:         33,
		ConnectType:  2,
		PlayerStatus: 2,
		AckCount:     2,
		SendCount:    2,
	}
	GetUC().Users["key2"] = &UserStruct{
		Ip:           "key2",
		Id:           "i2",
		Name:         "name2",
		Ping:         34,
		ConnectType:  1,
		PlayerStatus: 2,
		AckCount:     2,
		SendCount:    2,
	}
	for i, j := range GetUC().Users {
		log.Infof("%s: %+v", i, j)
	}
	MakeUDPServer()
	MakeProcServer()

	time.Sleep(1000 * time.Second)

}
