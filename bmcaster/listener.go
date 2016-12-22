package main

import (
	"bytes"
	"encoding/gob"
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
	var m Msg
	for {
		b := make([]byte, MaxDatagramSize)
		n, src, err := l.ReadFromUDP(b)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}
		log.Println(n, "bytes read from", src)
		buf := bytes.NewBuffer(b)
		decoder := gob.NewDecoder(buf)
		err = decoder.Decode(&m)
		if err != nil {
			log.Fatal("gob Decode failed", err)
		}
		log.Println("Received message", m.MsgID)
		//TODO check whether value already exists in store?
		store.Add(m)
		SaveAsJson(m, StoreDir)
	}
}
