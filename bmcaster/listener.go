package main

import (
	"github.com/golang/protobuf/proto"
	"log"
	"math/rand"
	"net"
	"time"
)

type Msg struct {
	MsgID int
	Val   string
}

func Listener(store *msgstore) {
	log.Println("Starting Listener")
	rand.Seed(time.Now().UnixNano())
	addr, err := net.ResolveUDPAddr("udp", SrvAddr)
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.ListenMulticastUDP("udp", nil, addr)
	l.SetReadBuffer(MaxDatagramSize)
	defer l.Close()
	m := new(NewMsg)
	b := make([]byte, MaxDatagramSize)
	for {
		n, src, err := l.ReadFromUDP(b)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}
		log.Println(n, "bytes read from", src)
		//TODO remove later. For testing Fetcher only
		if rand.Intn(100) < MCastDropPercent {
			continue
		}
		err = proto.Unmarshal(b[:n], m)
		if err != nil {
			log.Fatal("protobuf Unmarshal failed", err)
		}
		id := m.GetHead().GetMsgId()
		log.Println("CONFIG-UPDATE-RECEIVED { \"update_id\" =", id, "}")
		//TODO check whether value already exists in store?
		store.Add(id)
		SaveAsFile(id, b[:n], StoreDir)
		m.Reset()
	}
}
