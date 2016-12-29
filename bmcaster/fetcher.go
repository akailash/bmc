package main

import (
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

func TimeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}

func NewTimeoutClient(connectTimeout time.Duration, readWriteTimeout time.Duration) *http.Client {

	return &http.Client{
		Transport: &http.Transport{
			Dial: TimeoutDialer(connectTimeout, readWriteTimeout),
		},
	}
}

func sendMsgToClient(conn net.Conn) {
	log.Println("Connection established")
	//Close the connection when the function exits
	defer conn.Close()
	data := make([]byte, MaxDatagramSize)
	//Read the data waiting on the connection and put it in the data buffer
	n, err := conn.Read(data)
	if err != nil {
		log.Println(err)
		return
	}
	protodata := new(Header)
	err = proto.Unmarshal(data[0:n], protodata)
	if err != nil {
		log.Println(err)
		return
	}
	b, err := GetMsgFromDisk(protodata.GetMsgId(), StoreDir)
	if err != nil {
		log.Println(err)
		return
	}
	n, err = conn.Write(b)
	log.Println("Sent", n, "bytes")
}

func TcpMsgServer() {
	log.Println("Starting TcpMsgServer")
	// listen on all interfaces
	ln, _ := net.Listen("tcp", FetcherPort)

	// accept connection on port
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go sendMsgToClient(conn)
	}

}

func TcpFetcher(store *msgstore, wanted *wantedstore) {
	log.Println("Starting TcpFetcher")
	for {
		time.Sleep(FetchDuration * time.Millisecond)
		//TODO should use reader lock on wanted but since writer lock is acquired within the loop its not done
		for k, srcs := range wanted.cache {
			log.Println("Fetcher fetching MsgID:", k)
			head := new(Header)
			head.MsgId = proto.Int64(k)
			head.MsgLength = proto.Int32(0)
			head.MsgType = Header_CONFIG.Enum()
			for _, src := range srcs {
				b, err := proto.Marshal(head)
				if err != nil {
					log.Println(err)
					break
				}
				log.Println("Fetcher fetching MsgID:", k, "from", src)

				conn, err := net.Dial("tcp", src+FetcherPort)
				if err != nil {
					log.Println("Error fetching ", k, "from", src, err)
					continue
				}
				n, err := conn.Write(b)
				if err != nil {
					log.Println(err)
					conn.Close()
					continue
				}
				log.Println("Wrote header ", n)
				buf := make([]byte, MaxDatagramSize)
				n, err = conn.Read(buf)
				if err != nil {
					log.Println(err)
					conn.Close()
					continue
				}
				log.Println("Read message ", n)
				conn.Close()
				m := new(NewMsg)
				err = proto.Unmarshal(buf[:n], m)
				if err != nil {
					log.Println(err)
					continue
				}
				log.Println("Downloaded message: ", m)
				wanted.Delete(k)
				store.Add(k)
				SaveAsFile(k, b, StoreDir)
				//Fetch next key
				break
			}

		}

	}

}

func MsgServer() {
	log.Println("Starting MsgServer")
	//	http.HandleFunc("/", JsonResponse)
	http.Handle("/", http.FileServer(http.Dir(StoreDir)))
	log.Fatal(http.ListenAndServe(FetcherPort, nil))
}

func Fetcher(store *msgstore, wanted *wantedstore) {
	log.Println("Starting Fetcher")
	client := NewTimeoutClient(HttpConnectTO*time.Second, HttpReadWriteTO*time.Second)
	for {
		time.Sleep(FetchDuration * time.Millisecond)
		//TODO should use reader lock on wanted but since writer lock is acquired within the loop its not done
		for k, srcs := range wanted.cache {
			log.Println("Fetcher fetching MsgID:", k)
			for _, src := range srcs {
				log.Println("Fetcher fetching MsgID:", k, "from", src)

				resp, err := client.Get("http://" + src + FetcherPort + "/" + strconv.FormatInt(k, 10))
				if err != nil {
					log.Println("Error fetching ", k, "from", src, err)
					continue
				}
				b, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Println(err)
					resp.Body.Close()
					continue
				}
				resp.Body.Close()
				m := new(NewMsg)
				err = proto.Unmarshal(b, m)
				if err != nil {
					log.Println(err)
					continue
				}
				log.Println("Downloaded message: ", m)
				wanted.Delete(k)
				store.Add(k)
				SaveAsFile(k, b, StoreDir)
				//Fetch next key
				break
			}

		}

	}

}
