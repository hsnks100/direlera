package main

import (
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

	s := NewService()
	go s.RunService()

	time.Sleep(1000 * time.Second)

}
