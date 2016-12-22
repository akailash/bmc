package main

import (
	"github.com/hashicorp/serf/serf"
	"log"
	"strings"
	"time"
)

const (
	SrvAddr         = "224.0.0.1:9999"
	MaxDatagramSize = 65507
	Memberlist      = "n1,n2,n3,n4,n5"
	Digestinterval  = time.Duration(5000) //Millisecond
	Digestratio     = 0.5                 //Ratio of nodes which should be send a digest msg
	Digestport      = ":6000"
	Cleantime       = time.Duration(30) //Second
	StoreDir        = "./bmstore/"
	FetcherPort     = ":3000"
	FetchDuration   = time.Duration(5000) //Millisecond
	HttpConnectTO   = time.Duration(1)    //Second
	HttpReadWriteTO = time.Duration(2)    //Second
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting BMcaster")
	/* Create the initial memberlist from a safe configuration.
	   Please reference the godoc for other default config types.
	   http://godoc.org/github.com/hashicorp/memberlist#Config
	*/
	list, err := serf.Create(serf.DefaultConfig())
	if err != nil {
		log.Printf("Failed to create memberlist: " + err.Error())
	}
	var store msgstore
	var wanted wantedstore

	// Join an existing cluster by specifying at least one known member.
	n, err := list.Join(strings.Split(Memberlist, ","), true)
	if err != nil {
		log.Printf("Failed to join cluster: " + err.Error())
	}
	log.Printf("%d Members existing.", n)

	// Ask for members of the cluster
	for _, member := range list.Members() {
		log.Printf("Member: %s %s\n", member.Name, member.Addr)
	}

	// Continue doing whatever you need, memberlist will maintain membership
	// information in the background. Delegates can be used for receiving
	// events when members join or leave.
	//Init maps in store
	store.cache = make(map[int]Msg)
	wanted.cache = make(map[int][]string)
	go MsgServer()
	go Fetcher(&store, &wanted)
	go Digestrx(&store, &wanted)
	go Listener(&store)
	go Digesttx(&store, list)
	for {
		//Clean up stores periodically
		time.Sleep(Cleantime * time.Second)
		store.Clean()
	}
}
