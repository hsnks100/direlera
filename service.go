package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net"
)

func SvcUserLogin(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcUserLogin ===============")
	/*
		real:		01 00 00 12 00 05 00 00 00 00 00 01 00 00 00 02 00 00 00 03 00 00 00                                                .
		me:         01 00 00 12 00 05 00 00 00 00 00 01 00 00 00 02 00 00 00 03 00 00 00
	*/
	fields := bytes.Split(ph.data, []byte{0})
	for i, j := range fields {
		log.Infof("SvcUserLogin %d: %s", i, j)
	}

	user := GetUC().Users[addr.String()]
	user.Name = string(fields[0])
	user.EmulName = string(fields[1])
	user.ConnectType = fields[2][0]

	v3 := MakeServerAck()
	send := Protocol{}
	send.header.MessageType = 0x05
	send.header.Seq = 0
	send.data = v3
	packet := make([]byte, 0)
	packet = append(packet, 1) // N = 1
	packet = append(packet, send.MakePacket()...)
	fmt.Printf("%s\n", hex.Dump(packet))
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
		{
			p := GetUC().MakeServerStatus(uint16(user.SendCount), user)
			user.Packets = append(user.Packets, p)

			packet := make([]byte, 0)
			packet = append(packet, 1) // N = 1
			packet = append(packet, p.MakePacket()...)
			server.WriteTo(packet, addr)
			user.SendCount += 1
		}
		// joined packet
		{
			randomId := fmt.Sprintf("%2x", GetUC().RandomId)
			GetUC().RandomId += 1
			for _, u := range GetUC().Users {
				p := Protocol{}
				p.header.Seq = uint16(u.SendCount)
				p.header.MessageType = 0x02
				p.data = make([]byte, 0)
				p.data = append(p.data, []byte(user.Name+"\x00")...)
				p.data = append(p.data, []byte(randomId)...)
				p.data = append(p.data, Uint32ToBytes(user.Ping)...)
				p.data = append(p.data, user.ConnectType)
				u.Packets = append(u.Packets, p)
				packet := make([]byte, 0)
				packet = append(packet, 1) // N = 1
				packet = append(packet, p.MakePacket()...)
				log.Infof("writeto: %s", addr)
				server.WriteTo(packet, u.IpAddr)
				u.SendCount += 1
			}
		}
		// server info
		{
			p := Protocol{}
			p.header.Seq = uint16(user.SendCount)
			p.header.MessageType = 0x17
			p.data = make([]byte, 0)
			p.data = append(p.data, []byte("Server"+"\x00")...)
			p.data = append(p.data, []byte("Dire's kaillera server^^"+"\x00")...)
			user.Packets = append(user.Packets, p)
			packet := make([]byte, 0)
			packet = append(packet, 1) // N = 1
			packet = append(packet, p.MakePacket()...)
			server.WriteTo(packet, addr)
			user.SendCount += 1
		}

	}
	fmt.Printf("%s\n", hex.Dump(ph.data))
}

func SvcChatMesg(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcChatMesg ===============")
	user := GetUC().Users[addr.String()]
	// chatmesg := ph.data[1:]
	for _, u := range GetUC().Users {
		p := Protocol{}
		p.header.Seq = uint16(u.SendCount)
		p.header.MessageType = 0x07
		p.data = append(p.data, []byte(user.Name+"\x00")...)
		p.data = append(p.data, ph.data[1:]...)
		u.Packets = append(user.Packets, p)
		packet := make([]byte, 0)
		packet = append(packet, 1) // N = 1
		packet = append(packet, p.MakePacket()...)
		server.WriteTo(packet, u.IpAddr)
		log.Infof("WriteTo: %s", u.IpAddr.String())
		u.SendCount += 1
	}
}

func SvcUserQuit(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcUserQuit ===============")
	user := GetUC().Users[addr.String()]
	clientMsg := ph.data[3:]

	// chatmesg := ph.data[1:]
	for _, u := range GetUC().Users {
		p := Protocol{}
		p.header.Seq = uint16(u.SendCount)
		p.header.MessageType = 0x01
		p.data = append(p.data, []byte(user.Name+"\x00")...)
		p.data = append(p.data, []byte(user.Id)...)
		p.data = append(p.data, []byte(clientMsg)...)
		u.Packets = append(user.Packets, p)
		packet := make([]byte, 0)
		packet = append(packet, 1) // N = 1
		packet = append(packet, p.MakePacket()...)
		server.WriteTo(packet, u.IpAddr)
		log.Infof("WriteTo: %s", u.IpAddr.String())
		u.SendCount += 1
	}
	delete(GetUC().Users, addr.String())
}

func SvcCreateGame(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcCreateGame ===============")
	// user := GetUC().Users[addr.String()]
	fields := bytes.Split(ph.data, []byte{0})
	for i, j := range fields {
		log.Infof("SvcCreateGame %d: %s", i, j)
	}
	for _, u := range GetUC().Users {
		p := Protocol{}
		p.header.Seq = uint16(u.SendCount)
		p.header.MessageType = 0x0A
		p.data = append(p.data, []byte(u.Name+"\x00")...)
		p.data = append(p.data, append(fields[1], 0)...)
		p.data = append(p.data, []byte(u.EmulName+"\x00")...)
		p.data = append(p.data, Uint32ToBytes(100)...) // Game id ??
		u.Packets = append(u.Packets, p)
		packet := make([]byte, 0)
		packet = append(packet, 1) // N = 1
		packet = append(packet, p.MakePacket()...)
		server.WriteTo(packet, u.IpAddr)
		log.Infof("WriteTo: %s", u.IpAddr.String())
		u.SendCount += 1
	}
}
