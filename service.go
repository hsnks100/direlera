package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
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
	log.Infof("svcuserlogin conntype : %+v", fields[2])
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
	user.SendPacket(server, send)
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
		user.SendPacket(server, send)

	} else {
		randomPing := uint32(rand.Intn(30))
		user.Ping = randomPing
		{
			p := GetUC().MakeServerStatus(uint16(user.SendCount), user)
			user.SendPacket(server, p)
		}
		// joined packet
		{
			for _, u := range GetUC().Users {
				p := Protocol{}
				p.header.Seq = uint16(u.SendCount)
				p.header.MessageType = 0x02
				p.data = make([]byte, 0)
				p.data = append(p.data, []byte(user.Name+"\x00")...)
				p.data = append(p.data, Uint16ToBytes(user.UserId)...)
				p.data = append(p.data, Uint32ToBytes(user.Ping)...)
				p.data = append(p.data, user.ConnectType)
				u.SendPacket(server, p)
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
			user.SendPacket(server, p)
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
		u.SendPacket(server, p)
	}
}
func SvcGameChat(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcGameChat ===============")
	user := GetUC().Users[addr.String()]
	cs := GetUC().Channels[user.GameRoomId]
	for i, _ := range cs.Players {
		u := GetUC().Users[i]
		p := Protocol{}
		p.header.Seq = uint16(u.SendCount)
		p.header.MessageType = 0x08
		p.data = append(p.data, []byte(user.Name+"\x00")...)
		p.data = append(p.data, ph.data[1:]...)
		u.SendPacket(server, p)
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
		p.data = append(p.data, Uint16ToBytes(user.UserId)...)
		p.data = append(p.data, []byte(clientMsg)...)
		u.SendPacket(server, p)
	}
	delete(GetUC().Users, addr.String())
}

var g_gameId = 0x1

func SvcCreateGame(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcCreateGame ===============")
	user := GetUC().Users[addr.String()]
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
		p.data = append(p.data, Uint32ToBytes(uint32(g_gameId))...) // Game id ??

		u.SendPacket(server, p)
	}
	cs := NewChannelStruct()
	cs.CreatorId = user.Name
	cs.EmulName = user.EmulName
	cs.GameId = uint32(g_gameId)
	user.GameRoomId = cs.GameId
	g_gameId += 1
	log.Infof("create gameid: %+v", Uint32ToBytes(cs.GameId))
	cs.GameName = string(fields[1])
	cs.GameStatus = 0
	cs.Players[user.IpAddr.String()] = struct{}{}
	GetUC().AddChannel(cs.GameId, cs)
	// update game status
	for _, u := range GetUC().Users {

		p := Protocol{}
		p.header.Seq = uint16(u.SendCount)
		p.header.MessageType = 0x0E
		p.data = append(p.data, 0)
		p.data = append(p.data, Uint32ToBytes(cs.GameId)...)
		p.data = append(p.data, cs.GameStatus)
		p.data = append(p.data, uint8(len(cs.Players)))
		p.data = append(p.data, 4)
		u.SendPacket(server, p)
	}
	// player info
	{
		p := Protocol{}
		p.header.Seq = uint16(user.SendCount)
		p.header.MessageType = 0x0D
		p.data = append(p.data, 0)
		p.data = append(p.data, Uint32ToBytes(uint32(len(cs.Players))-1)...)
		for i, _ := range cs.Players {
			roomUser := GetUC().Users[i]
			p.data = append(p.data, []byte(roomUser.Name+"\x00")...)
			p.data = append(p.data, Uint32ToBytes(roomUser.Ping)...)
			p.data = append(p.data, Uint16ToBytes(roomUser.UserId)...)
			p.data = append(p.data, roomUser.ConnectType)
		}
		user.SendPacket(server, p)
	}
	// Join Game Noti
	for _, u := range GetUC().Users {
		p := Protocol{}
		p.header.Seq = uint16(u.SendCount)
		p.header.MessageType = 0x0C
		p.data = append(p.data, 0)
		p.data = append(p.data, Uint32ToBytes(cs.GameId)...)
		p.data = append(p.data, []byte(user.Name+"\x00")...)
		p.data = append(p.data, Uint32ToBytes(user.Ping)...)
		p.data = append(p.data, Uint16ToBytes(user.UserId)...)
		p.data = append(p.data, user.ConnectType)
		u.SendPacket(server, p)
	}
	// server info
	{
		p := Protocol{}
		p.header.Seq = uint16(user.SendCount)
		p.header.MessageType = 0x17
		p.data = make([]byte, 0)
		p.data = append(p.data, []byte("Server"+"\x00")...)
		p.data = append(p.data, []byte(fmt.Sprintf("%s Creates Room: %s\x00", user.Name, fields[1]))...)
		user.SendPacket(server, p)
	}
}

