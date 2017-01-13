package main

import (
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/memberlist"
	"log"
	"math/rand"
	"net"
	"time"
)

type Digest struct {
	Ranges []keyrange
}

func (m Digest) makeProtobuf() *NewDigest {
	p := &NewDigest{}
	for _, r := range m.Ranges {
		pRange := new(Range)
		pRange.LLimit = proto.Int64(r.Llimit)
		pRange.ULimit = proto.Int64(r.Ulimit)
		p.Ranges = append(p.Ranges, pRange)
	}

	return p
}

func Digesttx(store *msgstore, list *memberlist.Memberlist) {
	log.Println("Starting Digesttx")
	for {
		time.Sleep(Digestinterval * time.Millisecond)
		if !store.IsEmpty() && list.NumMembers() > 1 {
			sendDigest(store, list)
		}
	}
}
func sendDigest(store *msgstore, list *memberlist.Memberlist) {
	//Get list of keys in store
	var digestmsg Digest
	digestmsg.Ranges = store.GetKeys()
	protodigest := digestmsg.makeProtobuf()
	sendbuf, err := proto.Marshal(protodigest)
	if err != nil {
		log.Fatal("protobuf Marshal failed", err)
	}
	//Get list of current members
	nodes := list.Members()
	self := list.LocalNode()
	//Send keys to random nodes in list
	i := 0
	numDigest := Digestratio * float32(len(nodes))
	log.Println("Sending Digest to", numDigest, "node(s) out of", len(nodes))
	for i < int(numDigest) {
		to := nodes[rand.Intn(len(nodes))]
		if to.Name == self.Name {
			//Do not send digest to self
			continue
		}
		addr, err := net.ResolveUDPAddr("udp", to.Name+Digestport)
		if err != nil {
			log.Printf("Cannot find host %s", to.Name)
			i++
			continue
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
func Digestrx(store *msgstore) {
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
		n, src, err := c.ReadFromUDP(b)
		if err != nil {
			log.Println("Error: ", err)
		}
		getDigest(b[:n], src.String(), store)
	}
}

func getDigest(b []byte, src string, store *msgstore) {
	//Decode keys
	protodigest := new(NewDigest)
	err := proto.Unmarshal(b, protodigest)
	if err != nil {
		log.Fatal("protobuf Unmarshal failed: ", err)
	}
	log.Println("Received digest from ", src)
	var digestmsg Digest
	for _, pRange := range protodigest.Ranges {
		var pkeyrange keyrange
		pkeyrange.Llimit = pRange.GetLLimit()
		pkeyrange.Ulimit = pRange.GetULimit()
		digestmsg.Ranges = append(digestmsg.Ranges, pkeyrange)
	}
	//Compare keys with store as well as disk
	unknown := store.DiffKeys(digestmsg.Ranges)
	if unknown != nil {
		host, _, _ := net.SplitHostPort(src)
		log.Println("unknown keys:", unknown, host)
		go TcpFetcher(store, unknown, host)
	}
}
