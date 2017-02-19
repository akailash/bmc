package main

import (
	"github.com/hashicorp/memberlist"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

//TODO remove when pprof is not needed
import _ "net/http/pprof"

const (
	SrvAddr          = "224.0.0.1:9999"
	MaxDatagramSize  = 65507
	Memberlist       = "bimodalmcast_node_1"
	Digestinterval   = time.Duration(5000) //Millisecond
	Digestratio      = 0.5                 //Ratio of nodes which should be send a digest msg
	Digestport       = ":6000"
	Cleantime        = time.Duration(30) //Second
	StoreDir         = "./bmstore/"
	LogDir           = "./configlog/"
	FetcherPort      = ":3000"
	MCastDropPercent = 20 //To simulate packet loss and trigger Fetcher
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
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	/*	if _, err := os.Stat(LogDir); err != nil {
			if os.IsNotExist(err) {
				os.Mkdir(LogDir, 0755)
			} else {
				log.Fatalln(err)
			}
		}
		f, err := os.OpenFile(LogDir+"config.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("error opening file: %v", err)
		} else {
			defer f.Close()
			log.SetOutput(f)
		}*/
	log.Println("Starting BMcaster")
	/* Create the initial memberlist from a safe configuration.
	   Please reference the godoc for other default config types.
	   http://godoc.org/github.com/hashicorp/memberlist#Config
	*/
	config := memberlist.DefaultLocalConfig()
	ipadd, err := externalIP()
	if err == nil && ipadd != "" {
		config.BindAddr = ipadd
	}
	list, err := memberlist.Create(config)
	if err != nil {
		log.Printf("Failed to create memberlist: " + err.Error())
	}
	//log.SetPrefix(list.LocalNode().Name + "\t")
	var store msgstore

	// Join an existing cluster by specifying at least one known member.
	n, err := list.Join(strings.Split(Memberlist, ","))
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

	//TODO remove when pprof is not needed
	go func() {
		log.Fatal(http.ListenAndServe(":8080", http.DefaultServeMux))
	}()

	for {
		//Clean up stores periodically
		time.Sleep(Cleantime * time.Second)
		store.Clean()
	}
}