func SvcJoinGame(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcJoinGame ===============")
	user := GetUC().Users[addr.String()]
	gameId := binary.LittleEndian.Uint32(ph.data[1:5])
	log.Infof("join gameid: %+v", Uint32ToBytes(gameId))
	connType := ph.data[12:13][0]
	cs := GetUC().Channels[gameId]
	cs.Players[user.IpAddr.String()] = struct{}{}
	user.GameRoomId = gameId
	_ = connType
	for _, u := range GetUC().Users {
		p := Protocol{}
		p.header.Seq = uint16(u.SendCount)
		p.header.MessageType = 0x0E
		p.data = append(p.data, 0)
		p.data = append(p.data, Uint32ToBytes(gameId)...)
		p.data = append(p.data, cs.GameStatus)
		p.data = append(p.data, uint8(len(cs.Players)))
		p.data = append(p.data, 4)
		user.SendPacket(server, p)
	}
	{
		p := Protocol{}
		p.header.Seq = uint16(user.SendCount)
		p.header.MessageType = 0x0D
		p.data = append(p.data, 0)
		p.data = append(p.data, Uint32ToBytes(uint32(len(cs.Players))-1)...)
		for i, _ := range cs.Players {
			if i != addr.String() {
				roomUser := GetUC().Users[i]
				p.data = append(p.data, []byte(roomUser.Name+"\x00")...)
				p.data = append(p.data, Uint32ToBytes(roomUser.Ping)...)
				p.data = append(p.data, Uint16ToBytes(roomUser.UserId)...)
				p.data = append(p.data, roomUser.ConnectType)
			}
		}
		user.SendPacket(server, p)
	}
	for _, u := range GetUC().Users {
		p := Protocol{}
		p.header.Seq = uint16(u.SendCount)
		p.header.MessageType = 0x0C
		p.data = append(p.data, 0)
		p.data = append(p.data, Uint32ToBytes(gameId)...)
		p.data = append(p.data, []byte(user.Name+"\x00")...)
		p.data = append(p.data, Uint32ToBytes(user.Ping)...)
		p.data = append(p.data, Uint16ToBytes(user.UserId)...)
		p.data = append(p.data, user.ConnectType)
		u.SendPacket(server, p)
	}
}

func SvcQuitGame(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcQuitGame ===============")
	user := GetUC().Users[addr.String()]
	log.Infof("[SvcQuitGame] %s", ph.data)
	if user == nil {
		return
	}
	if len(ph.data) == 3 && ph.data[0] == 0x00 && ph.data[1] == 0xff && ph.data[2] == 0xff {
		cs := GetUC().Channels[user.GameRoomId]
		if cs == nil {
			return
		}
		log.Infof("delete cs: %+v", cs)
		delete(cs.Players, user.IpAddr.String())
		closeGame := false
		if len(cs.Players) == 0 {
			err := GetUC().DeleteChannel(user.GameRoomId)
			closeGame = true
			if err != nil {
				log.Infof("Delete Channel Error")
			} else {
				log.Infof("Delete Channel Ok")
			}
		}

		if closeGame {
			log.Infof("closeGame Close")
			// Game Status Noti
			for _, u := range GetUC().Users {
				p := Protocol{}
				p.header.Seq = uint16(u.SendCount)
				p.header.MessageType = 0x10
				p.data = append(p.data, 0)
				p.data = append(p.data, Uint32ToBytes(cs.GameId)...)
				u.SendPacket(server, p)
			}
		} else {
			log.Infof("closeGame Leave")
			// Game Status Noti
			for _, u := range GetUC().Users {
				p := Protocol{}
				p.header.Seq = uint16(u.SendCount)
				p.header.MessageType = 0x0E
				p.data = append(p.data, 0)
				p.data = append(p.data, Uint32ToBytes(cs.GameId)...)
				p.data = append(p.data, cs.GameStatus)
				p.data = append(p.data, uint8(len(cs.Players)))
				p.data = append(p.data, 4)
				u.SendPacket(server, p)
			}

		}
		for _, u := range GetUC().Users {
			p := Protocol{}
			p.header.Seq = uint16(user.SendCount)
			p.header.MessageType = 0x0B
			p.data = append(p.data, []byte(user.Name+"\x00")...)
			p.data = append(p.data, Uint16ToBytes(user.UserId)...)
			u.SendPacket(server, p)
		}
	}
	// user := GetUC().Users[addr.String()]
	// gameId := binary.LittleEndian.Uint32(ph.data[1:5])
	// log.Infof("join gameid: %+v", Uint32ToBytes(gameId))
}

func SvcKeepAlive(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcKeepAlive ===============")
}
func SvcKickUserFromGame(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcKickUserFromGame ===============")
}
func SvcStartGame(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcStartGame ===============")
}
func SvcGameData(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcGameData ===============")
}
func SvcGameCache(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcGameCache ===============")
}
func SvcDropGame(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcDropGame ===============")
}
