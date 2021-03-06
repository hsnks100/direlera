package main

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"sync"
)

type UserStruct struct {
	IpAddr       net.Addr
	UserId       uint16
	Name         string
	EmulName     string
	Ping         uint32
	ConnectType  uint8
	PlayerStatus uint8
	AckCount     uint32
	SendCount    int32
	CurSeq       int
	GameRoomId   uint32
	RoomOrder    int
	Inputs       []byte
	Packets      []Protocol // 보낸 패킷들
	CallCnt      int
	CallCntTime  int64
	LastInput    []byte

	CachePosition    uint8
	IncomingGameData map[uint8][]byte
	RequireFrame     int
	// 보내기전에 HitCache 에 조회해보고 있으면 value 보내고
	//                                   없으면 GameData 보냄.
	OutcomingGameCache     map[uint8][]byte
	OutcomingHitCache      map[string]uint8
	OutcomingCachePosition uint8
}

func (u *UserStruct) ResetOutcoming() {
	u.OutcomingCachePosition = 0
	u.OutcomingGameCache = map[uint8][]byte{}
	u.OutcomingHitCache = map[string]uint8{}
	u.Inputs = make([]byte, 0)

	u.CachePosition = 0
	u.IncomingGameData = map[uint8][]byte{}
	u.RequireFrame = 0
}

func NewUserStruct() *UserStruct {
	return &UserStruct{Packets: make([]Protocol, 0), CurSeq: 0, Inputs: make([]byte, 0), IncomingGameData: map[uint8][]byte{}, LastInput: []byte{},
		OutcomingGameCache: map[uint8][]byte{}, OutcomingHitCache: map[string]uint8{}}
}

func (u *UserStruct) SendPacket(server net.PacketConn, p Protocol) {
	p.header.Seq = uint16(u.SendCount)
	u.Packets = append(u.Packets, p)
	packet := make([]byte, 0)
	packet = append(packet, 1) // N = 1
	packet = append(packet, p.MakePacket()...)
	server.WriteTo(packet, u.IpAddr)
	// log.Infof("WriteTo: %s", u.IpAddr.String())
	u.SendCount += 1
}

type ChannelStruct struct {
	GameName      string
	GameId        uint32
	EmulName      string
	CreatorId     string
	Players       map[string]struct{}
	GameStatus    uint8
	CachePosition int
}

func NewChannelStruct() *ChannelStruct {
	return &ChannelStruct{Players: map[string]struct{}{}}
}

type UserChannel struct {
	Users      map[string]*UserStruct
	Channels   map[uint32]*ChannelStruct
	NextUserId uint16
}

var instance *UserChannel
var once sync.Once

func GetUC() *UserChannel {
	once.Do(func() {
		instance = &UserChannel{
			Users:      map[string]*UserStruct{},
			Channels:   map[uint32]*ChannelStruct{},
			NextUserId: 0x01,
		}
	})
	return instance
}

func (t *UserChannel) AddChannel(ch uint32, u *ChannelStruct) error {
	if _, ok := t.Channels[ch]; !ok {
		t.Channels[ch] = u
	} else {
		return errors.New("exist")
	}
	return nil
}

func (t *UserChannel) DeleteChannel(ch uint32) error {
	if _, ok := t.Channels[ch]; !ok {
		return errors.New("don't exist")
	} else {
		delete(t.Channels, ch)
	}
	return nil
}
func (t *UserChannel) AddUser(ipaddr net.Addr, u *UserStruct) error {
	if _, ok := t.Users[ipaddr.String()]; !ok {
		u.IpAddr = ipaddr
		t.Users[ipaddr.String()] = u
	} else {
		log.Infof("exist user")
		return errors.New("exist")
	}
	return nil
}

func Uint32ToBytes(i uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, i)
	return b
}
func Uint16ToBytes(i uint16) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, i)
	return b
}
func (t *UserChannel) MakeServerStatus(seq uint16, user *UserStruct) Protocol {
	ret := make([]byte, 0)
	ret = append(ret, 0)
	ret = append(ret, Uint32ToBytes(uint32(len(GetUC().Users)-1))...)
	ret = append(ret, Uint32ToBytes(uint32(len(GetUC().Channels)))...)

	for _, j := range GetUC().Users {
		// 본인은 제외함.
		if j.IpAddr.String() != user.IpAddr.String() {
			log.Infof("Make ServerStatus User %s", j.Name)
			ret = append(ret, []byte(j.Name+"\x00")...)
			ret = append(ret, Uint32ToBytes(j.Ping)...)
			// ret = append(ret, 0)
			// ret = append(ret, j.ConnectType)
			ret = append(ret, j.PlayerStatus)
			ret = append(ret, Uint16ToBytes(j.UserId)...)
			ret = append(ret, j.ConnectType)
		}
	}
	for _, j := range GetUC().Channels {
		ret = append(ret, []byte(j.GameName+"\x00")...)
		ret = append(ret, Uint32ToBytes(j.GameId)...)
		ret = append(ret, []byte(j.EmulName+"\x00")...)
		ret = append(ret, []byte(j.CreatorId+"\x00")...)
		ret = append(ret, []byte(fmt.Sprintf("%d/%d\x00", len(j.Players), 4))...)
		ret = append(ret, j.GameStatus)
	}
	fmt.Printf("%s\n", hex.Dump(ret))
	p := Protocol{}
	p.data = ret
	p.header.MessageType = 0x04
	p.header.Seq = seq
	return p
}
