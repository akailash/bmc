package main

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"log"
	"math/rand"
	"net"
	"time"
)

const (
	srvAddr         = "224.0.0.1:9999"
	maxDatagramSize = 8192
	interval        = time.Duration(1000) //Millisecond
	maxLength       = int32(100)          //Length of msg
)

type Msg struct {
	MsgID int
	Val   string
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func RandStringBytesMaskImprSrc(n int32) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax
	// characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func RandomMsgGen(forever bool, repeat int64) {
	log.Println("Generating Random Messages")
	addr, err := net.ResolveUDPAddr("udp", srvAddr)
	if err != nil {
		log.Fatal(err)
	}
	c, err := net.DialUDP("udp", nil, addr)
	defer c.Close()
	if err != nil {
		log.Fatal(err)
	}

	var id int64 = 1
	for forever != (id <= repeat) {
		m := &NewMsg{
			Head: &Header{
				MsgId:     proto.Int64(id),
				MsgLength: proto.Int32(maxLength),
				MsgType:   Header_CONFIG.Enum(),
			},
			Config: &ConfigMsg{
				Data: []byte(RandStringBytesMaskImprSrc(maxLength)),
			},
		}
		//m := Msg{id, RandStringBytesMaskImprSrc(maxLength)}
		log.Println("Sending message", m)
		sendbuf, err := proto.Marshal(m)
		if err != nil {
			log.Fatal("protobuf Marshal failed", err)
		}
		_, err = c.Write(sendbuf) /*fire it out an existing udp
		connection*/
		if err != nil {
			log.Fatal(err, "sending slice")
		}
		time.Sleep(interval * time.Millisecond)
		id++
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Multicaster")
	RandomMsgGen(true, 0)
	fmt.Println("done")
}
