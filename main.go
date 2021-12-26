package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	// prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var log = logrus.New()

func init() {
	// logrus.For
	// log.Formatter = new(prefixed.TextFormatter)
	log.Level = logrus.InfoLevel
	Formatter := new(logrus.TextFormatter)
	Formatter.TimestampFormat = "2006-01-02T15:04:05.999999999Z07:00"
	Formatter.FullTimestamp = true
	log.SetFormatter(Formatter)
	// 2006-01-02T15:04:05.999999999Z07:00
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
					user := GetUC().Users[clientAddress.String()]
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
							SvcUserQuit(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x03 {
							GetUC().AddUser(clientAddress, NewUserStruct())
							GetUC().NextUserId += 1
							GetUC().Users[clientAddress.String()].UserId = GetUC().NextUserId
							GetUC().Users[clientAddress.String()].PlayerStatus = 1
							SvcUserLogin(server, clientAddress, msgtype) // , rawdata[:])
						} else if msgtype.header.MessageType == 0x06 {
							SvcAck(server, clientAddress, msgtype) // , rawdata[:])
						} else if msgtype.header.MessageType == 0x07 {
							SvcChatMesg(server, clientAddress, msgtype) // , rawdata[:])
						} else if msgtype.header.MessageType == 0x08 {
							SvcGameChat(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x09 {
							SvcKeepAlive(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x0a {
							SvcCreateGame(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x0b {
							SvcQuitGame(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x0c {
							SvcJoinGame(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x0f {
							SvcKickUserFromGame(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x11 {
							SvcStartGame(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x12 {
							SvcGameData(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x13 {
							SvcGameCache(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x14 {
							SvcDropGame(server, clientAddress, msgtype)
						} else if msgtype.header.MessageType == 0x15 {
							SvcReadyToPlaySignal(server, clientAddress, msgtype)
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
