package main

import (
	"github.com/golang/protobuf/proto"
	"log"
	"net"
)

type Msg struct {
	MsgID int
	Val   string
}

func Listener(store *msgstore) {
	log.Println("Starting Listener")
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
		err = proto.Unmarshal(b[:n], m)
		if err != nil {
			log.Fatal("protobuf Unmarshal failed", err)
		}
		log.Println("Received message", m.Head.GetMsgId())
		//TODO check whether value already exists in store?
		store.Add(m)
		SaveAsFile(m, StoreDir)
		m.Reset()
	}
}
