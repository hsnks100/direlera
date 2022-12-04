package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
)

type Service struct {
	userChannel *UserChannel
}

func NewService() *Service {
	s := &Service{userChannel: NewUserChannel()}
	return s
}

func (s *Service) RunService() {
	server, err := net.ListenPacket("udp", "0.0.0.0:27999")
	if err != nil {
		log.Fatal(err)
	}
	func() {
		for {
			buf := make([]byte, 1024)
			n, clientAddress, err := server.ReadFrom(buf)
			if err != nil {
				continue
			}
			// fmt.Printf("<== received: %+v ( %s ) / %s / %s\n", buf[:n], string(buf[:n]), "from", clientAddress)
			ph := ProtocolPackets{}
			buf2 := bytes.NewBuffer(buf[:ProtocolPacketsSize])
			err = binary.Read(buf2, binary.LittleEndian, &ph)
			// N 만큼의 프로토콜 배열로 다 얻어옴. 그리고 유저가 처리해야할 시퀀스 번호를 가지고 메시지를 구함.
			messages := GetProtocolFromBytes(buf[1:n])
			if len(messages) >= 1 {
				processCount := 0
				// 처리해야할 메시지가 여러개인 경우 여러개 처리함.
				for {
					user := s.userChannel.Users[clientAddress.String()]
					msgtype := Protocol{}
					match := false
					userSeq := -1
					if user != nil {
						userSeq = user.CurSeq
					}
					log.Debugf("<<<<<<<<<<<<< recovery: want seq: %d ", userSeq+1)
					for _, j := range messages {
						if int(j.header.Seq) == userSeq+1 {
							msgtype = j
							match = true
							if user != nil {
								user.CurSeq = userSeq + 1
							}

							break
						}
						// log.Infof("seq: %d", j.header.Seq)
					}
					log.Debugf("recovery >>>>>>>>>>>>>>> ")
					if match {
						processCount += 1
						if msgtype.header.MessageType == 0x01 {
							s.SvcUserQuit(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x03 {
							s.userChannel.AddUser(clientAddress, NewUserStruct())
							s.userChannel.NextUserId += 1
							s.userChannel.Users[clientAddress.String()].UserId = s.userChannel.NextUserId
							s.userChannel.Users[clientAddress.String()].PlayerStatus = 1
							s.SvcUserLogin(server, clientAddress, msgtype) // , rawdata[:])
						} else if msgtype.header.MessageType == 0x06 {
							s.SvcAck(server, clientAddress, msgtype) // , rawdata[:])
						} else if msgtype.header.MessageType == 0x07 {
							s.SvcChatMesg(server, clientAddress, msgtype) // , rawdata[:])
						} else if msgtype.header.MessageType == 0x08 {
							s.SvcGameChat(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x09 {
							s.SvcKeepAlive(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x0a {
							s.SvcCreateGame(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x0b {
							s.SvcQuitGame(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x0c {
							s.SvcJoinGame(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x0f {
							s.SvcKickUserFromGame(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x11 {
							s.SvcStartGame(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x12 {
							s.SvcGameData(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x13 {
							s.SvcGameCache(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x14 {
							s.SvcDropGame(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x15 {
							s.SvcReadyToPlaySignal(server, clientAddress, msgtype)
						} else {
							log.Infof("unknown Message[%2x]: %+v", msgtype.header.MessageType, msgtype)
						}
					} else {
						// log.Infof("match finish, proc count: %d", processCount)
						break
					}

				}
			} else {
				break
			}
		}
	}()
}

func (s *Service) SvcUserLogin(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcUserLogin ===============")
	/*
		real:		01 00 00 12 00 05 00 00 00 00 00 01 00 00 00 02 00 00 00 03 00 00 00                                                .
		me:         01 00 00 12 00 05 00 00 00 00 00 01 00 00 00 02 00 00 00 03 00 00 00
	*/
	fields := bytes.Split(ph.data, []byte{0})
	for i, j := range fields {
		log.Infof("SvcUserLogin %d: %s", i, j)
	}

	user := s.userChannel.Users[addr.String()]
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

func (s *Service) SvcAck(server net.PacketConn, addr net.Addr, ph Protocol) { // Header, body []byte) {
	log.Infof("================ SvcAck ===============")
	user := s.userChannel.Users[addr.String()]
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
			p := s.userChannel.MakeServerStatus(uint16(user.SendCount), user)
			user.SendPacket(server, p)
		}
		// joined packet
		{
			for _, u := range s.userChannel.Users {
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

func (s *Service) SvcChatMesg(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcChatMesg ===============")
	user := s.userChannel.Users[addr.String()]
	// chatmesg := ph.data[1:]
	for _, u := range s.userChannel.Users {
		p := Protocol{}
		p.header.Seq = uint16(u.SendCount)
		p.header.MessageType = 0x07
		p.data = append(p.data, []byte(user.Name+"\x00")...)
		p.data = append(p.data, ph.data[1:]...)
		u.SendPacket(server, p)
	}
}
func (s *Service) SvcGameChat(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcGameChat ===============")
	user := s.userChannel.Users[addr.String()]
	cs := s.userChannel.Channels[user.GameRoomId]
	for _, j := range cs.Players {
		u := s.userChannel.Users[j]
		p := Protocol{}
		p.header.Seq = uint16(u.SendCount)
		p.header.MessageType = 0x08
		p.data = append(p.data, []byte(user.Name+"\x00")...)
		p.data = append(p.data, ph.data[1:]...)
		u.SendPacket(server, p)
	}
}

func (s *Service) SvcUserQuit(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcUserQuit ===============")
	user := s.userChannel.Users[addr.String()]
	clientMsg := ph.data[3:]

	// chatmesg := ph.data[1:]
	for _, u := range s.userChannel.Users {
		p := Protocol{}
		p.header.Seq = uint16(u.SendCount)
		p.header.MessageType = 0x01
		p.data = append(p.data, []byte(user.Name+"\x00")...)
		p.data = append(p.data, Uint16ToBytes(user.UserId)...)
		p.data = append(p.data, []byte(clientMsg)...)
		u.SendPacket(server, p)
	}
	delete(s.userChannel.Users, addr.String())
}

var g_gameId = 0x1

func (s *Service) SvcCreateGame(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcCreateGame ===============")
	user := s.userChannel.Users[addr.String()]
	fields := bytes.Split(ph.data, []byte{0})
	for i, j := range fields {
		log.Infof("SvcCreateGame %d: %s", i, j)
	}
	for _, u := range s.userChannel.Users {
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
	cs.Players = append(cs.Players, user.IpAddr.String())
	s.userChannel.AddChannel(cs.GameId, cs)
	// update game status
	for _, u := range s.userChannel.Users {
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
		for _, j := range cs.Players {
			roomUser := s.userChannel.Users[j]
			p.data = append(p.data, []byte(roomUser.Name+"\x00")...)
			p.data = append(p.data, Uint32ToBytes(roomUser.Ping)...)
			p.data = append(p.data, Uint16ToBytes(roomUser.UserId)...)
			p.data = append(p.data, roomUser.ConnectType)
		}
		user.SendPacket(server, p)
	}
	// Join Game Noti
	for _, u := range s.userChannel.Users {
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

func (s *Service) SvcJoinGame(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcJoinGame ===============")
	user := s.userChannel.Users[addr.String()]
	gameId := binary.LittleEndian.Uint32(ph.data[1:5])
	log.Infof("join gameid: %+v", Uint32ToBytes(gameId))
	connType := ph.data[12:13][0]
	cs := s.userChannel.Channels[gameId]
	cs.Players = append(cs.Players, user.IpAddr.String())
	user.GameRoomId = gameId
	_ = connType
	for _, u := range s.userChannel.Users {
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
		for _, j := range cs.Players {
			if j != addr.String() {
				roomUser := s.userChannel.Users[j]
				p.data = append(p.data, []byte(roomUser.Name+"\x00")...)
				p.data = append(p.data, Uint32ToBytes(roomUser.Ping)...)
				p.data = append(p.data, Uint16ToBytes(roomUser.UserId)...)
				p.data = append(p.data, roomUser.ConnectType)
			}
		}
		user.SendPacket(server, p)
	}
	for _, u := range s.userChannel.Users {
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

func (s *Service) SvcQuitGame(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcQuitGame ===============")
	user := s.userChannel.Users[addr.String()]
	log.Infof("[SvcQuitGame] %s", ph.data)
	if user == nil {
		return
	}
	if len(ph.data) == 3 && ph.data[0] == 0x00 && ph.data[1] == 0xff && ph.data[2] == 0xff {
		cs := s.userChannel.Channels[user.GameRoomId]
		if cs == nil {
			return
		}
		log.Infof("delete cs: %+v", cs)
		tmp := cs.Players[:0]
		for _, j := range cs.Players {
			if j != user.IpAddr.String() {
				tmp = append(tmp, j)
			}
		}
		cs.Players = tmp
		closeGame := false
		if len(cs.Players) == 0 {
			err := s.userChannel.DeleteChannel(user.GameRoomId)
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
			for _, u := range s.userChannel.Users {
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
			for _, u := range s.userChannel.Users {
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
		for _, u := range s.userChannel.Users {
			p := Protocol{}
			p.header.Seq = uint16(user.SendCount)
			p.header.MessageType = 0x0B
			p.data = append(p.data, []byte(user.Name+"\x00")...)
			p.data = append(p.data, Uint16ToBytes(user.UserId)...)
			u.SendPacket(server, p)
		}
	}
	// user := s.userChannel.Users[addr.String()]
	// gameId := binary.LittleEndian.Uint32(ph.data[1:5])
	// log.Infof("join gameid: %+v", Uint32ToBytes(gameId))
}

func (s *Service) SvcKeepAlive(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcKeepAlive ===============")
}
func (s *Service) SvcKickUserFromGame(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcKickUserFromGame ===============")
}
func (s *Service) SvcStartGame(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcStartGame ===============")
	if len(ph.data) == 5 && ph.data[0] == 0x00 && ph.data[1] == 0xff && ph.data[2] == 0xff && ph.data[3] == 0xff && ph.data[4] == 0xff {
		user := s.userChannel.Users[addr.String()]
		cs := s.userChannel.Channels[user.GameRoomId]
		cs.GameStatus = 2
		order := uint8(0)
		for _, u := range s.userChannel.Users {
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
		for _, j := range cs.Players {

			order += 1
			u := s.userChannel.Users[j]
			u.RoomOrder = int(order)
			p := Protocol{}
			p.header.MessageType = 0x11
			p.data = append(p.data, 0)
			p.data = append(p.data, Uint16ToBytes(1)...)
			p.data = append(p.data, order)
			u.PlayerOrder = int(order)
			tt := len(cs.Players)
			p.data = append(p.data, byte(tt))
			u.ResetOutcoming()
			u.SendPacket(server, p)
		}
	}
}
func (s *Service) SvcReadyToPlaySignal(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcReadyToPlaySignal ===============")
	fmt.Println(hex.Dump(ph.data))
	if len(ph.data) == 1 && ph.data[0] == 0x00 {
		user := s.userChannel.Users[addr.String()]
		cs := s.userChannel.Channels[user.GameRoomId]
		cs.GameStatus = 1
		for _, u := range s.userChannel.Users {
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
		for _, j := range cs.Players {
			u := s.userChannel.Users[j]
			p := Protocol{}
			p.header.MessageType = 0x15
			p.data = append(p.data, 0)
			u.SendPacket(server, p)
		}
	}
}

var seq bool = false

func (s *Service) SvcGameData(server net.PacketConn, addr net.Addr, ph Protocol) {
	seq = false
	log.Infof("================ SvcGameData %s===============", addr.String())
	fmt.Println(hex.Dump(ph.data))
	if len(ph.data) < 3 {
		return
	}
	gameDataLength := binary.LittleEndian.Uint16(ph.data[1:3])
	if len(ph.data) < 3+int(gameDataLength) {
		return
	}
	gameData := ph.data[3 : 3+gameDataLength]
	user := s.userChannel.Users[addr.String()]
	// user.Inputs = gameData
	user.cacheSystem.PutData(gameData)

	cs := s.userChannel.Channels[user.GameRoomId]
	for _, j := range cs.Players {
		u := s.userChannel.Users[j]
		u.PlayersInput[user.PlayerOrder-1] = append(u.PlayersInput[user.PlayerOrder-1], gameData...)
		// fmt.Printf("insert gameData player: %d, playername: %s, %+v, total length: %d\n", user.PlayerOrder-1, user.Name, gameData, len(u.PlayersInput[user.PlayerOrder-1]))
	}

	s.InputProcess2(server, addr, ph)
}

func (s *Service) SvcGameCache(server net.PacketConn, addr net.Addr, ph Protocol) {
	// log.Infof("================ SvcGameCache ===============")
	// fmt.Println(hex.Dump(ph.data))
	if len(ph.data) != 2 {
		return
	}
	cachePosition := ph.data[1]
	user := s.userChannel.Users[addr.String()]
	inputData, err := user.cacheSystem.GetData(cachePosition)
	if err != nil {
		fmt.Println("no cache")
		return
	}
	cs := s.userChannel.Channels[user.GameRoomId]
	for _, j := range cs.Players {
		u := s.userChannel.Users[j]
		u.PlayersInput[user.PlayerOrder-1] = append(u.PlayersInput[user.PlayerOrder-1], inputData...)
		// fmt.Printf("cache insert gameData player: %d, playername: %s, %+v, total length: %d\n", user.PlayerOrder-1, user.Name, inputData, len(u.PlayersInput[user.PlayerOrder-1]))
	}
	s.InputProcess2(server, addr, ph)

}

func genInput(cs *ChannelStruct, user *UserStruct) []byte {
	allInput := true
	requireBytes := int(user.ConnectType) * 2
	for i := 0; i < len(cs.Players); i++ {
		if len(user.PlayersInput[i]) < requireBytes {
			// fmt.Printf("player [%d]'s input: %+v\n", i, user.PlayersInput[i])
			allInput = false
			break
		}
	}
	if !allInput {
		return nil
	}
	totalSplits := [][][]byte{}
	for i := 0; i < len(cs.Players); i++ {
		s := split(user.PlayersInput[i][:requireBytes], 2)
		totalSplits = append(totalSplits, s)
		user.PlayersInput[i] = user.PlayersInput[i][requireBytes:]
	}
	totalSum := []byte{}
	if allInput {
		for {
			for i := 0; i < len(totalSplits); i++ {
				j := totalSplits[i]
				totalSum = append(totalSum, j[0]...)
				totalSplits[i] = j[1:]
			}
			allzero := true
			for i := 0; i < len(totalSplits); i++ {
				if len(totalSplits[i]) != 0 {
					allzero = false
					break
				}
			}
			if allzero {
				break
			}
		}
	} else {
		return nil
	}
	return totalSum
}
func (s *Service) InputProcess2(server net.PacketConn, addr net.Addr, ph Protocol) {
	user := s.userChannel.Users[addr.String()]
	cs := s.userChannel.Channels[user.GameRoomId]
	// 각 플레이어마다 커넥션  타입에 맞게 패킷 생성해야함.
	for _, j := range cs.Players {
		u := s.userChannel.Users[j]
		sendData := genInput(cs, u)
		if len(sendData) > 0 {
			// 보낼 데이터가 캐시에 있는가?
			cachePos, err := u.putCache.GetCachePosition(sendData)
			if err != nil {
				u.putCache.PutData(sendData)
				p := Protocol{}
				p.header.MessageType = 0x12
				p.data = append(p.data, 0)
				p.data = append(p.data, Uint16ToBytes(uint16(len(sendData)))...)
				p.data = append(p.data, sendData...)
				u.SendPacket(server, p)
			} else {
				p := Protocol{}
				// game cache
				p.header.MessageType = 0x13
				p.data = append(p.data, 0)
				p.data = append(p.data, cachePos)
				u.SendPacket(server, p)
			}
		}
	}
}

func (s *Service) SvcDropGame(server net.PacketConn, addr net.Addr, ph Protocol) {
	log.Infof("================ SvcDropGame ===============")
	if len(ph.data) != 2 {
		return
	}
	if ph.data[0] != 0 || ph.data[1] != 0 {
		return
	}
	user := s.userChannel.Users[addr.String()]
	cs := s.userChannel.Channels[user.GameRoomId]
	cs.GameStatus = 0
	for _, u := range s.userChannel.Users {
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
	for _, j := range cs.Players {
		u := s.userChannel.Users[j]
		p := Protocol{}
		// game data
		p.header.MessageType = 0x14
		p.data = append(p.data, []byte(user.Name+"\x00")...)
		p.data = append(p.data, byte(user.RoomOrder))
		u.SendPacket(server, p)
	}
	log.Infof("================ SvcDropGame Ok ===============")
}

func split(buf []byte, lim int) [][]byte {
	var chunk []byte
	chunks := make([][]byte, 0, len(buf)/lim+1)
	for len(buf) >= lim {
		chunk, buf = buf[:lim], buf[lim:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:len(buf)])
	}
	return chunks
}
