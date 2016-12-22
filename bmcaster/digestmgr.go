package main

import (
	"bytes"
	"encoding/gob"
	"github.com/hashicorp/serf/serf"
	"log"
	"math/rand"
	"net"
	"time"
)

type Digest struct {
	Ranges []keyrange
	Keys   []int
}

func Digesttx(store *msgstore, list *serf.Serf) {
	log.Println("Starting Digesttx")
	sendbuf := make([]byte, MaxDatagramSize)
	for {
		time.Sleep(Digestinterval * time.Millisecond)
		//Get list of keys in store
		var digestmsg Digest
		digestmsg.Ranges, digestmsg.Keys = store.GetKeys()
		udpsendbuf := new(bytes.Buffer)
		encoder := gob.NewEncoder(udpsendbuf)
		err := encoder.Encode(digestmsg)
		if err != nil {
			log.Fatal("gob Encode failed", err)
		}
		n, err := udpsendbuf.Read(sendbuf)
		if err != nil {
			log.Fatal(err, "filling slice from buffer")
		}

		//Get list of current members
		nodes := list.Members()
		//Send keys to random nodes in list
		i := 0
		numDigest := Digestratio * float32(len(nodes))
		log.Println("Sending Digest to", numDigest, "node(s) out of", len(nodes))
		for i < int(numDigest) {
			to := nodes[rand.Intn(len(nodes))]
			addr, err := net.ResolveUDPAddr("udp", to.Name+Digestport)
			if err != nil {
				log.Fatal(err)
			}
			c, err := net.DialUDP("udp", nil, addr)
			if err != nil {
				log.Fatal(err)
			}
			c.Write(sendbuf[:n])
			log.Println("Sent Digest to", to.Name)
			i++
			c.Close()
		}
	}
}

func Digestrx(store *msgstore, wanted *wantedstore) {
	log.Println("Starting Digestrx")
	addr, err := net.ResolveUDPAddr("udp", Digestport)

	/* Now listen at selected port */
	c, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Println("Error: ", err)
	}

	defer c.Close()
	b := make([]byte, MaxDatagramSize)

	for {
		for {
			n, src, err := c.ReadFromUDP(b)
			if err != nil {
				log.Println("Error: ", err)
			}

			//Decode keys
			buf := bytes.NewBuffer(b[:n])
			decoder := gob.NewDecoder(buf)
			var digestmsg Digest
			err = decoder.Decode(&digestmsg)
			if err != nil {
				log.Fatal("gob Decode failed: ", err, " No. of bytes in digest: ", n)
			}
			log.Println("Received ", n, " bytes digest from ", src)
			//Compare keys with store as well as disk
			unknown := store.DiffKeys(digestmsg.Ranges, digestmsg.Keys)
			if unknown != nil {
				host := src.String()
				host, _, _ = net.SplitHostPort(src.String())
				log.Println("unknown keys:", unknown, host)
				wanted.Add(unknown, host)
			}
		}
	}
}
