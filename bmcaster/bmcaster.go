package main

import (
	"github.com/hashicorp/serf/serf"
	"log"
	"net"
	"strings"
	"time"
)

const (
	SrvAddr         = "224.0.0.1:9999"
	MaxDatagramSize = 65507
	Memberlist      = "n0,n1,n2,n3,n4,n5"
	Digestinterval  = time.Duration(5000) //Millisecond
	Digestratio     = 0.5                 //Ratio of nodes which should be send a digest msg
	Digestport      = ":6000"
	Cleantime       = time.Duration(30) //Second
	StoreDir        = "./bmstore/"
	FetcherPort     = ":3000"
	HttpConnectTO   = time.Duration(1) //Second
	HttpReadWriteTO = time.Duration(2) //Second
)

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting BMcaster")
	/* Create the initial memberlist from a safe configuration.
	   Please reference the godoc for other default config types.
	   http://godoc.org/github.com/hashicorp/memberlist#Config
	*/
	config := serf.DefaultConfig()
	ipadd, err := externalIP()
	if err == nil && ipadd != "" {
		config.MemberlistConfig.BindAddr = ipadd
	}
	list, err := serf.Create(config)
	if err != nil {
		log.Printf("Failed to create memberlist: " + err.Error())
	}
	var store msgstore

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
	go TcpMsgServer()
	go Digestrx(&store)
	go Listener(&store)
	go Digesttx(&store, list)
	for {
		//Clean up stores periodically
		time.Sleep(Cleantime * time.Second)
		store.Clean()
	}
}
