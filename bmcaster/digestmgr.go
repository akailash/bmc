package main

import (
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/serf/serf"
	"log"
	"math/rand"
	"net"
	"time"
)

type Digest struct {
	Ranges []keyrange
	Keys   []int64
}

func (m Digest) makeProtobuf() *NewDigest {
	p := &NewDigest{
		Keys: m.Keys,
	}
	for _, r := range m.Ranges {
		pRange := new(Range)
		pRange.LLimit = proto.Int64(r.Llimit)
		pRange.ULimit = proto.Int64(r.Ulimit)
		p.Ranges = append(p.Ranges, pRange)
	}

	return p
}

func Digesttx(store *msgstore, list *serf.Serf) {
	log.Println("Starting Digesttx")
	for {
		time.Sleep(Digestinterval * time.Millisecond)
		//Get list of keys in store
		var digestmsg Digest
		digestmsg.Ranges, digestmsg.Keys = store.GetKeys()
		protodigest := digestmsg.makeProtobuf()
		sendbuf, err := proto.Marshal(protodigest)
		if err != nil {
			log.Fatal("protobuf Marshal failed", err)
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
			c.Write(sendbuf)
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
			protodigest := new(NewDigest)
			err = proto.Unmarshal(b[:n], protodigest)
			if err != nil {
				log.Fatal("protobuf Unmarshal failed: ", err, " No. of bytes in digest: ", n)
			}
			log.Println("Received ", n, " bytes digest from ", src)
			var digestmsg Digest
			digestmsg.Keys = protodigest.GetKeys()
			for _, pRange := range protodigest.Ranges {
				var pkeyrange keyrange
				pkeyrange.Llimit = pRange.GetLLimit()
				pkeyrange.Ulimit = pRange.GetULimit()
				digestmsg.Ranges = append(digestmsg.Ranges, pkeyrange)
			}
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
