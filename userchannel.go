package main

import (
	"errors"
	"sync"
)

type UserStruct struct {
	Id           string
	Name         string
	Ping         uint32
	ConnectType  uint32
	PlayerStatus uint32
	AckCount     uint32
	SendCount    int32
	Packets      [][]byte
}

func NewUserStruct() *UserStruct {
	return &UserStruct{Packets: make([][]byte, 0)}
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
}

var instance *UserChannel
var once sync.Once

func GetUC() *UserChannel {
	once.Do(func() {
		instance = &UserChannel{}
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

func (t *UserChannel) AddUser(userid string, u UserStruct) error {
	if _, ok := t.Users[userid]; !ok {
		t.Users[userid] = &u
	} else {
		return errors.New("exist")
	}
	return nil
}
