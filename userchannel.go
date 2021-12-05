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
	Id           string
	Name         string
	EmulName     string
	Ping         uint32
	ConnectType  uint8
	PlayerStatus uint8
	AckCount     uint32
	SendCount    int32
	Packets      []Protocol
}

func NewUserStruct() *UserStruct {
	return &UserStruct{Packets: make([]Protocol, 0)}
}

type ChannelStruct struct {
	GameName   string
	GameId     string
	EmulName   string
	CreatorId  string
	Players    map[string]struct{}
	GameStatus uint32
}

func NewChannelStruct() *ChannelStruct {
	return &ChannelStruct{Players: map[string]struct{}{}}
}

type UserChannel struct {
	Users    map[string]*UserStruct
	Channels map[string]*ChannelStruct
	RandomId int
}

var instance *UserChannel
var once sync.Once

func GetUC() *UserChannel {
	once.Do(func() {
		instance = &UserChannel{
			Users:    map[string]*UserStruct{},
			Channels: map[string]*ChannelStruct{},
		}
	})
	return instance
}

func (t *UserChannel) AddChannel(ch string, u ChannelStruct) error {
	if _, ok := t.Channels[ch]; !ok {
		t.Channels[ch] = &u
	} else {
		return errors.New("exist")
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
func (t *UserChannel) MakeServerStatus(seq uint16, user *UserStruct) Protocol {
	ret := make([]byte, 0)
	ret = append(ret, 0)
	ret = append(ret, Uint32ToBytes(uint32(len(GetUC().Users)-1))...)
	ret = append(ret, Uint32ToBytes(0)...)

	for _, j := range GetUC().Users {
		// 본인은 제외함.
		log.Infof("MakeStatus %s != %s", j.IpAddr.String(), user.IpAddr.String())
		if j.IpAddr.String() != user.IpAddr.String() {
			ret = append(ret, []byte(j.Name+"\x00")...)
			ret = append(ret, Uint32ToBytes(j.Ping)...)
			ret = append(ret, j.ConnectType)
			ret = append(ret, []byte(j.Id[:2])...)
			ret = append(ret, j.PlayerStatus)
		}
	}
	fmt.Printf("%s\n", hex.Dump(ret))
	p := Protocol{}
	p.data = ret
	p.header.MessageType = 0x04
	p.header.Seq = seq
	return p
}
