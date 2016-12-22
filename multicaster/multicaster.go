package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
)

const (
	srvAddr         = "224.0.0.1:9999"
	maxDatagramSize = 8192
	interval        = time.Duration(1000) //Millisecond
	maxLength       = 100                 //LEngth of msg
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

func RandStringBytesMaskImprSrc(n int) string {
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

func RandomMsgGen(srvAddr string, interval time.Duration, maxLength int) {
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

	id := 1
	for {
		udpsendbuf := new(bytes.Buffer)
		encoder := gob.NewEncoder(udpsendbuf)
		m := Msg{id, RandStringBytesMaskImprSrc(maxLength)}
		log.Println("Sending message", m)
		err = encoder.Encode(m)
		if err != nil {
			log.Fatal("gob Encode failed", err)
		}
		sendbuf := make([]byte, 1600)
		n, err := udpsendbuf.Read(sendbuf)
		if err != nil {
			log.Fatal(err, "filling slice from buffer")
		}
		//log.Println("Sending bytes", sendbuf[:n])
		n, err = c.Write(sendbuf[:n]) /*fire it out an existing udp
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
	RandomMsgGen(srvAddr, interval, maxLength)
	fmt.Println("done")
}
