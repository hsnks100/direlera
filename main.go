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

type ProtocolHeader struct {
	N           uint8
	Seq         uint16
	Length      uint16
	MessageType uint8
}

const (
	ProtocolHeaderSize = 6
)

type ProtocolBody struct {
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
			ph := ProtocolHeader{}
			buf2 := bytes.NewBuffer(buf[:ProtocolHeaderSize])
			err = binary.Read(buf2, binary.LittleEndian, &ph)
			fmt.Printf("ph: %+v\n", ph)
			fmt.Printf("header size: %d", ProtocolHeaderSize)
			rawdata := buf[ProtocolHeaderSize:n]
			// fmt.Printf("%+v\n", rawdata)

			if ph.MessageType == 0x03 {
				SvcUserLogin(server, clientAddress, ph, rawdata)
			} else if ph.MessageType == 0x06 {
				SvcAck(server, clientAddress, ph, rawdata[:ph.Length])
			}
		}
	}()
	return server.LocalAddr()
}

func SvcUserLogin(server net.PacketConn, addr net.Addr, ph ProtocolHeader, body []byte) {
	log.Infof("================ SvcUserLogin ===============")
	/*
		real:		01 00 00 12 00 05 00 00 00 00 00 01 00 00 00 02 00 00 00 03 00 00 00                                                .
		me:         01 00 00 12 00 05 00 00 00 00 00 01 00 00 00 02 00 00 00 03 00 00 00
	*/
	v3 := MakeServerAck(0)
	v1 := make([][]byte, 0)
	v1 = append(v1, v3)
	send := MakeMergePacket(v1)
	fmt.Printf("%s\n", hex.Dump(send))
	server.WriteTo(send, addr)
}

func SvcAck(server net.PacketConn, addr net.Addr, ph ProtocolHeader, body []byte) {
	log.Infof("================ SvcAck ===============")
	fmt.Printf("%s\n", hex.Dump(body))
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
	MakeUDPServer()
	MakeProcServer()

	time.Sleep(1000 * time.Second)

}
